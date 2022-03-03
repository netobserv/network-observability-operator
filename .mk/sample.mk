# Deploy the sample FlowCollector CR
.PHONY: deploy-sample-cr
deploy-sample-cr:
	sed -e 's~:main~:$(VERSION)~' ./config/samples/flows_v1alpha1_flowcollector.yaml | kubectl apply -f - || true

# Undeploy the sample FlowCollector CR
.PHONY: undeploy-sample-cr
undeploy-sample-cr:
	sed -e 's~:main~:$(VERSION)~' ./config/samples/flows_v1alpha1_flowcollector.yaml | kubectl --ignore-not-found=true delete -f - || true
