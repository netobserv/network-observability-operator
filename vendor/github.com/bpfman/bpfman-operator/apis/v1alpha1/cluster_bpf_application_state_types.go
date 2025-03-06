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

// ClBpfApplicationProgramState defines the desired state of BpfApplication
// +union
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'XDP' ?  has(self.xdp) : !has(self.xdp)",message="xdp configuration is required when type is xdp, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TC' ?  has(self.tc) : !has(self.tc)",message="tc configuration is required when type is tc, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TCX' ?  has(self.tcx) : !has(self.tcx)",message="tcx configuration is required when type is tcx, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'FEntry' ?  has(self.fentry) : !has(self.fentry)",message="fentry configuration is required when type is fentry, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'FExit' ?  has(self.fexit) : !has(self.fexit)",message="fexit configuration is required when type is fexit, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'KProbe' ?  has(self.kprobe) : !has(self.kprobe)",message="kprobe configuration is required when type is kprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'KRetProbe' ?  has(self.kretprobe) : !has(self.kretprobe)",message="kretprobe configuration is required when type is kretprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'UProbe' ?  has(self.uprobe) : !has(self.uprobe)",message="uprobe configuration is required when type is uprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'URetProbe' ?  has(self.uretprobe) : !has(self.uretprobe)",message="uretprobe configuration is required when type is uretprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TracePoint' ?  has(self.tracepoint) : !has(self.tracepoint)",message="tracepoint configuration is required when type is tracepoint, and forbidden otherwise"
type ClBpfApplicationProgramState struct {
	BpfProgramStateCommon `json:",inline"`
	// type specifies the bpf program type
	// +unionDiscriminator
	// +required
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"FEntry";"FExit";"KProbe";"KRetProbe";"UProbe";"URetProbe";"TracePoint"
	Type EBPFProgType `json:"type"`

	// xdp defines the desired state of the application's XdpPrograms.
	// +unionMember
	// +optional
	XDP *ClXdpProgramInfoState `json:"xdp,omitempty"`

	// tc defines the desired state of the application's TcPrograms.
	// +unionMember
	// +optional
	TC *ClTcProgramInfoState `json:"tc,omitempty"`

	// tcx defines the desired state of the application's TcxPrograms.
	// +unionMember
	// +optional
	TCX *ClTcxProgramInfoState `json:"tcx,omitempty"`

	// fentry defines the desired state of the application's FentryPrograms.
	// +unionMember
	// +optional
	FEntry *ClFentryProgramInfoState `json:"fentry,omitempty"`

	// fexit defines the desired state of the application's FexitPrograms.
	// +unionMember
	// +optional
	FExit *ClFexitProgramInfoState `json:"fexit,omitempty"`

	// kprobe defines the desired state of the application's KprobePrograms.
	// +unionMember
	// +optional
	KProbe *ClKprobeProgramInfoState `json:"kprobe,omitempty"`

	// kretprobe defines the desired state of the application's KprobePrograms.
	// +unionMember
	// +optional
	KRetProbe *ClKretprobeProgramInfoState `json:"kretprobe,omitempty"`

	// uprobe defines the desired state of the application's UprobePrograms.
	// +unionMember
	// +optional
	UProbe *ClUprobeProgramInfoState `json:"uprobe,omitempty"`

	// uretprobe defines the desired state of the application's UretprobePrograms.
	// +unionMember
	// +optional
	URetProbe *ClUprobeProgramInfoState `json:"uretprobe,omitempty"`

	// tracepoint defines the desired state of the application's TracepointPrograms.
	// +unionMember
	// +optional
	TracePoint *ClTracepointProgramInfoState `json:"tracepoint,omitempty"`
}

// BpfApplicationSpec defines the desired state of BpfApplication
type ClBpfApplicationStateSpec struct {
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
	Programs []ClBpfApplicationProgramState `json:"programs,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterBpfApplicationState contains the per-node state of a BpfApplication.
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=".spec.node"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterBpfApplicationState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClBpfApplicationStateSpec `json:"spec,omitempty"`
	Status BpfAppStatus              `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterBpfApplicationStateList contains a list of BpfApplicationState objects
type ClusterBpfApplicationStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBpfApplicationState `json:"items"`
}

func (an ClusterBpfApplicationState) GetName() string {
	return an.Name
}

func (an ClusterBpfApplicationState) GetUID() metav1types.UID {
	return an.UID
}

func (an ClusterBpfApplicationState) GetAnnotations() map[string]string {
	return an.Annotations
}

func (an ClusterBpfApplicationState) GetLabels() map[string]string {
	return an.Labels
}

func (an ClusterBpfApplicationState) GetStatus() *BpfAppStatus {
	return &an.Status
}

func (an ClusterBpfApplicationState) GetClientObject() client.Object {
	return &an
}

func (anl ClusterBpfApplicationStateList) GetItems() []ClusterBpfApplicationState {
	return anl.Items
}
