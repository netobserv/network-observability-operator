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

// All fields are required unless explicitly marked optional
package v1alpha1

// ClKprobeProgramInfo contains the information for the kprobe program
type ClKprobeProgramInfo struct {
	// The list of points to which the program should be attached.  The list items
	// are optional and may be udated after the bpf program has been loaded
	// +optional
	Links []ClKprobeAttachInfo `json:"links,omitempty"`
}

type ClKprobeAttachInfo struct {
	// function to attach the kprobe to.
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Function string `json:"function"`

	// offset added to the address of the function for kprobe.
	// The offset must be zero for kretprobes.
	// TODO: Add a webhook to enforce kretprobe offset=0.
	// See: https://github.com/bpfman/bpfman-operator/issues/403
	// +optional
	// +kubebuilder:default:=0
	Offset uint64 `json:"offset"`
}

type ClKprobeProgramInfoState struct {
	// List of attach points for the BPF program on the given node. Each entry
	// in *AttachInfoState represents a specific, unique attach point that is
	// derived from *AttachInfo by fully expanding any selectors.  Each entry
	// also contains information about the attach point required by the
	// reconciler
	// +optional
	Links []ClKprobeAttachInfoState `json:"links,omitempty"`
}

type ClKprobeAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`

	// Function to attach the kprobe to.
	Function string `json:"function"`

	// Offset added to the address of the function for kprobe.
	Offset uint64 `json:"offset"`
}
