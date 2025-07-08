/*
Copyright 2022.

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

type ClTracepointProgramInfo struct {
	// links is an optional field and is the list of attachment points to which the
	// Tracepoint program should be attached. The Tracepoint program is loaded in
	// kernel memory when the BPF Application CRD is created and the selected
	// Kubernetes nodes are active. The Tracepoint program will not be triggered
	// until the program has also been attached to an attachment point described in
	// this list. Items may be added or removed from the list at any point, causing
	// the Tracepoint program to be attached or detached.
	//
	// The attachment point for a Tracepoint program is a one of a predefined set
	// of Linux kernel functions.
	// +optional
	Links []ClTracepointAttachInfo `json:"links,omitempty"`
}

type ClTracepointAttachInfo struct {
	// name is a required field and specifies the name of the Linux kernel
	// Tracepoint to attach the eBPF program. name must not be an empty string,
	// must not exceed 64 characters in length, must start with alpha characters
	// and must only contain alphanumeric characters.
	// +required
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Name string `json:"name"`
}

type ClTracepointProgramInfoState struct {
	// links is a list of attachment points for the Tracepoint program. Each entry
	// in the list includes a linkStatus, which indicates if the attachment was
	// successful or not on this node, a linkId, which is the kernel ID for the
	// link if successfully attached, and other attachment specific data.
	// +optional
	Links []ClTracepointAttachInfoState `json:"links,omitempty"`
}

type ClTracepointAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`

	// The name of a kernel tracepoint to attach the bpf program to.
	// name is the provisioned name of the Linux kernel tracepoint function the
	// Tracepoint program should be attached.
	// +required
	Name string `json:"name"`
}
