##@ Local (Kind)

.PHONY: create-kind-cluster
create-kind-cluster: $(KIND) ## Create kind cluster
	-$(KIND) create cluster --config config/kubernetes/kind.config.yaml
	kubectl cluster-info --context kind-kind

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

local-run: create-kind-cluster local-redeploy ## local-redeploy + run the operator locally
	@echo "====> Running the operator locally (in background process)"
	go run ./main.go &
	@echo "====> Waiting for flowlogs-pipeline pod to be ready"
	while : ; do kubectl get ds flowlogs-pipeline && break; sleep 1; done
	kubectl wait --timeout=180s --for=condition=ready pod -l app=flowlogs-pipeline
	@echo "====> Getting first pod in the demon-set"
	first_pod=$$(kubectl get pods --selector=app=flowlogs-pipeline -o jsonpath='{.items[0].metadata.name}'); \
	kubectl expose pod $$first_pod --name=flowlogs-pipeline-netflows --protocol=UDP --port=2056 --target-port=2056; \
	echo "====> Sending simulated logs to pod $$first_pod exposed as service flowlogs-pipeline-netflows"
	kubectl create deployment netflow-simulator --image=networkstatic/nflow-generator:latest -- /etc/nflow/nflow-generator --target=flowlogs-pipeline-netflows.network-observability.svc.cluster.local --port=2056 --spike=http
	kubectl wait --timeout=180s --for=condition=ready pod -l app=netflow-simulator
	@echo "====> Operator process info"
	@PID=$$(pgrep --oldest --full "main.go"); echo -e "\n===> The operator is running in process $$PID\nTo stop the operator process use: pkill -p $$PID"
	@echo "====> Done"
