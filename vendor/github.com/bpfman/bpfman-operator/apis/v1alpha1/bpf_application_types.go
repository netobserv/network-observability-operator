/*
Copyright 2024 The bpfman Authors.

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

// BpfApplicationProgram defines the desired state of BpfApplication
// +union
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'XDP' ?  has(self.xdp) : !has(self.xdp)",message="xdp configuration is required when type is xdp, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TC' ?  has(self.tc) : !has(self.tc)",message="tc configuration is required when type is tc, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'TCX' ?  has(self.tcx) : !has(self.tcx)",message="tcx configuration is required when type is TCX, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'UProbe' ?  has(self.uprobe) : !has(self.uprobe)",message="uprobe configuration is required when type is uprobe, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'URetProbe' ?  has(self.uretprobe) : !has(self.uretprobe)",message="uretprobe configuration is required when type is uretprobe, and forbidden otherwise"
type BpfApplicationProgram struct {
	// name is a required field and is the name of the function that is the entry
	// point for the eBPF program. name must not be an empty string, must not
	// exceed 64 characters in length, must start with alpha characters and must
	// only contain alphanumeric characters.
	// +required
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Name string `json:"name"`

	// type is a required field used to specify the type of the eBPF program.
	//
	// Allowed values are:
	//   TC, TCX, UProbe, URetProbe, XDP
	//
	// When set to TC, the eBPF program can attach to network devices (interfaces).
	// The program can be attached on either packet ingress or egress, so the
	// program will be called on every incoming or outgoing packet seen by the
	// network device. When using the TC program type, the tc field is required.
	// See tc for more details on TC programs.
	//
	// When set to TCX, the eBPF program can attach to network devices
	// (interfaces). The program can be attached on either packet ingress or
	// egress, so the program will be called on every incoming or outgoing packet
	// seen by the network device. When using the TCX program type, the tcx field
	// is required. See tcx for more details on TCX programs.
	//
	// When set to UProbe, the program can attach in user-space. The UProbe is
	// attached to a binary, library or function name, and optionally an offset in
	// the code. When using the UProbe program type, the uprobe field is required.
	// See uprobe for more details on UProbe programs.
	//
	// When set to URetProbe, the program can attach in user-space.
	// The URetProbe is attached to the return of a binary, library or function
	// name, and optionally an offset in the code.  When using the URetProbe
	// program type, the uretprobe field is required. See uretprobe for more
	// details on URetProbe programs.
	//
	// When set to XDP, the eBPF program can attach to network devices (interfaces)
	// and will be called on every incoming packet received by the network device.
	// When using the XDP program type, the xdp field is required. See xdp for more
	// details on XDP programs.
	// +unionDiscriminator
	// +required
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"UProbe";"URetProbe"
	Type EBPFProgType `json:"type"`

	// xdp is an optional field, but required when the type field is set to XDP.
	// xdp defines the desired state of the application's XDP programs. XDP program
	// can be attached to network devices (interfaces) and will be called on every
	// incoming packet received by the network device. The XDP attachment point is
	// just after the packet has been received off the wire, but before the Linux
	// kernel has allocated an sk_buff, which is used to pass the packet through
	// the kernel networking stack.
	// +unionMember
	// +optional
	XDP *XdpProgramInfo `json:"xdp,omitempty"`

	// tc is an optional field, but required when the type field is set to TC. tc
	// defines the desired state of the application's TC programs. TC programs are
	// attached to network devices (interfaces). The program can be attached on
	// either packet ingress or egress, so the program will be called on every
	// incoming or outgoing packet seen by the network device. The TC attachment
	// point is in Linux's Traffic Control (tc) subsystem, which is after the
	// Linux kernel has allocated an sk_buff. TCX is newer implementation of TC
	// with enhanced performance and better support for running multiple programs
	// on a given network device. This makes TC useful for packet classification
	// actions.
	// +unionMember
	// +optional
	TC *TcProgramInfo `json:"tc,omitempty"`

	// tcx is an optional field, but required when the type field is set to TCX.
	// tcx defines the desired state of the application's TCX programs. TCX
	// programs are attached to network devices (interfaces). The program can be
	// attached on either packet ingress or egress, so the program will be called
	// on every incoming or outgoing packet seen by the network device. The TCX
	// attachment point is in Linux's Traffic Control (tc) subsystem, which is
	// after the Linux kernel has allocated an sk_buff. This makes TCX useful for
	// packet classification actions. TCX is a newer implementation of TC with
	// enhanced performance and better support for running multiple programs on a
	// given network device.
	// +unionMember
	// +optional
	TCX *TcxProgramInfo `json:"tcx,omitempty"`

	// uprobe is an optional field, but required when the type field is set to
	// UProbe. uprobe defines the desired state of the application's UProbe
	// programs. UProbe programs are user-space probes. A target must be provided,
	// which is the library name or absolute path to a binary or library where the
	// probe is attached. Optionally, a function name can also be provided to
	// provide finer granularity on where the probe is attached. They can be
	// attached at any point in the binary, library or function using the optional
	// offset field. However, caution must be taken when using the offset, ensuring
	// the offset is still in the desired bytecode.
	// +unionMember
	// +optional
	UProbe *UprobeProgramInfo `json:"uprobe,omitempty"`

	// uretprobe is an optional field, but required when the type field is set to
	// URetProbe. uretprobe defines the desired state of the application's
	// URetProbe programs. URetProbe programs are user-space probes. A target must
	// be provided, which is the library name or absolute path to a binary or
	// library where the probe is attached. Optionally, a function name can also be
	// provided to provide finer granularity on where the probe is attached. They
	// are attached to the return point of the binary, library or function, but can
	// be set anywhere using the optional offset field. However, caution must be
	// taken when using the offset, ensuring the offset is still in the desired
	// bytecode.
	// +unionMember
	// +optional
	URetProbe *UprobeProgramInfo `json:"uretprobe,omitempty"`
}

// spec defines the desired state of the BpfApplication. The BpfApplication
// describes the set of one or more namespace scoped eBPF programs that should
// be loaded for a given application and attributes for how they should be
// loaded. eBPF programs that are grouped together under the same
// BpfApplication instance can share maps and global data between the eBPF
// programs loaded on the same Kubernetes Node.
type BpfApplicationSpec struct {
	BpfAppCommon `json:",inline"`

	// programs is a required field and is the list of eBPF programs in a BPF
	// Application CRD that should be loaded in kernel memory. At least one entry
	// is required. eBPF programs in this list will be loaded on the system based
	// the nodeSelector. Even if an eBPF program is loaded in kernel memory, it
	// cannot be triggered until an attachment point is provided. The different
	// program types have different ways of attaching. The attachment points can be
	// added at creation time or modified (added or removed) at a later time to
	// activate or deactivate the eBPF program as desired.
	// CAUTION: When programs are added or removed from the list, that requires all
	// programs in the list to be reloaded, which could be temporarily service
	// effecting. For this reason, modifying the list is currently not allowed.
	// +required
	// +kubebuilder:validation:MinItems:=1
	Programs []BpfApplicationProgram `json:"programs,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// BpfApplication is the schema for the namespace scoped BPF Applications API.
// This API allows applications to use bpfman to load and attach one or more
// eBPF programs on a Kubernetes cluster.
//
// The bpfApplication.status field reports the overall status of the
// BpfApplication CRD. A given BpfApplication CRD can result in loading and
// attaching multiple eBPF programs on multiple nodes, so this status is just a
// summary. More granular per-node status details can be found in the
// corresponding BpfApplicationState CRD that bpfman creates for each node.
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=`.spec.nodeselector`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BpfApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BpfApplicationSpec `json:"spec,omitempty"`
	Status BpfAppStatus       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// BpfApplicationList contains a list of BpfApplications
type BpfApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BpfApplication `json:"items"`
}
