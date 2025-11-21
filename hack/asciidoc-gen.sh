#!/bin/sh

set -e

mkdir -p _tmp
oc get --raw /openapi/v2 | jq . > _tmp/openapi.1.json

jq '.definitions |= ({"io.netobserv.flows.v1beta2.FlowCollector", "io.netobserv.flows.v1alpha1.FlowMetric", "io.netobserv.flows.v1alpha1.FlowCollectorSlice"})
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.status)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.metadata."$ref")
  | .definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.metadata += {type:"object"}
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.agent.properties.ipfix)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.agent.properties.ebpf.properties.resources.properties.claims)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.agent.properties.ebpf.properties.advanced.properties.scheduling.properties.affinity.properties)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.agent.properties.ebpf.properties.advanced.properties.scheduling.properties.tolerations.items)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.processor.properties.resources.properties.claims)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.processor.properties.advanced.properties.scheduling.properties.affinity.properties)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.processor.properties.advanced.properties.scheduling.properties.tolerations.items)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.consolePlugin.properties.resources.properties.claims)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.consolePlugin.properties.autoscaler.properties)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.consolePlugin.properties.advanced.properties.scheduling.properties.affinity.properties)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.consolePlugin.properties.advanced.properties.scheduling.properties.tolerations.items)
  | del(.definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.processor.properties.kafkaConsumerAutoscaler.properties)
  | .definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.consolePlugin.properties.autoscaler.description |= . + " Refer to HorizontalPodAutoscaler documentation (autoscaling/v2)."
  | .definitions."io.netobserv.flows.v1beta2.FlowCollector".properties.spec.properties.processor.properties.kafkaConsumerAutoscaler.description |= . + " Refer to HorizontalPodAutoscaler documentation (autoscaling/v2)."
  | del(.definitions."io.netobserv.flows.v1alpha1.FlowMetric".properties.status)
  | del(.definitions."io.netobserv.flows.v1alpha1.FlowMetric".properties.metadata."$ref")
  | .definitions."io.netobserv.flows.v1alpha1.FlowMetric".properties.metadata += {type:"object"}
  | del(.definitions."io.netobserv.flows.v1alpha1.FlowCollectorSlice".properties.status)
  | del(.definitions."io.netobserv.flows.v1alpha1.FlowCollectorSlice".properties.metadata."$ref")
  | .definitions."io.netobserv.flows.v1alpha1.FlowCollectorSlice".properties.metadata += {type:"object"}' \
  _tmp/openapi.1.json > _tmp/openapi.2.json

openshift-apidocs-gen build -c hack/asciidoc-gen-config.yaml _tmp/openapi.2.json


amend_doc() {
  local filename=$1

  mv _tmp/flows_netobserv_io/$filename docs/$filename

  sed -i -r 's/^:_content-type: ASSEMBLY$/:_mod-docs-content-type: REFERENCE/' docs/$filename
  sed -i -r 's/^\[id="flowcollector-flows-netobserv-io-v.+"\]$/[id="network-observability-flowcollector-api-specifications_{context}"]/' docs/$filename
  sed -i -r 's/= FlowCollector \[flows.netobserv.io.*/= FlowCollector API specifications/' docs/$filename
  sed -i -r '/^:toc: macro$/d ' docs/$filename
  sed -i -r '/^:toc-title:$/d ' docs/$filename
  sed -i -r '/^toc::\[\]$/d ' docs/$filename
  sed -i -r '/^== Specification$/d ' docs/$filename
  sed -i -r 's/^==/=/g' docs/$filename
  sed -i -r '/^= API endpoints/Q' docs/$filename
  sed -i -r 's/OpenShift/{product-title}/g' docs/$filename
  sed -i -r 's/\<NetObserv\>/Network Observability/g' docs/$filename
  sed -i -r 's/<br>/ +\n/g' docs/$filename
  sed -i -r 's/<i>/_/g' docs/$filename
  sed -i -r 's/<\/i>/_/g' docs/$filename
  sed -i -r 's/ may / might /g' docs/$filename
  # Our asciidoc gen doesn't handle arrays very well, producing duplicate fields... so remove one of them
  sed -i -r '/^\| `.+\[\]`$/,+3d' docs/$filename
}

amend_doc "flowcollector-flows-netobserv-io-v1beta2.adoc"
amend_doc "flowmetric-flows-netobserv-io-v1alpha1.adoc"
amend_doc "flowcollectorslice-flows-netobserv-io-v1alpha1.adoc"
