# API Reference

Packages:

- [flows.netobserv.io/v1alpha1](#flowsnetobserviov1alpha1)

# flows.netobserv.io/v1alpha1

Resource Types:

- [FlowCollector](#flowcollector)




## FlowCollector
<sup><sup>[↩ Parent](#flowsnetobserviov1alpha1 )</sup></sup>






FlowCollector is the Schema for the flowcollectors API, which pilots and configures netflow collection.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>flows.netobserv.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FlowCollector</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspec">spec</a></b></td>
        <td>object</td>
        <td>
          FlowCollectorSpec defines the desired state of FlowCollector<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorstatus">status</a></b></td>
        <td>object</td>
        <td>
          FlowCollectorStatus defines the observed state of FlowCollector<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec
<sup><sup>[↩ Parent](#flowcollector)</sup></sup>



FlowCollectorSpec defines the desired state of FlowCollector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>agent</b></td>
        <td>enum</td>
        <td>
          Select the flows tracing agent. Possible values are "ipfix" (default) to use the IPFIX collector, or "ebpf" to use NetObserv eBPF agent. When using IPFIX with OVN-Kubernetes CNI, NetObserv will configure OVN's IPFIX exporter. Other CNIs are not supported, they could work but necessitate manual configuration.<br/>
          <br/>
            <i>Enum</i>: ipfix, ebpf<br/>
            <i>Default</i>: ipfix<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecclusternetworkoperator">clusterNetworkOperator</a></b></td>
        <td>object</td>
        <td>
          Settings related to the OpenShift Cluster Network Operator, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsoleplugin">consolePlugin</a></b></td>
        <td>object</td>
        <td>
          Settings related to the OpenShift Console plugin, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecebpf">ebpf</a></b></td>
        <td>object</td>
        <td>
          Settings related to eBPF-based flow reporter when the "agent" property is set to "ebpf".<br/>
          <br/>
            <i>Default</i>: map[imagePullPolicy:IfNotPresent]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipeline">flowlogsPipeline</a></b></td>
        <td>object</td>
        <td>
          Settings related to the flowlogs-pipeline component, which collects and enriches the flows, and produces metrics.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecipfix">ipfix</a></b></td>
        <td>object</td>
        <td>
          Settings related to IPFIX-based flow reporter when the "agent" property is set to "ipfix".<br/>
          <br/>
            <i>Default</i>: map[sampling:400]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Kafka can provide better scalability, resiliency and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecloki">loki</a></b></td>
        <td>object</td>
        <td>
          Settings related to the Loki client, used as a flow store.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where NetObserv pods are deployed. If empty, the namespace of the operator is going to be used.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecovnkubernetes">ovnKubernetes</a></b></td>
        <td>object</td>
        <td>
          Settings related to OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.clusterNetworkOperator
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to the OpenShift Cluster Network Operator, when available.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace  where the configmap is going to be deployed.<br/>
          <br/>
            <i>Default</i>: openshift-network-operator<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to the OpenShift Console plugin, when available.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>register</b></td>
        <td>boolean</td>
        <td>
          Automatically register the provided console plugin with the OpenShift Console operator. When set to false, you can still register it manually by editing console.operator.openshift.io/cluster. E.g: oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "network-observability-plugin"}]'<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpa">hpa</a></b></td>
        <td>object</td>
        <td>
          HPA spec of an horizontal pod autoscaler to set up for the plugin Deployment.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image is the plugin image (including domain and tag)<br/>
          <br/>
            <i>Default</i>: quay.io/netobserv/network-observability-console-plugin:main<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          ImagePullPolicy is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          LogLevel defines the log level for the console plugin backend<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the plugin service port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9001<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginportnaming">portNaming</a></b></td>
        <td>object</td>
        <td>
          Configuration of the port to service name translation<br/>
          <br/>
            <i>Default</i>: map[enable:true]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          Replicas defines the number of replicas (pods) to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginresources">resources</a></b></td>
        <td>object</td>
        <td>
          Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:100Mi] requests:map[cpu:100m memory:50Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



HPA spec of an horizontal pod autoscaler to set up for the plugin Deployment.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>maxReplicas</b></td>
        <td>integer</td>
        <td>
          upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindex">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          minReplicas is the lower limit for the number of replicas to which the autoscaler can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured.  Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpa)</sup></sup>



MetricSpec specifies how to scale based on a single metric (only `type` and one other matching field should be set at once).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type is the type of metric source.  It should be one of "ContainerResource", "External", "Object", "Pods" or "Resource", each mapping to a matching field in the object. Note: "ContainerResource" type is available on when the feature-gate HPAContainerMetrics is enabled<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexcontainerresource">containerResource</a></b></td>
        <td>object</td>
        <td>
          container resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexexternal">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobject">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexpods">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindex)</sup></sup>



container resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>container</b></td>
        <td>string</td>
        <td>
          container is the name of the container in the pods of the scaling target<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the resource in question.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexcontainerresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexcontainerresource)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindex)</sup></sup>



external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexexternalmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexexternaltarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexexternal)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexexternalmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexexternalmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexexternalmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexexternalmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexexternal)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindex)</sup></sup>



object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobjectdescribedobject">describedObject</a></b></td>
        <td>object</td>
        <td>
          CrossVersionObjectReference contains enough information to let you identify the referred resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobjectmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobjecttarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexobject)</sup></sup>



CrossVersionObjectReference contains enough information to let you identify the referred resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexobject)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobjectmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexobjectmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexobjectmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexobjectmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexobject)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindex)</sup></sup>



pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexpodsmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexpodstarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexpods)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexpodsmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexpodsmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexpodsmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexpodsmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexpods)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindex)</sup></sup>



resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the resource in question.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginhpametricsindexresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.hpa.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginhpametricsindexresource)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.portNaming
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



Configuration of the port to service name translation

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Should this feature be enabled<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portNames</b></td>
        <td>map[string]string</td>
        <td>
          Additional port name to use in the console E.g. portNames: {"3100": "loki"}<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.resources
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.ebpf
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to eBPF-based flow reporter when the "agent" property is set to "ebpf".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>cacheActiveTimeout</b></td>
        <td>string</td>
        <td>
          CacheActiveTimeout is the max period during which the reporter will aggregate flows before sending<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1000<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          Env allows passing custom environment variables to the NetObserv Agent. Useful for passing some very concrete performance-tuning options (e.g. GOGC, GOMAXPROCS) that shouldn't be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug/support scenarios.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>excludeInterfaces</b></td>
        <td>[]string</td>
        <td>
          ExcludeInterfaces contains the interface names that will be excluded from flow tracing. If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression, otherwise it will be matched as a case-sensitive string.<br/>
          <br/>
            <i>Default</i>: [lo]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image is the NetObserv Agent image (including domain and tag)<br/>
          <br/>
            <i>Default</i>: quay.io/netobserv/netobserv-ebpf-agent:main<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          ImagePullPolicy is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>interfaces</b></td>
        <td>[]string</td>
        <td>
          Interfaces contains the interface names from where flows will be collected. If empty, the agent will fetch all the interfaces in the system, excepting the ones listed in ExcludeInterfaces. If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression, otherwise it will be matched as a case-sensitive string.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          LogLevel defines the log level for the NetObserv eBPF Agent<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          Privileged mode for the eBPF Agent container. If false, the operator will add the following capabilities to the container, to enable its correct operation: BPF, PERFMON, NET_ADMIN, SYS_RESOURCE.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecebpfresources">resources</a></b></td>
        <td>object</td>
        <td>
          Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. 0 or 1 means disabled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.ebpf.resources
<sup><sup>[↩ Parent](#flowcollectorspecebpf)</sup></sup>



Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to the flowlogs-pipeline component, which collects and enriches the flows, and produces metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dropUnusedFields</b></td>
        <td>boolean</td>
        <td>
          Set true to drop fields that are known to be unused by OVS, in order to save storage space.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableKubeProbes</b></td>
        <td>boolean</td>
        <td>
          EnableKubeProbes is a flag to enable or disable Kubernetes liveness/readiness probes<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>healthPort</b></td>
        <td>integer</td>
        <td>
          HealthPort is a collector HTTP port in the Pod that exposes the health check API<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 8080<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpa">hpa</a></b></td>
        <td>object</td>
        <td>
          HPA spec of an horizontal pod autoscaler to set up for the collector Deployment. Ignored for DaemonSet.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ignoreMetrics</b></td>
        <td>[]string</td>
        <td>
          IgnoreMetrics is a list of tags to specify which metrics to ignore<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image is the collector image (including domain and tag)<br/>
          <br/>
            <i>Default</i>: quay.io/netobserv/flowlogs-pipeline:main<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          ImagePullPolicy is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          Kind is the workload kind, either DaemonSet or Deployment. When DaemonSet is used, each pod will receive flows from the node it is running on. When Deployment is used, the flows traffic received from nodes will be load-balanced. Note that in such a case, the number of replicas should be less or equal to the number of nodes, as extra-pods would be unused due to session affinity with the node IP. When using Kafka, this option only affects the flowlogs-pipeline ingester, not the transformer.<br/>
          <br/>
            <i>Enum</i>: DaemonSet, Deployment<br/>
            <i>Default</i>: DaemonSet<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          LogLevel defines the log level for the collector runtime<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the collector port: either a service port for Deployment kind, or host port for DaemonSet kind By conventions, some value are not authorized port must not be below 1024 and must not equal this values: 4789,6081,500, and 4500<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 2055<br/>
            <i>Minimum</i>: 1025<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>prometheusPort</b></td>
        <td>integer</td>
        <td>
          PrometheusPort is the prometheus HTTP port: this port exposes prometheus metrics<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9102<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          Replicas defines the number of replicas (pods) to start for Deployment kind. Ignored for DaemonSet.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelineresources">resources</a></b></td>
        <td>object</td>
        <td>
          Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:300Mi] requests:map[cpu:100m memory:100Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipeline)</sup></sup>



HPA spec of an horizontal pod autoscaler to set up for the collector Deployment. Ignored for DaemonSet.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>maxReplicas</b></td>
        <td>integer</td>
        <td>
          upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindex">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          minReplicas is the lower limit for the number of replicas to which the autoscaler can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured.  Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpa)</sup></sup>



MetricSpec specifies how to scale based on a single metric (only `type` and one other matching field should be set at once).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type is the type of metric source.  It should be one of "ContainerResource", "External", "Object", "Pods" or "Resource", each mapping to a matching field in the object. Note: "ContainerResource" type is available on when the feature-gate HPAContainerMetrics is enabled<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexcontainerresource">containerResource</a></b></td>
        <td>object</td>
        <td>
          container resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexexternal">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobject">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexpods">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindex)</sup></sup>



container resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>container</b></td>
        <td>string</td>
        <td>
          container is the name of the container in the pods of the scaling target<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the resource in question.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexcontainerresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexcontainerresource)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindex)</sup></sup>



external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexexternalmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexexternaltarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexexternal)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexexternalmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexexternalmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexexternalmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexexternalmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexexternal)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindex)</sup></sup>



object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobjectdescribedobject">describedObject</a></b></td>
        <td>object</td>
        <td>
          CrossVersionObjectReference contains enough information to let you identify the referred resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobjectmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobjecttarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexobject)</sup></sup>



CrossVersionObjectReference contains enough information to let you identify the referred resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexobject)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobjectmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexobjectmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexobjectmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexobjectmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexobject)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindex)</sup></sup>



pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexpodsmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexpodstarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexpods)</sup></sup>



metric identifies the target metric by name and selector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the given metric<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexpodsmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexpodsmetric)</sup></sup>



selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexpodsmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexpodsmetricselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexpods)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindex)</sup></sup>



resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the resource in question.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipelinehpametricsindexresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.hpa.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipelinehpametricsindexresource)</sup></sup>



target specifies the target value for the given metric

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type represents whether the metric type is Utilization, Value, or AverageValue<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>averageUtilization</b></td>
        <td>integer</td>
        <td>
          averageUtilization is the target value of the average of the resource metric across all relevant pods, represented as a percentage of the requested value of the resource for the pods. Currently only valid for Resource metric source type<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>averageValue</b></td>
        <td>int or string</td>
        <td>
          averageValue is the target value of the average of the metric across all relevant pods (as a quantity)<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>int or string</td>
        <td>
          value is the target value of the metric (as a quantity).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.flowlogsPipeline.resources
<sup><sup>[↩ Parent](#flowcollectorspecflowlogspipeline)</sup></sup>



Compute Resources required by this container. Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.ipfix
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to IPFIX-based flow reporter when the "agent" property is set to "ipfix".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>cacheActiveTimeout</b></td>
        <td>string</td>
        <td>
          CacheActiveTimeout is the max period during which the reporter will aggregate flows before sending<br/>
          <br/>
            <i>Default</i>: 20s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>forceSampleAll</b></td>
        <td>boolean</td>
        <td>
          It is not recommended to sample all the traffic with IPFIX, as it may generate cluster instability. If you REALLY want to do that, set this flag to true. Use at your own risks. When it is set to true, the value of "sampling" is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. To ensure cluster stability, it is not possible to set a value below 2. If you really want to sample every packet, which may impact the cluster stability, refer to "forceSampleAll". Alternatively, you can use the eBPF Agent instead of IPFIX.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 2<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Kafka can provide better scalability, resiliency and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>address</b></td>
        <td>string</td>
        <td>
          Address of the Kafka server<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>topic</b></td>
        <td>string</td>
        <td>
          Kafka topic to use. It must exist, NetObserv will not create it.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Set true to use Kafka as part of the flow collection pipeline. When enabled, the pipeline is split in two parts: ingestion and transformation, connected by Kafka. The ingestion is either done by a specific flowlogs-pipeline workload, or by the eBPF agent, depending on the value of `spec.agent`. The transformation is done by a new flowlogs-pipeline deployment.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls
<sup><sup>[↩ Parent](#flowcollectorspeckafka)</sup></sup>



TLS client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspeckafkatlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          CA certificate reference<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Enable TLS<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Skip client-side verification of the server certificate<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          User certificate reference<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls)</sup></sup>



CA certificate reference

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>certFile</b></td>
        <td>string</td>
        <td>
          Certificate file name within the ConfigMap / Secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          Certificate private key file name within the ConfigMap / Secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the ConfigMap or Secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Reference type: configmap or secret<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls)</sup></sup>



User certificate reference

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>certFile</b></td>
        <td>string</td>
        <td>
          Certificate file name within the ConfigMap / Secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          Certificate private key file name within the ConfigMap / Secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the ConfigMap or Secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Reference type: configmap or secret<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to the Loki client, used as a flow store.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>batchSize</b></td>
        <td>integer</td>
        <td>
          BatchSize is max batch size (in bytes) of logs to accumulate before sending<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: 102400<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>batchWait</b></td>
        <td>string</td>
        <td>
          BatchWait is max time to wait before sending a batch<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxBackoff</b></td>
        <td>string</td>
        <td>
          MaxBackoff is the maximum backoff time for client connection between retries<br/>
          <br/>
            <i>Default</i>: 300s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxRetries</b></td>
        <td>integer</td>
        <td>
          MaxRetries is the maximum number of retries for client connections<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minBackoff</b></td>
        <td>string</td>
        <td>
          MinBackoff is the initial backoff time for client connection between retries<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>querierUrl</b></td>
        <td>string</td>
        <td>
          QuerierURL specifies the address of the Loki querier service, in case it is different from the Loki ingester URL. If empty, the URL value will be used (assuming that the Loki ingester and querier are int he same host).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendAuthToken</b></td>
        <td>boolean</td>
        <td>
          SendAuthToken is a flag to enable or disable Authorization header from service account secret It allows authentication to loki operator gateway<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>staticLabels</b></td>
        <td>map[string]string</td>
        <td>
          StaticLabels is a map of common labels to set on each flow<br/>
          <br/>
            <i>Default</i>: map[app:netobserv-flowcollector]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenantID</b></td>
        <td>string</td>
        <td>
          TenantID is the Loki X-Scope-OrgID that identifies the tenant for each request. it will be ignored if instanceSpec is specified<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>string</td>
        <td>
          Timeout is the maximum time connection / request limit A Timeout of zero means no timeout.<br/>
          <br/>
            <i>Default</i>: 10s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokitls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          URL is the address of an existing Loki service to push the flows to.<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls
<sup><sup>[↩ Parent](#flowcollectorspecloki)</sup></sup>



TLS client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorspeclokitlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          CA certificate reference<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Enable TLS<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Skip client-side verification of the server certificate<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokitlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          User certificate reference<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokitls)</sup></sup>



CA certificate reference

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>certFile</b></td>
        <td>string</td>
        <td>
          Certificate file name within the ConfigMap / Secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          Certificate private key file name within the ConfigMap / Secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the ConfigMap or Secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Reference type: configmap or secret<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokitls)</sup></sup>



User certificate reference

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>certFile</b></td>
        <td>string</td>
        <td>
          Certificate file name within the ConfigMap / Secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          Certificate private key file name within the ConfigMap / Secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the ConfigMap or Secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Reference type: configmap or secret<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.ovnKubernetes
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Settings related to OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Name of the container to configure for IPFIX.<br/>
          <br/>
            <i>Default</i>: ovnkube-node<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>daemonSetName</b></td>
        <td>string</td>
        <td>
          Name of the DaemonSet controlling the OVN-Kubernetes pods.<br/>
          <br/>
            <i>Default</i>: ovnkube-node<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where OVN-Kubernetes pods are deployed.<br/>
          <br/>
            <i>Default</i>: ovn-kubernetes<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.status
<sup><sup>[↩ Parent](#flowcollector)</sup></sup>



FlowCollectorStatus defines the observed state of FlowCollector

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#flowcollectorstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Conditions represent the latest available observations of an object's state<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where console plugin and flowlogs-pipeline have been deployed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.status.conditions[index]
<sup><sup>[↩ Parent](#flowcollectorstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource. --- This struct is intended for direct use as an array at the field path .status.conditions.  For example, type FooStatus struct{     // Represents the observations of a foo's current state.     // Known .status.conditions.type are: "Available", "Progressing", and "Degraded"     // +patchMergeKey=type     // +patchStrategy=merge     // +listType=map     // +listMapKey=type     Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"` 
     // other fields }

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition. This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase. --- Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be useful (see .node.status.conditions), the ability to deconflict is important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>