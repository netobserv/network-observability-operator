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

// +union
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'XDP' ?  has(self.xdp) : !has(self.xdp)",message="xdp configuration is required when type is xdp, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TC' ?  has(self.tc) : !has(self.tc)",message="tc configuration is required when type is tc, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TCX' ?  has(self.tcx) : !has(self.tcx)",message="tcx configuration is required when type is TCX, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'UProbe' ?  has(self.uprobe) : !has(self.uprobe)",message="uprobe configuration is required when type is uprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'URetProbe' ?  has(self.uretprobe) : !has(self.uretprobe)",message="uretprobe configuration is required when type is uretprobe, and forbidden otherwise"
type BpfApplicationProgramState struct {
	BpfProgramStateCommon `json:",inline"`

	// type specifies the provisioned eBPF program type for this program entry.
	// Type will be one of:
	//   TC, TCX, UProbe, URetProbe, XDP
	//
	// When set to TC, the tc object will be populated with the eBPF program data
	// associated with a TC program.
	//
	// When set to TCX, the tcx object will be populated with the eBPF program
	// data associated with a TCX program.
	//
	// When set to UProbe, the uprobe object will be populated with the eBPF
	// program data associated with a UProbe program.
	//
	// When set to URetProbe, the uretprobe object will be populated with the eBPF
	// program data associated with a URetProbe program.
	//
	// When set to XDP, the xdp object will be populated with the eBPF program data
	// associated with a URetProbe program.
	// +unionDiscriminator
	// +required
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"UProbe";"URetProbe"
	Type EBPFProgType `json:"type"`

	// xdp contains the attachment data for an XDP program when type is set to XDP.
	// +unionMember
	// +optional
	XDP *XdpProgramInfoState `json:"xdp,omitempty"`

	// tc contains the attachment data for a TC program when type is set to TC.
	// +unionMember
	// +optional
	TC *TcProgramInfoState `json:"tc,omitempty"`

	// tcx contains the attachment data for a TCX program when type is set to TCX.
	// +unionMember
	// +optional
	TCX *TcxProgramInfoState `json:"tcx,omitempty"`

	// uprobe contains the attachment data for a UProbe program when type is set to
	// UProbe.
	// +unionMember
	// +optional
	UProbe *UprobeProgramInfoState `json:"uprobe,omitempty"`

	// uretprobe contains the attachment data for a URetProbe program when type is
	// set to URetProbe.
	// +unionMember
	// +optional
	URetProbe *UprobeProgramInfoState `json:"uretprobe,omitempty"`
}

type BpfApplicationStateStatus struct {
	// UpdateCount tracks the number of times the BpfApplicationState object has
	// been updated. The bpfman agent initializes it to 1 when it creates the
	// object, and then increments it before each subsequent update. It serves
	// as a lightweight sequence number to verify that the API server is serving
	// the most recent version of the object before beginning a new Reconcile
	// operation.
	UpdateCount int64 `json:"updateCount"`
	// node is the name of the Kubernets node for this BpfApplicationState.
	Node string `json:"node"`
	// appLoadStatus reflects the status of loading the eBPF application on the
	// given node.
	//
	// NotLoaded is a temporary state that is assigned when a
	// ClusterBpfApplicationState is created and the initial reconcile is being
	// processed.
	//
	// LoadSuccess is returned if all the programs have been loaded with no
	// errors.
	//
	// LoadError is returned if one or more programs encountered an error and
	// were not loaded.
	//
	// NotSelected is returned if this application did not select to run on this
	// Kubernetes node.
	//
	// UnloadSuccess is returned when all the programs were successfully
	// unloaded.
	//
	// UnloadError is returned if one or more programs encountered an error when
	// being unloaded.
	AppLoadStatus AppLoadStatus `json:"appLoadStatus"`
	// programs is a list of eBPF programs contained in the parent BpfApplication
	// instance. Each entry in the list contains the derived program attributes as
	// well as the attach status for each program on the given Kubernetes node.
	Programs []BpfApplicationProgramState `json:"programs,omitempty"`
	// conditions contains the summary state of the BpfApplication for the given
	// Kubernetes node. If one or more programs failed to load or attach to the
	// designated attachment point, the condition will report the error. If more
	// than one error has occurred, condition will contain the first error
	// encountered.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// BpfApplicationState contains the state of a BpfApplication instance for a
// given Kubernetes node. When a user creates a BpfApplication instance, bpfman
// creates a BpfApplicationState instance for each node in a Kubernetes
// cluster.
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=".status.node"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BpfApplicationState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// status reflects the status of a BpfApplication instance for the given node.
	// appLoadStatus and conditions provide an overall status for the given node,
	// while each item in the programs list provides a per eBPF program status for
	// the given node.
	Status BpfApplicationStateStatus `json:"status,omitempty"`
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

func (an BpfApplicationState) GetConditions() []metav1.Condition {
	return an.Status.Conditions
}

func (an BpfApplicationState) GetClientObject() client.Object {
	return &an
}

func (anl BpfApplicationStateList) GetItems() []BpfApplicationState {
	return anl.Items
}
