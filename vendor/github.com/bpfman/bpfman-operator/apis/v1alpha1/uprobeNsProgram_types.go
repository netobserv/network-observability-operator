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

// UprobeNsProgram is the Schema for the UprobeNsPrograms API
// +kubebuilder:printcolumn:name="BpfFunctionName",type=string,JSONPath=`.spec.bpffunctionname`
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=`.spec.nodeselector`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="FunctionName",type=string,JSONPath=`.spec.func_name`,priority=1
// +kubebuilder:printcolumn:name="Offset",type=integer,JSONPath=`.spec.offset`,priority=1
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target`,priority=1
// +kubebuilder:printcolumn:name="RetProbe",type=boolean,JSONPath=`.spec.retprobe`,priority=1
// +kubebuilder:printcolumn:name="Pid",type=integer,JSONPath=`.spec.pid`,priority=1
type UprobeNsProgram struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UprobeNsProgramSpec `json:"spec"`
	// +optional
	Status UprobeProgramStatus `json:"status,omitempty"`
}

// UprobeNsProgramSpec defines the desired state of UprobeProgram
type UprobeNsProgramSpec struct {
	UprobeNsProgramInfo `json:",inline"`
	BpfAppCommon        `json:",inline"`
}

// UprobeProgramInfo contains the information about the uprobe program
type UprobeNsProgramInfo struct {
	BpfProgramCommon `json:",inline"`

	// Function to attach the uprobe to.
	// +optional
	FunctionName string `json:"func_name"`

	// Offset added to the address of the function for uprobe.
	// +optional
	// +kubebuilder:default:=0
	Offset uint64 `json:"offset"`

	// Library name or the absolute path to a binary or library.
	Target string `json:"target"`

	// Whether the program is a uretprobe.  Default is false
	// +optional
	// +kubebuilder:default:=false
	RetProbe bool `json:"retprobe"`

	// Only execute uprobe for given process identification number (PID). If PID
	// is not provided, uprobe executes for all PIDs.
	// +optional
	Pid int32 `json:"pid"`

	// Containers identifies the set of containers in which to attach the uprobe.
	// If Containers is not specified, the uprobe will be attached in the
	// bpfman-agent container.  The ContainerNsSelector is very flexible and even
	// allows the selection of all containers in a cluster.  If an attempt is
	// made to attach uprobes to too many containers, it can have a negative
	// impact on on the cluster.
	Containers ContainerNsSelector `json:"containers"`
}

// +kubebuilder:object:root=true
// UprobeNsProgramList contains a list of UprobeNsPrograms
type UprobeNsProgramList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UprobeNsProgram `json:"items"`
}
