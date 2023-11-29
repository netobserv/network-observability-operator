/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1beta2

import (
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	AgentIPFIX            = "IPFIX"
	AgentEBPF             = "EBPF"
	DeploymentModelDirect = "DIRECT"
	DeploymentModelKafka  = "KAFKA"
)

// Please notice that the FlowCollectorSpec's properties MUST redefine one of the default
// values to force the definition of the section when it is not provided by the manifest.
// This will cause that the remaining default fields will be set according to their definition.
// Otherwise, omitting the sections in the manifest would lead to zero-valued properties.
// This is a workaround for the related issue:
// https://github.com/kubernetes-sigs/controller-tools/issues/622

// Defines the desired state of the FlowCollector resource.
// <br><br>
// *: the mention of "unsupported", or "deprecated" for a feature throughout this document means that this feature
// is not officially supported by Red Hat. It might have been, for instance, contributed by the community
// and accepted without a formal agreement for maintenance. The product maintainers might provide some support
// for these features as a best effort only.
type FlowCollectorSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// Namespace where NetObserv pods are deployed.
	// +kubebuilder:default:=netobserv
	Namespace string `json:"namespace,omitempty"`

	// Agent configuration for flows extraction.
	Agent FlowCollectorAgent `json:"agent,omitempty"`

	// `processor` defines the settings of the component that receives the flows from the agent,
	// enriches them, generates metrics, and forwards them to the Loki persistence layer and/or any available exporter.
	Processor FlowCollectorFLP `json:"processor,omitempty"`

	// `loki`, the flow store, client settings.
	Loki FlowCollectorLoki `json:"loki,omitempty"`

	// `consolePlugin` defines the settings related to the OpenShift Console plugin, when available.
	ConsolePlugin FlowCollectorConsolePlugin `json:"consolePlugin,omitempty"`

	// `deploymentModel` defines the desired type of deployment for flow processing. Possible values are:<br>
	// - `DIRECT` (default) to make the flow processor listening directly from the agents.<br>
	// - `KAFKA` to make flows sent to a Kafka pipeline before consumption by the processor.<br>
	// Kafka can provide better scalability, resiliency, and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="DIRECT";"KAFKA"
	// +kubebuilder:default:=DIRECT
	DeploymentModel string `json:"deploymentModel,omitempty"`

	// Kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the `spec.deploymentModel` is `KAFKA`.
	// +optional
	Kafka FlowCollectorKafka `json:"kafka,omitempty"`

	// `exporters` define additional optional exporters for custom consumption or storage.
	// +optional
	Exporters []*FlowCollectorExporter `json:"exporters"`
}

// `FlowCollectorAgent` is a discriminated union that allows to select either ipfix or ebpf, but does not
// allow defining both fields.
// +union
type FlowCollectorAgent struct {
	// `type` selects the flows tracing agent. Possible values are:<br>
	// - `EBPF` (default) to use NetObserv eBPF agent.<br>
	// - `IPFIX` [deprecated (*)] - to use the legacy IPFIX collector.<br>
	// `EBPF` is recommended as it offers better performances and should work regardless of the CNI installed on the cluster.
	// `IPFIX` works with OVN-Kubernetes CNI (other CNIs could work if they support exporting IPFIX,
	// but they would require manual configuration).
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="EBPF";"IPFIX"
	// +kubebuilder:default:=EBPF
	Type string `json:"type,omitempty"`

	// `ipfix` [deprecated (*)] - describes the settings related to the IPFIX-based flow reporter when `spec.agent.type`
	// is set to `IPFIX`.
	// +optional
	IPFIX FlowCollectorIPFIX `json:"ipfix,omitempty"`

	// `ebpf` describes the settings related to the eBPF-based flow reporter when `spec.agent.type`
	// is set to `EBPF`.
	// +optional
	EBPF FlowCollectorEBPF `json:"ebpf,omitempty"`
}

// `FlowCollectorIPFIX` defines a FlowCollector that uses IPFIX on OVN-Kubernetes to collect the
// flows information
type FlowCollectorIPFIX struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="20s"
	// `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending.
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty" mapstructure:"cacheActiveTimeout,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=400
	// `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows.
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty" mapstructure:"cacheMaxFlows,omitempty"`

	//+kubebuilder:validation:Minimum=2
	//+kubebuilder:default:=400
	// `sampling` is the sampling rate on the reporter. 100 means one flow on 100 is sent.
	// To ensure cluster stability, it is not possible to set a value below 2.
	// If you really want to sample every packet, which might impact the cluster stability,
	// refer to `forceSampleAll`. Alternatively, you can use the eBPF Agent instead of IPFIX.
	Sampling int32 `json:"sampling,omitempty" mapstructure:"sampling,omitempty"`

	//+kubebuilder:default:=false
	// `forceSampleAll` allows disabling sampling in the IPFIX-based flow reporter.
	// It is not recommended to sample all the traffic with IPFIX, as it might generate cluster instability.
	// If you REALLY want to do that, set this flag to `true`. Use at your own risk.
	// When it is set to `true`, the value of `sampling` is ignored.
	ForceSampleAll bool `json:"forceSampleAll,omitempty" mapstructure:"-"`

	// `clusterNetworkOperator` defines the settings related to the OpenShift Cluster Network Operator, when available.
	ClusterNetworkOperator ClusterNetworkOperatorConfig `json:"clusterNetworkOperator,omitempty" mapstructure:"-"`

	// `ovnKubernetes` defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.
	OVNKubernetes OVNKubernetesConfig `json:"ovnKubernetes,omitempty" mapstructure:"-"`
}

// Agent feature, can be one of:<br>
// - `PacketDrop`, to track packet drops.<br>
// - `DNSTracking`, to track specific information on DNS traffic.<br>
// - `FlowRTT`, to track TCP latency. [Unsupported (*)].<br>
// +kubebuilder:validation:Enum:="PacketDrop";"DNSTracking";"FlowRTT"
type AgentFeature string

const (
	PacketDrop  AgentFeature = "PacketDrop"
	DNSTracking AgentFeature = "DNSTracking"
	FlowRTT     AgentFeature = "FlowRTT"
)

// `FlowCollectorEBPF` defines a FlowCollector that uses eBPF to collect the flows information
type FlowCollectorEBPF struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// `imagePullPolicy` is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"800Mi"}}
	// `resources` are the compute resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Sampling rate of the flow reporter. 100 means one flow on 100 is sent. 0 or 1 means all flows are sampled.
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=50
	//+optional
	Sampling *int32 `json:"sampling,omitempty"`

	// `cacheActiveTimeout` is the max period during which the reporter aggregates flows before sending.
	// Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load,
	// however you can expect higher memory consumption and an increased latency in the flow collection.
	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="5s"
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty"`

	// `cacheMaxFlows` is the max number of flows in an aggregate; when reached, the reporter sends the flows.
	// Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load,
	// however you can expect higher memory consumption and an increased latency in the flow collection.
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=100000
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty"`

	// `interfaces` contains the interface names from where flows are collected. If empty, the agent
	// fetches all the interfaces in the system, excepting the ones listed in ExcludeInterfaces.
	// An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression.
	// Otherwise it is matched as a case-sensitive string.
	//+optional
	Interfaces []string `json:"interfaces"`

	// `excludeInterfaces` contains the interface names that are excluded from flow tracing.
	// An entry enclosed by slashes, such as `/br-/`, is matched as a regular expression.
	// Otherwise it is matched as a case-sensitive string.
	//+kubebuilder:default=lo;
	//+optional
	ExcludeInterfaces []string `json:"excludeInterfaces"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// `logLevel` defines the log level for the NetObserv eBPF Agent
	LogLevel string `json:"logLevel,omitempty"`

	// Privileged mode for the eBPF Agent container. In general this setting can be ignored or set to `false`:
	// in that case, the operator sets granular capabilities (BPF, PERFMON, NET_ADMIN, SYS_RESOURCE)
	// to the container, to enable its correct operation.
	// If for some reason these capabilities cannot be set, such as if an old kernel version not knowing CAP_BPF
	// is in use, then you can turn on this mode for more global privileges.
	// +optional
	Privileged bool `json:"privileged,omitempty"`

	//+kubebuilder:default:=10485760
	// +optional
	// `kafkaBatchSize` limits the maximum size of a request in bytes before being sent to a partition. Ignored when not using Kafka. Default: 10MB.
	KafkaBatchSize int `json:"kafkaBatchSize"`

	// `debug` allows setting some aspects of the internal configuration of the eBPF agent.
	// This section is aimed exclusively for debugging and fine-grained performance optimizations,
	// such as `GOGC` and `GOMAXPROCS` env vars. Users setting its values do it at their own risk.
	// +optional
	Debug DebugConfig `json:"debug,omitempty"`

	// List of additional features to enable. They are all disabled by default. Enabling additional features might have performance impacts. Possible values are:<br>
	// - `PacketDrop`: enable the packets drop flows logging feature. This feature requires mounting
	// the kernel debug filesystem, so the eBPF pod has to run as privileged.
	// If the `spec.agent.eBPF.privileged` parameter is not set, an error is reported.<br>
	// - `DNSTracking`: enable the DNS tracking feature.<br>
	// - `FlowRTT` [unsupported (*)]: enable flow latency (RTT) calculations in the eBPF agent during TCP handshakes. This feature better works with `sampling` set to 1.<br>
	// +optional
	Features []AgentFeature `json:"features,omitempty"`
}

// `FlowCollectorKafka` defines the desired Kafka config of FlowCollector
type FlowCollectorKafka struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=""
	// Address of the Kafka server
	Address string `json:"address"`

	//+kubebuilder:default:=""
	// Kafka topic to use. It must exist. NetObserv does not create it.
	Topic string `json:"topic"`

	// TLS client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.
	// +optional
	TLS ClientTLS `json:"tls"`

	// SASL authentication configuration. [Unsupported (*)].
	// +optional
	SASL SASLConfig `json:"sasl"`
}

type FlowCollectorIPFIXReceiver struct {
	//+kubebuilder:default:=""
	// Address of the IPFIX external receiver
	TargetHost string `json:"targetHost"`

	// Port for the IPFIX external receiver
	TargetPort int `json:"targetPort"`

	// Transport protocol (`TCP` or `UDP`) to be used for the IPFIX connection, defaults to `TCP`.
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="TCP";"UDP"
	// +optional
	Transport string `json:"transport,omitempty"`
}

const (
	ServerTLSDisabled = "DISABLED"
	ServerTLSProvided = "PROVIDED"
	ServerTLSAuto     = "AUTO"
)

type ServerTLSConfigType string

// `ServerTLS` define the TLS configuration, server side
type ServerTLS struct {
	// Select the type of TLS configuration:<br>
	// - `DISABLED` (default) to not configure TLS for the endpoint.
	// - `PROVIDED` to manually provide cert file and a key file.
	// - `AUTO` to use OpenShift auto generated certificate using annotations.
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="DISABLED";"PROVIDED";"AUTO"
	// +kubebuilder:validation:Required
	//+kubebuilder:default:="DISABLED"
	Type ServerTLSConfigType `json:"type,omitempty"`

	// TLS configuration when `type` is set to `PROVIDED`.
	// +optional
	Provided *CertificateReference `json:"provided"`

	//+kubebuilder:default:=false
	// `insecureSkipVerify` allows skipping client-side verification of the provided certificate.
	// If set to `true`, the `providedCaFile` field is ignored.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// Reference to the CA file when `type` is set to `PROVIDED`.
	// +optional
	ProvidedCaFile *FileReference `json:"providedCaFile,omitempty"`
}

// `MetricsServerConfig` define the metrics server endpoint configuration for Prometheus scraper
type MetricsServerConfig struct {

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9102
	// The prometheus HTTP port
	Port int32 `json:"port,omitempty"`

	// TLS configuration.
	// +optional
	TLS ServerTLS `json:"tls"`
}

const (
	AlertNoFlows   = "NetObservNoFlows"
	AlertLokiError = "NetObservLokiError"
)

// Name of a processor alert.
// Possible values are:<br>
// - `NetObservNoFlows`, which is triggered when no flows are being observed for a certain period.<br>
// - `NetObservLokiError`, which is triggered when flows are being dropped due to Loki errors.<br>
// +kubebuilder:validation:Enum:="NetObservNoFlows";"NetObservLokiError"
type FLPAlert string

// Metric name. More information in https://github.com/netobserv/network-observability-operator/blob/main/docs/Metrics.md.
// +kubebuilder:validation:Enum:="namespace_egress_bytes_total";"namespace_egress_packets_total";"namespace_ingress_bytes_total";"namespace_ingress_packets_total";"namespace_flows_total";"node_egress_bytes_total";"node_egress_packets_total";"node_ingress_bytes_total";"node_ingress_packets_total";"node_flows_total";"workload_egress_bytes_total";"workload_egress_packets_total";"workload_ingress_bytes_total";"workload_ingress_packets_total";"workload_flows_total";"namespace_drop_bytes_total";"namespace_drop_packets_total";"node_drop_bytes_total";"node_drop_packets_total";"workload_drop_bytes_total";"workload_drop_packets_total";"namespace_rtt_seconds";"node_rtt_seconds";"workload_rtt_seconds";"namespace_dns_latency_seconds";"node_dns_latency_seconds";"workload_dns_latency_seconds"
type FLPMetric string

// `FLPMetrics` define the desired FLP configuration regarding metrics
type FLPMetrics struct {
	// Metrics server endpoint configuration for Prometheus scraper
	// +optional
	Server MetricsServerConfig `json:"server,omitempty"`

	// `includeList` is a list of metric names to specify which ones to generate.
	// The names correspond to the names in Prometheus without the prefix. For example,
	// `namespace_egress_packets_total` will show up as `netobserv_namespace_egress_packets_total` in Prometheus.
	// Note that the more metrics you add, the bigger is the impact on Prometheus workload resources.
	// Metrics enabled by default are:
	// `namespace_flows_total`, `node_ingress_bytes_total`, `workload_ingress_bytes_total`, `namespace_drop_packets_total` (when `PacketDrop` feature is enabled),
	// `namespace_rtt_seconds` (when `FlowRTT` feature is enabled), `namespace_dns_latency_seconds` (when `DNSTracking` feature is enabled).
	// More information, with full list of available metrics: https://github.com/netobserv/network-observability-operator/blob/main/docs/Metrics.md
	// +optional
	IncludeList *[]FLPMetric `json:"includeList,omitempty"`

	// `disableAlerts` is a list of alerts that should be disabled.
	// Possible values are:<br>
	// `NetObservNoFlows`, which is triggered when no flows are being observed for a certain period.<br>
	// `NetObservLokiError`, which is triggered when flows are being dropped due to Loki errors.<br>
	// +optional
	DisableAlerts []FLPAlert `json:"disableAlerts"`
}

const (
	LogTypeFlows              = "FLOWS"
	LogTypeConversations      = "CONVERSATIONS"
	LogTypeEndedConversations = "ENDED_CONVERSATIONS"
	LogTypeAll                = "ALL"
)

// `FlowCollectorFLP` defines the desired flowlogs-pipeline state of FlowCollector
type FlowCollectorFLP struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Minimum=1025
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=2055
	// Port of the flow collector (host port).
	// By convention, some values are forbidden. It must be greater than 1024 and different from
	// 4500, 4789 and 6081.
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=8080
	// `healthPort` is a collector HTTP port in the Pod that exposes the health check API
	HealthPort int32 `json:"healthPort,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=65535
	//+optional
	// `profilePort` allows setting up a Go pprof profiler listening to this port
	ProfilePort int32 `json:"profilePort,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// `imagePullPolicy` is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// `Metrics` define the processor configuration regarding metrics
	Metrics FLPMetrics `json:"metrics,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// `logLevel` of the processor runtime
	LogLevel string `json:"logLevel,omitempty"`

	//+kubebuilder:default:={requests:{memory:"100Mi",cpu:"100m"},limits:{memory:"800Mi"}}
	// `resources` are the compute resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:default:=true
	// `enableKubeProbes` is a flag to enable or disable Kubernetes liveness and readiness probes
	EnableKubeProbes *bool `json:"enableKubeProbes,omitempty"`

	//+kubebuilder:default:=true
	// `dropUnusedFields` allows, when set to `true`, to drop fields that are known to be unused by OVS, to save storage space.
	DropUnusedFields *bool `json:"dropUnusedFields,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=3
	// `kafkaConsumerReplicas` defines the number of replicas (pods) to start for `flowlogs-pipeline-transformer`, which consumes Kafka messages.
	// This setting is ignored when Kafka is disabled.
	KafkaConsumerReplicas *int32 `json:"kafkaConsumerReplicas,omitempty"`

	// `kafkaConsumerAutoscaler` is the spec of a horizontal pod autoscaler to set up for `flowlogs-pipeline-transformer`, which consumes Kafka messages.
	// This setting is ignored when Kafka is disabled.
	// +optional
	KafkaConsumerAutoscaler FlowCollectorHPA `json:"kafkaConsumerAutoscaler,omitempty"`

	//+kubebuilder:default:=1000
	// +optional
	// `kafkaConsumerQueueCapacity` defines the capacity of the internal message queue used in the Kafka consumer client. Ignored when not using Kafka.
	KafkaConsumerQueueCapacity int `json:"kafkaConsumerQueueCapacity"`

	//+kubebuilder:default:=10485760
	// +optional
	// `kafkaConsumerBatchSize` indicates to the broker the maximum batch size, in bytes, that the consumer accepts. Ignored when not using Kafka. Default: 10MB.
	KafkaConsumerBatchSize int `json:"kafkaConsumerBatchSize"`

	// `logTypes` defines the desired record types to generate. Possible values are:<br>
	// - `FLOWS` (default) to export regular network flows<br>
	// - `CONVERSATIONS` to generate events for started conversations, ended conversations as well as periodic "tick" updates<br>
	// - `ENDED_CONVERSATIONS` to generate only ended conversations events<br>
	// - `ALL` to generate both network flows and all conversations events<br>
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:="FLOWS";"CONVERSATIONS";"ENDED_CONVERSATIONS";"ALL"
	// +kubebuilder:default:=FLOWS
	LogTypes *string `json:"logTypes,omitempty"`

	//+kubebuilder:default:="30s"
	// +optional
	// `conversationHeartbeatInterval` is the time to wait between "tick" events of a conversation
	ConversationHeartbeatInterval *metav1.Duration `json:"conversationHeartbeatInterval,omitempty"`

	//+kubebuilder:default:="10s"
	// +optional
	// `conversationEndTimeout` is the time to wait after a network flow is received, to consider the conversation ended.
	// This delay is ignored when a FIN packet is collected for TCP flows (see `conversationTerminatingTimeout` instead).
	ConversationEndTimeout *metav1.Duration `json:"conversationEndTimeout,omitempty"`

	//+kubebuilder:default:="5s"
	// +optional
	// `conversationTerminatingTimeout` is the time to wait from detected FIN flag to end a conversation. Only relevant for TCP flows.
	ConversationTerminatingTimeout *metav1.Duration `json:"conversationTerminatingTimeout,omitempty"`

	//+kubebuilder:default:=""
	// +optional
	// `clusterName` is the name of the cluster to appear in the flows data. This is useful in a multi-cluster context. When using OpenShift, leave empty to make it automatically determined.
	ClusterName string `json:"clusterName,omitempty"`

	//+kubebuilder:default:=false
	// Set `multiClusterDeployment` to `true` to enable multi clusters feature. This will add clusterName label to flows data
	MultiClusterDeployment *bool `json:"multiClusterDeployment,omitempty"`

	// `debug` allows setting some aspects of the internal configuration of the flow processor.
	// This section is aimed exclusively for debugging and fine-grained performance optimizations,
	// such as `GOGC` and `GOMAXPROCS` env vars. Users setting its values do it at their own risk.
	// +optional
	Debug DebugConfig `json:"debug,omitempty"`
}

const (
	HPAStatusDisabled = "DISABLED"
	HPAStatusEnabled  = "ENABLED"
)

type FlowCollectorHPA struct {
	// +kubebuilder:validation:Enum:=DISABLED;ENABLED
	// +kubebuilder:default:=DISABLED
	// `status` describes the desired status regarding deploying an horizontal pod autoscaler.<br>
	// - `DISABLED` does not deploy an horizontal pod autoscaler.<br>
	// - `ENABLED` deploys an horizontal pod autoscaler.<br>
	Status string `json:"status,omitempty"`

	// `minReplicas` is the lower limit for the number of replicas to which the autoscaler
	// can scale down. It defaults to 1 pod. minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured. Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`

	// `maxReplicas` is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	// +kubebuilder:default:=3
	// +optional
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`

	// Metrics used by the pod autoscaler
	// +optional
	Metrics []ascv2.MetricSpec `json:"metrics"`
}

const (
	LokiAuthDisabled         = "DISABLED"
	LokiAuthUseHostToken     = "HOST"
	LokiAuthForwardUserToken = "FORWARD"
)

// `LokiManualParams` defines the full connection parameters to Loki.
type LokiManualParams struct {
	//+kubebuilder:default:="http://loki:3100/"
	// `ingesterUrl` is the address of an existing Loki ingester service to push the flows to. When using the Loki Operator,
	// set it to the Loki gateway service with the `network` tenant set in path, for example
	// https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.
	IngesterURL string `json:"ingesterUrl,omitempty"`

	//+kubebuilder:default:="http://loki:3100/"
	// `querierUrl` specifies the address of the Loki querier service.
	// When using the Loki Operator, set it to the Loki gateway service with the `network` tenant set in path, for example
	// https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.
	QuerierURL string `json:"querierUrl,omitempty"`

	//+kubebuilder:validation:optional
	// `statusUrl` specifies the address of the Loki `/ready`, `/metrics` and `/config` endpoints, in case it is different from the
	// Loki querier URL. If empty, the `querierUrl` value is used.
	// This is useful to show error messages and some context in the frontend.
	// When using the Loki Operator, set it to the Loki HTTP query frontend service, for example
	// https://loki-query-frontend-http.netobserv.svc:3100/.
	// `statusTLS` configuration is used when `statusUrl` is set.
	StatusURL string `json:"statusUrl,omitempty"`

	//+kubebuilder:default:="netobserv"
	// `tenantID` is the Loki `X-Scope-OrgID` that identifies the tenant for each request.
	// When using the Loki Operator, set it to `network`, which corresponds to a special tenant mode.
	TenantID string `json:"tenantID,omitempty"`

	//+kubebuilder:validation:Enum:="DISABLED";"HOST";"FORWARD"
	//+kubebuilder:default:="DISABLED"
	// `authToken` describes the way to get a token to authenticate to Loki.<br>
	// - `DISABLED` does not send any token with the request.<br>
	// - `FORWARD` forwards the user token for authorization.<br>
	// - `HOST` [deprecated (*)] - uses the local pod service account to authenticate to Loki.<br>
	// When using the Loki Operator, this must be set to `FORWARD`.
	AuthToken string `json:"authToken,omitempty"`

	// TLS client configuration for Loki URL.
	// +optional
	TLS ClientTLS `json:"tls"`

	// TLS client configuration for Loki status URL.
	// +optional
	StatusTLS ClientTLS `json:"statusTls"`
}

// LokiMicroservicesParams is the configuration for microservices Loki (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#microservices-mode)
type LokiMicroservicesParams struct {
	//+kubebuilder:default:="http://loki-distributor:3100/"
	// `ingesterUrl` is the address of an existing Loki ingester service to push the flows to.
	IngesterURL string `json:"ingesterUrl,omitempty"`

	//+kubebuilder:default:="http://loki-query-frontend:3100/"
	// `querierURL` specifies the address of the Loki querier service.
	QuerierURL string `json:"querierUrl,omitempty"`

	//+kubebuilder:default:="netobserv"
	// `tenantID` is the Loki `X-Scope-OrgID` header that identifies the tenant for each request.
	TenantID string `json:"tenantID,omitempty"`

	// TLS client configuration for Loki URL.
	// +optional
	TLS ClientTLS `json:"tls"`
}

// LokiMonolithParams is the configuration for monolithic Loki (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#monolithic-mode)
type LokiMonolithParams struct {
	//+kubebuilder:default:="http://loki:3100/"
	// `url` is the unique address of an existing Loki service that points to both the ingester and the querier.
	URL string `json:"url,omitempty"`

	//+kubebuilder:default:="netobserv"
	// `tenantID` is the Loki `X-Scope-OrgID` header that identifies the tenant for each request.
	TenantID string `json:"tenantID,omitempty"`

	// TLS client configuration for Loki URL.
	// +optional
	TLS ClientTLS `json:"tls"`
}

// LokiStackRef defines the name and namespace of the LokiStack instance
type LokiStackRef struct {
	// Name of an existing LokiStack resource to use.
	//+kubebuilder:default:="loki"
	//+kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Namespace where this `LokiStack` resource is located. If omited, it is assumed to be the same as `spec.namespace`.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type LokiMode string

const (
	LokiModeManual        LokiMode = "Manual"
	LokiModeLokiStack     LokiMode = "LokiStack"
	LokiModeMonolithic    LokiMode = "Monolithic"
	LokiModeMicroservices LokiMode = "Microservices"
)

// `FlowCollectorLoki` defines the desired state for FlowCollector's Loki client.
type FlowCollectorLoki struct {
	// `mode` must be set according to the installation mode of Loki:<br>
	// - Use "LokiStack" when Loki is managed using the Loki Operator<br>
	// - Use "Monolithic" when Loki is installed as a monolithic workload<br>
	// - Use "Microservices" when Loki is installed as microservices, but without Loki Operator<br>
	// - Use "Manual" if none of the options above match your setup<br>
	//+unionDiscriminator
	//+kubebuilder:validation:Enum=Manual;LokiStack;Monolithic;Microservices
	//+kubebuilder:default:="Monolithic"
	//+kubebuilder:validation:Required
	Mode LokiMode `json:"mode,omitempty"`

	// Loki configuration for "Manual" mode. This is the most flexible configuration.
	// It is ignored for other modes.
	// +optional
	Manual LokiManualParams `json:"manual,omitempty"`

	// Loki configuration for "Microservices" mode.
	// Use this option when Loki is installed using the microservices deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#microservices-mode).
	// It is ignored for other modes.
	// +optional
	Microservices LokiMicroservicesParams `json:"microservices,omitempty"`

	// Loki configuration for "Monolithic" mode.
	// Use this option when Loki is installed using the monolithic deployment mode (https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/#monolithic-mode).
	// It is ignored for other modes.
	// +optional
	Monolithic LokiMonolithParams `json:"monolithic,omitempty"`

	// Loki configuration for "LokiStack" mode. This is useful for an easy loki-operator configuration.
	// It is ignored for other modes.
	// +optional
	LokiStack LokiStackRef `json:"lokiStack,omitempty"`

	//+kubebuilder:default:=true
	// Set `enable` to `true` to store flows in Loki. It is required for the OpenShift Console plugin installation.
	Enable *bool `json:"enable,omitempty"`

	//+kubebuilder:default:="1s"
	// `batchWait` is the maximum time to wait before sending a batch.
	BatchWait *metav1.Duration `json:"batchWait,omitempty"` // Warning: keep as pointer, else default is ignored

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=102400
	// `batchSize` is the maximum batch size (in bytes) of logs to accumulate before sending.
	BatchSize int64 `json:"batchSize,omitempty"`

	//+kubebuilder:default:="10s"
	// `timeout` is the maximum time connection / request limit.
	// A timeout of zero means no timeout.
	Timeout *metav1.Duration `json:"timeout,omitempty"` // Warning: keep as pointer, else default is ignored

	//+kubebuilder:default="1s"
	// `minBackoff` is the initial backoff time for client connection between retries.
	MinBackoff *metav1.Duration `json:"minBackoff,omitempty"` // Warning: keep as pointer, else default is ignored

	//+kubebuilder:default="5s"
	// `maxBackoff` is the maximum backoff time for client connection between retries.
	MaxBackoff *metav1.Duration `json:"maxBackoff,omitempty"` // Warning: keep as pointer, else default is ignored

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=2
	// `maxRetries` is the maximum number of retries for client connections.
	MaxRetries *int32 `json:"maxRetries,omitempty"`

	//+kubebuilder:default:={"app":"netobserv-flowcollector"}
	// +optional
	// `staticLabels` is a map of common labels to set on each flow.
	StaticLabels map[string]string `json:"staticLabels"`
}

// FlowCollectorConsolePlugin defines the desired ConsolePlugin state of FlowCollector
type FlowCollectorConsolePlugin struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=true
	// Enables the console plugin deployment.
	// `spec.Loki.enable` must also be `true`
	Enable *bool `json:"enable,omitempty"`

	//+kubebuilder:default:=true
	// `register` allows, when set to `true`, to automatically register the provided console plugin with the OpenShift Console operator.
	// When set to `false`, you can still register it manually by editing console.operator.openshift.io/cluster with the following command:
	// `oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "netobserv-plugin"}]'`
	Register *bool `json:"register,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	// `replicas` defines the number of replicas (pods) to start.
	Replicas *int32 `json:"replicas,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9001
	// `port` is the plugin service port. Do not use 9002, which is reserved for metrics.
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// `imagePullPolicy` is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"100Mi"}}
	// `resources`, in terms of compute resources, required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// `logLevel` for the console plugin backend
	LogLevel string `json:"logLevel,omitempty"`

	// `autoscaler` spec of a horizontal pod autoscaler to set up for the plugin Deployment.
	// +optional
	Autoscaler FlowCollectorHPA `json:"autoscaler,omitempty"`

	//+kubebuilder:default:={enable:true}
	// `portNaming` defines the configuration of the port-to-service name translation
	PortNaming ConsolePluginPortConfig `json:"portNaming,omitempty"`

	//+kubebuilder:default:={{name:"Applications",filter:{"src_namespace!":"openshift-,netobserv","dst_namespace!":"openshift-,netobserv"},default:true},{name:"Infrastructure",filter:{"src_namespace":"openshift-,netobserv","dst_namespace":"openshift-,netobserv"}},{name:"Pods network",filter:{"src_kind":"Pod","dst_kind":"Pod"},default:true},{name:"Services network",filter:{"dst_kind":"Service"}}}
	// +optional
	// `quickFilters` configures quick filter presets for the Console plugin
	QuickFilters []QuickFilter `json:"quickFilters"`
}

// Configuration of the port to service name translation feature of the console plugin
type ConsolePluginPortConfig struct {
	//+kubebuilder:default:=true
	// Enable the console plugin port-to-service name translation
	Enable *bool `json:"enable,omitempty"`

	// `portNames` defines additional port names to use in the console,
	// for example, `portNames: {"3100": "loki"}`.
	// +optional
	PortNames map[string]string `json:"portNames" yaml:"portNames"`
}

// `QuickFilter` defines preset configuration for Console's quick filters
type QuickFilter struct {
	// Name of the filter, that is displayed in the Console
	// +kubebuilder:MinLength:=1
	Name string `json:"name"`
	// `filter` is a set of keys and values to be set when this filter is selected. Each key can relate to a list of values using a coma-separated string,
	// for example, `filter: {"src_namespace": "namespace1,namespace2"}`.
	// +kubebuilder:MinProperties:=1
	Filter map[string]string `json:"filter"`
	// `default` defines whether this filter should be active by default or not
	// +optional
	Default bool `json:"default,omitempty"`
}

// `ClusterNetworkOperatorConfig` defines the desired configuration related to the Cluster Network Configuration
type ClusterNetworkOperatorConfig struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=openshift-network-operator
	// Namespace  where the config map is going to be deployed.
	Namespace string `json:"namespace,omitempty"`
}

// `OVNKubernetesConfig` defines the desired configuration related to the OVN-Kubernetes network provider, when Cluster Network Operator isn't installed.
type OVNKubernetesConfig struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=ovn-kubernetes
	// Namespace where OVN-Kubernetes pods are deployed.
	Namespace string `json:"namespace,omitempty"`

	//+kubebuilder:default:=ovnkube-node
	// `daemonSetName` defines the name of the DaemonSet controlling the OVN-Kubernetes pods.
	DaemonSetName string `json:"daemonSetName,omitempty"`

	//+kubebuilder:default:=ovnkube-node
	// `containerName` defines the name of the container to configure for IPFIX.
	ContainerName string `json:"containerName,omitempty"`
}

type MountableType string

const (
	RefTypeSecret    MountableType = "secret"
	RefTypeConfigMap MountableType = "configmap"
)

type FileReference struct {
	//+kubebuilder:validation:Enum=configmap;secret
	// Type for the file reference: "configmap" or "secret"
	Type MountableType `json:"type,omitempty"`

	// Name of the config map or secret containing the file
	Name string `json:"name,omitempty"`

	// Namespace of the config map or secret containing the file. If omitted, the default is to use the same namespace as where NetObserv is deployed.
	// If the namespace is different, the config map or the secret is copied so that it can be mounted as required.
	// +optional
	//+kubebuilder:default:=""
	Namespace string `json:"namespace,omitempty"`

	// File name within the config map or secret
	File string `json:"file,omitempty"`
}

type CertificateReference struct {
	//+kubebuilder:validation:Enum=configmap;secret
	// Type for the certificate reference: `configmap` or `secret`
	Type MountableType `json:"type,omitempty"`

	// Name of the config map or secret containing certificates
	Name string `json:"name,omitempty"`

	// Namespace of the config map or secret containing certificates. If omitted, the default is to use the same namespace as where NetObserv is deployed.
	// If the namespace is different, the config map or the secret is copied so that it can be mounted as required.
	// +optional
	//+kubebuilder:default:=""
	Namespace string `json:"namespace,omitempty"`

	// `certFile` defines the path to the certificate file name within the config map or secret
	CertFile string `json:"certFile,omitempty"`

	// `certKey` defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.
	// +optional
	CertKey string `json:"certKey,omitempty"`
}

// `ClientTLS` defines TLS client configuration
type ClientTLS struct {
	//+kubebuilder:default:=false
	// Enable TLS
	Enable bool `json:"enable,omitempty"`

	//+kubebuilder:default:=false
	// `insecureSkipVerify` allows skipping client-side verification of the server certificate.
	// If set to `true`, the `caCert` field is ignored.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// `caCert` defines the reference of the certificate for the Certificate Authority
	CACert CertificateReference `json:"caCert,omitempty"`

	// `userCert` defines the user certificate reference and is used for mTLS (you can ignore it when using one-way TLS)
	// +optional
	UserCert CertificateReference `json:"userCert,omitempty"`
}

type SASLType string

const (
	SASLDisabled    SASLType = "DISABLED"
	SASLPlain       SASLType = "PLAIN"
	SASLScramSHA512 SASLType = "SCRAM-SHA512"
)

// `SASLConfig` defines SASL configuration
type SASLConfig struct {
	//+kubebuilder:validation:Enum=DISABLED;PLAIN;SCRAM-SHA512
	//+kubebuilder:default:=DISABLED
	// Type of SASL authentication to use, or `DISABLED` if SASL is not used
	Type SASLType `json:"type,omitempty"`

	// Reference to the secret or config map containing the client ID
	ClientIDReference FileReference `json:"clientIDReference,omitempty"`

	// Reference to the secret or config map containing the client secret
	ClientSecretReference FileReference `json:"clientSecretReference,omitempty"`
}

// `DebugConfig` allows tweaking some aspects of the internal configuration of the agent and FLP.
// They are aimed exclusively for debugging. Users setting these values do it at their own risk.
type DebugConfig struct {
	// `env` allows passing custom environment variables to underlying components. Useful for passing
	// some very concrete performance-tuning options, such as `GOGC` and `GOMAXPROCS`, that should not be
	// publicly exposed as part of the FlowCollector descriptor, as they are only useful
	// in edge debug or support scenarios.
	//+optional
	Env map[string]string `json:"env,omitempty"`
}

// Add more exporter types below
type ExporterType string

const (
	KafkaExporter ExporterType = "KAFKA"
	IpfixExporter ExporterType = "IPFIX"
)

// `FlowCollectorExporter` defines an additional exporter to send enriched flows to.
type FlowCollectorExporter struct {
	// `type` selects the type of exporters. The available options are `KAFKA` and `IPFIX`.
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="KAFKA";"IPFIX"
	// +kubebuilder:validation:Required
	Type ExporterType `json:"type"`

	// Kafka configuration, such as the address and topic, to send enriched flows to.
	// +optional
	Kafka FlowCollectorKafka `json:"kafka,omitempty"`

	// IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.
	// +optional
	IPFIX FlowCollectorIPFIXReceiver `json:"ipfix,omitempty"`
}

// `FlowCollectorStatus` defines the observed state of FlowCollector
type FlowCollectorStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// `conditions` represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// Namespace where console plugin and flowlogs-pipeline have been deployed.
	Namespace string `json:"namespace,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Agent",type="string",JSONPath=`.spec.agent.type`
// +kubebuilder:printcolumn:name="Sampling (EBPF)",type="string",JSONPath=`.spec.agent.ebpf.sampling`
// +kubebuilder:printcolumn:name="Deployment Model",type="string",JSONPath=`.spec.deploymentModel`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// `FlowCollector` is the schema for the network flows collection API, which pilots and configures the underlying deployments.
type FlowCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowCollectorSpec   `json:"spec,omitempty"`
	Status FlowCollectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// `FlowCollectorList` contains a list of FlowCollector
type FlowCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowCollector{}, &FlowCollectorList{})
}
