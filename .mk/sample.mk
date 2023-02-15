# Deploy the sample FlowCollector CR
.PHONY: deploy-sample-cr
deploy-sample-cr:
	@echo -e "\n==> Deploy sample CR"
ifeq (main,$(VERSION))
	kubectl apply -f ./config/samples/flows_v1beta1_flowcollector.yaml || true
else
	kubectl apply -f ./config/samples/flows_v1beta1_flowcollector_versioned.yaml || true
endif

# Undeploy the sample FlowCollector CR
.PHONY: undeploy-sample-cr
undeploy-sample-cr:
	@echo -e "\n==> Undeploy sample CR"
	kubectl --ignore-not-found=true delete flowcollector cluster || true

# Deploy sample workload
.PHONY: deploy-sample-workload
deploy-sample-workload:
	@echo -e "\n==> Deploy sample workload"
	-kubectl create namespace sample-workload
	oc adm policy add-scc-to-user privileged system:serviceaccount:sample-workload:default
	kubectl -n sample-workload apply -f  https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/release/v0.4.1/release/kubernetes-manifests.yaml
	kubectl -n sample-workload run syn-flood --privileged --image=bilalcaliskan/syn-flood:latest --restart=Never -- --host frontend-external.sample-workload.svc.cluster.local --port 80 --floodType syn

# undeploy sample workload
.PHONY: undeploy-sample-workload
undeploy-sample-workload:
	@echo -e "\n==> Undeploy sample workload"
	kubectl -n sample-workload delete --ignore-not-found=true -f  https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/release/v0.4.1/release/kubernetes-manifests.yaml
	-kubectl -n sample-workload delete --ignore-not-found=true pod syn-flood
	-kubectl delete --ignore-not-found=true namespace sample-workload
