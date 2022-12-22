##@ Local (Kind)

.PHONY: create-kind-cluster
create-kind-cluster: $(KIND) ## Create kind cluster
	-$(KIND) create cluster --config config/kubernetes/kind.config.yaml
	kubectl cluster-info --context kind-kind

.PHONY: delete-kind-cluster
delete-kind-cluster: $(KIND) ## Delete kind cluster
	$(KIND) delete cluster

.PHONY: local-deploy
local-deploy: create-kind-cluster deploy-all  ## Local deploy (kind, loki, grafana, example-cr and sample-workload excluding the operator)

.PHONY: clean-leftovers
clean-leftovers:
	-PID=$$(pgrep --oldest --full "main.go"); pkill -P $$PID; pkill $$PID
	-kubectl delete namespace netobserv

.PHONY: local-redeploy
local-redeploy: clean-leftovers undeploy-all deploy-all  ## Local re-deploy (loki, grafana, example-cr and sample-workload excluding the operator)

.PHONY: local-undeploy
local-undeploy: clean-leftovers undeploy-all delete-kind-cluster  ## Local cleanup

local-run: create-kind-cluster local-redeploy local-deploy-operator ## local-redeploy + run the operator locally

.PHONY: local-deploy-operator
local-deploy-operator:
# TODO: restore traffic generator, but using IPFIX instead of NFv5 (could be inspired from https://github.com/netobserv/flowlogs-pipeline/blob/main/pkg/test/ipfix.go)
	@echo "====> Running the operator locally (in background process)"
	go run ./main.go \
		-ebpf-agent-image=quay.io/netobserv/netobserv-ebpf-agent:main \
		-flowlogs-pipeline-image=quay.io/netobserv/flowlogs-pipeline:main \
		-console-plugin-image=quay.io/netobserv/network-observability-console-plugin:main &
	@echo "====> Waiting for flowlogs-pipeline pod to be ready"
	while : ; do kubectl get ds flowlogs-pipeline && break; sleep 1; done
	kubectl wait --timeout=180s --for=condition=ready pod -l app=flowlogs-pipeline
	@echo "====> Operator process info"
	@PID=$$(pgrep --oldest --full "main.go"); echo -e "\n===> The operator is running in process $$PID\nTo stop the operator process use: pkill -p $$PID"
	@echo "====> Done"
