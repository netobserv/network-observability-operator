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

// All fields are required unless explicitly marked optional
package v1alpha1

// UprobeProgramInfo contains the information for the uprobe program
type UprobeProgramInfo struct {
	// links is The list of points to which the program should be attached.  The list items
	// are optional and may be updated after the bpf program has been loaded
	// +optional
	Links []UprobeAttachInfo `json:"links,omitempty"`
}

type UprobeAttachInfo struct {
	// function to attach the uprobe to.
	// +optional
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Function string `json:"function,omitempty"`

	// offset added to the address of the function for uprobe.
	// +optional
	// +kubebuilder:default:=0
	Offset uint64 `json:"offset"`

	// target is the Library name or the absolute path to a binary or library.
	Target string `json:"target"`

	// pid is only execute uprobe for given process identification number (PID). If PID
	// is not provided, uprobe executes for all PIDs.
	// +optional
	Pid *int32 `json:"pid,omitempty"`

	// containers identify the set of containers in which to attach the
	// uprobe.
	Containers ContainerSelector `json:"containers"`
}

type UprobeProgramInfoState struct {
	// List of attach points for the BPF program on the given node. Each entry
	// in *AttachInfoState represents a specific, unique attach point that is
	// derived from *AttachInfo by fully expanding any selectors.  Each entry
	// also contains information about the attach point required by the
	// reconciler
	// +optional
	Links []UprobeAttachInfoState `json:"links,omitempty"`
}

type UprobeAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`

	// function to attach the uprobe to.
	// +optional
	Function string `json:"function,omitempty"`

	// offset added to the address of the function for uprobe.
	// +optional
	// +kubebuilder:default:=0
	Offset uint64 `json:"offset"`

	// target is the library name or the absolute path to a binary or library.
	Target string `json:"target"`

	// pid is Only execute uprobe for given process identification number (PID). If PID
	// is not provided, uprobe executes for all PIDs.
	// +optional
	Pid *int32 `json:"pid,omitempty"`

	// containerPid is container pid to attach the uprobe program in.
	// +optional
	ContainerPid int32 `json:"containerPid,omitempty"`
}
