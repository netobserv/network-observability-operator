##@ Development

# use default cluster storage class
DEFAULT_SC := $(shell kubectl get storageclass -o=jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}')

.PHONY: deploy-loki
deploy-loki: ## Deploy loki.
	@echo -e "\n==> Deploy loki"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl config set-context --current --namespace=$(NAMESPACE)
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/1-storage.yaml | kubectl create -f - || true
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/2-loki.yaml	 | kubectl create -f - || true
	kubectl wait --timeout=180s --for=condition=ready pod -l app=loki
	-pkill --oldest --full "3100:3100"
ifeq (true, $(PORT_FWD))
	-kubectl port-forward --address 0.0.0.0 svc/loki 3100:3100 2>&1 >/dev/null &
	@echo -e "\n===> loki endpoint is available on http://localhost:3100\n"
endif

.PHONY: undeploy-loki-tls
undeploy-loki-tls:
	@echo -e "\n==> Undeploy tls loki"
	kubectl config set-context --current --namespace=$(NAMESPACE)
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/2-loki-tls.yaml	 | kubectl --ignore-not-found=true  delete -f - || true
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/1-storage.yaml | kubectl --ignore-not-found=true  delete -f - || true
	-pkill --oldest --full "3100:3100"

.PHONY: deploy-loki-tls
deploy-loki-tls:
	@echo -e "\n==> Deploy tls loki"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl config set-context --current --namespace=$(NAMESPACE)
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/1-storage.yaml | kubectl create -f - || true
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/2-loki-tls.yaml	 | kubectl create -f - || true
	kubectl wait --timeout=180s --for=condition=ready pod -l app=loki
	-pkill --oldest --full "3100:3100"
ifeq (true, $(PORT_FWD))
	-kubectl port-forward --address 0.0.0.0 svc/loki 3100:3100 2>&1 >/dev/null &
	@echo -e "\n===> loki endpoint is available on http://localhost:3100\n"
endif

.PHONY: undeploy-loki
undeploy-loki: ## Undeploy loki.
	@echo -e "\n==> Undeploy loki"
	kubectl config set-context --current --namespace=$(NAMESPACE)
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/2-loki.yaml	 | kubectl --ignore-not-found=true  delete -f - || true
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/1-storage.yaml | kubectl --ignore-not-found=true  delete -f - || true
	-pkill --oldest --full "3100:3100"

.PHONY: deploy-kafka
deploy-kafka:
	@echo -e "\n==> Deploy default Kafka. Get more help on https://github.com/netobserv/documents/blob/main/kafka.md"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f "https://strimzi.io/install/latest?namespace="$(NAMESPACE) -n $(NAMESPACE)
	kubectl apply -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/metrics-config.yaml" -n $(NAMESPACE)
	curl -s -L "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/default.yaml" | envsubst | kubectl apply -n $(NAMESPACE) -f -
	@echo -e "\n==>Using storage class ${DEFAULT_SC}"
	kubectl apply -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/topic.yaml" -n $(NAMESPACE)
	kubectl wait --timeout=180s --for=condition=ready kafkatopic network-flows -n $(NAMESPACE)

.PHONY: deploy-kafka-tls
deploy-kafka-tls:
	@echo -e "\n==> Deploy Kafka with mTLS. Get more help on https://github.com/netobserv/documents/blob/main/kafka.md"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f "https://strimzi.io/install/latest?namespace="$(NAMESPACE) -n $(NAMESPACE)
	kubectl apply -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/metrics-config.yaml" -n $(NAMESPACE)
	curl -s -L "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/tls.yaml" | envsubst | kubectl apply -n $(NAMESPACE) -f - 
	@echo -e "\n==>Using storage class ${DEFAULT_SC}"
	kubectl apply -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/topic.yaml" -n $(NAMESPACE)
	kubectl apply -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/user.yaml" -n $(NAMESPACE)
	kubectl wait --timeout=180s --for=condition=ready kafkauser flp-kafka -n $(NAMESPACE)

.PHONY: undeploy-kafka
undeploy-kafka: ## Undeploy kafka.
	@echo -e "\n==> Undeploy kafka"
	kubectl delete -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka/topic.yaml" -n $(NAMESPACE) --ignore-not-found=true
	kubectl delete kafkauser flp-kafka -n $(NAMESPACE) --ignore-not-found=true
	kubectl delete kafka kafka-cluster -n $(NAMESPACE) --ignore-not-found=true
	kubectl delete -f "https://strimzi.io/install/latest?namespace="$(NAMESPACE) -n $(NAMESPACE) --ignore-not-found=true

.PHONY: fix-ebpf-kafka-tls
fix-ebpf-kafka-tls:
	@echo -e "\n==> Fix eBPF with Kafka on TLS: copying secrets to privileged namespace"
	kubectl get secret flp-kafka -n $(NAMESPACE) -o yaml | yq 'del(.metadata)' | yq '.metadata.name = "flp-kafka"' | kubectl apply -n "$(NAMESPACE)-privileged" -f -
	kubectl get secret kafka-cluster-cluster-ca-cert -n $(NAMESPACE) -o yaml | yq 'del(.metadata)' | yq '.metadata.name = "kafka-cluster-cluster-ca-cert"' | kubectl apply -n "$(NAMESPACE)-privileged" -f -
	@echo -e "\n===> Restarting eBPF pods"
	kubectl delete pods -n "$(NAMESPACE)-privileged" --all --force

.PHONY: deploy-grafana
deploy-grafana: ## Deploy grafana.
	@echo -e "\n==> Deploy grafana"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl config set-context --current --namespace=$(NAMESPACE)
	./hack/deploy-grafana.sh $(NAMESPACE)
	-pkill --oldest --full "3000:3000"
ifeq (true, $(PORT_FWD))
	-kubectl port-forward --address 0.0.0.0 svc/grafana 3000:3000 2>&1 >/dev/null &
	@echo -e "\n===> grafana ui is available (user: admin password: admin) on http://localhost:3000\n"
endif

.PHONY: undeploy-grafana
undeploy-grafana: ## Undeploy grafana.
	@echo -e "\n==> Undeploy grafana"
	kubectl config set-context --current --namespace=$(NAMESPACE)
	kubectl delete --ignore-not-found=true deployment grafana
	kubectl delete --ignore-not-found=true service grafana
	kubectl delete --ignore-not-found=true configMap grafana-datasources
	-pkill --oldest --full "3000:3000"

.PHONY: deploy-infra
deploy-infra: manifests generate fmt lint deploy-loki deploy-grafana install

.PHONY: deploy-all
deploy-all: deploy-infra deploy-sample-cr deploy-sample-workload

.PHONY: undeploy-infra
undeploy-infra: undeploy-loki undeploy-grafana uninstall

.PHONY: undeploy-all
undeploy-all: undeploy-infra undeploy-sample-cr undeploy-sample-workload

.PHONY: deploy-prometheus
deploy-prometheus: ## Deploy prometheus.
	@echo -e "\n==> Deploy prometheus"
	kubectl apply -f config/kubernetes/deployment-prometheus.yaml
	kubectl rollout status "deploy/prometheus" --timeout=600s
	-pkill --oldest --full "9090:9090"
ifeq (true, $(PORT_FWD))
	kubectl port-forward --address 0.0.0.0 svc/prometheus 9090:9090 2>&1 >/dev/null &
	@echo -e "\n===> prometheus ui is available on http://localhost:9090\n"
endif

.PHONY: undeploy-prometheus
undeploy-prometheus: ## Undeploy prometheus.
	@echo -e "\n==> Undeploy prometheus"
	kubectl --ignore-not-found=true delete -f config/kubernetes/deployment-prometheus.yaml || true
	-pkill --oldest --full "9090:9090"

.PHONY: get-related-images
get-related-images:
	kubectl set env deployment netobserv-controller-manager --list

.PHONY: set-agent-image
set-agent-image:
	kubectl set env deployment netobserv-controller-manager RELATED_IMAGE_EBPF_AGENT=quay.io/$(USER)/netobserv-ebpf-agent:$(VERSION)
	@echo -e "\n==> Redeploying..."
	kubectl rollout status -n $(NAMESPACE) --timeout=60s deployment netobserv-controller-manager
	kubectl wait -n $(NAMESPACE) --timeout=60s --for condition=Available=True deployment netobserv-controller-manager
	@echo -e "\n==> Wait a moment before agents are fully redeployed"

.PHONY: set-flp-image
set-flp-image:
	kubectl set env deployment netobserv-controller-manager RELATED_IMAGE_FLOWLOGS_PIPELINE=quay.io/$(USER)/flowlogs-pipeline:$(VERSION)
	@echo -e "\n==> Redeploying..."
	kubectl rollout status -n $(NAMESPACE) --timeout=60s deployment netobserv-controller-manager
	kubectl wait -n $(NAMESPACE) --timeout=60s --for condition=Available=True deployment netobserv-controller-manager
	@echo -e "\n==> Wait a moment before FLP is fully redeployed"

.PHONY: set-plugin-image
set-plugin-image:
	kubectl set env deployment netobserv-controller-manager RELATED_IMAGE_CONSOLE_PLUGIN=quay.io/$(USER)/network-observability-console-plugin:$(VERSION)
	@echo -e "\n==> Redeploying..."
	kubectl rollout status -n $(NAMESPACE) --timeout=60s deployment netobserv-controller-manager
	kubectl wait -n $(NAMESPACE) --timeout=60s --for condition=Available=True deployment netobserv-controller-manager
	kubectl rollout status -n $(NAMESPACE) --timeout=60s deployment netobserv-plugin
	kubectl wait -n $(NAMESPACE) --timeout=60s --for condition=Available=True deployment netobserv-plugin
