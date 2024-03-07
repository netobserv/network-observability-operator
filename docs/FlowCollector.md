# API Reference

Packages:

- [flows.netobserv.io/v1beta1](#flowsnetobserviov1beta1)
- [flows.netobserv.io/v1beta2](#flowsnetobserviov1beta2)

# flows.netobserv.io/v1beta1

Resource Types:

- [FlowCollector](#flowcollector)




## FlowCollector
<sup><sup>[↩ Parent](#flowsnetobserviov1beta1 )</sup></sup>






`FlowCollector` is the schema for the network flows collection API, which pilots and configures the underlying deployments.

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
      <td>flows.netobserv.io/v1beta1</td>
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
          Defines the desired state of the FlowCollector resource. <br><br> *: the mention of "unsupported", or "deprecated" for a feature throughout this document means that this feature is not officially supported by Red Hat. It might have been, for example, contributed by the community and accepted without a formal agreement for maintenance. The product maintainers might provide some support for these features as a best effort only.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorstatus">status</a></b></td>
        <td>object</td>
        <td>
          `FlowCollectorStatus` defines the observed state of FlowCollector<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec
<sup><sup>[↩ Parent](#flowcollector)</sup></sup>



Defines the desired state of the FlowCollector resource. <br><br> *: the mention of "unsupported", or "deprecated" for a feature throughout this document means that this feature is not officially supported by Red Hat. It might have been, for example, contributed by the community and accepted without a formal agreement for maintenance. The product maintainers might provide some support for these features as a best effort only.

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
        <td><b><a href="#flowcollectorspecagent">agent</a></b></td>
        <td>object</td>
        <td>
          Agent configuration for flows extraction.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsoleplugin">consolePlugin</a></b></td>
        <td>object</td>
        <td>
          `consolePlugin` defines the settings related to the OpenShift Console plugin, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>deploymentModel</b></td>
        <td>enum</td>
        <td>
          `deploymentModel` defines the desired type of deployment for flow processing. Possible values are:<br> - `DIRECT` (default) to make the flow processor listening directly from the agents.<br> - `KAFKA` to make flows sent to a Kafka pipeline before consumption by the processor.<br> Kafka can provide better scalability, resiliency, and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).<br/>
          <br/>
            <i>Enum</i>: DIRECT, KAFKA<br/>
            <i>Default</i>: DIRECT<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindex">exporters</a></b></td>
        <td>[]object</td>
        <td>
          `exporters` define additional optional exporters for custom consumption or storage.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the `spec.deploymentModel` is `KAFKA`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecloki">loki</a></b></td>
        <td>object</td>
        <td>
          `loki`, the flow store, client settings.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where NetObserv pods are deployed.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessor">processor</a></b></td>
        <td>object</td>
        <td>
          `processor` defines the settings of the component that receives the flows from the agent, enriches them, generates metrics, and forwards them to the Loki persistence layer and/or any available exporter.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Agent configuration for flows extraction.

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
        <td><b><a href="#flowcollectorspecagentebpf">ebpf</a></b></td>
        <td>object</td>
        <td>
          `ebpf` describes the settings related to the eBPF-based flow reporter when `spec.agent.type` is set to `EBPF`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfix">ipfix</a></b></td>
        <td>object</td>
        <td>
          `ipfix` [deprecated (*)] - describes the settings related to the IPFIX-based flow reporter when `spec.agent.type` is set to `IPFIX`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          `type` [deprecated (*)] selects the flows tracing agent. The only possible value is `EBPF` (default), to use NetObserv eBPF agent.<br> Previously, using an IPFIX collector was allowed, but was deprecated and it is now removed.<br> Setting `IPFIX` is ignored and still use the eBPF Agent. Since there is only a single option here, this field will be remove in a future API version.<br/>
          <br/>
            <i>Enum</i>: EBPF, IPFIX<br/>
            <i>Default</i>: EBPF<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf
<sup><sup>[↩ Parent](#flowcollectorspecagent)</sup></sup>



`ebpf` describes the settings related to the eBPF-based flow reporter when `spec.agent.type` is set to `EBPF`.

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
          `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending. Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load, however you can expect higher memory consumption and an increased latency in the flow collection.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows. Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load, however you can expect higher memory consumption and an increased latency in the flow collection.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100000<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfdebug">debug</a></b></td>
        <td>object</td>
        <td>
          `debug` allows setting some aspects of the internal configuration of the eBPF agent. This section is aimed exclusively for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>excludeInterfaces</b></td>
        <td>[]string</td>
        <td>
          `excludeInterfaces` contains the interface names that are excluded from flow tracing. An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression. Otherwise it is matched as a case-sensitive string.<br/>
          <br/>
            <i>Default</i>: [lo]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>features</b></td>
        <td>[]enum</td>
        <td>
          List of additional features to enable. They are all disabled by default. Enabling additional features might have performance impacts. Possible values are:<br> - `PacketDrop`: enable the packets drop flows logging feature. This feature requires mounting the kernel debug filesystem, so the eBPF pod has to run as privileged. If the `spec.agent.ebpf.privileged` parameter is not set, an error is reported.<br> - `DNSTracking`: enable the DNS tracking feature.<br> - `FlowRTT` [unsupported (*)]: enable flow latency (RTT) calculations in the eBPF agent during TCP handshakes. This feature better works with `sampling` set to 1.<br><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>interfaces</b></td>
        <td>[]string</td>
        <td>
          `interfaces` contains the interface names from where flows are collected. If empty, the agent fetches all the interfaces in the system, excepting the ones listed in ExcludeInterfaces. An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression. Otherwise it is matched as a case-sensitive string.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBatchSize</b></td>
        <td>integer</td>
        <td>
          `kafkaBatchSize` limits the maximum size of a request in bytes before being sent to a partition. Ignored when not using Kafka. Default: 1MB.<br/>
          <br/>
            <i>Default</i>: 1048576<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` defines the log level for the NetObserv eBPF Agent<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetrics">metrics</a></b></td>
        <td>object</td>
        <td>
          `metrics` defines the eBPF agent configuration regarding metrics<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          Privileged mode for the eBPF Agent container. When ignored or set to `false`, the operator sets granular capabilities (BPF, PERFMON, NET_ADMIN, SYS_RESOURCE) to the container. If for some reason these capabilities cannot be set, such as if an old kernel version not knowing CAP_BPF is in use, then you can turn on this mode for more global privileges. Some agent features require the privileged mode, such as packet drops tracking (see `features`) and SR-IOV support.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfresources">resources</a></b></td>
        <td>object</td>
        <td>
          `resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:800Mi] requests:map[cpu:100m memory:50Mi]]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          Sampling rate of the flow reporter. 100 means one flow on 100 is sent. 0 or 1 means all flows are sampled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 50<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.debug
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf)</sup></sup>



`debug` allows setting some aspects of the internal configuration of the eBPF agent. This section is aimed exclusively for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.

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
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          `env` allows passing custom environment variables to underlying components. Useful for passing some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf)</sup></sup>



`metrics` defines the eBPF agent configuration regarding metrics

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
          Set `enable` to `true` to enable eBPF agent metrics collection.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsserver">server</a></b></td>
        <td>object</td>
        <td>
          Metrics server endpoint configuration for Prometheus scraper<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetrics)</sup></sup>



Metrics server endpoint configuration for Prometheus scraper

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
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The prometheus HTTP port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9102<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS configuration.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsserver)</sup></sup>



TLS configuration.

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
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          `insecureSkipVerify` allows skipping client-side verification of the provided certificate. If set to `true`, the `providedCaFile` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertlsprovided">provided</a></b></td>
        <td>object</td>
        <td>
          TLS configuration when `type` is set to `PROVIDED`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertlsprovidedcafile">providedCaFile</a></b></td>
        <td>object</td>
        <td>
          Reference to the CA file when `type` is set to `PROVIDED`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Select the type of TLS configuration:<br> - `DISABLED` (default) to not configure TLS for the endpoint. - `PROVIDED` to manually provide cert file and a key file. - `AUTO` to use OpenShift auto generated certificate using annotations.<br/>
          <br/>
            <i>Enum</i>: DISABLED, PROVIDED, AUTO<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls.provided
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsservertls)</sup></sup>



TLS configuration when `type` is set to `PROVIDED`.

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls.providedCaFile
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsservertls)</sup></sup>



Reference to the CA file when `type` is set to `PROVIDED`.

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.resources
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf)</sup></sup>



`resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecagentebpfresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix
<sup><sup>[↩ Parent](#flowcollectorspecagent)</sup></sup>



`ipfix` [deprecated (*)] - describes the settings related to the IPFIX-based flow reporter when `spec.agent.type` is set to `IPFIX`.

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
          `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending.<br/>
          <br/>
            <i>Default</i>: 20s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfixclusternetworkoperator">clusterNetworkOperator</a></b></td>
        <td>object</td>
        <td>
          `clusterNetworkOperator` defines the settings related to the OpenShift Cluster Network Operator, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>forceSampleAll</b></td>
        <td>boolean</td>
        <td>
          `forceSampleAll` allows disabling sampling in the IPFIX-based flow reporter. It is not recommended to sample all the traffic with IPFIX, as it might generate cluster instability. If you REALLY want to do that, set this flag to `true`. Use at your own risk. When it is set to `true`, the value of `sampling` is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfixovnkubernetes">ovnKubernetes</a></b></td>
        <td>object</td>
        <td>
          `ovnKubernetes` defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          `sampling` is the sampling rate on the reporter. 100 means one flow on 100 is sent. To ensure cluster stability, it is not possible to set a value below 2. If you really want to sample every packet, which might impact the cluster stability, refer to `forceSampleAll`. Alternatively, you can use the eBPF Agent instead of IPFIX.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 2<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix.clusterNetworkOperator
<sup><sup>[↩ Parent](#flowcollectorspecagentipfix)</sup></sup>



`clusterNetworkOperator` defines the settings related to the OpenShift Cluster Network Operator, when available.

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
          Namespace  where the config map is going to be deployed.<br/>
          <br/>
            <i>Default</i>: openshift-network-operator<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix.ovnKubernetes
<sup><sup>[↩ Parent](#flowcollectorspecagentipfix)</sup></sup>



`ovnKubernetes` defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.

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
          `containerName` defines the name of the container to configure for IPFIX.<br/>
          <br/>
            <i>Default</i>: ovnkube-node<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>daemonSetName</b></td>
        <td>string</td>
        <td>
          `daemonSetName` defines the name of the DaemonSet controlling the OVN-Kubernetes pods.<br/>
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


### FlowCollector.spec.consolePlugin
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



`consolePlugin` defines the settings related to the OpenShift Console plugin, when available.

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
        <td><b><a href="#flowcollectorspecconsolepluginautoscaler">autoscaler</a></b></td>
        <td>object</td>
        <td>
          `autoscaler` spec of a horizontal pod autoscaler to set up for the plugin Deployment.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Enables the console plugin deployment. `spec.loki.enable` must also be `true`<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` for the console plugin backend<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          `port` is the plugin service port. Do not use 9002, which is reserved for metrics.<br/>
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
          `portNaming` defines the configuration of the port-to-service name translation<br/>
          <br/>
            <i>Default</i>: map[enable:true]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginquickfiltersindex">quickFilters</a></b></td>
        <td>[]object</td>
        <td>
          `quickFilters` configures quick filter presets for the Console plugin<br/>
          <br/>
            <i>Default</i>: [map[default:true filter:map[flow_layer:app] name:Applications] map[filter:map[flow_layer:infra] name:Infrastructure] map[default:true filter:map[dst_kind:Pod src_kind:Pod] name:Pods network] map[filter:map[dst_kind:Service] name:Services network]]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>register</b></td>
        <td>boolean</td>
        <td>
          `register` allows, when set to `true`, to automatically register the provided console plugin with the OpenShift Console operator. When set to `false`, you can still register it manually by editing console.operator.openshift.io/cluster with the following command: `oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "netobserv-plugin"}]'`<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          `replicas` defines the number of replicas (pods) to start.<br/>
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
          `resources`, in terms of compute resources, required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:100Mi] requests:map[cpu:100m memory:50Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



`autoscaler` spec of a horizontal pod autoscaler to set up for the plugin Deployment.

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
          `maxReplicas` is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindex">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          `minReplicas` is the lower limit for the number of replicas to which the autoscaler can scale down. It defaults to 1 pod. minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured. Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          `status` describes the desired status regarding deploying an horizontal pod autoscaler.<br> - `DISABLED` does not deploy an horizontal pod autoscaler.<br> - `ENABLED` deploys an horizontal pod autoscaler.<br><br/>
          <br/>
            <i>Enum</i>: DISABLED, ENABLED<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscaler)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresource">containerResource</a></b></td>
        <td>object</td>
        <td>
          containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternal">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobject">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpods">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex)</sup></sup>



containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresource)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternaltarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternal)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselector)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternal)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectdescribedobject">describedObject</a></b></td>
        <td>object</td>
        <td>
          describedObject specifies the descriptions of a object,such as kind,name apiVersion<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjecttarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject)</sup></sup>



describedObject specifies the descriptions of a object,such as kind,name apiVersion

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
          kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          apiVersion is the API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselector)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodstarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpods)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselector)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpods)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexresource)</sup></sup>



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



`portNaming` defines the configuration of the port-to-service name translation

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
          Enable the console plugin port-to-service name translation<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portNames</b></td>
        <td>map[string]string</td>
        <td>
          `portNames` defines additional port names to use in the console, for example, `portNames: {"3100": "loki"}`.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.quickFilters[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



`QuickFilter` defines preset configuration for Console's quick filters

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
        <td><b>filter</b></td>
        <td>map[string]string</td>
        <td>
          `filter` is a set of keys and values to be set when this filter is selected. Each key can relate to a list of values using a coma-separated string, for example, `filter: {"src_namespace": "namespace1,namespace2"}`.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the filter, that is displayed in the Console<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>default</b></td>
        <td>boolean</td>
        <td>
          `default` defines whether this filter should be active by default or not<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.resources
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin)</sup></sup>



`resources`, in terms of compute resources, required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecconsolepluginresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index]
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



`FlowCollectorExporter` defines an additional exporter to send enriched flows to.

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
        <td>enum</td>
        <td>
          `type` selects the type of exporters. The available options are `KAFKA` and `IPFIX`.<br/>
          <br/>
            <i>Enum</i>: KAFKA, IPFIX<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexipfix">ipfix</a></b></td>
        <td>object</td>
        <td>
          IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafka">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka configuration, such as the address and topic, to send enriched flows to.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].ipfix
<sup><sup>[↩ Parent](#flowcollectorspecexportersindex)</sup></sup>



IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.

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
        <td><b>targetHost</b></td>
        <td>string</td>
        <td>
          Address of the IPFIX external receiver<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetPort</b></td>
        <td>integer</td>
        <td>
          Port for the IPFIX external receiver<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>transport</b></td>
        <td>enum</td>
        <td>
          Transport protocol (`TCP` or `UDP`) to be used for the IPFIX connection, defaults to `TCP`.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka
<sup><sup>[↩ Parent](#flowcollectorspecexportersindex)</sup></sup>



Kafka configuration, such as the address and topic, to send enriched flows to.

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
          Kafka topic to use. It must exist. NetObserv does not create it.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkasasl">sasl</a></b></td>
        <td>object</td>
        <td>
          SASL authentication configuration. [Unsupported (*)].<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkatls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafka)</sup></sup>



SASL authentication configuration. [Unsupported (*)].

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
        <td><b><a href="#flowcollectorspecexportersindexkafkasaslclientidreference">clientIDReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client ID<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkasaslclientsecretreference">clientSecretReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type of SASL authentication to use, or `DISABLED` if SASL is not used<br/>
          <br/>
            <i>Enum</i>: DISABLED, PLAIN, SCRAM-SHA512<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl.clientIDReference
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkasasl)</sup></sup>



Reference to the secret or config map containing the client ID

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl.clientSecretReference
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkasasl)</sup></sup>



Reference to the secret or config map containing the client secret

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafka)</sup></sup>



TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.

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
        <td><b><a href="#flowcollectorspecexportersindexkafkatlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkatlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkatls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkatls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the `spec.deploymentModel` is `KAFKA`.

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
          Kafka topic to use. It must exist. NetObserv does not create it.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkasasl">sasl</a></b></td>
        <td>object</td>
        <td>
          SASL authentication configuration. [Unsupported (*)].<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl
<sup><sup>[↩ Parent](#flowcollectorspeckafka)</sup></sup>



SASL authentication configuration. [Unsupported (*)].

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
        <td><b><a href="#flowcollectorspeckafkasaslclientidreference">clientIDReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client ID<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkasaslclientsecretreference">clientSecretReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type of SASL authentication to use, or `DISABLED` if SASL is not used<br/>
          <br/>
            <i>Enum</i>: DISABLED, PLAIN, SCRAM-SHA512<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl.clientIDReference
<sup><sup>[↩ Parent](#flowcollectorspeckafkasasl)</sup></sup>



Reference to the secret or config map containing the client ID

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl.clientSecretReference
<sup><sup>[↩ Parent](#flowcollectorspeckafkasasl)</sup></sup>



Reference to the secret or config map containing the client secret

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls
<sup><sup>[↩ Parent](#flowcollectorspeckafka)</sup></sup>



TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.

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
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



`loki`, the flow store, client settings.

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
        <td><b>authToken</b></td>
        <td>enum</td>
        <td>
          `authToken` describes the way to get a token to authenticate to Loki.<br> - `DISABLED` does not send any token with the request.<br> - `FORWARD` forwards the user token for authorization.<br> - `HOST` [deprecated (*)] - uses the local pod service account to authenticate to Loki.<br> When using the Loki Operator, this must be set to `FORWARD`.<br/>
          <br/>
            <i>Enum</i>: DISABLED, HOST, FORWARD<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>batchSize</b></td>
        <td>integer</td>
        <td>
          `batchSize` is the maximum batch size (in bytes) of logs to accumulate before sending.<br/>
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
          `batchWait` is the maximum time to wait before sending a batch.<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Set `enable` to `true` to store flows in Loki. It is required for the OpenShift Console plugin installation.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxBackoff</b></td>
        <td>string</td>
        <td>
          `maxBackoff` is the maximum backoff time for client connection between retries.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxRetries</b></td>
        <td>integer</td>
        <td>
          `maxRetries` is the maximum number of retries for client connections.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 2<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minBackoff</b></td>
        <td>string</td>
        <td>
          `minBackoff` is the initial backoff time for client connection between retries.<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>querierUrl</b></td>
        <td>string</td>
        <td>
          `querierURL` specifies the address of the Loki querier service, in case it is different from the Loki ingester URL. If empty, the URL value is used (assuming that the Loki ingester and querier are in the same server). When using the Loki Operator, do not set it, since ingestion and queries use the Loki gateway.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readTimeout</b></td>
        <td>string</td>
        <td>
          `readTimeout` is the maximum loki query total time limit. A timeout of zero means no timeout.<br/>
          <br/>
            <i>Default</i>: 30s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>staticLabels</b></td>
        <td>map[string]string</td>
        <td>
          `staticLabels` is a map of common labels to set on each flow.<br/>
          <br/>
            <i>Default</i>: map[app:netobserv-flowcollector]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokistatustls">statusTls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki status URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>statusUrl</b></td>
        <td>string</td>
        <td>
          `statusURL` specifies the address of the Loki `/ready`, `/metrics` and `/config` endpoints, in case it is different from the Loki querier URL. If empty, the `querierURL` value is used. This is useful to show error messages and some context in the frontend. When using the Loki Operator, set it to the Loki HTTP query frontend service, for example https://loki-query-frontend-http.netobserv.svc:3100/. `statusTLS` configuration is used when `statusUrl` is set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenantID</b></td>
        <td>string</td>
        <td>
          `tenantID` is the Loki `X-Scope-OrgID` that identifies the tenant for each request. When using the Loki Operator, set it to `network`, which corresponds to a special tenant mode.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>string</td>
        <td>
          `timeout` is the maximum processor time connection / request limit. A timeout of zero means no timeout.<br/>
          <br/>
            <i>Default</i>: 10s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokitls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          `url` is the address of an existing Loki service to push the flows to. When using the Loki Operator, set it to the Loki gateway service with the `network` tenant set in path, for example https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.statusTls
<sup><sup>[↩ Parent](#flowcollectorspecloki)</sup></sup>



TLS client configuration for Loki status URL.

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
        <td><b><a href="#flowcollectorspeclokistatustlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokistatustlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.statusTls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokistatustls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.statusTls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokistatustls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls
<sup><sup>[↩ Parent](#flowcollectorspecloki)</sup></sup>



TLS client configuration for Loki URL.

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
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokitlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokitls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokitls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor
<sup><sup>[↩ Parent](#flowcollectorspec)</sup></sup>



`processor` defines the settings of the component that receives the flows from the agent, enriches them, generates metrics, and forwards them to the Loki persistence layer and/or any available exporter.

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
        <td><b>addZone</b></td>
        <td>boolean</td>
        <td>
          `addZone` allows availability zone awareness by labelling flows with their source and destination zones. This feature requires the "topology.kubernetes.io/zone" label to be set on nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusterName</b></td>
        <td>string</td>
        <td>
          `clusterName` is the name of the cluster to appear in the flows data. This is useful in a multi-cluster context. When using OpenShift, leave empty to make it automatically determined.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conversationEndTimeout</b></td>
        <td>string</td>
        <td>
          `conversationEndTimeout` is the time to wait after a network flow is received, to consider the conversation ended. This delay is ignored when a FIN packet is collected for TCP flows (see `conversationTerminatingTimeout` instead).<br/>
          <br/>
            <i>Default</i>: 10s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conversationHeartbeatInterval</b></td>
        <td>string</td>
        <td>
          `conversationHeartbeatInterval` is the time to wait between "tick" events of a conversation<br/>
          <br/>
            <i>Default</i>: 30s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conversationTerminatingTimeout</b></td>
        <td>string</td>
        <td>
          `conversationTerminatingTimeout` is the time to wait from detected FIN flag to end a conversation. Only relevant for TCP flows.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessordebug">debug</a></b></td>
        <td>object</td>
        <td>
          `debug` allows setting some aspects of the internal configuration of the flow processor. This section is aimed exclusively for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dropUnusedFields</b></td>
        <td>boolean</td>
        <td>
          `dropUnusedFields` [deprecated (*)] this setting is not used anymore.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableKubeProbes</b></td>
        <td>boolean</td>
        <td>
          `enableKubeProbes` is a flag to enable or disable Kubernetes liveness and readiness probes<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>healthPort</b></td>
        <td>integer</td>
        <td>
          `healthPort` is a collector HTTP port in the Pod that exposes the health check API<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 8080<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscaler">kafkaConsumerAutoscaler</a></b></td>
        <td>object</td>
        <td>
          `kafkaConsumerAutoscaler` is the spec of a horizontal pod autoscaler to set up for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerBatchSize</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerBatchSize` indicates to the broker the maximum batch size, in bytes, that the consumer accepts. Ignored when not using Kafka. Default: 10MB.<br/>
          <br/>
            <i>Default</i>: 10485760<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerQueueCapacity</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerQueueCapacity` defines the capacity of the internal message queue used in the Kafka consumer client. Ignored when not using Kafka.<br/>
          <br/>
            <i>Default</i>: 1000<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerReplicas</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerReplicas` defines the number of replicas (pods) to start for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` of the processor runtime<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logTypes</b></td>
        <td>enum</td>
        <td>
          `logTypes` defines the desired record types to generate. Possible values are:<br> - `FLOWS` (default) to export regular network flows<br> - `CONVERSATIONS` to generate events for started conversations, ended conversations as well as periodic "tick" updates<br> - `ENDED_CONVERSATIONS` to generate only ended conversations events<br> - `ALL` to generate both network flows and all conversations events<br><br/>
          <br/>
            <i>Enum</i>: FLOWS, CONVERSATIONS, ENDED_CONVERSATIONS, ALL<br/>
            <i>Default</i>: FLOWS<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetrics">metrics</a></b></td>
        <td>object</td>
        <td>
          `Metrics` define the processor configuration regarding metrics<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>multiClusterDeployment</b></td>
        <td>boolean</td>
        <td>
          Set `multiClusterDeployment` to `true` to enable multi clusters feature. This adds clusterName label to flows data<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          [Deprecated (*)] Port of the flow collector (host port). It is not used anymore and will be removed in a future version.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 2055<br/>
            <i>Minimum</i>: 1025<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>profilePort</b></td>
        <td>integer</td>
        <td>
          `profilePort` allows setting up a Go pprof profiler listening to this port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorresources">resources</a></b></td>
        <td>object</td>
        <td>
          `resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:800Mi] requests:map[cpu:100m memory:100Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.debug
<sup><sup>[↩ Parent](#flowcollectorspecprocessor)</sup></sup>



`debug` allows setting some aspects of the internal configuration of the flow processor. This section is aimed exclusively for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.

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
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          `env` allows passing custom environment variables to underlying components. Useful for passing some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler
<sup><sup>[↩ Parent](#flowcollectorspecprocessor)</sup></sup>



`kafkaConsumerAutoscaler` is the spec of a horizontal pod autoscaler to set up for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.

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
          `maxReplicas` is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          `minReplicas` is the lower limit for the number of replicas to which the autoscaler can scale down. It defaults to 1 pod. minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured. Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          `status` describes the desired status regarding deploying an horizontal pod autoscaler.<br> - `DISABLED` does not deploy an horizontal pod autoscaler.<br> - `ENABLED` deploys an horizontal pod autoscaler.<br><br/>
          <br/>
            <i>Enum</i>: DISABLED, ENABLED<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscaler)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresource">containerResource</a></b></td>
        <td>object</td>
        <td>
          containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex)</sup></sup>



containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresource)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternaltarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselector)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectdescribedobject">describedObject</a></b></td>
        <td>object</td>
        <td>
          describedObject specifies the descriptions of a object,such as kind,name apiVersion<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjecttarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject)</sup></sup>



describedObject specifies the descriptions of a object,such as kind,name apiVersion

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
          kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          apiVersion is the API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselector)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetric">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodstarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetric)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselectormatchexpressionsindex">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselector)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresourcetarget">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresource)</sup></sup>



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


### FlowCollector.spec.processor.metrics
<sup><sup>[↩ Parent](#flowcollectorspecprocessor)</sup></sup>



`Metrics` define the processor configuration regarding metrics

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
        <td><b>disableAlerts</b></td>
        <td>[]enum</td>
        <td>
          `disableAlerts` is a list of alerts that should be disabled. Possible values are:<br> `NetObservNoFlows`, which is triggered when no flows are being observed for a certain period.<br> `NetObservLokiError`, which is triggered when flows are being dropped due to Loki errors.<br><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ignoreTags</b></td>
        <td>[]string</td>
        <td>
          `ignoreTags` [deprecated (*)] is a list of tags to specify which metrics to ignore. Each metric is associated with a list of tags. More details in https://github.com/netobserv/network-observability-operator/tree/main/controllers/flowlogspipeline/metrics_definitions . Available tags are: `egress`, `ingress`, `flows`, `bytes`, `packets`, `namespaces`, `nodes`, `workloads`, `nodes-flows`, `namespaces-flows`, `workloads-flows`. Namespace-based metrics are covered by both `workloads` and `namespaces` tags, hence it is recommended to always ignore one of them (`workloads` offering a finer granularity).<br> Deprecation notice: use `includeList` instead.<br/>
          <br/>
            <i>Default</i>: [egress packets nodes-flows namespaces-flows workloads-flows namespaces]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>includeList</b></td>
        <td>[]enum</td>
        <td>
          `includeList` is a list of metric names to specify which ones to generate. The names correspond to the names in Prometheus without the prefix. For example, `namespace_egress_packets_total` will show up as `netobserv_namespace_egress_packets_total` in Prometheus. Note that the more metrics you add, the bigger is the impact on Prometheus workload resources. Metrics enabled by default are: `namespace_flows_total`, `node_ingress_bytes_total`, `workload_ingress_bytes_total`, `namespace_drop_packets_total` (when `PacketDrop` feature is enabled), `namespace_rtt_seconds` (when `FlowRTT` feature is enabled), `namespace_dns_latency_seconds` (when `DNSTracking` feature is enabled). More information, with full list of available metrics: https://github.com/netobserv/network-observability-operator/blob/main/docs/Metrics.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsserver">server</a></b></td>
        <td>object</td>
        <td>
          Metrics server endpoint configuration for Prometheus scraper<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetrics)</sup></sup>



Metrics server endpoint configuration for Prometheus scraper

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
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The prometheus HTTP port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9102<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS configuration.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsserver)</sup></sup>



TLS configuration.

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
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          `insecureSkipVerify` allows skipping client-side verification of the provided certificate. If set to `true`, the `providedCaFile` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertlsprovided">provided</a></b></td>
        <td>object</td>
        <td>
          TLS configuration when `type` is set to `PROVIDED`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertlsprovidedcafile">providedCaFile</a></b></td>
        <td>object</td>
        <td>
          Reference to the CA file when `type` is set to `PROVIDED`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Select the type of TLS configuration:<br> - `DISABLED` (default) to not configure TLS for the endpoint. - `PROVIDED` to manually provide cert file and a key file. - `AUTO` to use OpenShift auto generated certificate using annotations.<br/>
          <br/>
            <i>Enum</i>: DISABLED, PROVIDED, AUTO<br/>
            <i>Default</i>: DISABLED<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls.provided
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsservertls)</sup></sup>



TLS configuration when `type` is set to `PROVIDED`.

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls.providedCaFile
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsservertls)</sup></sup>



Reference to the CA file when `type` is set to `PROVIDED`.

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.resources
<sup><sup>[↩ Parent](#flowcollectorspecprocessor)</sup></sup>



`resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecprocessorresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.status
<sup><sup>[↩ Parent](#flowcollector)</sup></sup>



`FlowCollectorStatus` defines the observed state of FlowCollector

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
          `conditions` represent the latest available observations of an object's state<br/>
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



Condition contains details for one aspect of the current state of this API Resource. --- This struct is intended for direct use as an array at the field path .status.conditions.  For example, 
 	type FooStatus struct{ 	    // Represents the observations of a foo's current state. 	    // Known .status.conditions.type are: "Available", "Progressing", and "Degraded" 	    // +patchMergeKey=type 	    // +patchStrategy=merge 	    // +listType=map 	    // +listMapKey=type 	    Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"` 
 	    // other fields 	}

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

# flows.netobserv.io/v1beta2

Resource Types:

- [FlowCollector](#flowcollector)




## FlowCollector
<sup><sup>[↩ Parent](#flowsnetobserviov1beta2 )</sup></sup>






`FlowCollector` is the schema for the network flows collection API, which pilots and configures the underlying deployments.

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
      <td>flows.netobserv.io/v1beta2</td>
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
        <td><b><a href="#flowcollectorspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          Defines the desired state of the FlowCollector resource. <br><br> *: the mention of "unsupported", or "deprecated" for a feature throughout this document means that this feature is not officially supported by Red Hat. It might have been, for example, contributed by the community and accepted without a formal agreement for maintenance. The product maintainers might provide some support for these features as a best effort only.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          `FlowCollectorStatus` defines the observed state of FlowCollector<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec
<sup><sup>[↩ Parent](#flowcollector-1)</sup></sup>



Defines the desired state of the FlowCollector resource. <br><br> *: the mention of "unsupported", or "deprecated" for a feature throughout this document means that this feature is not officially supported by Red Hat. It might have been, for example, contributed by the community and accepted without a formal agreement for maintenance. The product maintainers might provide some support for these features as a best effort only.

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
        <td><b><a href="#flowcollectorspecagent-1">agent</a></b></td>
        <td>object</td>
        <td>
          Agent configuration for flows extraction.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsoleplugin-1">consolePlugin</a></b></td>
        <td>object</td>
        <td>
          `consolePlugin` defines the settings related to the OpenShift Console plugin, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>deploymentModel</b></td>
        <td>enum</td>
        <td>
          `deploymentModel` defines the desired type of deployment for flow processing. Possible values are:<br> - `Direct` (default) to make the flow processor listening directly from the agents.<br> - `Kafka` to make flows sent to a Kafka pipeline before consumption by the processor.<br> Kafka can provide better scalability, resiliency, and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).<br/>
          <br/>
            <i>Enum</i>: Direct, Kafka<br/>
            <i>Default</i>: Direct<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindex-1">exporters</a></b></td>
        <td>[]object</td>
        <td>
          `exporters` define additional optional exporters for custom consumption or storage.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafka-1">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the `spec.deploymentModel` is `Kafka`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecloki-1">loki</a></b></td>
        <td>object</td>
        <td>
          `loki`, the flow store, client settings.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where NetObserv pods are deployed.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessor-1">processor</a></b></td>
        <td>object</td>
        <td>
          `processor` defines the settings of the component that receives the flows from the agent, enriches them, generates metrics, and forwards them to the Loki persistence layer and/or any available exporter.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



Agent configuration for flows extraction.

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
        <td><b><a href="#flowcollectorspecagentebpf-1">ebpf</a></b></td>
        <td>object</td>
        <td>
          `ebpf` describes the settings related to the eBPF-based flow reporter when `spec.agent.type` is set to `eBPF`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfix-1">ipfix</a></b></td>
        <td>object</td>
        <td>
          `ipfix` [deprecated (*)] - describes the settings related to the IPFIX-based flow reporter when `spec.agent.type` is set to `IPFIX`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          `type` [deprecated (*)] selects the flows tracing agent. The only possible value is `eBPF` (default), to use NetObserv eBPF agent.<br> Previously, using an IPFIX collector was allowed, but was deprecated and it is now removed.<br> Setting `IPFIX` is ignored and still use the eBPF Agent. Since there is only a single option here, this field will be remove in a future API version.<br/>
          <br/>
            <i>Enum</i>: eBPF, IPFIX<br/>
            <i>Default</i>: eBPF<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf
<sup><sup>[↩ Parent](#flowcollectorspecagent-1)</sup></sup>



`ebpf` describes the settings related to the eBPF-based flow reporter when `spec.agent.type` is set to `eBPF`.

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
        <td><b><a href="#flowcollectorspecagentebpfadvanced">advanced</a></b></td>
        <td>object</td>
        <td>
          `advanced` allows setting some aspects of the internal configuration of the eBPF agent. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheActiveTimeout</b></td>
        <td>string</td>
        <td>
          `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending. Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load, however you can expect higher memory consumption and an increased latency in the flow collection.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows. Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load, however you can expect higher memory consumption and an increased latency in the flow collection.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100000<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>excludeInterfaces</b></td>
        <td>[]string</td>
        <td>
          `excludeInterfaces` contains the interface names that are excluded from flow tracing. An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression. Otherwise it is matched as a case-sensitive string.<br/>
          <br/>
            <i>Default</i>: [lo]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>features</b></td>
        <td>[]enum</td>
        <td>
          List of additional features to enable. They are all disabled by default. Enabling additional features might have performance impacts. Possible values are:<br> - `PacketDrop`: enable the packets drop flows logging feature. This feature requires mounting the kernel debug filesystem, so the eBPF pod has to run as privileged. If the `spec.agent.ebpf.privileged` parameter is not set, an error is reported.<br> - `DNSTracking`: enable the DNS tracking feature.<br> - `FlowRTT`: enable flow latency (RTT) calculations in the eBPF agent during TCP handshakes. This feature better works with `sampling` set to 1.<br><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>interfaces</b></td>
        <td>[]string</td>
        <td>
          `interfaces` contains the interface names from where flows are collected. If empty, the agent fetches all the interfaces in the system, excepting the ones listed in ExcludeInterfaces. An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression. Otherwise it is matched as a case-sensitive string.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaBatchSize</b></td>
        <td>integer</td>
        <td>
          `kafkaBatchSize` limits the maximum size of a request in bytes before being sent to a partition. Ignored when not using Kafka. Default: 1MB.<br/>
          <br/>
            <i>Default</i>: 1048576<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` defines the log level for the NetObserv eBPF Agent<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetrics-1">metrics</a></b></td>
        <td>object</td>
        <td>
          `metrics` defines the eBPF agent configuration regarding metrics<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          Privileged mode for the eBPF Agent container. When ignored or set to `false`, the operator sets granular capabilities (BPF, PERFMON, NET_ADMIN, SYS_RESOURCE) to the container. If for some reason these capabilities cannot be set, such as if an old kernel version not knowing CAP_BPF is in use, then you can turn on this mode for more global privileges. Some agent features require the privileged mode, such as packet drops tracking (see `features`) and SR-IOV support.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfresources-1">resources</a></b></td>
        <td>object</td>
        <td>
          `resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:800Mi] requests:map[cpu:100m memory:50Mi]]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          Sampling rate of the flow reporter. 100 means one flow on 100 is sent. 0 or 1 means all flows are sampled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 50<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.advanced
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf-1)</sup></sup>



`advanced` allows setting some aspects of the internal configuration of the eBPF agent. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.

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
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          `env` allows passing custom environment variables to underlying components. Useful for passing some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf-1)</sup></sup>



`metrics` defines the eBPF agent configuration regarding metrics

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
          Set `enable` to `true` to enable eBPF agent metrics collection.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsserver-1">server</a></b></td>
        <td>object</td>
        <td>
          Metrics server endpoint configuration for Prometheus scraper<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetrics-1)</sup></sup>



Metrics server endpoint configuration for Prometheus scraper

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
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The prometheus HTTP port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9102<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertls-1">tls</a></b></td>
        <td>object</td>
        <td>
          TLS configuration.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsserver-1)</sup></sup>



TLS configuration.

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
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          `insecureSkipVerify` allows skipping client-side verification of the provided certificate. If set to `true`, the `providedCaFile` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertlsprovided-1">provided</a></b></td>
        <td>object</td>
        <td>
          TLS configuration when `type` is set to `Provided`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentebpfmetricsservertlsprovidedcafile-1">providedCaFile</a></b></td>
        <td>object</td>
        <td>
          Reference to the CA file when `type` is set to `Provided`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Select the type of TLS configuration:<br> - `Disabled` (default) to not configure TLS for the endpoint. - `Provided` to manually provide cert file and a key file. - `Auto` to use OpenShift auto generated certificate using annotations.<br/>
          <br/>
            <i>Enum</i>: Disabled, Provided, Auto<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls.provided
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsservertls-1)</sup></sup>



TLS configuration when `type` is set to `Provided`.

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.metrics.server.tls.providedCaFile
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfmetricsservertls-1)</sup></sup>



Reference to the CA file when `type` is set to `Provided`.

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.resources
<sup><sup>[↩ Parent](#flowcollectorspecagentebpf-1)</sup></sup>



`resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecagentebpfresourcesclaimsindex-1">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ebpf.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecagentebpfresources-1)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix
<sup><sup>[↩ Parent](#flowcollectorspecagent-1)</sup></sup>



`ipfix` [deprecated (*)] - describes the settings related to the IPFIX-based flow reporter when `spec.agent.type` is set to `IPFIX`.

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
          `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending.<br/>
          <br/>
            <i>Default</i>: 20s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>cacheMaxFlows</b></td>
        <td>integer</td>
        <td>
          `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfixclusternetworkoperator-1">clusterNetworkOperator</a></b></td>
        <td>object</td>
        <td>
          `clusterNetworkOperator` defines the settings related to the OpenShift Cluster Network Operator, when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>forceSampleAll</b></td>
        <td>boolean</td>
        <td>
          `forceSampleAll` allows disabling sampling in the IPFIX-based flow reporter. It is not recommended to sample all the traffic with IPFIX, as it might generate cluster instability. If you REALLY want to do that, set this flag to `true`. Use at your own risk. When it is set to `true`, the value of `sampling` is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecagentipfixovnkubernetes-1">ovnKubernetes</a></b></td>
        <td>object</td>
        <td>
          `ovnKubernetes` defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sampling</b></td>
        <td>integer</td>
        <td>
          `sampling` is the sampling rate on the reporter. 100 means one flow on 100 is sent. To ensure cluster stability, it is not possible to set a value below 2. If you really want to sample every packet, which might impact the cluster stability, refer to `forceSampleAll`. Alternatively, you can use the eBPF Agent instead of IPFIX.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 400<br/>
            <i>Minimum</i>: 2<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix.clusterNetworkOperator
<sup><sup>[↩ Parent](#flowcollectorspecagentipfix-1)</sup></sup>



`clusterNetworkOperator` defines the settings related to the OpenShift Cluster Network Operator, when available.

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
          Namespace  where the config map is going to be deployed.<br/>
          <br/>
            <i>Default</i>: openshift-network-operator<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.agent.ipfix.ovnKubernetes
<sup><sup>[↩ Parent](#flowcollectorspecagentipfix-1)</sup></sup>



`ovnKubernetes` defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.

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
          `containerName` defines the name of the container to configure for IPFIX.<br/>
          <br/>
            <i>Default</i>: ovnkube-node<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>daemonSetName</b></td>
        <td>string</td>
        <td>
          `daemonSetName` defines the name of the DaemonSet controlling the OVN-Kubernetes pods.<br/>
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


### FlowCollector.spec.consolePlugin
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



`consolePlugin` defines the settings related to the OpenShift Console plugin, when available.

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
        <td><b><a href="#flowcollectorspecconsolepluginadvanced">advanced</a></b></td>
        <td>object</td>
        <td>
          `advanced` allows setting some aspects of the internal configuration of the console plugin. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscaler-1">autoscaler</a></b></td>
        <td>object</td>
        <td>
          `autoscaler` spec of a horizontal pod autoscaler to set up for the plugin Deployment.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Enables the console plugin deployment. `spec.loki.enable` must also be `true`<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` for the console plugin backend<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginportnaming-1">portNaming</a></b></td>
        <td>object</td>
        <td>
          `portNaming` defines the configuration of the port-to-service name translation<br/>
          <br/>
            <i>Default</i>: map[enable:true]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginquickfiltersindex-1">quickFilters</a></b></td>
        <td>[]object</td>
        <td>
          `quickFilters` configures quick filter presets for the Console plugin<br/>
          <br/>
            <i>Default</i>: [map[default:true filter:map[flow_layer:app] name:Applications] map[filter:map[flow_layer:infra] name:Infrastructure] map[default:true filter:map[dst_kind:Pod src_kind:Pod] name:Pods network] map[filter:map[dst_kind:Service] name:Services network]]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          `replicas` defines the number of replicas (pods) to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginresources-1">resources</a></b></td>
        <td>object</td>
        <td>
          `resources`, in terms of compute resources, required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:100Mi] requests:map[cpu:100m memory:50Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.advanced
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin-1)</sup></sup>



`advanced` allows setting some aspects of the internal configuration of the console plugin. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.

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
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          `args` allows passing custom arguments to underlying components. Useful for overriding some parameters, such as an url or a configuration path, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          `env` allows passing custom environment variables to underlying components. Useful for passing some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          `port` is the plugin service port. Do not use 9002, which is reserved for metrics.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9001<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>register</b></td>
        <td>boolean</td>
        <td>
          `register` allows, when set to `true`, to automatically register the provided console plugin with the OpenShift Console operator. When set to `false`, you can still register it manually by editing console.operator.openshift.io/cluster with the following command: `oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "netobserv-plugin"}]'`<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin-1)</sup></sup>



`autoscaler` spec of a horizontal pod autoscaler to set up for the plugin Deployment.

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
          `maxReplicas` is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindex-1">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          `minReplicas` is the lower limit for the number of replicas to which the autoscaler can scale down. It defaults to 1 pod. minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured. Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          `status` describes the desired status regarding deploying an horizontal pod autoscaler.<br> - `Disabled` does not deploy an horizontal pod autoscaler.<br> - `Enabled` deploys an horizontal pod autoscaler.<br><br/>
          <br/>
            <i>Enum</i>: Disabled, Enabled<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscaler-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresource-1">containerResource</a></b></td>
        <td>object</td>
        <td>
          containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternal-1">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobject-1">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpods-1">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexresource-1">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex-1)</sup></sup>



containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresourcetarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexcontainerresource-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternaltarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternal-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternalmetricselector-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexexternal-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectdescribedobject-1">describedObject</a></b></td>
        <td>object</td>
        <td>
          describedObject specifies the descriptions of a object,such as kind,name apiVersion<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjecttarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject-1)</sup></sup>



describedObject specifies the descriptions of a object,such as kind,name apiVersion

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
          kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          apiVersion is the API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobjectmetricselector-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexobject-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodstarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpods-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpodsmetricselector-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexpods-1)</sup></sup>



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


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecconsolepluginautoscalermetricsindexresourcetarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.autoscaler.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginautoscalermetricsindexresource-1)</sup></sup>



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
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin-1)</sup></sup>



`portNaming` defines the configuration of the port-to-service name translation

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
          Enable the console plugin port-to-service name translation<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portNames</b></td>
        <td>map[string]string</td>
        <td>
          `portNames` defines additional port names to use in the console, for example, `portNames: {"3100": "loki"}`.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.quickFilters[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin-1)</sup></sup>



`QuickFilter` defines preset configuration for Console's quick filters

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
        <td><b>filter</b></td>
        <td>map[string]string</td>
        <td>
          `filter` is a set of keys and values to be set when this filter is selected. Each key can relate to a list of values using a coma-separated string, for example, `filter: {"src_namespace": "namespace1,namespace2"}`.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the filter, that is displayed in the Console<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>default</b></td>
        <td>boolean</td>
        <td>
          `default` defines whether this filter should be active by default or not<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.resources
<sup><sup>[↩ Parent](#flowcollectorspecconsoleplugin-1)</sup></sup>



`resources`, in terms of compute resources, required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecconsolepluginresourcesclaimsindex-1">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.consolePlugin.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecconsolepluginresources-1)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index]
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



`FlowCollectorExporter` defines an additional exporter to send enriched flows to.

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
        <td>enum</td>
        <td>
          `type` selects the type of exporters. The available options are `Kafka` and `IPFIX`.<br/>
          <br/>
            <i>Enum</i>: Kafka, IPFIX<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexipfix-1">ipfix</a></b></td>
        <td>object</td>
        <td>
          IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafka-1">kafka</a></b></td>
        <td>object</td>
        <td>
          Kafka configuration, such as the address and topic, to send enriched flows to.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].ipfix
<sup><sup>[↩ Parent](#flowcollectorspecexportersindex-1)</sup></sup>



IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.

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
        <td><b>targetHost</b></td>
        <td>string</td>
        <td>
          Address of the IPFIX external receiver<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetPort</b></td>
        <td>integer</td>
        <td>
          Port for the IPFIX external receiver<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>transport</b></td>
        <td>enum</td>
        <td>
          Transport protocol (`TCP` or `UDP`) to be used for the IPFIX connection, defaults to `TCP`.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka
<sup><sup>[↩ Parent](#flowcollectorspecexportersindex-1)</sup></sup>



Kafka configuration, such as the address and topic, to send enriched flows to.

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
          Kafka topic to use. It must exist. NetObserv does not create it.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkasasl-1">sasl</a></b></td>
        <td>object</td>
        <td>
          SASL authentication configuration. [Unsupported (*)].<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkatls-1">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafka-1)</sup></sup>



SASL authentication configuration. [Unsupported (*)].

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
        <td><b><a href="#flowcollectorspecexportersindexkafkasaslclientidreference-1">clientIDReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client ID<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkasaslclientsecretreference-1">clientSecretReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type of SASL authentication to use, or `Disabled` if SASL is not used<br/>
          <br/>
            <i>Enum</i>: Disabled, Plain, ScramSHA512<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl.clientIDReference
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkasasl-1)</sup></sup>



Reference to the secret or config map containing the client ID

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.sasl.clientSecretReference
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkasasl-1)</sup></sup>



Reference to the secret or config map containing the client secret

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafka-1)</sup></sup>



TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.

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
        <td><b><a href="#flowcollectorspecexportersindexkafkatlscacert-1">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecexportersindexkafkatlsusercert-1">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkatls-1)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.exporters[index].kafka.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspecexportersindexkafkatls-1)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the `spec.deploymentModel` is `Kafka`.

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
          Kafka topic to use. It must exist. NetObserv does not create it.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkasasl-1">sasl</a></b></td>
        <td>object</td>
        <td>
          SASL authentication configuration. [Unsupported (*)].<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatls-1">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl
<sup><sup>[↩ Parent](#flowcollectorspeckafka-1)</sup></sup>



SASL authentication configuration. [Unsupported (*)].

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
        <td><b><a href="#flowcollectorspeckafkasaslclientidreference-1">clientIDReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client ID<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkasaslclientsecretreference-1">clientSecretReference</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret or config map containing the client secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type of SASL authentication to use, or `Disabled` if SASL is not used<br/>
          <br/>
            <i>Enum</i>: Disabled, Plain, ScramSHA512<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl.clientIDReference
<sup><sup>[↩ Parent](#flowcollectorspeckafkasasl-1)</sup></sup>



Reference to the secret or config map containing the client ID

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.sasl.clientSecretReference
<sup><sup>[↩ Parent](#flowcollectorspeckafkasasl-1)</sup></sup>



Reference to the secret or config map containing the client secret

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls
<sup><sup>[↩ Parent](#flowcollectorspeckafka-1)</sup></sup>



TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.

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
        <td><b><a href="#flowcollectorspeckafkatlscacert-1">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeckafkatlsusercert-1">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls-1)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.kafka.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeckafkatls-1)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



`loki`, the flow store, client settings.

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
        <td><b><a href="#flowcollectorspeclokiadvanced">advanced</a></b></td>
        <td>object</td>
        <td>
          `advanced` allows setting some aspects of the internal configuration of the Loki clients. This section is aimed mostly for debugging and fine-grained performance optimizations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enable</b></td>
        <td>boolean</td>
        <td>
          Set `enable` to `true` to store flows in Loki. It is required for the OpenShift Console plugin installation.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokilokistack">lokiStack</a></b></td>
        <td>object</td>
        <td>
          Loki configuration for `LokiStack` mode. This is useful for an easy loki-operator configuration. It is ignored for other modes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimanual">manual</a></b></td>
        <td>object</td>
        <td>
          Loki configuration for `Manual` mode. This is the most flexible configuration. It is ignored for other modes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimicroservices">microservices</a></b></td>
        <td>object</td>
        <td>
          Loki configuration for `Microservices` mode. Use this option when Loki is installed using the microservices deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#microservices-mode). It is ignored for other modes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>enum</td>
        <td>
          `mode` must be set according to the installation mode of Loki:<br> - Use `LokiStack` when Loki is managed using the Loki Operator<br> - Use `Monolithic` when Loki is installed as a monolithic workload<br> - Use `Microservices` when Loki is installed as microservices, but without Loki Operator<br> - Use `Manual` if none of the options above match your setup<br><br/>
          <br/>
            <i>Enum</i>: Manual, LokiStack, Monolithic, Microservices<br/>
            <i>Default</i>: Monolithic<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimonolithic">monolithic</a></b></td>
        <td>object</td>
        <td>
          Loki configuration for `Monolithic` mode. Use this option when Loki is installed using the monolithic deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#monolithic-mode). It is ignored for other modes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readTimeout</b></td>
        <td>string</td>
        <td>
          `readTimeout` is the maximum console plugin loki query total time limit. A timeout of zero means no timeout.<br/>
          <br/>
            <i>Default</i>: 30s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeBatchSize</b></td>
        <td>integer</td>
        <td>
          `writeBatchSize` is the maximum batch size (in bytes) of Loki logs to accumulate before sending.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: 102400<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeBatchWait</b></td>
        <td>string</td>
        <td>
          `writeBatchWait` is the maximum time to wait before sending a Loki batch.<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeTimeout</b></td>
        <td>string</td>
        <td>
          `writeTimeout` is the maximum Loki time connection / request limit. A timeout of zero means no timeout.<br/>
          <br/>
            <i>Default</i>: 10s<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.advanced
<sup><sup>[↩ Parent](#flowcollectorspecloki-1)</sup></sup>



`advanced` allows setting some aspects of the internal configuration of the Loki clients. This section is aimed mostly for debugging and fine-grained performance optimizations.

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
        <td><b>staticLabels</b></td>
        <td>map[string]string</td>
        <td>
          `staticLabels` is a map of common labels to set on each flow in Loki storage.<br/>
          <br/>
            <i>Default</i>: map[app:netobserv-flowcollector]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeMaxBackoff</b></td>
        <td>string</td>
        <td>
          `writeMaxBackoff` is the maximum backoff time for Loki client connection between retries.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeMaxRetries</b></td>
        <td>integer</td>
        <td>
          `writeMaxRetries` is the maximum number of retries for Loki client connections.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 2<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>writeMinBackoff</b></td>
        <td>string</td>
        <td>
          `writeMinBackoff` is the initial backoff time for Loki client connection between retries.<br/>
          <br/>
            <i>Default</i>: 1s<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.lokiStack
<sup><sup>[↩ Parent](#flowcollectorspecloki-1)</sup></sup>



Loki configuration for `LokiStack` mode. This is useful for an easy loki-operator configuration. It is ignored for other modes.

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
          Name of an existing LokiStack resource to use.<br/>
          <br/>
            <i>Default</i>: loki<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where this `LokiStack` resource is located. If omited, it is assumed to be the same as `spec.namespace`.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual
<sup><sup>[↩ Parent](#flowcollectorspecloki-1)</sup></sup>



Loki configuration for `Manual` mode. This is the most flexible configuration. It is ignored for other modes.

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
        <td><b>authToken</b></td>
        <td>enum</td>
        <td>
          `authToken` describes the way to get a token to authenticate to Loki.<br> - `Disabled` does not send any token with the request.<br> - `Forward` forwards the user token for authorization.<br> - `Host` [deprecated (*)] - uses the local pod service account to authenticate to Loki.<br> When using the Loki Operator, this must be set to `Forward`.<br/>
          <br/>
            <i>Enum</i>: Disabled, Host, Forward<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ingesterUrl</b></td>
        <td>string</td>
        <td>
          `ingesterUrl` is the address of an existing Loki ingester service to push the flows to. When using the Loki Operator, set it to the Loki gateway service with the `network` tenant set in path, for example https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>querierUrl</b></td>
        <td>string</td>
        <td>
          `querierUrl` specifies the address of the Loki querier service. When using the Loki Operator, set it to the Loki gateway service with the `network` tenant set in path, for example https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimanualstatustls">statusTls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki status URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>statusUrl</b></td>
        <td>string</td>
        <td>
          `statusUrl` specifies the address of the Loki `/ready`, `/metrics` and `/config` endpoints, in case it is different from the Loki querier URL. If empty, the `querierUrl` value is used. This is useful to show error messages and some context in the frontend. When using the Loki Operator, set it to the Loki HTTP query frontend service, for example https://loki-query-frontend-http.netobserv.svc:3100/. `statusTLS` configuration is used when `statusUrl` is set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenantID</b></td>
        <td>string</td>
        <td>
          `tenantID` is the Loki `X-Scope-OrgID` that identifies the tenant for each request. When using the Loki Operator, set it to `network`, which corresponds to a special tenant mode.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimanualtls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki URL.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.statusTls
<sup><sup>[↩ Parent](#flowcollectorspeclokimanual)</sup></sup>



TLS client configuration for Loki status URL.

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
        <td><b><a href="#flowcollectorspeclokimanualstatustlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimanualstatustlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.statusTls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimanualstatustls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.statusTls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimanualstatustls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.tls
<sup><sup>[↩ Parent](#flowcollectorspeclokimanual)</sup></sup>



TLS client configuration for Loki URL.

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
        <td><b><a href="#flowcollectorspeclokimanualtlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimanualtlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimanualtls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.manual.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimanualtls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.microservices
<sup><sup>[↩ Parent](#flowcollectorspecloki-1)</sup></sup>



Loki configuration for `Microservices` mode. Use this option when Loki is installed using the microservices deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#microservices-mode). It is ignored for other modes.

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
        <td><b>ingesterUrl</b></td>
        <td>string</td>
        <td>
          `ingesterUrl` is the address of an existing Loki ingester service to push the flows to.<br/>
          <br/>
            <i>Default</i>: http://loki-distributor:3100/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>querierUrl</b></td>
        <td>string</td>
        <td>
          `querierURL` specifies the address of the Loki querier service.<br/>
          <br/>
            <i>Default</i>: http://loki-query-frontend:3100/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenantID</b></td>
        <td>string</td>
        <td>
          `tenantID` is the Loki `X-Scope-OrgID` header that identifies the tenant for each request.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimicroservicestls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki URL.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.microservices.tls
<sup><sup>[↩ Parent](#flowcollectorspeclokimicroservices)</sup></sup>



TLS client configuration for Loki URL.

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
        <td><b><a href="#flowcollectorspeclokimicroservicestlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimicroservicestlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.microservices.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimicroservicestls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.microservices.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimicroservicestls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.monolithic
<sup><sup>[↩ Parent](#flowcollectorspecloki-1)</sup></sup>



Loki configuration for `Monolithic` mode. Use this option when Loki is installed using the monolithic deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#monolithic-mode). It is ignored for other modes.

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
        <td><b>tenantID</b></td>
        <td>string</td>
        <td>
          `tenantID` is the Loki `X-Scope-OrgID` header that identifies the tenant for each request.<br/>
          <br/>
            <i>Default</i>: netobserv<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimonolithictls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS client configuration for Loki URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          `url` is the unique address of an existing Loki service that points to both the ingester and the querier.<br/>
          <br/>
            <i>Default</i>: http://loki:3100/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.monolithic.tls
<sup><sup>[↩ Parent](#flowcollectorspeclokimonolithic)</sup></sup>



TLS client configuration for Loki URL.

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
        <td><b><a href="#flowcollectorspeclokimonolithictlscacert">caCert</a></b></td>
        <td>object</td>
        <td>
          `caCert` defines the reference of the certificate for the Certificate Authority<br/>
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
          `insecureSkipVerify` allows skipping client-side verification of the server certificate. If set to `true`, the `caCert` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspeclokimonolithictlsusercert">userCert</a></b></td>
        <td>object</td>
        <td>
          `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.monolithic.tls.caCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimonolithictls)</sup></sup>



`caCert` defines the reference of the certificate for the Certificate Authority

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.loki.monolithic.tls.userCert
<sup><sup>[↩ Parent](#flowcollectorspeclokimonolithictls)</sup></sup>



`userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor
<sup><sup>[↩ Parent](#flowcollectorspec-1)</sup></sup>



`processor` defines the settings of the component that receives the flows from the agent, enriches them, generates metrics, and forwards them to the Loki persistence layer and/or any available exporter.

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
        <td><b>addZone</b></td>
        <td>boolean</td>
        <td>
          `addZone` allows availability zone awareness by labelling flows with their source and destination zones. This feature requires the "topology.kubernetes.io/zone" label to be set on nodes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessoradvanced">advanced</a></b></td>
        <td>object</td>
        <td>
          `advanced` allows setting some aspects of the internal configuration of the flow processor. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusterName</b></td>
        <td>string</td>
        <td>
          `clusterName` is the name of the cluster to appear in the flows data. This is useful in a multi-cluster context. When using OpenShift, leave empty to make it automatically determined.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>enum</td>
        <td>
          `imagePullPolicy` is the Kubernetes pull policy for the image defined above<br/>
          <br/>
            <i>Enum</i>: IfNotPresent, Always, Never<br/>
            <i>Default</i>: IfNotPresent<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscaler-1">kafkaConsumerAutoscaler</a></b></td>
        <td>object</td>
        <td>
          `kafkaConsumerAutoscaler` is the spec of a horizontal pod autoscaler to set up for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerBatchSize</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerBatchSize` indicates to the broker the maximum batch size, in bytes, that the consumer accepts. Ignored when not using Kafka. Default: 10MB.<br/>
          <br/>
            <i>Default</i>: 10485760<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerQueueCapacity</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerQueueCapacity` defines the capacity of the internal message queue used in the Kafka consumer client. Ignored when not using Kafka.<br/>
          <br/>
            <i>Default</i>: 1000<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kafkaConsumerReplicas</b></td>
        <td>integer</td>
        <td>
          `kafkaConsumerReplicas` defines the number of replicas (pods) to start for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logLevel</b></td>
        <td>enum</td>
        <td>
          `logLevel` of the processor runtime<br/>
          <br/>
            <i>Enum</i>: trace, debug, info, warn, error, fatal, panic<br/>
            <i>Default</i>: info<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logTypes</b></td>
        <td>enum</td>
        <td>
          `logTypes` defines the desired record types to generate. Possible values are:<br> - `Flows` (default) to export regular network flows<br> - `Conversations` to generate events for started conversations, ended conversations as well as periodic "tick" updates<br> - `EndedConversations` to generate only ended conversations events<br> - `All` to generate both network flows and all conversations events<br><br/>
          <br/>
            <i>Enum</i>: Flows, Conversations, EndedConversations, All<br/>
            <i>Default</i>: Flows<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetrics-1">metrics</a></b></td>
        <td>object</td>
        <td>
          `Metrics` define the processor configuration regarding metrics<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>multiClusterDeployment</b></td>
        <td>boolean</td>
        <td>
          Set `multiClusterDeployment` to `true` to enable multi clusters feature. This adds `clusterName` label to flows data<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorresources-1">resources</a></b></td>
        <td>object</td>
        <td>
          `resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
          <br/>
            <i>Default</i>: map[limits:map[memory:800Mi] requests:map[cpu:100m memory:100Mi]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.advanced
<sup><sup>[↩ Parent](#flowcollectorspecprocessor-1)</sup></sup>



`advanced` allows setting some aspects of the internal configuration of the flow processor. This section is aimed mostly for debugging and fine-grained performance optimizations, such as `GOGC` and `GOMAXPROCS` env vars. Set these values at your own risk.

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
        <td><b>conversationEndTimeout</b></td>
        <td>string</td>
        <td>
          `conversationEndTimeout` is the time to wait after a network flow is received, to consider the conversation ended. This delay is ignored when a FIN packet is collected for TCP flows (see `conversationTerminatingTimeout` instead).<br/>
          <br/>
            <i>Default</i>: 10s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conversationHeartbeatInterval</b></td>
        <td>string</td>
        <td>
          `conversationHeartbeatInterval` is the time to wait between "tick" events of a conversation<br/>
          <br/>
            <i>Default</i>: 30s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>conversationTerminatingTimeout</b></td>
        <td>string</td>
        <td>
          `conversationTerminatingTimeout` is the time to wait from detected FIN flag to end a conversation. Only relevant for TCP flows.<br/>
          <br/>
            <i>Default</i>: 5s<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dropUnusedFields</b></td>
        <td>boolean</td>
        <td>
          `dropUnusedFields` [deprecated (*)] this setting is not used anymore.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableKubeProbes</b></td>
        <td>boolean</td>
        <td>
          `enableKubeProbes` is a flag to enable or disable Kubernetes liveness and readiness probes<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>env</b></td>
        <td>map[string]string</td>
        <td>
          `env` allows passing custom environment variables to underlying components. Useful for passing some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be publicly exposed as part of the FlowCollector descriptor, as they are only useful in edge debug or support scenarios.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>healthPort</b></td>
        <td>integer</td>
        <td>
          `healthPort` is a collector HTTP port in the Pod that exposes the health check API<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 8080<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          [Deprecated (*)] Port of the flow collector (host port). It is not used anymore and will be removed in a future version.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 2055<br/>
            <i>Minimum</i>: 1025<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>profilePort</b></td>
        <td>integer</td>
        <td>
          `profilePort` allows setting up a Go pprof profiler listening to this port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 6060<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler
<sup><sup>[↩ Parent](#flowcollectorspecprocessor-1)</sup></sup>



`kafkaConsumerAutoscaler` is the spec of a horizontal pod autoscaler to set up for `flowlogs-pipeline-transformer`, which consumes Kafka messages. This setting is ignored when Kafka is disabled.

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
          `maxReplicas` is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1">metrics</a></b></td>
        <td>[]object</td>
        <td>
          Metrics used by the pod autoscaler<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minReplicas</b></td>
        <td>integer</td>
        <td>
          `minReplicas` is the lower limit for the number of replicas to which the autoscaler can scale down. It defaults to 1 pod. minReplicas is allowed to be 0 if the alpha feature gate HPAScaleToZero is enabled and at least one Object or External metric is configured. Scaling is active as long as at least one metric value is available.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          `status` describes the desired status regarding deploying an horizontal pod autoscaler.<br> - `Disabled` does not deploy an horizontal pod autoscaler.<br> - `Enabled` deploys an horizontal pod autoscaler.<br><br/>
          <br/>
            <i>Enum</i>: Disabled, Enabled<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscaler-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresource-1">containerResource</a></b></td>
        <td>object</td>
        <td>
          containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal-1">external</a></b></td>
        <td>object</td>
        <td>
          external refers to a global metric that is not associated with any Kubernetes object. It allows autoscaling based on information coming from components running outside of cluster (for example length of queue in cloud messaging service, or QPS from loadbalancer running outside of cluster).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject-1">object</a></b></td>
        <td>object</td>
        <td>
          object refers to a metric describing a single kubernetes object (for example, hits-per-second on an Ingress object).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods-1">pods</a></b></td>
        <td>object</td>
        <td>
          pods refers to a metric describing each pod in the current scale target (for example, transactions-processed-per-second).  The values will be averaged together before being compared to the target value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresource-1">resource</a></b></td>
        <td>object</td>
        <td>
          resource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing each pod in the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].containerResource
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1)</sup></sup>



containerResource refers to a resource metric (such as those specified in requests and limits) known to Kubernetes describing a single container in each pod of the current scale target (e.g. CPU or memory). Such metrics are built in to Kubernetes, and have special scaling options on top of those available to normal per-pod metrics using the "pods" source. This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.

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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresourcetarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].containerResource.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexcontainerresource-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternaltarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternalmetricselector-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].external.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexexternal-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectdescribedobject-1">describedObject</a></b></td>
        <td>object</td>
        <td>
          describedObject specifies the descriptions of a object,such as kind,name apiVersion<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjecttarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.describedObject
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject-1)</sup></sup>



describedObject specifies the descriptions of a object,such as kind,name apiVersion

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
          kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          apiVersion is the API version of the referent<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobjectmetricselector-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].object.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexobject-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetric-1">metric</a></b></td>
        <td>object</td>
        <td>
          metric identifies the target metric by name and selector<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodstarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselector-1">selector</a></b></td>
        <td>object</td>
        <td>
          selector is the string-encoded form of a standard kubernetes label selector for the given metric When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping. When unset, just the metricName will be used to gather metrics.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric.selector
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetric-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselectormatchexpressionsindex-1">matchExpressions</a></b></td>
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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.metric.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpodsmetricselector-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].pods.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexpods-1)</sup></sup>



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


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].resource
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindex-1)</sup></sup>



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
        <td><b><a href="#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresourcetarget-1">target</a></b></td>
        <td>object</td>
        <td>
          target specifies the target value for the given metric<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.kafkaConsumerAutoscaler.metrics[index].resource.target
<sup><sup>[↩ Parent](#flowcollectorspecprocessorkafkaconsumerautoscalermetricsindexresource-1)</sup></sup>



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


### FlowCollector.spec.processor.metrics
<sup><sup>[↩ Parent](#flowcollectorspecprocessor-1)</sup></sup>



`Metrics` define the processor configuration regarding metrics

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
        <td><b>disableAlerts</b></td>
        <td>[]enum</td>
        <td>
          `disableAlerts` is a list of alerts that should be disabled. Possible values are:<br> `NetObservNoFlows`, which is triggered when no flows are being observed for a certain period.<br> `NetObservLokiError`, which is triggered when flows are being dropped due to Loki errors.<br><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>includeList</b></td>
        <td>[]enum</td>
        <td>
          `includeList` is a list of metric names to specify which ones to generate. The names correspond to the names in Prometheus without the prefix. For example, `namespace_egress_packets_total` shows up as `netobserv_namespace_egress_packets_total` in Prometheus. Note that the more metrics you add, the bigger is the impact on Prometheus workload resources. Metrics enabled by default are: `namespace_flows_total`, `node_ingress_bytes_total`, `workload_ingress_bytes_total`, `namespace_drop_packets_total` (when `PacketDrop` feature is enabled), `namespace_rtt_seconds` (when `FlowRTT` feature is enabled), `namespace_dns_latency_seconds` (when `DNSTracking` feature is enabled). More information, with full list of available metrics: https://github.com/netobserv/network-observability-operator/blob/main/docs/Metrics.md<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsserver-1">server</a></b></td>
        <td>object</td>
        <td>
          Metrics server endpoint configuration for Prometheus scraper<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetrics-1)</sup></sup>



Metrics server endpoint configuration for Prometheus scraper

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
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The prometheus HTTP port<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 9102<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertls-1">tls</a></b></td>
        <td>object</td>
        <td>
          TLS configuration.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsserver-1)</sup></sup>



TLS configuration.

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
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          `insecureSkipVerify` allows skipping client-side verification of the provided certificate. If set to `true`, the `providedCaFile` field is ignored.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertlsprovided-1">provided</a></b></td>
        <td>object</td>
        <td>
          TLS configuration when `type` is set to `Provided`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flowcollectorspecprocessormetricsservertlsprovidedcafile-1">providedCaFile</a></b></td>
        <td>object</td>
        <td>
          Reference to the CA file when `type` is set to `Provided`.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Select the type of TLS configuration:<br> - `Disabled` (default) to not configure TLS for the endpoint. - `Provided` to manually provide cert file and a key file. - `Auto` to use OpenShift auto generated certificate using annotations.<br/>
          <br/>
            <i>Enum</i>: Disabled, Provided, Auto<br/>
            <i>Default</i>: Disabled<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls.provided
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsservertls-1)</sup></sup>



TLS configuration when `type` is set to `Provided`.

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
          `certFile` defines the path to the certificate file name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certKey</b></td>
        <td>string</td>
        <td>
          `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing certificates<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the certificate reference: `configmap` or `secret`<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.metrics.server.tls.providedCaFile
<sup><sup>[↩ Parent](#flowcollectorspecprocessormetricsservertls-1)</sup></sup>



Reference to the CA file when `type` is set to `Provided`.

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
        <td><b>file</b></td>
        <td>string</td>
        <td>
          File name within the config map or secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the config map or secret containing the file<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed. If the namespace is different, the config map or the secret is copied so that it can be mounted as required.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type for the file reference: "configmap" or "secret"<br/>
          <br/>
            <i>Enum</i>: configmap, secret<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.resources
<sup><sup>[↩ Parent](#flowcollectorspecprocessor-1)</sup></sup>



`resources` are the compute resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

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
        <td><b><a href="#flowcollectorspecprocessorresourcesclaimsindex-1">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
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
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.spec.processor.resources.claims[index]
<sup><sup>[↩ Parent](#flowcollectorspecprocessorresources-1)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

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
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### FlowCollector.status
<sup><sup>[↩ Parent](#flowcollector-1)</sup></sup>



`FlowCollectorStatus` defines the observed state of FlowCollector

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
        <td><b><a href="#flowcollectorstatusconditionsindex-1">conditions</a></b></td>
        <td>[]object</td>
        <td>
          `conditions` represent the latest available observations of an object's state<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where console plugin and flowlogs-pipeline have been deployed. Deprecated: annotations are used instead<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlowCollector.status.conditions[index]
<sup><sup>[↩ Parent](#flowcollectorstatus-1)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource. --- This struct is intended for direct use as an array at the field path .status.conditions.  For example, 
 	type FooStatus struct{ 	    // Represents the observations of a foo's current state. 	    // Known .status.conditions.type are: "Available", "Progressing", and "Degraded" 	    // +patchMergeKey=type 	    // +patchStrategy=merge 	    // +listType=map 	    // +listMapKey=type 	    Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"` 
 	    // other fields 	}

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