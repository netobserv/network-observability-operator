/*
Copyright 2021.

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
package v1alpha1

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

// FlowCollectorSpec defines the desired state of FlowCollector
type FlowCollectorSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// namespace where NetObserv pods are deployed.
	// If empty, the namespace of the operator is going to be used.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// agent for flows extraction.
	// +kubebuilder:default:={type:"EBPF"}
	Agent FlowCollectorAgent `json:"agent"`

	// processor defines the settings of the component that receives the flows from the agent,
	// enriches them, and forwards them to the Loki persistence layer.
	Processor FlowCollectorFLP `json:"processor,omitempty"`

	// loki, the flow store, client settings.
	Loki FlowCollectorLoki `json:"loki,omitempty"`

	// consolePlugin defines the settings related to the OpenShift Console plugin, when available.
	ConsolePlugin FlowCollectorConsolePlugin `json:"consolePlugin,omitempty"`

	// deploymentModel defines the desired type of deployment for flow processing. Possible values are "DIRECT" (default) to make
	// the flow processor listening directly from the agents, or "KAFKA" to make flows sent to a Kafka pipeline before consumption
	// by the processor.
	// Kafka can provide better scalability, resiliency and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="DIRECT";"KAFKA"
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=DIRECT
	DeploymentModel string `json:"deploymentModel"`

	// kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline. Available when the "spec.deploymentModel" is "KAFKA".
	// +optional
	Kafka FlowCollectorKafka `json:"kafka,omitempty"`

	// exporters defines additional optional exporters for custom consumption or storage. This is an experimental feature. Currently, only KAFKA exporter is available.
	// +optional
	Exporters []*FlowCollectorExporter `json:"exporters"`
}

// FlowCollectorAgent is a discriminated union that allows to select either ipfix or ebpf, but does not
// allow defining both fields.
// +union
type FlowCollectorAgent struct {
	// type selects the flows tracing agent. Possible values are "EBPF" (default) to use NetObserv eBPF agent,
	// "IPFIX" to use the legacy IPFIX collector. "EBPF" is recommended in most cases as it offers better
	// performances and should work regardless of the CNI installed on the cluster.
	// "IPFIX" works with OVN-Kubernetes CNI (other CNIs could work if they support exporting IPFIX,
	// but they would require manual configuration).
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="EBPF";"IPFIX"
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=EBPF
	Type string `json:"type"`

	// ipfix describes the settings related to the IPFIX-based flow reporter when the "agent.type"
	// property is set to "IPFIX".
	// +optional
	IPFIX FlowCollectorIPFIX `json:"ipfix,omitempty"`

	// ebpf describes the settings related to the eBPF-based flow reporter when the "agent.type"
	// property is set to "EBPF".
	// +optional
	EBPF FlowCollectorEBPF `json:"ebpf,omitempty"`
}

// FlowCollectorIPFIX defines a FlowCollector that uses IPFIX on OVN-Kubernetes to collect the
// flows information
type FlowCollectorIPFIX struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="20s"
	// cacheActiveTimeout is the max period during which the reporter will aggregate flows before sending
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty" mapstructure:"cacheActiveTimeout,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=400
	// cacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty" mapstructure:"cacheMaxFlows,omitempty"`

	//+kubebuilder:validation:Minimum=2
	//+kubebuilder:default:=400
	// sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent.
	// To ensure cluster stability, it is not possible to set a value below 2.
	// If you really want to sample every packet, which might impact the cluster stability,
	// refer to "forceSampleAll". Alternatively, you can use the eBPF Agent instead of IPFIX.
	Sampling int32 `json:"sampling,omitempty" mapstructure:"sampling,omitempty"`

	//+kubebuilder:default:=false
	// forceSampleAll allows disabling sampling in the IPFIX-based flow reporter.
	// It is not recommended to sample all the traffic with IPFIX, as it might generate cluster instability.
	// If you REALLY want to do that, set this flag to true. Use at your own risk.
	// When it is set to true, the value of "sampling" is ignored.
	ForceSampleAll bool `json:"forceSampleAll,omitempty" mapstructure:"-"`

	// clusterNetworkOperator defines the settings related to the OpenShift Cluster Network Operator, when available.
	ClusterNetworkOperator ClusterNetworkOperatorConfig `json:"clusterNetworkOperator,omitempty" mapstructure:"-"`

	// ovnKubernetes defines the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.
	OVNKubernetes OVNKubernetesConfig `json:"ovnKubernetes,omitempty" mapstructure:"-"`
}

// FlowCollectorEBPF defines a FlowCollector that uses eBPF to collect the flows information
type FlowCollectorEBPF struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"800Mi"}}
	// resources are the compute resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// sampling rate of the flow reporter. 100 means one flow on 100 is sent. 0 or 1 means all flows are sampled.
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=50
	//+optional
	Sampling *int32 `json:"sampling,omitempty"`

	// cacheActiveTimeout is the max period during which the reporter will aggregate flows before sending.
	// Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load,
	// however you can expect higher memory consumption and an increased latency in the flow collection.
	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="5s"
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty"`

	// cacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows.
	// Increasing `cacheMaxFlows` and `cacheActiveTimeout` can decrease the network traffic overhead and the CPU load,
	// however you can expect higher memory consumption and an increased latency in the flow collection.
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=100000
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty"`

	// interfaces contains the interface names from where flows will be collected. If empty, the agent
	// will fetch all the interfaces in the system, excepting the ones listed in ExcludeInterfaces.
	// If an entry is enclosed by slashes (such as `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+optional
	Interfaces []string `json:"interfaces,omitempty"`

	// excludeInterfaces contains the interface names that will be excluded from flow tracing.
	// If an entry is enclosed by slashes (such as `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+kubebuilder:default=lo;
	ExcludeInterfaces []string `json:"excludeInterfaces,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel defines the log level for the NetObserv eBPF Agent
	LogLevel string `json:"logLevel,omitempty"`

	// privileged mode for the eBPF Agent container. In general this setting can be ignored or set to false:
	// in that case, the operator will set granular capabilities (BPF, PERFMON, NET_ADMIN, SYS_RESOURCE)
	// to the container, to enable its correct operation.
	// If for some reason these capabilities cannot be set (for example old kernel version not knowing CAP_BPF)
	// then you can turn on this mode for more global privileges.
	// +optional
	Privileged bool `json:"privileged,omitempty"`

	//+kubebuilder:default:=10485760
	// +optional
	// kafkaBatchSize limits the maximum size of a request in bytes before being sent to a partition. Ignored when not using Kafka. Default: 10MB.
	KafkaBatchSize int `json:"kafkaBatchSize"`

	// Debug allows setting some aspects of the internal configuration of the eBPF agent.
	// This section is aimed exclusively for debugging and fine-grained performance optimizations
	// (for example GOGC, GOMAXPROCS env vars). Users setting its values do it at their own risk.
	// +optional
	Debug DebugConfig `json:"debug,omitempty"`
}

// FlowCollectorKafka defines the desired Kafka config of FlowCollector
type FlowCollectorKafka struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=""
	// address of the Kafka server
	Address string `json:"address"`

	//+kubebuilder:default:=""
	// kafka topic to use. It must exist, NetObserv will not create it.
	Topic string `json:"topic"`

	// tls client configuration. When using TLS, verify that the address matches the Kafka port used for TLS, generally 9093.
	// Note that, when eBPF agents are used, Kafka certificate needs to be copied in the agent namespace (by default it's netobserv-privileged).
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

// ServerTLS define the TLS configuration, server side
type ServerTLS struct {
	// Select the type of TLS configuration
	// "DISABLED" (default) to not configure TLS for the endpoint, "PROVIDED" to manually provide cert file and a key file,
	// and "AUTO" to use OpenShift auto generated certificate using annotations
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="DISABLED";"PROVIDED";"AUTO"
	// +kubebuilder:validation:Required
	//+kubebuilder:default:="DISABLED"
	Type ServerTLSConfigType `json:"type,omitempty"`

	// TLS configuration.
	// +optional
	Provided *CertificateReference `json:"provided"`
}

// MetricsServerConfig define the metrics server endpoint configuration for Prometheus scraper
type MetricsServerConfig struct {

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9102
	// the prometheus HTTP port
	Port int32 `json:"port,omitempty"`

	// TLS configuration.
	// +optional
	TLS ServerTLS `json:"tls"`
}

// FLPMetrics define the desired FLP configuration regarding metrics
type FLPMetrics struct {
	// metricsServer endpoint configuration for Prometheus scraper
	// +optional
	Server MetricsServerConfig `json:"server,omitempty"`

	// ignoreTags is a list of tags to specify which metrics to ignore. Each metric is associated with a list of tags. More details in https://github.com/netobserv/network-observability-operator/tree/main/controllers/flowlogspipeline/metrics_definitions .
	// Available tags are: egress, ingress, flows, bytes, packets, namespaces, nodes, workloads
	//+kubebuilder:default:={"egress","packets"}
	IgnoreTags []string `json:"ignoreTags,omitempty"`
}

// FlowCollectorFLP defines the desired flowlogs-pipeline state of FlowCollector
type FlowCollectorFLP struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Minimum=1025
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=2055
	// port of the flow collector (host port)
	// By conventions, some value are not authorized port must not be below 1024 and must not equal this values:
	// 4789,6081,500, and 4500
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=8080
	// healthPort is a collector HTTP port in the Pod that exposes the health check API
	HealthPort int32 `json:"healthPort,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=65535
	//+optional
	// profilePort allows setting up a Go pprof profiler listening to this port
	ProfilePort int32 `json:"profilePort,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Metrics define the processor configuration regarding metrics
	Metrics FLPMetrics `json:"metrics,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel of the collector runtime
	LogLevel string `json:"logLevel,omitempty"`

	//+kubebuilder:default:={requests:{memory:"100Mi",cpu:"100m"},limits:{memory:"800Mi"}}
	// resources are the compute resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:default:=true
	// enableKubeProbes is a flag to enable or disable Kubernetes liveness and readiness probes
	EnableKubeProbes bool `json:"enableKubeProbes,omitempty"`

	//+kubebuilder:default:=true
	// dropUnusedFields allows, when set to true, to drop fields that are known to be unused by OVS, in order to save storage space.
	DropUnusedFields bool `json:"dropUnusedFields,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=3
	// kafkaConsumerReplicas defines the number of replicas (pods) to start for flowlogs-pipeline-transformer, which consumes Kafka messages.
	// This setting is ignored when Kafka is disabled.
	KafkaConsumerReplicas int32 `json:"kafkaConsumerReplicas,omitempty"`

	// kafkaConsumerAutoscaler spec of a horizontal pod autoscaler to set up for flowlogs-pipeline-transformer, which consumes Kafka messages.
	// This setting is ignored when Kafka is disabled.
	// +optional
	KafkaConsumerAutoscaler FlowCollectorHPA `json:"kafkaConsumerAutoscaler,omitempty"`

	//+kubebuilder:default:=1000
	// +optional
	// kafkaConsumerQueueCapacity defines the capacity of the internal message queue used in the Kafka consumer client. Ignored when not using Kafka.
	KafkaConsumerQueueCapacity int `json:"kafkaConsumerQueueCapacity"`

	//+kubebuilder:default:=10485760
	// +optional
	// kafkaConsumerBatchSize indicates to the broker the maximum batch size, in bytes, that the consumer will accept. Ignored when not using Kafka. Default: 10MB.
	KafkaConsumerBatchSize int `json:"kafkaConsumerBatchSize"`

	// Debug allows setting some aspects of the internal configuration of the flow processor.
	// This section is aimed exclusively for debugging and fine-grained performance optimizations
	// (for example GOGC, GOMAXPROCS env vars). Users setting its values do it at their own risk.
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
	// Status describe the desired status regarding deploying an horizontal pod autoscaler
	// DISABLED will not deploy an horizontal pod autoscaler
	// ENABLED will deploy an horizontal pod autoscaler
	Status string `json:"status,omitempty"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// maxReplicas is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	// +kubebuilder:default:=3
	// +optional
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// metrics used by the pod autoscaler
	// +optional
	Metrics []ascv2.MetricSpec `json:"metrics"`
}

const (
	LokiAuthDisabled         = "DISABLED"
	LokiAuthUseHostToken     = "HOST"
	LokiAuthForwardUserToken = "FORWARD"
)

// FlowCollectorLoki defines the desired state for FlowCollector's Loki client.
type FlowCollectorLoki struct {
	//+kubebuilder:default:="http://loki:3100/"
	// url is the address of an existing Loki service to push the flows to. When using the Loki Operator,
	// set it to the Loki gateway service with the `network` tenant set in path, for example
	// https://loki-gateway-http.netobserv.svc:8080/api/logs/v1/network.
	URL string `json:"url,omitempty"`

	//+kubebuilder:validation:optional
	// querierURL specifies the address of the Loki querier service, in case it is different from the
	// Loki ingester URL. If empty, the URL value will be used (assuming that the Loki ingester
	// and querier are in the same server). When using the Loki Operator, do not set it, since
	// ingestion and queries use the Loki gateway.
	QuerierURL string `json:"querierUrl,omitempty"`

	//+kubebuilder:validation:optional
	// statusURL specifies the address of the Loki /ready /metrics /config endpoints, in case it is different from the
	// Loki querier URL. If empty, the QuerierURL value will be used.
	// This is useful to show error messages and some context in the frontend.
	// When using the Loki Operator, set it to the Loki HTTP query frontend service, for example
	// https://loki-query-frontend-http.netobserv.svc:3100/.
	StatusURL string `json:"statusUrl,omitempty"`

	//+kubebuilder:default:="netobserv"
	// tenantID is the Loki X-Scope-OrgID that identifies the tenant for each request.
	// When using the Loki Operator, set it to `network`, which corresponds to a special tenant mode.
	TenantID string `json:"tenantID,omitempty"`

	// +kubebuilder:validation:Enum:="DISABLED";"HOST";"FORWARD"
	//+kubebuilder:default:="DISABLED"
	// AuthToken describe the way to get a token to authenticate to Loki.
	// DISABLED will not send any token with the request.
	// HOST will use the local pod service account to authenticate to Loki.
	// FORWARD will forward user token, in this mode, pod that are not receiving user request like the processor will use the local pod service account. Similar to HOST mode.
	// When using the Loki Operator, set it to `HOST` or `FORWARD`.
	AuthToken string `json:"authToken,omitempty"`

	//+kubebuilder:default:="1s"
	// batchWait is max time to wait before sending a batch.
	BatchWait metav1.Duration `json:"batchWait,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=102400
	// batchSize is max batch size (in bytes) of logs to accumulate before sending.
	BatchSize int64 `json:"batchSize,omitempty"`

	//+kubebuilder:default:="10s"
	// timeout is the maximum time connection / request limit.
	// A Timeout of zero means no timeout.
	Timeout metav1.Duration `json:"timeout,omitempty"`

	//+kubebuilder:default:="1s"
	// minBackoff is the initial backoff time for client connection between retries.
	MinBackoff metav1.Duration `json:"minBackoff,omitempty"`

	//+kubebuilder:default:="5s"
	// maxBackoff is the maximum backoff time for client connection between retries.
	MaxBackoff metav1.Duration `json:"maxBackoff,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=2
	// maxRetries is the maximum number of retries for client connections.
	MaxRetries int32 `json:"maxRetries,omitempty"`

	//+kubebuilder:default:={"app":"netobserv-flowcollector"}
	// staticLabels is a map of common labels to set on each flow.
	StaticLabels map[string]string `json:"staticLabels,omitempty"`

	// tls client configuration.
	// +optional
	TLS ClientTLS `json:"tls"`
}

// FlowCollectorConsolePlugin defines the desired ConsolePlugin state of FlowCollector
type FlowCollectorConsolePlugin struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=true
	// register allows, when set to true, to automatically register the provided console plugin with the OpenShift Console operator.
	// When set to false, you can still register it manually by editing console.operator.openshift.io/cluster.
	// E.g: oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "netobserv-plugin"}]'
	Register bool `json:"register"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	// replicas defines the number of replicas (pods) to start.
	Replicas int32 `json:"replicas,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9001
	// port is the plugin service port
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"100Mi"}}
	// resources, in terms of compute resources, required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel for the console plugin backend
	LogLevel string `json:"logLevel,omitempty"`

	// autoscaler spec of a horizontal pod autoscaler to set up for the plugin Deployment.
	// +optional
	Autoscaler FlowCollectorHPA `json:"autoscaler,omitempty"`

	//+kubebuilder:default:={enable:true}
	// portNaming defines the configuration of the port-to-service name translation
	PortNaming ConsolePluginPortConfig `json:"portNaming,omitempty"`

	//+kubebuilder:default:={{name:"Applications",filter:{"src_namespace!":"openshift-,netobserv","dst_namespace!":"openshift-,netobserv"},default:true},{name:"Infrastructure",filter:{"src_namespace":"openshift-,netobserv","dst_namespace":"openshift-,netobserv"}},{name:"Pods network",filter:{"src_kind":"Pod","dst_kind":"Pod"},default:true},{name:"Services network",filter:{"dst_kind":"Service"}}}
	// quickFilters configures quick filter presets for the Console plugin
	QuickFilters []QuickFilter `json:"quickFilters,omitempty"`
}

// Configuration of the port to service name translation feature of the console plugin
type ConsolePluginPortConfig struct {
	//+kubebuilder:default:=true
	// enable the console plugin port-to-service name translation
	Enable bool `json:"enable,omitempty"`

	// portNames defines additional port names to use in the console.
	// Example: portNames: {"3100": "loki"}
	// +optional
	PortNames map[string]string `json:"portNames,omitempty" yaml:"portNames,omitempty"`
}

// QuickFilter defines preset configuration for Console's quick filters
type QuickFilter struct {
	// name of the filter, that will be displayed in Console
	// +kubebuilder:MinLength:=1
	Name string `json:"name"`
	// filter is a set of keys and values to be set when this filter is selected. Each key can relate to a list of values using a coma-separated string.
	// Example: filter: {"src_namespace": "namespace1,namespace2"}
	// +kubebuilder:MinProperties:=1
	Filter map[string]string `json:"filter"`
	// default defines whether this filter should be active by default or not
	// +optional
	Default bool `json:"default,omitempty"`
}

// ClusterNetworkOperatorConfig defines the desired configuration related to the Cluster Network Configuration
type ClusterNetworkOperatorConfig struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=openshift-network-operator
	// namespace  where the config map is going to be deployed.
	Namespace string `json:"namespace,omitempty"`
}

// OVNKubernetesConfig defines the desired configuration related to the OVN-Kubernetes network provider, when Cluster Network Operator isn't installed.
type OVNKubernetesConfig struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=ovn-kubernetes
	// namespace where OVN-Kubernetes pods are deployed.
	Namespace string `json:"namespace,omitempty"`

	//+kubebuilder:default:=ovnkube-node
	// daemonSetName defines the name of the DaemonSet controlling the OVN-Kubernetes pods.
	DaemonSetName string `json:"daemonSetName,omitempty"`

	//+kubebuilder:default:=ovnkube-node
	// containerName defines the name of the container to configure for IPFIX.
	ContainerName string `json:"containerName,omitempty"`
}

type MountableType string

const (
	CertRefTypeSecret    MountableType = "secret"
	CertRefTypeConfigMap MountableType = "configmap"
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
	// type for the certificate reference: "configmap" or "secret"
	Type MountableType `json:"type,omitempty"`

	// name of the config map or secret containing certificates
	Name string `json:"name,omitempty"`

	// certFile defines the path to the certificate file name within the config map or secret
	CertFile string `json:"certFile,omitempty"`

	// certKey defines the path to the certificate private key file name within the config map or secret. Omit when the key is not necessary.
	// +optional
	CertKey string `json:"certKey,omitempty"`

	// namespace of the config map or secret containing certificates. If omitted, assumes same namespace as where NetObserv is deployed.
	// If the namespace is different, the config map or the secret will be copied so that it can be mounted as required.
	// +optional
	//+kubebuilder:default:=""
	Namespace string `json:"namespace,omitempty"`
}

// ClientTLS defines TLS client configuration
type ClientTLS struct {
	//+kubebuilder:default:=false
	// enable TLS
	Enable bool `json:"enable,omitempty"`

	//+kubebuilder:default:=false
	// insecureSkipVerify allows skipping client-side verification of the server certificate
	// If set to true, CACert field will be ignored
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// caCert defines the reference of the certificate for the Certificate Authority
	CACert CertificateReference `json:"caCert,omitempty"`

	// userCert defines the user certificate reference, used for mTLS (you can ignore it when using regular, one-way TLS)
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

// DebugConfig allows tweaking some aspects of the internal configuration of the agent and FLP.
// They are aimed exclusively for debugging. Users setting these values do it at their own risk.
type DebugConfig struct {
	// env allows passing custom environment variables to the NetObserv Agent. Useful for passing
	// some very concrete performance-tuning options (such as GOGC, GOMAXPROCS) that shouldn't be
	// publicly exposed as part of the FlowCollector descriptor, as they are only useful
	// in edge debug and support scenarios.
	//+optional
	Env map[string]string `json:"env,omitempty"`
}

// Add more exporter types below
type ExporterType string

const (
	KafkaExporter ExporterType = "KAFKA"
)

// FlowCollectorExporter defines an additional exporter to send enriched flows to
type FlowCollectorExporter struct {
	// `type` selects the type of exporters. The available options are `KAFKA` and `IPFIX`.
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="KAFKA";"IPFIX"
	// +kubebuilder:validation:Required
	Type ExporterType `json:"type"`

	// kafka configuration, such as address or topic, to send enriched flows to.
	// +optional
	Kafka FlowCollectorKafka `json:"kafka,omitempty"`

	// IPFIX configuration, such as the IP address and port to send enriched IPFIX flows to.
	// +optional
	IPFIX FlowCollectorIPFIXReceiver `json:"ipfix,omitempty"`
}

// FlowCollectorStatus defines the observed state of FlowCollector
type FlowCollectorStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// namespace where console plugin and flowlogs-pipeline have been deployed.
	Namespace string `json:"namespace,omitempty"`
}

// +kubebuilder:deprecatedversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Agent",type="string",JSONPath=`.spec.agent.type`
// +kubebuilder:printcolumn:name="Sampling (EBPF)",type="string",JSONPath=`.spec.agent.ebpf.sampling`
// +kubebuilder:printcolumn:name="Deployment Model",type="string",JSONPath=`.spec.deploymentModel`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].reason`

// FlowCollector is the Schema for the flowcollectors API, which pilots and configures netflow collection.
//
// Deprecated: This package will be removed in one of the next releases.
type FlowCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowCollectorSpec   `json:"spec,omitempty"`
	Status FlowCollectorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FlowCollectorList contains a list of FlowCollector
//
// Deprecated: This package will be removed in one of the next releases.
type FlowCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowCollector{}, &FlowCollectorList{})
}
