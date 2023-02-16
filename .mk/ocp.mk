##@ OCP

.PHONY: ocp-expose-infra
ocp-expose-infra:
	oc expose service grafana || true
	@grafana_url=$$(oc get route grafana -o jsonpath='{.spec.host}'); \
	echo -e "\nAccess grafana on OCP using: http://"$$grafana_url"\n"
	oc expose service loki || true
	@loki_url=$$(oc get route loki -o jsonpath='{.spec.host}'); \
	echo -e "\nAccess loki on OCP using: http://"$$loki_url"\n"

.PHONY: ocp-expose-all
ocp-expose-all: ocp-expose-infra
	oc expose -n sample-workload service frontend-external || true
	@sample_workload_url=$$(oc get -n sample-workload route frontend-external -o jsonpath='{.spec.host}'); \
	echo -e "\nAccess sample workload on OCP using: http://"$$sample_workload_url"\n"

.PHONY: ocp-deploy-infra
ocp-deploy-infra: deploy-infra ocp-expose-infra ## OCP infra. deploy (only loki and grafana excluding the operator)

.PHONY: ocp-deploy
ocp-deploy: deploy-all ocp-expose-all ## OCP deploy (loki, grafana, example-cr and sample-workload excluding the operator)

.PHONY: ocp-undeploy
ocp-undeploy: undeploy-all  ## OCP cleanup

.PHONY: ocp-run
ocp-run: ocp-undeploy ocp-deploy ocp-deploy-operator  ## OCP-deploy + run the operator locally

.PHONY: ocp-deploy-operator
ocp-deploy-operator: ## run flp from the operator
	@echo "====> Enable netobserv-plugin in OCP console"
	oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins", "value": ["netobserv-plugin"]}]'
	$(MAKE) deploy

.PHONY: ocp-undeploy-operator
ocp-undeploy-operator: ## stop the operator locally
	oc delete service flowlogs-pipeline-prom || true
	oc delete ds flowlogs-pipeline || true
	oc delete servicemonitor flowlogs-pipeline-monitor || true
	oc delete service netobserv-plugin || true
	oc delete deployment netobserv-plugin || true
	oc delete servicemonitor netobserv-console-plugin || true
	$(MAKE) undeploy

.PHONY: ocp-refresh-ovs
ocp-refresh-ovs:
	@echo "====> Re-applying OVS configuration to speed-up templates sync"
	./hack/refresh-ovs.sh
