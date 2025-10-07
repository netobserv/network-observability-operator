/*
Copyright 2023 The bpfman Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Config holds the configuration for bpfman-operator.
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// Spec defines the desired state of the bpfman-operator.
type ConfigSpec struct {
	// Agent holds the configuration for the bpfman agent.
	// +required
	Agent AgentSpec `json:"agent,omitempty"`
	// Configuration holds the content of bpfman.toml.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Configuration string `json:"configuration"`
	// Image holds the image of the bpfman DaemonSets.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`
	// LogLevel holds the log level for the bpfman-operator.
	// +optional
	LogLevel string `json:"logLevel,omitempty"`
	// Namespace holds the namespace where bpfman-operator resources shall be
	// deployed.
	Namespace string `json:"namespace,omitempty"`
}

// AgentSpec defines the desired state of the bpfman agent.
type AgentSpec struct {
	// HealthProbePort holds the health probe bind port for the bpfman agent.
	// +optional
	// +kubebuilder:default=8175
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	HealthProbePort int `json:"healthProbePort"`
	// Image holds the image for the bpfman agent.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`
	// LogLevel holds the log level for the bpfman agent.
	// +optional
	LogLevel string `json:"logLevel,omitempty"`
}

// status reflects the status of the bpfman-operator configuration.
type ConfigStatus struct {
	// conditions store the status conditions of the bpfman-operator.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// ConfigList contains a list of Configs.
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}
