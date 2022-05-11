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
	AgentIPFIX = "ipfix"
	AgentEBPF  = "ebpf"
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

	//+kubebuilder:default:=""
	// Namespace where console plugin and collector pods are going to be deployed.
	// If empty, the namespace of the operator is going to be used
	Namespace string `json:"namespace,omitempty"`

	//+kubebuilder:validation:Enum=ipfix;ebpf
	//+kubebuilder:default:=ipfix
	// Agent selects the flows' tracing agent. Possible values are "ipfix" (default) to use
	// the OpenVSwitch IPFIX collector (only valid if your cluster uses OVN-Kubernetes CNI) or
	// "ebpf" to use NetObserv's eBPF agent. The eBPF agent is not officially released yet, it
	// is provided as a preview.
	Agent string `json:"agent"`

	// IPFIX contains the settings of an IPFIX-based flow reporter when the "agent" property is set
	// to "ipfix".
	// defined if the ebpf section is already defined
	// +kubebuilder:default:={sampling:400}
	IPFIX FlowCollectorIPFIX `json:"ipfix,omitempty"`

	// EBPF contains the settings of an eBPF-based flow reporter  when the "agent" property is set
	// to "ebpf".
	// +kubebuilder:default={imagePullPolicy:"IfNotPresent"}
	EBPF FlowCollectorEBPF `json:"ebpf,omitempty"`

	// FlowlogsPipeline contains settings related to the flowlogs-pipeline component
	FlowlogsPipeline FlowCollectorFLP `json:"flowlogsPipeline,omitempty"`

	// Loki contains settings related to the loki client
	Loki FlowCollectorLoki `json:"loki,omitempty"`

	// ConsolePlugin contains settings related to the console dynamic plugin
	ConsolePlugin FlowCollectorConsolePlugin `json:"consolePlugin,omitempty"`

	// ClusterNetworkOperator contains settings related to the cluster network operator
	ClusterNetworkOperator ClusterNetworkOperator `json:"clusterNetworkOperator,omitempty"`
}

// FlowCollectorIPFIX defines a FlowCollector that uses IPFIX on OVN-Kubernetes to collect the
// flows information
type FlowCollectorIPFIX struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="60s"
	// CacheActiveTimeout is the max period during which the reporter will aggregate flows before sending
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty" mapstructure:"cacheActiveTimeout,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=100
	// CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty" mapstructure:"cacheMaxFlows,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=400
	// Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. 0 means disabled.
	Sampling int32 `json:"sampling,omitempty" mapstructure:"sampling,omitempty"`
}

// FlowCollectorEBPF defines a FlowCollector that uses eBPF to collect the flows information
type FlowCollectorEBPF struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:="quay.io/netobserv/netobserv-ebpf-agent:main"
	// Image is the NetObserv Agent image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// ImagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. 0 or 1 means disabled.
	//+optional
	Sampling int32 `json:"sampling,omitempty"`

	// CacheActiveTimeout is the max period during which the reporter will aggregate flows before sending
	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="5s"
	CacheActiveTimeout string `json:"cacheActiveTimeout,omitempty"`

	// CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=1000
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty"`

	// Interfaces contains the interface names from where flows will be collected. If empty, the agent
	// will fetch all the interfaces in the system, excepting the ones listed in ExcludeInterfaces.
	// If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+optional
	Interfaces []string `json:"interfaces,omitempty"`

	// ExcludeInterfaces contains the interface names that will be excluded from flow tracing.
	// If an entry is enclosed by slashes (e.g. `/br-/`), it will match as regular expression,
	// otherwise it will be matched as a case-sensitive string.
	//+kubebuilder:default=lo;
	ExcludeInterfaces []string `json:"excludeInterfaces,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// LogLevel defines the log level for the NetObserv eBPF Agent
	LogLevel string `json:"logLevel,omitempty"`

	// Env allows passing custom environment variables to the NetObserv Agent. Useful for passing
	// some very concrete performance-tuning options (e.g. GOGC, GOMAXPROCS) that shouldn't be
	// publicly exposed as part of the FlowCollector descriptor, as they are only useful
	// in edge debug/support scenarios.
	//+optional
	Env map[string]string `json:"env,omitempty"`

	// Privileged mode for the eBPF Agent container. If false, the operator will add the following
	// capabilities to the container, to enable its correct operation:
	// BPF, PERFMON, NET_ADMIN, SYS_RESOURCE.
	// +optional
	Privileged bool `json:"privileged,omitempty"`
}

// FlowCollectorFLP defines the desired flowlogs-pipeline state of FlowCollector
type FlowCollectorFLP struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=DaemonSet;Deployment
	//+kubebuilder:default:=DaemonSet
	// Kind is the workload kind, either DaemonSet or Deployment
	Kind string `json:"kind,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	// Replicas defines the number of replicas (pods) to start for Deployment kind. Ignored for DaemonSet.
	Replicas int32 `json:"replicas,omitempty"`

	// HPA spec of an horizontal pod autoscaler to set up for the collector Deployment. Ignored for DaemonSet.
	// +optional
	HPA *FlowCollectorHPA `json:"hpa,omitempty"`

	//+kubebuilder:validation:Minimum=1025
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=2055
	// Port is the collector port: either a service port for Deployment kind, or host port for DaemonSet kind
	// By conventions, some value are not authorized port must not be below 1024 and must not equal this values:
	// 4789,6081,500, and 4500
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=8080
	// HealthPort is a collector HTTP port in the Pod that exposes the health check API
	HealthPort int32 `json:"healthPort,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9090
	// PrometheusPort is the prometheus HTTP port: this port exposes prometheus metrics
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	//+kubebuilder:default:="quay.io/netobserv/flowlogs-pipeline:main"
	// Image is the collector image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// ImagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// LogLevel defines the log level for the collector runtime
	LogLevel string `json:"logLevel,omitempty"`

	//+kubebuilder:default:={requests:{memory:"100Mi",cpu:"100m"},limits:{memory:"300Mi"}}
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:default:=true
	// EnableKubeProbes is a flag to enable or disable Kubernetes liveness/readiness probes
	EnableKubeProbes bool `json:"enableKubeProbes,omitempty"`
}

type FlowCollectorHPA struct {
	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// Metrics used by the pod autoscaler
	// +optional
	Metrics []ascv2.MetricSpec `json:"metrics"`
}

// FlowCollectorLoki defines the desired state for FlowCollector's Loki client
type FlowCollectorLoki struct {
	//+kubebuilder:default:="http://loki:3100/"
	// URL is the address of an existing Loki service to push the flows to.
	URL string `json:"url,omitempty"`

	//+kubebuilder:validation:optional
	// QuerierURL specifies the address of the Loki querier service, in case it is different from the
	// Loki ingester URL. If empty, the URL value will be used (assuming that the Loki ingester
	// and querier are int he same host).
	QuerierURL string `json:"querierUrl,omitempty"`

	//+kubebuilder:default:="1s"
	// BatchWait is max time to wait before sending a batch
	BatchWait metav1.Duration `json:"batchWait,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:default:=102400
	// BatchSize is max batch size (in bytes) of logs to accumulate before sending
	BatchSize int64 `json:"batchSize,omitempty"`

	//+kubebuilder:default:="10s"
	// Timeout is the maximum time connection / request limit
	// A Timeout of zero means no timeout.
	Timeout metav1.Duration `json:"timeout,omitempty"`

	//+kubebuilder:default:="1s"
	// MinBackoff is the initial backoff time for client connection between retries
	MinBackoff metav1.Duration `json:"minBackoff,omitempty"`

	//+kubebuilder:default:="300s"
	// MaxBackoff is the maximum backoff time for client connection between retries
	MaxBackoff metav1.Duration `json:"maxBackoff,omitempty"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=10
	// MaxRetries is the maximum number of retries for client connections
	MaxRetries int32 `json:"maxRetries,omitempty"`

	//+kubebuilder:default:={"app":"netobserv-flowcollector"}
	// StaticLabels is a map of common labels to set on each flow
	StaticLabels map[string]string `json:"staticLabels,omitempty"`
}

// FlowCollectorConsolePlugin defines the desired ConsolePlugin state of FlowCollector
type FlowCollectorConsolePlugin struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=true
	// Automatically register the provided console plugin with the OpenShift Console operator.
	// When set to false, you can still register it manually by editing console.operator.openshift.io/cluster.
	// E.g: oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "network-observability-plugin"}]'
	Register bool `json:"register"`

	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:default:=1
	// Replicas defines the number of replicas (pods) to start.
	Replicas int32 `json:"replicas,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=9001
	// Port is the plugin service port
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:default:="quay.io/netobserv/network-observability-console-plugin:main"
	// Image is the plugin image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	//+kubebuilder:default:=IfNotPresent
	// ImagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:={requests:{memory:"50Mi",cpu:"100m"},limits:{memory:"100Mi"}}
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	//+kubebuilder:default:=info
	// LogLevel defines the log level for the console plugin backend
	LogLevel string `json:"logLevel,omitempty"`

	// HPA spec of an horizontal pod autoscaler to set up for the plugin Deployment.
	// +optional
	HPA *FlowCollectorHPA `json:"hpa,omitempty"`

	//+kubebuilder:default:={enable:true}
	// Configuration of the port to service name translation
	PortNaming ConsolePluginPortConfig `json:"portNaming,omitempty"`
}

// Configuration of the port to service name translation feature of the console plugin
type ConsolePluginPortConfig struct {
	//+kubebuilder:default:=true
	// Should this feature be enabled
	Enable bool `json:"enable,omitempty"`

	// Additional port name to use in the console
	// E.g. portNames: {"3100": "loki"}
	// +optional
	PortNames map[string]string `json:"portNames,omitempty" yaml:"portNames,omitempty"`
}

// CNO defines the desired configuration related to the Cluster Network Configuration
type ClusterNetworkOperator struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=openshift-network-operator
	// Namespace  where the configmap is going to be deployed.
	Namespace string `json:"namespace,omitempty"`
}

// FlowCollectorStatus defines the observed state of FlowCollector
type FlowCollectorStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Namespace where console plugin and flowlogs-pipeline have been deployed.
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

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
