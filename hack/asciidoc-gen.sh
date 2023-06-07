#!/bin/sh

set -e

mkdir -p _tmp
oc get --raw /openapi/v2 | jq . > _tmp/openapi.json

jq '.definitions |= ({"io.netobserv.flows.v1beta1.FlowCollector"})
  | del(.definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.status)
  | del(.definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.metadata."$ref")
  | .definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.metadata += {type:"object"}
  | del(.definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.spec.properties.consolePlugin.properties.autoscaler.properties)
  | del(.definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.spec.properties.processor.properties.kafkaConsumerAutoscaler.properties)
  | .definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.spec.properties.consolePlugin.properties.autoscaler.description |= . + " Please refer to HorizontalPodAutoscaler documentation (autoscaling/v2)."
  | .definitions."io.netobserv.flows.v1beta1.FlowCollector".properties.spec.properties.processor.properties.kafkaConsumerAutoscaler.description |= . + " Please refer to HorizontalPodAutoscaler documentation (autoscaling/v2)."' \
  _tmp/openapi.json > _tmp/openapi-amended.json

openshift-apidocs-gen build -c hack/asciidoc-gen-config.yaml _tmp/openapi-amended.json

ADOC=docs/flowcollector-flows-netobserv-io-v1beta1.adoc

mv _tmp/flows_netobserv_io/flowcollector-flows-netobserv-io-v1beta1.adoc $ADOC

sed -i -r 's/^:_content-type: ASSEMBLY$/:_content-type: REFERENCE/' $ADOC
sed -i -r 's/^\[id="flowcollector-flows-netobserv-io-v.+"\]$/[id="network-observability-flowcollector-api-specifications_{context}"]/' $ADOC
sed -i -r 's/= FlowCollector \[flows.netobserv.io.*/= FlowCollector API specifications/' $ADOC
sed -i -r '/^:toc: macro$/d ' $ADOC
sed -i -r '/^:toc-title:$/d ' $ADOC
sed -i -r '/^toc::\[\]$/d ' $ADOC
sed -i -r '/^== Specification$/d ' $ADOC
sed -i -r 's/^==/=/g' $ADOC
sed -i -r '/^= API endpoints/Q' $ADOC
sed -i -r 's/OpenShift/{product-title}/' $ADOC
sed -i -r 's/<br>/ +\n/g' $ADOC
sed -i -r 's/<i>/_/g' $ADOC
sed -i -r 's/<\/i>/_/g' $ADOC
