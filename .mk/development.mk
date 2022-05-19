##@ Development

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

.PHONY: undeploy-loki
undeploy-loki: ## Undeploy loki.
	@echo -e "\n==> Undeploy loki"
	kubectl config set-context --current --namespace=$(NAMESPACE)
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/2-loki.yaml	 | kubectl --ignore-not-found=true  delete -f - || true
	curl -S -L https://raw.githubusercontent.com/netobserv/documents/main/examples/zero-click-loki/1-storage.yaml | kubectl --ignore-not-found=true  delete -f - || true
	-pkill --oldest --full "3100:3100"

.PHONY: deploy-kafka
deploy-kafka:
	@echo -e "\n==> Deploy kafka"
	kubectl create namespace $(NAMESPACE)  --dry-run=client -o yaml | kubectl apply -f -
	kubectl create -f "https://strimzi.io/install/latest?namespace="$(NAMESPACE) -n $(NAMESPACE)
	kubectl create -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka-cluster.yaml" -n $(NAMESPACE)
	kubectl create -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka-topic.yaml" -n $(NAMESPACE)

.PHONY: undeploy-kafka
undeploy-kafka: ## Undeploy loki.
	@echo -e "\n==> Undeploy kafka"
	kubectl delete -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka-topic.yaml" -n $(NAMESPACE)
	kubectl delete -f "https://raw.githubusercontent.com/netobserv/documents/main/examples/kafka-cluster.yaml" -n $(NAMESPACE)
	kubectl delete -f "https://strimzi.io/install/latest?namespace="$(NAMESPACE) -n $(NAMESPACE)

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

.PHONY: deploy-all
deploy-all: manifests generate fmt lint deploy-loki deploy-grafana install deploy-sample-cr

.PHONY: undeploy-all
undeploy-all: undeploy-loki undeploy-grafana uninstall undeploy-sample-cr
