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
	metav1types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BpfApplicationProgramState defines the desired state of BpfApplication
// +union
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'XDP' ?  has(self.xdp) : !has(self.xdp)",message="xdp configuration is required when type is xdp, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TC' ?  has(self.tc) : !has(self.tc)",message="tc configuration is required when type is tc, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TCX' ?  has(self.tcx) : !has(self.tcx)",message="tcx configuration is required when type is TCX, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'UProbe' ?  has(self.uprobe) : !has(self.uprobe)",message="uprobe configuration is required when type is uprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'URetProbe' ?  has(self.uretprobe) : !has(self.uretprobe)",message="uretprobe configuration is required when type is uretprobe, and forbidden otherwise"
type BpfApplicationProgramState struct {
	BpfProgramStateCommon `json:",inline"`

	// type specifies the bpf program type
	// +unionDiscriminator
	// +required
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"UProbe";"URetProbe"
	Type EBPFProgType `json:"type"`

	// xdp defines the desired state of the application's XdpPrograms.
	// +unionMember
	// +optional
	XDP *XdpProgramInfoState `json:"xdp,omitempty"`

	// tc defines the desired state of the application's TcPrograms.
	// +unionMember
	// +optional
	TC *TcProgramInfoState `json:"tc,omitempty"`

	// tcx defines the desired state of the application's TcxPrograms.
	// +unionMember
	// +optional
	TCX *TcxProgramInfoState `json:"tcx,omitempty"`

	// uprobe defines the desired state of the application's UprobePrograms.
	// +unionMember
	// +optional
	UProbe *UprobeProgramInfoState `json:"uprobe,omitempty"`

	// uretprobe defines the desired state of the application's UretprobePrograms.
	// +unionMember
	// +optional
	URetProbe *UprobeProgramInfoState `json:"uretprobe,omitempty"`
}

// BpfApplicationSpec defines the desired state of BpfApplication
type BpfApplicationStateSpec struct {
	// node is the name of the node for this BpfApplicationStateSpec.
	Node string `json:"node"`
	// updateCount is the number of times the BpfApplicationState has been updated. Set to 1
	// when the object is created, then it is incremented prior to each update.
	// This allows us to verify that the API server has the updated object prior
	// to starting a new Reconcile operation.
	UpdateCount int64 `json:"updateCount"`
	// appLoadStatus reflects the status of loading the bpf application on the
	// given node.
	AppLoadStatus AppLoadStatus `json:"appLoadStatus"`
	// programs is a list of bpf programs contained in the parent application.
	// It is a map from the bpf program name to BpfApplicationProgramState
	// elements.
	Programs []BpfApplicationProgramState `json:"programs,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// BpfApplicationState contains the per-node state of a BpfApplication.
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=".spec.node"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BpfApplicationState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BpfApplicationStateSpec `json:"spec,omitempty"`
	Status BpfAppStatus            `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// BpfApplicationStateList contains a list of BpfApplicationState objects
type BpfApplicationStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BpfApplicationState `json:"items"`
}

func (an BpfApplicationState) GetName() string {
	return an.Name
}

func (an BpfApplicationState) GetUID() metav1types.UID {
	return an.UID
}

func (an BpfApplicationState) GetAnnotations() map[string]string {
	return an.Annotations
}

func (an BpfApplicationState) GetLabels() map[string]string {
	return an.Labels
}

func (an BpfApplicationState) GetStatus() *BpfAppStatus {
	return &an.Status
}

func (an BpfApplicationState) GetClientObject() client.Object {
	return &an
}

func (anl BpfApplicationStateList) GetItems() []BpfApplicationState {
	return anl.Items
}
