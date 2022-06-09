##@ Local (Kind)

.PHONY: create-kind-cluster
create-kind-cluster: $(KIND) ## Create kind cluster
	-$(KIND) create cluster --config config/kubernetes/kind.config.yaml
	kubectl cluster-info --context kind-kind
	# TODO: wait for all services to come up.

.PHONY: delete-kind-cluster
delete-kind-cluster: $(KIND) ## Delete kind cluster
	$(KIND) delete cluster

.PHONY: local-deploy
local-deploy: create-kind-cluster deploy-all  ## Local deploy (kind, loki, grafana and example-cr excluding the operator)

.PHONY: clean-leftovers
clean-leftovers:
	-PID=$$(pgrep --oldest --full "main.go"); pkill -P $$PID; pkill $$PID
	-kubectl delete namespace network-observability

.PHONY: local-redeploy
local-redeploy: clean-leftovers undeploy-all deploy-all  ## Local re-deploy (loki, grafana and example-cr excluding the operator)

.PHONY: local-undeploy
local-undeploy: clean-leftovers undeploy-all delete-kind-cluster  ## Local cleanup

local-run: create-kind-cluster local-redeploy deploy-flp ## local-redeploy + run the operator locally
# TODO: restore traffic generator, but using IPFIX instead of NFv5 (could be inspired from https://github.com/netobserv/flowlogs-pipeline/blob/main/pkg/test/ipfix.go)

.PHONY: deploy-flp
deploy-flp: ## Deploy flp
	@echo "====> Running the operator locally (in background process)"
	# temporarily change flowlogs-pipeline:main to flowlogs-pipeline:custom in config file
	sed -i 's~flowlogs-pipeline:main~flowlogs-pipeline:custom~' config/samples/flows_v1alpha1_flowcollector.yaml
	go run ./main.go &
	@echo "====> Waiting for flowlogs-pipeline pod to be ready"
	while : ; do kubectl get ds flowlogs-pipeline && break; sleep 1; done
	kubectl wait --timeout=180s --for=condition=ready pod -l app=flowlogs-pipeline
	@echo "====> Getting first pod in the demon-set"
	first_pod=$$(kubectl get pods --selector=app=flowlogs-pipeline -o jsonpath='{.items[0].metadata.name}'); \
	kubectl expose pod $$first_pod --name=flowlogs-pipeline-metrics --protocol=TCP --port=9102 --target-port=9102; \
	echo "====> $$first_pod exposed as service flowlogs-pipeline-metrics for prometheus"
	first_pod=$$(kubectl get pods --selector=app=flowlogs-pipeline -o jsonpath='{.items[0].metadata.name}'); \
	kubectl expose pod $$first_pod --name=flowlogs-pipeline-netflows --protocol=UDP --port=2056 --target-port=2056; \
	echo "====> $$first_pod exposed as service flowlogs-pipeline-netflows for simulated network flows"
	@echo "====> Operator process info"
	@PID=$$(pgrep --oldest --full "main.go"); echo -e "\n===> The operator is running in process $$PID\nTo stop the operator process use: pkill -p $$PID"
	sed -i 's~flowlogs-pipeline:custom~flowlogs-pipeline:main~' config/samples/flows_v1alpha1_flowcollector.yaml
	@echo "====> Done"

.PHONY: undeploy-flp
undeploy-flp: ## stop the the operator locally
	-PID=$$(pgrep --oldest --full "main.go"); pkill -P $$PID; pkill $$PID
	kubectl delete service flowlogs-pipeline-metrics || true
	kubectl delete service flowlogs-pipeline-netflows || true
	kubectl delete ds flowlogs-pipeline || true
	sed -i 's~flowlogs-pipeline:custom~flowlogs-pipeline:main~' config/samples/flows_v1alpha1_flowcollector.yaml

