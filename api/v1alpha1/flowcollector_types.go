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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlowCollectorSpec defines the desired state of FlowCollector
type FlowCollectorSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:default:=""
	// Namespace where console plugin and goflowkube pods are going to be deployed.
	// If empty, the namespace of the operator is going to be used
	Namespace string `json:"namespace,omitempty"`

	// IPFIX contains IPFIX-related settings for the flow reporter
	IPFIX FlowCollectorIPFIX `json:"ipfix,omitempty"`

	// GoflowKube contains settings related to goflow-kube
	GoflowKube FlowCollectorGoflowKube `json:"goflowkube,omitempty"`

	// Loki contains settings related to the loki client
	Loki FlowCollectorLoki `json:"loki,omitempty"`

	// ConsolePlugin contains settings related to the console dynamic plugin
	ConsolePlugin FlowCollectorConsolePlugin `json:"consolePlugin,omitempty"`

	// CNO contains settings related to the cluster network operator
	CNO ClusterNetworkOperator `json:"cno,omitempty"`
}

// FlowCollectorIPFIX defines the desired IPFIX state of FlowCollector
type FlowCollectorIPFIX struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Pattern:=^\d+(ns|ms|s|m)?$
	//+kubebuilder:default:="10s"
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

// FlowCollectorGoflowKube defines the desired goflow-kube state of FlowCollector
type FlowCollectorGoflowKube struct {
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

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	//+kubebuilder:default:=2055
	// Port is the collector port: either a service port for Deployment kind, or host port for DaemonSet kind
	Port int32 `json:"port,omitempty"`

	//+kubebuilder:default:="quay.io/netobserv/goflow2-kube:main"
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

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	//+kubebuilder:default:=false
	// PrintOutput is a debug flag to print flows exported in kube-enricher logs
	PrintOutput bool `json:"printOutput,omitempty"`
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
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	// +optional
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty" protobuf:"varint,4,opt,name=targetCPUUtilizationPercentage"`
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

	//+kubebuilder:default:="TimeFlowEnd"
	// TimestampLabel is the label to use for time indexing in Loki. E.g. "TimeReceived", "TimeFlowStart", "TimeFlowEnd".
	TimestampLabel string `json:"timestampLabel,omitempty"`
}

// FlowCollectorConsolePlugin defines the desired ConsolePlugin state of FlowCollector
type FlowCollectorConsolePlugin struct {
	// Important: Run "make generate" to regenerate code after modifying this file

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

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
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

	// Namespace where console plugin and goflowkube have been deployed.
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// FlowCollector is the Schema for the flowcollectors API
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
