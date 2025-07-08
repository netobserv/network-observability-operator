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

// EBPFProgType defines the supported eBPF program types
type EBPFProgType string

const (
	// ProgTypeXDP refers to the XDP program type.
	ProgTypeXDP EBPFProgType = "XDP"

	// ProgTypeTC refers to the TC program type.
	ProgTypeTC EBPFProgType = "TC"

	// ProgTypeTCX refers to the TCX program type.
	ProgTypeTCX EBPFProgType = "TCX"

	// ProgTypeFentry refers to the Fentry program type.
	ProgTypeFentry EBPFProgType = "FEntry"

	// ProgTypeFexit refers to the Fexit program type.
	ProgTypeFexit EBPFProgType = "FExit"

	// ProgTypeKprobe refers to the Kprobe program type.
	ProgTypeKprobe EBPFProgType = "KProbe"

	// ProgTypeKretprobe refers to the Kretprobe program type.
	ProgTypeKretprobe EBPFProgType = "KRetProbe"

	// ProgTypeUprobe refers to the Uprobe program type.
	ProgTypeUprobe EBPFProgType = "UProbe"

	// ProgTypeUretprobe refers to the Uretprobe program type.
	ProgTypeUretprobe EBPFProgType = "URetProbe"

	// ProgTypeTracepoint refers to the Tracepoint program type.
	ProgTypeTracepoint EBPFProgType = "TracePoint"
)

type TCDirectionType string

const (
	TCIngress TCDirectionType = "Ingress"
	TCEgress  TCDirectionType = "Egress"
)

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
type ClBpfApplicationProgram struct {
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
	//   FEntry, FExit, KProbe, KRetProbe, TC, TCX, TracePoint, UProbe, URetProbe,
	//   XDP
	//
	// When set to FEntry, the program is attached to the entry of a Linux kernel
	// function or to another eBPF program function. When using the FEntry program
	// type, the fentry field is required. See fentry for more details on FEntry
	// programs.
	//
	// When set to FExit, the program is attached to the exit of a Linux kernel
	// function or to another eBPF program function. When using the FExit program
	// type, the fexit field is required. See fexit for more details on FExit
	// programs.
	//
	// When set to KProbe, the program is attached to entry of a Linux kernel
	// function. When using the KProbe program type, the kprobe field is required.
	// See kprobe for more details on KProbe programs.
	//
	// When set to KRetProbe, the program is attached to exit of a Linux kernel
	// function. When using the KRetProbe program type, the kretprobe field is
	// required. See kretprobe for more details on KRetProbe programs.
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
	// When set to Tracepoint, the program can attach to one of the predefined set
	// of Linux kernel functions. When using the Tracepoint program type, the
	// tracepoint field is required. See tracepoint for more details on Tracepoint
	// programs.
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
	// +kubebuilder:validation:Enum:="XDP";"TC";"TCX";"FEntry";"FExit";"KProbe";"KRetProbe";"UProbe";"URetProbe";"TracePoint"
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
	XDP *ClXdpProgramInfo `json:"xdp,omitempty"`

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
	TC *ClTcProgramInfo `json:"tc,omitempty"`

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
	TCX *ClTcxProgramInfo `json:"tcx,omitempty"`

	// fentry is an optional field, but required when the type field is set to
	// FEntry. fentry defines the desired state of the application's FEntry
	// programs. FEntry programs are attached to the entry of a Linux kernel
	// function or to another eBPF program function. They are attached to the first
	// instruction, before control passes to the function. FEntry programs are
	// similar to KProbe programs, but have higher performance.
	// +unionMember
	// +optional
	FEntry *ClFentryProgramInfo `json:"fentry,omitempty"`

	// fexit is an optional field, but required when the type field is set to
	// FExit. fexit defines the desired state of the application's FExit programs.
	// FExit programs are attached to the exit of a Linux kernel function or to
	// another eBPF program function. The program is invoked when the function
	// returns, independent of where in the function that occurs. FExit programs
	// are similar to KRetProbe programs, but get invoked with the input arguments
	// and the return values. They also have higher performance over KRetProbe
	// programs.
	// +unionMember
	// +optional
	FExit *ClFexitProgramInfo `json:"fexit,omitempty"`

	// kprobe is an optional field, but required when the type field is set to
	// KProbe. kprobe defines the desired state of the application's Kprobe
	// programs. KProbe programs are attached to a Linux kernel function. Unlike
	// FEntry programs, which must always be attached at the entry point of a Linux
	// kernel function, KProbe programs can be attached at any point in the
	// function using the optional offset field. However, caution must be taken
	// when using the offset, ensuring the offset is still in the function
	// bytecode. FEntry programs have less overhead than KProbe programs.
	// +unionMember
	// +optional
	KProbe *ClKprobeProgramInfo `json:"kprobe,omitempty"`

	// kretprobe is an optional field, but required when the type field is set to
	// KRetProbe. kretprobe defines the desired state of the application's
	// KRetProbe programs. KRetProbe programs are attached to the exit of a Linux
	// kernel function. FExit programs have less overhead than KRetProbe programs
	// and FExit programs have access to both the input arguments as well as the
	// return values. KRetProbes only have access to the return values.
	// +unionMember
	// +optional
	KRetProbe *ClKretprobeProgramInfo `json:"kretprobe,omitempty"`

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
	UProbe *ClUprobeProgramInfo `json:"uprobe,omitempty"`

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
	URetProbe *ClUprobeProgramInfo `json:"uretprobe,omitempty"`

	// tracepoint is an optional field, but required when the type field is set to
	// Tracepoint. tracepoint defines the desired state of the application's
	// Tracepoint programs. Whereas KProbes attach to dynamically to any Linux
	// kernel function, Tracepoint programs are programs that can only be attached
	// at predefined locations in the Linux kernel. Use the following command to
	// see the available attachment points:
	//  `sudo find /sys/kernel/debug/tracing/events -type d`
	// While KProbes are more flexible in where in the kernel the probe can be
	// attached, the functions and data structure rely on the kernel your system is
	// running. Tracepoints tend to be more stable across kernel versions and are
	// better for portability.
	// +unionMember
	// +optional
	TracePoint *ClTracepointProgramInfo `json:"tracepoint,omitempty"`
}

// spec defines the desired state of the ClusterBpfApplication. The
// ClusterBpfApplication describes the set of one or more cluster scoped eBPF
// programs that should be loaded for a given application and attributes for
// how they should be loaded. eBPF programs that are grouped together under the
// same ClusterBpfApplication instance can share maps and global data between
// the eBPF programs loaded on the same Kubernetes Node.
type ClBpfApplicationSpec struct {
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
	Programs []ClBpfApplicationProgram `json:"programs"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterBpfApplication is the schema for the cluster scoped BPF Applications
// API. This API allows applications to use bpfman to load and attach one or
// more eBPF programs on a Kubernetes cluster.
//
// The clusterBpfApplication.status field reports the overall status of the
// ClusterBpfApplication CRD. A given ClusterBpfApplication CRD can result in
// loading and attaching multiple eBPF programs on multiple nodes, so this
// status is just a summary. More granular per-node status details can be
// found in the corresponding ClusterBpfApplicationState CRD that bpfman
// creates for each node.
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=`.spec.nodeselector`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterBpfApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClBpfApplicationSpec `json:"spec,omitempty"`
	Status BpfAppStatus         `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterBpfApplicationList contains a list of BpfApplications
type ClusterBpfApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBpfApplication `json:"items"`
}
