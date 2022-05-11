##@ OCP

.PHONY: ocp-expose
ocp-expose:
	oc expose service grafana || true
	@grafana_url=$$(oc get route grafana -o jsonpath='{.spec.host}'); \
	echo -e "\nAccess grafana on OCP using: http://"$$grafana_url"\n"
	oc expose service loki || true
	@loki_url=$$(oc get route loki -o jsonpath='{.spec.host}'); \
	echo -e "\nAccess loki on OCP using: http://"$$loki_url"\n"

.PHONY: ocp-deploy
ocp-deploy: deploy-all ocp-expose ## OCP deploy (loki, grafana and example-cr excluding the operator)

.PHONY: ocp-undeploy
ocp-undeploy: undeploy-all  ## OCP cleanup

.PHONY: ocp-run
ocp-run: ocp-undeploy ocp-deploy   ## OCP-deploy + run the operator locally
	@echo "====> Enable network-observability-plugin in OCP console"
	oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins", "value": ["network-observability-plugin"]}]'
	@echo "====> Running the operator locally"
	go run ./main.go

