/*
Copyright 2024.

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

// All fields are required unless explicitly marked optional
// +kubebuilder:validation:Required
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced

// TcxNsProgram is the Schema for the TcxNsProgram API
// +kubebuilder:printcolumn:name="BpfFunctionName",type=string,JSONPath=`.spec.bpffunctionname`
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=`.spec.nodeselector`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Direction",type=string,JSONPath=`.spec.direction`,priority=1
// +kubebuilder:printcolumn:name="InterfaceSelector",type=string,JSONPath=`.spec.interfaceselector`,priority=1
// +kubebuilder:printcolumn:name="Position",type=string,JSONPath=`.spec.position`,priority=1
// +kubebuilder:printcolumn:name="Priority",type=string,JSONPath=`.spec.priority`,priority=1
type TcxNsProgram struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TcxNsProgramSpec `json:"spec"`
	// +optional
	Status TcxProgramStatus `json:"status,omitempty"`
}

// TcxNsProgramSpec defines the desired state of TcxNsProgram
type TcxNsProgramSpec struct {
	TcxNsProgramInfo `json:",inline"`
	BpfAppCommon     `json:",inline"`
}

// TcxNsProgramInfo defines the TCX Ns Program details
type TcxNsProgramInfo struct {
	BpfProgramCommon `json:",inline"`

	// Selector to determine the network interface (or interfaces)
	InterfaceSelector InterfaceSelector `json:"interfaceselector"`

	// Containers identifies the set of containers in which to attach the eBPF
	// program.
	Containers ContainerNsSelector `json:"containers"`

	// Direction specifies the direction of traffic the tcx program should
	// attach to for a given network device.
	// +kubebuilder:validation:Enum=ingress;egress
	Direction string `json:"direction"`

	// Priority specifies the priority of the tc program in relation to
	// other programs of the same type with the same attach point. It is a value
	// from 0 to 1000 where lower values have higher precedence.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	Priority int32 `json:"priority"`
}

// +kubebuilder:object:root=true
// TcxNsProgramList contains a list of TcxNsPrograms
type TcxNsProgramList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TcxNsProgram `json:"items"`
}
