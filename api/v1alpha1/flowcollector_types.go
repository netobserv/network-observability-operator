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
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	AgentIPFIX = "IPFIX"
	AgentEBPF  = "EBPF"
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

	// agent for flows' extraction.
	// +kubebuilder:default:={type:"EBPF"}
	Agent FlowCollectorAgent `json:"agent"`

	// flowlogsPipeline settings. It which receives the metrics from the agent, enriches the flows,
	// and produces metrics.
	FlowlogsPipeline FlowCollectorFLP `json:"flowlogsPipeline,omitempty"`

	// loki, the flow store, client settings.
	Loki FlowCollectorLoki `json:"loki,omitempty"`

	// kafka configuration, allowing to use Kafka as a broker as part of the flow collection pipeline.
	// Kafka can provide better scalability, resiliency and high availability (for more details, see https://www.redhat.com/en/topics/integration/what-is-apache-kafka).
	// +optional
	Kafka FlowCollectorKafka `json:"kafka,omitempty"`

	// consolePlugin define the settings related to the OpenShift Console plugin, when available.
	ConsolePlugin FlowCollectorConsolePlugin `json:"consolePlugin,omitempty"`

	// clusterNetworkOperator define the settings related to the OpenShift Cluster Network Operator, when available.
	ClusterNetworkOperator ClusterNetworkOperatorConfig `json:"clusterNetworkOperator,omitempty"`

	// ovnKubernetes define the settings of the OVN-Kubernetes CNI, when available. This configuration is used when using OVN's IPFIX exports, without OpenShift. When using OpenShift, refer to the `clusterNetworkOperator` property instead.
	OVNKubernetes OVNKubernetesConfig `json:"ovnKubernetes,omitempty"`
}

// FlowCollectorAgent is a discriminated union that allows to select either ipfix or ebpf, but does not
// allow defining both fields.
// +union
type FlowCollectorAgent struct {
	// type selects the flows tracing agent. Possible values are "IPFIX" (default) to use
	// the IPFIX collector, or "EBPF" to use NetObserv eBPF agent. When using IPFIX with OVN-Kubernetes
	// CNI, NetObserv will configure OVN's IPFIX exporter. Other CNIs are not supported, they could
	// work but require manual configuration.
	// +unionDiscriminator
	// +kubebuilder:validation:Enum:="IPFIX";"EBPF"
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
	// If you really want to sample every packet, which may impact the cluster stability,
	// refer to "forceSampleAll". Alternatively, you can use the eBPF Agent instead of IPFIX.
	Sampling int32 `json:"sampling,omitempty" mapstructure:"sampling,omitempty"`

	//+kubebuilder:default:=false
	// forceSampleAll allows disabling sampling in the IPFIX-based flow reporter.
	// It is not recommended to sample all the traffic with IPFIX, as it may generate cluster instability.
	// If you REALLY want to do that, set this flag to true. Use at your own risks.
	// When it is set to true, the value of "sampling" is ignored.
	ForceSampleAll bool `json:"forceSampleAll,omitempty" mapstructure:"-"`
}

// FlowCollectorEBPF defines a FlowCollector that uses eBPF to collect the flows information
type FlowCollectorEBPF struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:="quay.io/netobserv/netobserv-ebpf-agent:main"
	// image is the NetObserv Agent image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// resources are the compute resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// sampling rate of the flow reporter. 100 means one flow on 100 is sent. 0 or 1 means all flows are sampled.
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=50
	//+optional
	Sampling int32 `json:"sampling,omitempty"`

	// cacheActiveTimeout is the max period during which the reporter will aggregate flows before sending
	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="5s"
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty"`

	// cacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=5000
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty"`

	// interfaces contains the interface names from where flows will be collected. If empty, the agent
	// will fetch all the interfaces in the system, excepting the ones listed in ExcludeInterfaces.
	// If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+optional
	Interfaces []string `json:"interfaces,omitempty"`

	// excludeInterfaces contains the interface names that will be excluded from flow tracing.
	// If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+kubebuilder:default=lo;
	ExcludeInterfaces []string `json:"excludeInterfaces,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel defines the log level for the NetObserv eBPF Agent
	LogLevel string `json:"logLevel,omitempty"`

	// env allows passing custom environment variables to the NetObserv Agent. Useful for passing
	// some very concrete performance-tuning options (e.g. GOGC, GOMAXPROCS) that shouldn't be
	// publicly exposed as part of the FlowCollector descriptor, as they are only useful
	// in edge debug/support scenarios.
	//+optional
	Env map[string]string `json:"env,omitempty"`

	// privileged mode for the eBPF Agent container. If false, the operator will add the following
	// capabilities to the container, to enable its correct operation:
	// BPF, PERFMON, NET_ADMIN, SYS_RESOURCE.
	// +optional
	Privileged bool `json:"privileged,omitempty"`
}

// FlowCollectorKafka defines the desired Kafka config of FlowCollector
type FlowCollectorKafka struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=false
	// enable Kafka. Set it to true to use Kafka as part of the flow collection pipeline. When enabled, the pipeline is split in two parts: ingestion and transformation, connected by Kafka.
	// The ingestion is either done by a specific flowlogs-pipeline workload, or by the eBPF agent, depending on the value of `spec.agent`.
	// The transformation is done by a new flowlogs-pipeline deployment.
	Enable bool `json:"enable,omitempty"`

	//+kubebuilder:default:=""
	// address of the Kafka server
	Address string `json:"address"`

	//+kubebuilder:default:=""
	// kafka topic to use. It must exist, NetObserv will not create it.
	Topic string `json:"topic"`

	// tls client configuration.
	// +optional
	TLS ClientTLS `json:"tls"`
}

// FlowCollectorFLP defines the desired flowlogs-pipeline state of FlowCollector
type FlowCollectorFLP struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=DaemonSet;Deployment
	//+kubebuilder:default:=DaemonSet
	// kind of the workload, either DaemonSet or Deployment. When DaemonSet is used, each pod will receive
	// flows from the node it is running on. When Deployment is used, the flows traffic received from nodes will
	// be load-balanced. Note that in such a case, the number of replicas should be less or equal to the number of
	// nodes, as extra-pods would be unused due to session affinity with the node IP.
	// When using Kafka, this option only affects the flowlogs-pipeline ingester, not the transformer.
	Kind string `json:"kind,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	// replicas defines the number of replicas (pods) to start for Deployment kind. Ignored for DaemonSet.
	Replicas int32 `json:"replicas,omitempty"`

	// hpa spec of a horizontal pod autoscaler to set up for the collector Deployment. Ignored for DaemonSet.
	// +optional
	HPA *FlowCollectorHPA `json:"hpa,omitempty"`

	//+kubebuilder:validation:Minimum=1025
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=2055
	// port of the flow collector: either a service port for Deployment kind, or host port for DaemonSet kind
	// By conventions, some value are not authorized port must not be below 1024 and must not equal this values:
	// 4789,6081,500, and 4500
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=8080
	// healthPort is a collector HTTP port in the Pod that exposes the health check API
	HealthPort int32 `json:"healthPort,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9102
	// prometheusPort is the prometheus HTTP port: this port exposes prometheus metrics
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	//+kubebuilder:default:="quay.io/netobserv/flowlogs-pipeline:main"
	// image of the collector container (including domain and tag)
	Image string `json:"image,omitempty"`

	// ignoreMetrics is a list of tags to specify which metrics to ignore
	IgnoreMetrics []string `json:"ignoreMetrics,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel of the collector runtime
	LogLevel string `json:"logLevel,omitempty"`

	//+kubebuilder:default:={requests:{memory:"100Mi",cpu:"100m"},limits:{memory:"300Mi"}}
	// resources are the compute resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:default:=true
	// enableKubeProbes is a flag to enable or disable Kubernetes liveness/readiness probes
	EnableKubeProbes bool `json:"enableKubeProbes,omitempty"`

	//+kubebuilder:default:=true
	// dropUnusedFields allows, when set to true, to drop fields that are known to be unused by OVS, in order to save storage space.
	DropUnusedFields bool `json:"dropUnusedFields,omitempty"`
}

type FlowCollectorHPA struct {
	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// maxReplicas is the upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// metrics used by the pod autoscaler
	// +optional
	Metrics []ascv2.MetricSpec `json:"metrics"`
}

// FlowCollectorLoki defines the desired state for FlowCollector's Loki client
type FlowCollectorLoki struct {
	//+kubebuilder:default:="http://loki:3100/"
	// url is the address of an existing Loki service to push the flows to.
	URL string `json:"url,omitempty"`

	//+kubebuilder:validation:optional
	// querierURL specifies the address of the Loki querier service, in case it is different from the
	// Loki ingester URL. If empty, the URL value will be used (assuming that the Loki ingester
	// and querier are in the same server).
	QuerierURL string `json:"querierUrl,omitempty"`

	//+kubebuilder:validation:optional
	// statusURL specifies the address of the Loki /ready /metrics /config endpoints, in case it is different from the
	// Loki querier URL. If empty, the QuerierURL value will be used.
	// This is useful to show error messages and some context in the frontend
	StatusURL string `json:"statusUrl,omitempty"`

	//+kubebuilder:default:="netobserv"
	// tenantID is the Loki X-Scope-OrgID that identifies the tenant for each request.
	// it will be ignored if instanceSpec is specified
	TenantID string `json:"tenantID,omitempty"`

	//+kubebuilder:default:=false
	// sendAuthToken is a flag to enable or disable Authorization header from service account secret
	// It allows authentication to loki operator gateway
	SendAuthToken bool `json:"sendAuthToken,omitempty"`

	//+kubebuilder:default:="1s"
	// batchWait is max time to wait before sending a batch
	BatchWait metav1.Duration `json:"batchWait,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=102400
	// batchSize is max batch size (in bytes) of logs to accumulate before sending
	BatchSize int64 `json:"batchSize,omitempty"`

	//+kubebuilder:default:="10s"
	// timeout is the maximum time connection / request limit
	// A Timeout of zero means no timeout.
	Timeout metav1.Duration `json:"timeout,omitempty"`

	//+kubebuilder:default:="1s"
	// minBackoff is the initial backoff time for client connection between retries
	MinBackoff metav1.Duration `json:"minBackoff,omitempty"`

	//+kubebuilder:default:="300s"
	// maxBackoff is the maximum backoff time for client connection between retries
	MaxBackoff metav1.Duration `json:"maxBackoff,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=10
	// maxRetries is the maximum number of retries for client connections
	MaxRetries int32 `json:"maxRetries,omitempty"`

	//+kubebuilder:default:={"app":"netobserv-flowcollector"}
	// staticLabels is a map of common labels to set on each flow
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
	// E.g: oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "network-observability-plugin"}]'
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

	//+kubebuilder:default:="quay.io/netobserv/network-observability-console-plugin:main"
	// image is the plugin image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// imagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"100Mi"}}
	// resources, in terms of compute resources, required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// logLevel for the console plugin backend
	LogLevel string `json:"logLevel,omitempty"`

	// hpa spec of a horizontal pod autoscaler to set up for the plugin Deployment.
	// +optional
	HPA *FlowCollectorHPA `json:"hpa,omitempty"`

	//+kubebuilder:default:={enable:true}
	// portNaming defines the configuration of the port-to-service name translation
	PortNaming ConsolePluginPortConfig `json:"portNaming,omitempty"`
}

// Configuration of the port to service name translation feature of the console plugin
type ConsolePluginPortConfig struct {
	//+kubebuilder:default:=true
	// enable the console plugin port-to-service name translation
	Enable bool `json:"enable,omitempty"`

	// portNames defines additional port names to use in the console
	// E.g. portNames: {"3100": "loki"}
	// +optional
	PortNames map[string]string `json:"portNames,omitempty" yaml:"portNames,omitempty"`
}

// ClusterNetworkOperatorConfig defines the desired configuration related to the Cluster Network Configuration
type ClusterNetworkOperatorConfig struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=openshift-network-operator
	// namespace  where the configmap is going to be deployed.
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

type CertificateReference struct {
	//+kubebuilder:validation:Enum=configmap;secret
	// type for the certificate reference: configmap or secret
	Type string `json:"type,omitempty"`

	// name of the ConfigMap or Secret containing certificates
	Name string `json:"name,omitempty"`

	// certFile defines the path to the certificate file name within the ConfigMap / Secret
	CertFile string `json:"certFile,omitempty"`

	// certKey defines the path to the certificate private key file name within the ConfigMap / Secret. Omit when the key is not necessary.
	// +optional
	CertKey string `json:"certKey,omitempty"`
}

// ClientTLS defines TLS client configuration
type ClientTLS struct {
	//+kubebuilder:default:=false
	// enable TLS
	Enable bool `json:"enable,omitempty"`

	//+kubebuilder:default:=false
	// insecureSkipVerify allows skipping client-side verification of the server certificate
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// caCert defines the reference of the certificate for the Certificate Authority
	CACert CertificateReference `json:"caCert,omitempty"`

	// userCert defines the user certificate reference
	// +optional
	UserCert CertificateReference `json:"userCert,omitempty"`
}

// FlowCollectorStatus defines the observed state of FlowCollector
type FlowCollectorStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// namespace where console plugin and flowlogs-pipeline have been deployed.
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Agent",type="string",JSONPath=`.spec.agent.type`
//+kubebuilder:printcolumn:name="Kafka",type="boolean",JSONPath=`.spec.kafka.enable`
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[*].reason"

// FlowCollector is the Schema for the flowcollectors API, which pilots and configures netflow collection.
type FlowCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowCollectorSpec   `json:"spec,omitempty"`
	Status FlowCollectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlowCollectorList contains a list of FlowCollector
type FlowCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowCollector{}, &FlowCollectorList{})
}
