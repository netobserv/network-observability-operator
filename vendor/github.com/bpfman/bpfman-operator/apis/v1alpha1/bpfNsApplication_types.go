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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BpfNsApplicationProgram defines the desired state of BpfApplication
// +union
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'XDP' ?  has(self.xdp) : !has(self.xdp)",message="xdp configuration is required when type is XDP, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TC' ?  has(self.tc) : !has(self.tc)",message="tc configuration is required when type is TC, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TCX' ?  has(self.tcx) : !has(self.tcx)",message="tcx configuration is required when type is TCX, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'Uprobe' ?  has(self.uprobe) : !has(self.uprobe)",message="uprobe configuration is required when type is Uprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'Uretprobe' ?  has(self.uretprobe) : !has(self.uretprobe)",message="uretprobe configuration is required when type is Uretprobe, and forbidden otherwise"
type BpfNsApplicationProgram struct {
	// Type specifies the bpf program type
	// +unionDiscriminator
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"Uprobe";"Uretprobe"
	Type EBPFProgType `json:"type,omitempty"`

	// xdp defines the desired state of the application's XdpNsPrograms.
	// +unionMember
	// +optional
	XDP *XdpNsProgramInfo `json:"xdp,omitempty"`

	// tc defines the desired state of the application's TcNsPrograms.
	// +unionMember
	// +optional
	TC *TcNsProgramInfo `json:"tc,omitempty"`

	// tcx defines the desired state of the application's TcxNsPrograms.
	// +unionMember
	// +optional
	TCX *TcxNsProgramInfo `json:"tcx,omitempty"`

	// uprobe defines the desired state of the application's UprobeNsPrograms.
	// +unionMember
	// +optional
	Uprobe *UprobeNsProgramInfo `json:"uprobe,omitempty"`

	// uretprobe defines the desired state of the application's UretprobeNsPrograms.
	// +unionMember
	// +optional
	Uretprobe *UprobeNsProgramInfo `json:"uretprobe,omitempty"`
}

// BpfApplicationSpec defines the desired state of BpfApplication
type BpfNsApplicationSpec struct {
	BpfAppCommon `json:",inline"`

	// Programs is a list of bpf programs supported for a specific application.
	// It's possible that the application can selectively choose which program(s)
	// to run from this list.
	// +kubebuilder:validation:MinItems:=1
	Programs []BpfNsApplicationProgram `json:"programs,omitempty"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced

// BpfNsApplication is the Schema for the bpfapplications API
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=`.spec.nodeselector`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BpfNsApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BpfNsApplicationSpec `json:"spec,omitempty"`
	Status BpfApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// BpfNsApplicationList contains a list of BpfNsApplications
type BpfNsApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BpfNsApplication `json:"items"`
}
