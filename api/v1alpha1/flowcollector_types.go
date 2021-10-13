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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlowCollectorSpec defines the desired state of FlowCollector
type FlowCollectorSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// IPFIX contains IPFIX-related settings for the flow reporter
	IPFIX FlowCollectorIPFIX `json:"ipfix,omitempty"`

	// Collector contains settings related to the flows collector
	Collector FlowCollectorCollector `json:"collector,omitempty"`

	// Loki contains settings related to the loki client
	Loki FlowCollectorLoki `json:"loki,omitempty"`
}

// FlowCollectorIPFIX defines the desired IPFIX state of FlowCollector
type FlowCollectorIPFIX struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// CacheActiveTimeout is the max period during which the reporter will aggregate flows before sending
	CacheActiveTimeout time.Duration `json:"cacheActiveTimeout,omitempty"`

	//+kubebuilder:validation:Minimum=0
	// CacheMaxFlows is the max number of flows in an aggregate; when reached, the reporter sends the flows
	CacheMaxFlows int32 `json:"cacheMaxFlows,omitempty"`

	//+kubebuilder:validation:Minimum=0
	// Sampling is the sampling rate on the reporter. 100 means one flow on 100 is sent. 0 means disabled.
	Sampling int32 `json:"sampling,omitempty"`
}

// FlowCollectorCollector defines the desired collector state of FlowCollector
type FlowCollectorCollector struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	//+kubebuilder:validation:Enum=DaemonSet;Deployment
	// Kind is the workload kind, either DaemonSet or Deployment
	Kind string `json:"kind,omitempty"`

	//+kubebuilder:validation:Minimum=0
	// Replicas defines the number of replicas (pods) to start for Deployment kind. Ignored for DaemonSet.
	Replicas int32 `json:"replicas,omitempty"`

	// TODO: HPA spec of an horizontal pod autoscaler to set up for the collector Deployment. Ignored for DaemonSet.
	// TODO: HPA interface{} `json:"hpa,omitempty"`

	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	// Port is the collector port: either a service port for Deployment kind, or host port for DaemonSet kind
	Port int32 `json:"port,omitempty"`

	// Image is the collector image (including domain and tag)
	Image string `json:"image,omitempty"`

	//+kubebuilder:validation:Enum=IfNotPresent;Always;Never
	// ImagePullPolicy is the Kubernetes pull policy for the image defined above
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:validation:Enum=trace;debug;info;warn;error;fatal;panic
	// LogLevel defines the log level for the collector runtime
	LogLevel string `json:"logLevel,omitempty"`
}

// FlowCollectorLoki defines the desired state for FlowCollector's Loki client
type FlowCollectorLoki struct {
	// URL is the address of an existing Loki service to push the flows to.
	URL string `json:"url,omitempty"`

	// BatchWait is max time to wait before sending a batch
	BatchWait time.Duration `json:"batchWait,omitempty"`

	//+kubebuilder:validation:Minimum=1
	// BatchSize is max batch size (in bytes) of logs to accumulate before sending
	BatchSize int64 `json:"batchSize,omitempty"`

	// MinBackoff is the initial backoff time for client connection between retries
	MinBackoff time.Duration `json:"minBackoff,omitempty"`

	// MaxBackoff is the maximum backoff time for client connection between retries
	MaxBackoff time.Duration `json:"maxBackoff,omitempty"`

	//+kubebuilder:validation:Minimum=0
	// MaxRetries is the maximum number of retries for client connections
	MaxRetries int32 `json:"maxRetries,omitempty"`

	// StaticLabels is a map of common labels to set on each flow
	StaticLabels map[string]string `json:"staticLabels,omitempty"`
}

// FlowCollectorStatus defines the observed state of FlowCollector
type FlowCollectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
