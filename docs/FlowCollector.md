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
          Agent selects the flows' tracing agent. Possible values are "ipfix" (default) to use the OpenVSwitch IPFIX collector (only valid if your cluster uses OVN-Kubernetes CNI) or "ebpf" to use NetObserv's eBPF agent. The eBPF agent is not officially released yet, it is provided as a preview.<br/>
          <br/>
            <i>Enum</i>: ipfix, ebpf<br/>
            <i>Default</i>: ipfix<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecclusternetworkoperator">clusterNetworkOperator</a></b></td>
        <td>object</td>
        <td>
          ClusterNetworkOperator contains settings related to the cluster network operator<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsoleplugin">consolePlugin</a></b></td>
        <td>object</td>
        <td>
          ConsolePlugin contains settings related to the console dynamic plugin<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecebpf">ebpf</a></b></td>
        <td>object</td>
        <td>
          EBPF contains the settings of an eBPF-based flow reporter  when the "agent" property is set to "ebpf".<br/>
          <br/>
            <i>Default</i>: map[imagePullPolicy:IfNotPresent]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecflowlogspipeline">flowlogsPipeline</a></b></td>
        <td>object</td>
        <td>
          FlowlogsPipeline contains settings related to the flowlogs-pipeline component<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafana">grafana</a></b></td>
        <td>object</td>
        <td>
          Grafana contains settings related to the grafana configuration. It will create subscription to install Grafana Operator. This feature is experimental, use it for development only<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecipfix">ipfix</a></b></td>
        <td>object</td>
        <td>
          IPFIX contains the settings of an IPFIX-based flow reporter when the "agent" property is set to "ipfix". defined if the ebpf section is already defined<br/>
          <br/>
            <i>Default</i>: map[sampling:400]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka contains settings related to the kafka configuration. It will create subscription to install Strimzi Operator. This feature is experimental, use it for development only<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecloki">loki</a></b></td>
        <td>object</td>
        <td>
          Loki contains settings related to the loki configuration. It will create subscription to install Loki Operator. This feature is experimental, use it for development only<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where console plugin and collector pods are going to be deployed. If empty, the namespace of the operator is going to be used<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprometheus">prometheus</a></b></td>
        <td>object</td>
        <td>
          Prometheus contains settings related to the prometheus configuration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.clusterNetworkOperator
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



ClusterNetworkOperator contains settings related to the cluster network operator

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



ConsolePlugin contains settings related to the console dynamic plugin

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



EBPF contains the settings of an eBPF-based flow reporter  when the "agent" property is set to "ebpf".

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



FlowlogsPipeline contains settings related to the flowlogs-pipeline component

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
          Kind is the workload kind, either DaemonSet or Deployment<br/>
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
            <i>Default</i>: 9090<br/>
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


### FlowCollector.spec.grafana
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Grafana contains settings related to the grafana configuration. It will create subscription to install Grafana Operator. This feature is experimental, use it for development only

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
        <td><b><a href="#flowcollectorspecgrafanadashboardspec">dashboardSpec</a></b></td>
        <td>object</td>
        <td>
          DashboardSpec is the Spec section of Dashboard object will be ignored if instanceSpec is not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespec">instanceSpec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec is the Spec section of Grafana instance<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.dashboardSpec
<sup><sup>[↩ Parent](#flowcollectorspecgrafana)</sup></sup>



DashboardSpec is the Spec section of Dashboard object will be ignored if instanceSpec is not set

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
        <td><b><a href="#flowcollectorspecgrafanadashboardspecconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key from a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.dashboardSpec.configMapRef
<sup><sup>[↩ Parent](#flowcollectorspecgrafanadashboardspec)</sup></sup>



Selects a key from a ConfigMap.

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
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec
<sup><sup>[↩ Parent](#flowcollectorspecgrafana)</sup></sup>



InstanceSpec is the Spec section of Grafana instance

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
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfig">config</a></b></td>
        <td>object</td>
        <td>
          GrafanaConfig is the configuration for grafana<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Should this feature be enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ingress</b></td>
        <td>boolean</td>
        <td>
          Ingress is a flag to create grafana route<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespec)</sup></sup>



GrafanaConfig is the configuration for grafana

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
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigalerting">alerting</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfiganalytics">analytics</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauth">auth</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthanonymous">auth.anonymous</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthazuread">auth.azuread</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthbasic">auth.basic</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthgeneric_oauth">auth.generic_oauth</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthgithub">auth.github</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthgitlab">auth.gitlab</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthgoogle">auth.google</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthldap">auth.ldap</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthokta">auth.okta</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthproxy">auth.proxy</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigauthsaml">auth.saml</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigdashboards">dashboards</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigdatabase">database</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigdataproxy">dataproxy</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigexternal_image_storage">external_image_storage</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigexternal_image_storageazure_blob">external_image_storage.azure_blob</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigexternal_image_storagegcs">external_image_storage.gcs</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigexternal_image_storages3">external_image_storage.s3</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigexternal_image_storagewebdav">external_image_storage.webdav</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigfeature_toggles">feature_toggles</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfiglive">live</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfiglog">log</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfiglogconsole">log.console</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfiglogfrontend">log.frontend</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigmetrics">metrics</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigmetricsgraphite">metrics.graphite</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigpanels">panels</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigpaths">paths</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigplugins">plugins</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigremote_cache">remote_cache</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigrendering">rendering</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigsecurity">security</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigserver">server</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigsmtp">smtp</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigsnapshots">snapshots</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigunified_alerting">unified_alerting</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecgrafanainstancespecconfigusers">users</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.alerting
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>concurrent_render_limit</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error_or_timeout</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>evaluation_timeout_seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>execute_alerts</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_attempts</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodata_or_nullvalues</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>notification_timeout_seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.analytics
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>check_for_updates</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>google_analytics_ua_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>reporting_enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>disable_login_form</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disable_signout_menu</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_cookie_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_maximum_inactive_lifetime_days</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_maximum_inactive_lifetime_duration</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_maximum_lifetime_days</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_maximum_lifetime_duration</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>oauth_auto_login</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>signout_redirect_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sigv4_auth_enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_rotation_interval_minutes</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.anonymous
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>org_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>org_role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.azuread
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_domains</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_groups</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.basic
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.generic_oauth
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_domains</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>api_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>email_attribute_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_attribute_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_attribute_strict</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tls_client_ca</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tls_client_cert</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tls_client_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tls_skip_verify_insecure</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.github
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_organizations</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>api_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>team_ids</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.gitlab
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_groups</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>api_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.google
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_domains</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.ldap
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>config_file</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.okta
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_domains</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_groups</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>api_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auth_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_attribute_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_attribute_strict</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>token_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.proxy
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>auto_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable_login_token</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>header_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>header_property</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>headers</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ldap_sync_ttl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>whitelist</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.auth.saml
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_idp_initiated</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowed_organizations</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_email</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_groups</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_login</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_org</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>assertion_attribute_role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certificate_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>idp_metadata_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_issue_delay</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>metadata_valid_duration</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>org_mapping</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>private_key_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>relay_state</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_values_admin</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_values_editor</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role_values_grafana_admin</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>signature_algorithm</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>single_logout</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.dashboards
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>default_home_dashboard_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>versions_to_keep</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.database
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>ca_cert_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cache_mode</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_cert_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client_key_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conn_max_lifetime</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>log_queries</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_idle_conn</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_open_conn</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>server_cert_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ssl_mode</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.dataproxy
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>logging</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>send_user_header</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.external_image_storage
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>provider</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.external_image_storage.azure_blob
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>account_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>account_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>container_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.external_image_storage.gcs
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key_file</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.external_image_storage.s3
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>access_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bucket_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>region</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.external_image_storage.webdav
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>public_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.feature_toggles
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.live
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allowed_origins</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_connections</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.log
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>filters</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.log.console
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>format</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.log.frontend
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>custom_endpoint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>log_endpoint_burst_limit</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>log_endpoint_requests_per_second_limit</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sample_rate</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sentry_dsn</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.metrics
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>basic_auth_password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>basic_auth_username</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>interval_seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.metrics.graphite
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>prefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.panels
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>disable_sanitize_html</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.paths
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>temp_data_lifetime</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.plugins
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>enable_alpha</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.remote_cache
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>connstr</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.rendering
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>callback_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>concurrent_render_request_limit</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>server_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.security
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>admin_password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>admin_user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allow_embedding</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cookie_samesite</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cookie_secure</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>data_source_proxy_whitelist</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disable_gravatar</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_remember_days</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>strict_transport_security</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>strict_transport_security_max_age_seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>strict_transport_security_preload</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>strict_transport_security_subdomains</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>x_content_type_options</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>x_xss_protection</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.server
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>cert_file</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cert_key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>domain</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable_gzip</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enforce_domain</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>http_addr</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>http_port</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>root_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>router_logging</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serve_from_sub_path</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>socket</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>static_root_path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.smtp
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>cert_file</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ehlo_identity</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>from_address</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>from_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key_file</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>skip_verify</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.snapshots
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>external_enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>external_snapshot_name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>external_snapshot_url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>snapshot_remove_expired</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.unified_alerting
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>evaluation_timeout</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>execute_alerts</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>max_attempts</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>min_interval</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.grafana.instanceSpec.config.users
<sup><sup>[↩ Parent](#flowcollectorspecgrafanainstancespecconfig)</sup></sup>





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
        <td><b>allow_org_create</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allow_sign_up</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auto_assign_org</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auto_assign_org_id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>auto_assign_org_role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>default_theme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editors_can_admin</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>login_hint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>password_hint</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>viewers_can_edit</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.ipfix
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



IPFIX contains the settings of an IPFIX-based flow reporter when the "agent" property is set to "ipfix". defined if the ebpf section is already defined

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
            <i>Default</i>: 60s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. 0 means disabled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Kafka contains settings related to the kafka configuration. It will create subscription to install Strimzi Operator. This feature is experimental, use it for development only

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
        <td><b><a href="#flowcollectorspeckafkainstancespec">instanceSpec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec is the Spec section of Kafka instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>topicName</b></td>
        <td>string</td>
        <td>
          The name of the topic.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatopicspec">topicSpec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec is the Spec section of Kafka topic<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec
<sup><sup>[↩ Parent](#flowcollectorspeckafka)</sup></sup>



InstanceSpec is the Spec section of Kafka instance

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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Configuration of the Kafka cluster.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeczookeeper">zookeeper</a></b></td>
        <td>object</td>
        <td>
          Configuration of the ZooKeeper cluster.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Should this feature be enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespec)</sup></sup>



Configuration of the Kafka cluster.

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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindex">listeners</a></b></td>
        <td>[]object</td>
        <td>
          Configures listeners of Kafka brokers.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          The number of pods in the cluster.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkastorage">storage</a></b></td>
        <td>object</td>
        <td>
          Storage configuration (disk). Cannot be updated.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>config</b></td>
        <td>JSON</td>
        <td>
          Kafka broker config properties<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          The kafka broker version. Defaults to {DefaultKafkaVersion}. Consult the user documentation to understand the process required to upgrade or downgrade the version.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafka)</sup></sup>





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
          Name of the listener. The name will be used to identify the listener and the related Kubernetes objects. The name has to be unique within given a Kafka cluster. The name can consist of lowercase characters and numbers and be up to 11 characters long.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port number used by the listener inside Kafka. The port number has to be unique within a given Kafka cluster. Allowed port numbers are 9092 and higher with the exception of ports 9404 and 9999, which are already used for Prometheus and JMX. Depending on the listener type, the port number might not be the same as the port number that connects Kafka clients.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>tls</b></td>
        <td>boolean</td>
        <td>
          Enables TLS encryption on the listener. This is a required property.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of the listener. Currently the supported types are `internal`, `route`, `loadbalancer`, `nodeport` and `ingress`. 
 * `internal` type exposes Kafka internally only within the Kubernetes cluster. * `route` type uses OpenShift Routes to expose Kafka. * `loadbalancer` type uses LoadBalancer type services to expose Kafka. * `nodeport` type uses NodePort type services to expose Kafka. * `ingress` type uses Kubernetes Nginx Ingress to expose Kafka.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexauthentication">authentication</a></b></td>
        <td>object</td>
        <td>
          Authentication configuration for this listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexconfiguration">configuration</a></b></td>
        <td>object</td>
        <td>
          Additional listener configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindex">networkPolicyPeers</a></b></td>
        <td>[]object</td>
        <td>
          List of peers which should be able to connect to this listener. Peers in this list are combined using a logical OR operation. If this field is empty or missing, all connections will be allowed for this listener. If this field is present and contains at least one item, the listener only allows the traffic which matches at least one item in this list.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].authentication
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindex)</sup></sup>



Authentication configuration for this listener.

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
          Authentication type. `oauth` type uses SASL OAUTHBEARER Authentication. `scram-sha-512` type uses SASL SCRAM-SHA-512 Authentication. `tls` type uses TLS Client Authentication. `tls` type is supported only on TLS listeners.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>accessTokenIsJwt</b></td>
        <td>boolean</td>
        <td>
          Configure whether the access token is treated as JWT. This must be set to `false` if the authorization server returns opaque tokens. Defaults to `true`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>checkAccessTokenType</b></td>
        <td>boolean</td>
        <td>
          Configure whether the access token type check is performed or not. This should be set to `false` if the authorization server does not include 'typ' claim in JWT token. Defaults to `true`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>checkAudience</b></td>
        <td>boolean</td>
        <td>
          Enable or disable audience checking. Audience checks identify the recipients of tokens. If audience checking is enabled, the OAuth Client ID also has to be configured using the `clientId` property. The Kafka broker will reject tokens that do not have its `clientId` in their `aud` (audience) claim.Default value is `false`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>checkIssuer</b></td>
        <td>boolean</td>
        <td>
          Enable or disable issuer checking. By default issuer is checked using the value configured by `validIssuerUri`. Default value is `true`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientAudience</b></td>
        <td>string</td>
        <td>
          The audience to use when making requests to the authorization server's token endpoint. Used for inter-broker authentication and for configuring OAuth 2.0 over PLAIN using the `clientId` and `secret` method.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientId</b></td>
        <td>string</td>
        <td>
          OAuth Client ID which the Kafka broker can use to authenticate against the authorization server and use the introspect endpoint URI.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientScope</b></td>
        <td>string</td>
        <td>
          The scope to use when making requests to the authorization server's token endpoint. Used for inter-broker authentication and for configuring OAuth 2.0 over PLAIN using the `clientId` and `secret` method.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexauthenticationclientsecret">clientSecret</a></b></td>
        <td>object</td>
        <td>
          Link to Kubernetes Secret containing the OAuth client secret which the Kafka broker can use to authenticate against the authorization server and use the introspect endpoint URI.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>customClaimCheck</b></td>
        <td>string</td>
        <td>
          JsonPath filter query to be applied to the JWT token or to the response of the introspection endpoint for additional token validation. Not set by default.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disableTlsHostnameVerification</b></td>
        <td>boolean</td>
        <td>
          Enable or disable TLS hostname verification. Default value is `false`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableECDSA</b></td>
        <td>boolean</td>
        <td>
          Enable or disable ECDSA support by installing BouncyCastle crypto provider. ECDSA support is always enabled. The BouncyCastle libraries are no longer packaged with Strimzi. Value is ignored.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableOauthBearer</b></td>
        <td>boolean</td>
        <td>
          Enable or disable OAuth authentication over SASL_OAUTHBEARER. Default value is `true`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enablePlain</b></td>
        <td>boolean</td>
        <td>
          Enable or disable OAuth authentication over SASL_PLAIN. There is no re-authentication support when this mechanism is used. Default value is `false`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fallbackUserNameClaim</b></td>
        <td>string</td>
        <td>
          The fallback username claim to be used for the user id if the claim specified by `userNameClaim` is not present. This is useful when `client_credentials` authentication only results in the client id being provided in another claim. It only takes effect if `userNameClaim` is set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fallbackUserNamePrefix</b></td>
        <td>string</td>
        <td>
          The prefix to use with the value of `fallbackUserNameClaim` to construct the user id. This only takes effect if `fallbackUserNameClaim` is true, and the value is present for the claim. Mapping usernames and client ids into the same user id space is useful in preventing name collisions.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>introspectionEndpointUri</b></td>
        <td>string</td>
        <td>
          URI of the token introspection endpoint which can be used to validate opaque non-JWT tokens.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jwksEndpointUri</b></td>
        <td>string</td>
        <td>
          URI of the JWKS certificate endpoint, which can be used for local JWT validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jwksExpirySeconds</b></td>
        <td>integer</td>
        <td>
          Configures how often are the JWKS certificates considered valid. The expiry interval has to be at least 60 seconds longer then the refresh interval specified in `jwksRefreshSeconds`. Defaults to 360 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jwksMinRefreshPauseSeconds</b></td>
        <td>integer</td>
        <td>
          The minimum pause between two consecutive refreshes. When an unknown signing key is encountered the refresh is scheduled immediately, but will always wait for this minimum pause. Defaults to 1 second.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jwksRefreshSeconds</b></td>
        <td>integer</td>
        <td>
          Configures how often are the JWKS certificates refreshed. The refresh interval has to be at least 60 seconds shorter then the expiry interval specified in `jwksExpirySeconds`. Defaults to 300 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxSecondsWithoutReauthentication</b></td>
        <td>integer</td>
        <td>
          Maximum number of seconds the authenticated session remains valid without re-authentication. This enables Apache Kafka re-authentication feature, and causes sessions to expire when the access token expires. If the access token expires before max time or if max time is reached, the client has to re-authenticate, otherwise the server will drop the connection. Not set by default - the authenticated session does not expire when the access token expires. This option only applies to SASL_OAUTHBEARER authentication mechanism (when `enableOauthBearer` is `true`).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexauthenticationtlstrustedcertificatesindex">tlsTrustedCertificates</a></b></td>
        <td>[]object</td>
        <td>
          Trusted certificates for TLS connection to the OAuth server.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tokenEndpointUri</b></td>
        <td>string</td>
        <td>
          URI of the Token Endpoint to use with SASL_PLAIN mechanism when the client authenticates with `clientId` and a `secret`. If set, the client can authenticate over SASL_PLAIN by either setting `username` to `clientId`, and setting `password` to client `secret`, or by setting `username` to account username, and `password` to access token prefixed with `$accessToken:`. If this option is not set, the `password` is always interpreted as an access token (without a prefix), and `username` as the account username (a so called 'no-client-credentials' mode).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>userInfoEndpointUri</b></td>
        <td>string</td>
        <td>
          URI of the User Info Endpoint to use as a fallback to obtaining the user id when the Introspection Endpoint does not return information that can be used for the user id.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>userNameClaim</b></td>
        <td>string</td>
        <td>
          Name of the claim from the JWT authentication token, Introspection Endpoint response or User Info Endpoint response which will be used to extract the user id. Defaults to `sub`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>validIssuerUri</b></td>
        <td>string</td>
        <td>
          URI of the token issuer used for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>validTokenType</b></td>
        <td>string</td>
        <td>
          Valid value for the `token_type` attribute returned by the Introspection Endpoint. No default value, and not checked by default.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].authentication.clientSecret
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexauthentication)</sup></sup>



Link to Kubernetes Secret containing the OAuth client secret which the Kafka broker can use to authenticate against the authorization server and use the introspect endpoint URI.

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
          The key under which the secret value is stored in the Kubernetes Secret.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          The name of the Kubernetes Secret containing the secret value.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].authentication.tlsTrustedCertificates[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexauthentication)</sup></sup>





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
        <td><b>certificate</b></td>
        <td>string</td>
        <td>
          The name of the file certificate in the Secret.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          The name of the Secret containing the certificate.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].configuration
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindex)</sup></sup>



Additional listener configuration.

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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexconfigurationbootstrap">bootstrap</a></b></td>
        <td>object</td>
        <td>
          Bootstrap configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexconfigurationbrokercertchainandkey">brokerCertChainAndKey</a></b></td>
        <td>object</td>
        <td>
          Reference to the `Secret` which holds the certificate and private key pair which will be used for this listener. The certificate can optionally contain the whole chain. This field can be used only with listeners with enabled TLS encryption.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexconfigurationbrokersindex">brokers</a></b></td>
        <td>[]object</td>
        <td>
          Per-broker configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          Configures the `Ingress` class that defines which `Ingress` controller will be used. This field can be used only with `ingress` type listener. If not specified, the default Ingress controller will be used.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalTrafficPolicy</b></td>
        <td>string</td>
        <td>
          Specifies whether the service routes external traffic to node-local or cluster-wide endpoints. `Cluster` may cause a second hop to another node and obscures the client source IP. `Local` avoids a second hop for LoadBalancer and Nodeport type services and preserves the client source IP (when supported by the infrastructure). If unspecified, Kubernetes will use `Cluster` as the default.This field can be used only with `loadbalancer` or `nodeport` type listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>finalizers</b></td>
        <td>[]string</td>
        <td>
          A list of finalizers which will be configured for the `LoadBalancer` type Services created for this listener. If supported by the platform, the finalizer `service.kubernetes.io/load-balancer-cleanup` to make sure that the external load balancer is deleted together with the service.For more information, see https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#garbage-collecting-load-balancers. This field can be used only with `loadbalancer` type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ipFamilies</b></td>
        <td>[]string</td>
        <td>
          Specifies the IP Families used by the service. Available options are `IPv4` and `IPv6. If unspecified, Kubernetes will choose the default value based on the `ipFamilyPolicy` setting. Available on Kubernetes 1.20 and newer.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ipFamilyPolicy</b></td>
        <td>string</td>
        <td>
          Specifies the IP Family Policy used by the service. Available options are `SingleStack`, `PreferDualStack` and `RequireDualStack`. `SingleStack` is for a single IP family. `PreferDualStack` is for two IP families on dual-stack configured clusters or a single IP family on single-stack clusters. `RequireDualStack` fails unless there are two IP families on dual-stack configured clusters. If unspecified, Kubernetes will choose the default value based on the service type. Available on Kubernetes 1.20 and newer.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerSourceRanges</b></td>
        <td>[]string</td>
        <td>
          A list of CIDR ranges (for example `10.0.0.0/8` or `130.211.204.1/32`) from which clients can connect to load balancer type listeners. If supported by the platform, traffic through the loadbalancer is restricted to the specified CIDR ranges. This field is applicable only for loadbalancer type services and is ignored if the cloud provider does not support the feature. For more information, see https://v1-17.docs.kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/. This field can be used only with `loadbalancer` type listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxConnectionCreationRate</b></td>
        <td>integer</td>
        <td>
          The maximum connection creation rate we allow in this listener at any time. New connections will be throttled if the limit is reached.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxConnections</b></td>
        <td>integer</td>
        <td>
          The maximum number of connections we allow for this listener in the broker at any time. New connections are blocked if the limit is reached.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>preferredNodePortAddressType</b></td>
        <td>string</td>
        <td>
          Defines which address type should be used as the node address. Available types are: `ExternalDNS`, `ExternalIP`, `InternalDNS`, `InternalIP` and `Hostname`. By default, the addresses will be used in the following order (the first one found will be used): 
 * `ExternalDNS` * `ExternalIP` * `InternalDNS` * `InternalIP` * `Hostname` 
 This field is used to select the preferred address type, which is checked first. If no address is found for this address type, the other types are checked in the default order. This field can only be used with `nodeport` type listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>useServiceDnsDomain</b></td>
        <td>boolean</td>
        <td>
          Configures whether the Kubernetes service DNS domain should be used or not. If set to `true`, the generated addresses will contain the service DNS domain suffix (by default `.cluster.local`, can be configured using environment variable `KUBERNETES_SERVICE_DNS_DOMAIN`). Defaults to `false`.This field can be used only with `internal` type listener.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].configuration.bootstrap
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexconfiguration)</sup></sup>



Bootstrap configuration.

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
        <td><b>alternativeNames</b></td>
        <td>[]string</td>
        <td>
          Additional alternative names for the bootstrap service. The alternative names will be added to the list of subject alternative names of the TLS certificates.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>annotations</b></td>
        <td>JSON</td>
        <td>
          Annotations that will be added to the `Ingress`, `Route`, or `Service` resource. You can use this field to configure DNS providers such as External DNS. This field can be used only with `loadbalancer`, `nodeport`, `route`, or `ingress` type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          The bootstrap host. This field will be used in the Ingress resource or in the Route resource to specify the desired hostname. This field can be used only with `route` (optional) or `ingress` (required) type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>JSON</td>
        <td>
          Labels that will be added to the `Ingress`, `Route`, or `Service` resource. This field can be used only with `loadbalancer`, `nodeport`, `route`, or `ingress` type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerIP</b></td>
        <td>string</td>
        <td>
          The loadbalancer is requested with the IP address specified in this field. This feature depends on whether the underlying cloud provider supports specifying the `loadBalancerIP` when a load balancer is created. This field is ignored if the cloud provider does not support the feature.This field can be used only with `loadbalancer` type listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodePort</b></td>
        <td>integer</td>
        <td>
          Node port for the bootstrap service. This field can be used only with `nodeport` type listener.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].configuration.brokerCertChainAndKey
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexconfiguration)</sup></sup>



Reference to the `Secret` which holds the certificate and private key pair which will be used for this listener. The certificate can optionally contain the whole chain. This field can be used only with listeners with enabled TLS encryption.

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
        <td><b>certificate</b></td>
        <td>string</td>
        <td>
          The name of the file certificate in the Secret.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The name of the private key in the Secret.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          The name of the Secret containing the certificate.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].configuration.brokers[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexconfiguration)</sup></sup>





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
        <td><b>broker</b></td>
        <td>integer</td>
        <td>
          ID of the kafka broker (broker identifier). Broker IDs start from 0 and correspond to the number of broker replicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>advertisedHost</b></td>
        <td>string</td>
        <td>
          The host name which will be used in the brokers' `advertised.brokers`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>advertisedPort</b></td>
        <td>integer</td>
        <td>
          The port number which will be used in the brokers' `advertised.brokers`.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>annotations</b></td>
        <td>JSON</td>
        <td>
          Annotations that will be added to the `Ingress` or `Service` resource. You can use this field to configure DNS providers such as External DNS. This field can be used only with `loadbalancer`, `nodeport`, or `ingress` type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          The broker host. This field will be used in the Ingress resource or in the Route resource to specify the desired hostname. This field can be used only with `route` (optional) or `ingress` (required) type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>JSON</td>
        <td>
          Labels that will be added to the `Ingress`, `Route`, or `Service` resource. This field can be used only with `loadbalancer`, `nodeport`, `route`, or `ingress` type listeners.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerIP</b></td>
        <td>string</td>
        <td>
          The loadbalancer is requested with the IP address specified in this field. This feature depends on whether the underlying cloud provider supports specifying the `loadBalancerIP` when a load balancer is created. This field is ignored if the cloud provider does not support the feature.This field can be used only with `loadbalancer` type listener.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodePort</b></td>
        <td>integer</td>
        <td>
          Node port for the per-broker service. This field can be used only with `nodeport` type listener.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindex)</sup></sup>





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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexipblock">ipBlock</a></b></td>
        <td>object</td>
        <td>
          IpBlock corresponds to the JSON schema field "ipBlock".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          NamespaceSelector corresponds to the JSON schema field "namespaceSelector".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexpodselector">podSelector</a></b></td>
        <td>object</td>
        <td>
          PodSelector corresponds to the JSON schema field "podSelector".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index].ipBlock
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindex)</sup></sup>



IpBlock corresponds to the JSON schema field "ipBlock".

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
        <td><b>cidr</b></td>
        <td>string</td>
        <td>
          Cidr corresponds to the JSON schema field "cidr".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>except</b></td>
        <td>[]string</td>
        <td>
          Except corresponds to the JSON schema field "except".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index].namespaceSelector
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindex)</sup></sup>



NamespaceSelector corresponds to the JSON schema field "namespaceSelector".

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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          MatchExpressions corresponds to the JSON schema field "matchExpressions".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>JSON</td>
        <td>
          MatchLabels corresponds to the JSON schema field "matchLabels".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index].namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexnamespaceselector)</sup></sup>





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
          Key corresponds to the JSON schema field "key".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Operator corresponds to the JSON schema field "operator".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          Values corresponds to the JSON schema field "values".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index].podSelector
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindex)</sup></sup>



PodSelector corresponds to the JSON schema field "podSelector".

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
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexpodselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          MatchExpressions corresponds to the JSON schema field "matchExpressions".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>JSON</td>
        <td>
          MatchLabels corresponds to the JSON schema field "matchLabels".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.listeners[index].networkPolicyPeers[index].podSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkalistenersindexnetworkpolicypeersindexpodselector)</sup></sup>





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
          Key corresponds to the JSON schema field "key".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          Operator corresponds to the JSON schema field "operator".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          Values corresponds to the JSON schema field "values".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.storage
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafka)</sup></sup>



Storage configuration (disk). Cannot be updated.

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
          Storage type, must be either 'ephemeral', 'persistent-claim', or 'jbod'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>deleteClaim</b></td>
        <td>boolean</td>
        <td>
          Specifies if the persistent volume claim has to be deleted when the cluster is un-deployed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          Storage identification number. It is mandatory only for storage volumes defined in a storage of type 'jbod'.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkastorageoverridesindex">overrides</a></b></td>
        <td>[]object</td>
        <td>
          Overrides for individual brokers. The `overrides` field allows to specify a different configuration for different brokers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>selector</b></td>
        <td>JSON</td>
        <td>
          Specifies a specific persistent volume to use. It contains key:value pairs representing labels for selecting such a volume.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>size</b></td>
        <td>string</td>
        <td>
          When type=persistent-claim, defines the size of the persistent volume claim (i.e 1Gi). Mandatory when type=persistent-claim.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>string</td>
        <td>
          When type=ephemeral, defines the total amount of local storage required for this EmptyDir volume (for example 1Gi).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkastoragevolumesindex">volumes</a></b></td>
        <td>[]object</td>
        <td>
          List of volumes as Storage objects representing the JBOD disks array.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.storage.overrides[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkastorage)</sup></sup>





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
        <td><b>broker</b></td>
        <td>integer</td>
        <td>
          Id of the kafka broker (broker identifier).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation for this broker.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.storage.volumes[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkastorage)</sup></sup>





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
          Storage type, must be either 'ephemeral' or 'persistent-claim'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>deleteClaim</b></td>
        <td>boolean</td>
        <td>
          Specifies if the persistent volume claim has to be deleted when the cluster is un-deployed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          Storage identification number. It is mandatory only for storage volumes defined in a storage of type 'jbod'.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeckafkastoragevolumesindexoverridesindex">overrides</a></b></td>
        <td>[]object</td>
        <td>
          Overrides for individual brokers. The `overrides` field allows to specify a different configuration for different brokers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>selector</b></td>
        <td>JSON</td>
        <td>
          Specifies a specific persistent volume to use. It contains key:value pairs representing labels for selecting such a volume.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>size</b></td>
        <td>string</td>
        <td>
          When type=persistent-claim, defines the size of the persistent volume claim (i.e 1Gi). Mandatory when type=persistent-claim.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>string</td>
        <td>
          When type=ephemeral, defines the total amount of local storage required for this EmptyDir volume (for example 1Gi).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.kafka.storage.volumes[index].overrides[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeckafkastoragevolumesindex)</sup></sup>





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
        <td><b>broker</b></td>
        <td>integer</td>
        <td>
          Id of the kafka broker (broker identifier).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation for this broker.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.zookeeper
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespec)</sup></sup>



Configuration of the ZooKeeper cluster.

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
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          The number of pods in the cluster.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeczookeeperstorage">storage</a></b></td>
        <td>object</td>
        <td>
          Storage configuration (disk). Cannot be updated.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.zookeeper.storage
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeczookeeper)</sup></sup>



Storage configuration (disk). Cannot be updated.

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
          Storage type, must be either 'ephemeral' or 'persistent-claim'.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>deleteClaim</b></td>
        <td>boolean</td>
        <td>
          Specifies if the persistent volume claim has to be deleted when the cluster is un-deployed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          Storage identification number. It is mandatory only for storage volumes defined in a storage of type 'jbod'.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkainstancespeczookeeperstorageoverridesindex">overrides</a></b></td>
        <td>[]object</td>
        <td>
          Overrides for individual brokers. The `overrides` field allows to specify a different configuration for different brokers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>selector</b></td>
        <td>JSON</td>
        <td>
          Specifies a specific persistent volume to use. It contains key:value pairs representing labels for selecting such a volume.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>size</b></td>
        <td>string</td>
        <td>
          When type=persistent-claim, defines the size of the persistent volume claim (i.e 1Gi). Mandatory when type=persistent-claim.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>string</td>
        <td>
          When type=ephemeral, defines the total amount of local storage required for this EmptyDir volume (for example 1Gi).<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.instanceSpec.zookeeper.storage.overrides[index]
<sup><sup>[↩ Parent](#flowcollectorspeckafkainstancespeczookeeperstorage)</sup></sup>





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
        <td><b>broker</b></td>
        <td>integer</td>
        <td>
          Id of the kafka broker (broker identifier).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The storage class to use for dynamic volume allocation for this broker.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.topicSpec
<sup><sup>[↩ Parent](#flowcollectorspeckafka)</sup></sup>



InstanceSpec is the Spec section of Kafka topic

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
        <td><b>config</b></td>
        <td>JSON</td>
        <td>
          The topic configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partitions</b></td>
        <td>integer</td>
        <td>
          The number of partitions the topic should have. This cannot be decreased after topic creation. It can be increased after topic creation, but it is important to understand the consequences that has, especially for topics with semantic partitioning. When absent this will default to the broker configuration for `num.partitions`.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          The number of replicas the topic should have. When absent this will default to the broker configuration for `default.replication.factor`.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Loki contains settings related to the loki configuration. It will create subscription to install Loki Operator. This feature is experimental, use it for development only

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
        <td><b><a href="#flowcollectorspeclokiinstancespec">instanceSpec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec is the Spec section of LokiStack instance. Experimental feature. If provided, the Loki operator will be automatically installed. At this moment, the Loki operator itself is experimental and hasn't been officially released by Grafana.<br/>
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
          QuerierURL specifies the address of the Loki querier service, in case it is different from the Loki ingester URL. If empty, the URL value will be used (assuming that the Loki ingester and querier are int he same host). it will be ignored if instanceSpec is specified<br/>
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
        <td><b>timestampLabel</b></td>
        <td>string</td>
        <td>
          TimestampLabel is the label to use for time indexing in Loki. E.g. "TimeReceived", "TimeFlowStart", "TimeFlowEnd".<br/>
          <br/>
            <i>Default</i>: TimeFlowEnd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          URL is the address of an existing Loki service to push the flows to. it will be ignored if instanceSpec is specified<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.instanceSpec
<sup><sup>[↩ Parent](#flowcollectorspecloki)</sup></sup>



InstanceSpec is the Spec section of LokiStack instance. Experimental feature. If provided, the Loki operator will be automatically installed. At this moment, the Loki operator itself is experimental and hasn't been officially released by Grafana.

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
        <td><b>replicationFactor</b></td>
        <td>integer</td>
        <td>
          ReplicationFactor defines the policy for log stream replication.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>size</b></td>
        <td>enum</td>
        <td>
          Size defines one of the support Loki deployment scale out sizes.<br/>
          <br/>
            <i>Enum</i>: 1x.extra-small, 1x.small, 1x.medium<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokiinstancespecstorage">storage</a></b></td>
        <td>object</td>
        <td>
          Storage defines the spec for the object storage endpoint to store logs.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          Storage class name defines the storage class for ingester/querier PVCs.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Should this feature be enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>managementState</b></td>
        <td>enum</td>
        <td>
          ManagementState defines if the CR should be managed by the operator or not. Default is managed.<br/>
          <br/>
            <i>Enum</i>: Managed, Unmanaged<br/>
            <i>Default</i>: Managed<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.instanceSpec.storage
<sup><sup>[↩ Parent](#flowcollectorspeclokiinstancespec)</sup></sup>



Storage defines the spec for the object storage endpoint to store logs.

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
        <td><b><a href="#flowcollectorspeclokiinstancespecstoragesecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret for object storage authentication. Name of a secret in the same namespace as the cluster logging operator.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.instanceSpec.storage.secret
<sup><sup>[↩ Parent](#flowcollectorspeclokiinstancespecstorage)</sup></sup>



Secret for object storage authentication. Name of a secret in the same namespace as the cluster logging operator.

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
          Name of a secret in the namespace configured for object storage secrets.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type of object storage that should be used<br/>
          <br/>
            <i>Enum</i>: azure, gcs, s3, swift<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.prometheus
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Prometheus contains settings related to the prometheus configuration

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
        <td><b><a href="#flowcollectorspecprometheusinstancespec">instanceSpec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec is the Spec section of Prometheus instance<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.prometheus.instanceSpec
<sup><sup>[↩ Parent](#flowcollectorspecprometheus)</sup></sup>



InstanceSpec is the Spec section of Prometheus instance

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
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          Number of replicas of each shard to deploy for a Prometheus deployment. Number of replicas multiplied by shards is the total number of Pods created.<br/>
          <br/>
            <i>Format</i>: int32<br/>
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
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where console plugin and flowlogs-pipeline have been deployed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>