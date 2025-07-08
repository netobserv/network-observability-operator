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

type ClFentryProgramInfo struct {
	ClFentryLoadInfo `json:",inline"`

	// links is an optional field and is a flag to indicate if the FEntry program
	// should be attached. The attachment point for a FEntry program is a Linux
	// kernel function. Unlike other eBPF program types, an FEntry program must be
	// provided with the target function at load time. The links field is optional,
	// but unlike other program types where it represents a list of attachment
	// points, for FEntry programs it contains at most one entry that determines
	// whether the program should be attached to the specified function. To attach
	// the program, add an entry to links with mode set to Attach. To detach it,
	// remove the entry from links.
	// +optional
	// +kubebuilder:validation:MaxItems=1
	Links []ClFentryAttachInfo `json:"links,omitempty"`
}

type ClFentryLoadInfo struct {
	// function is a required field and specifies the name of the Linux kernel
	// function to attach the FEntry program. function must not be an empty string,
	// must not exceed 64 characters in length, must start with alpha characters
	// and must only contain alphanumeric characters.
	// +required
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Function string `json:"function"`
}

type AttachTypeAttach string

const (
	Attach AttachTypeAttach = "Attach"
)

type ClFentryAttachInfo struct {
	// mode is a required field. When set to Attach, the FEntry program will
	// attempt to be attached. To detach the FEntry program, remove the link entry.
	// +required
	// +kubebuilder:validation:Enum=Attach;
	Mode AttachTypeAttach `json:"mode"`
}

type ClFentryProgramInfoState struct {
	ClFentryLoadInfo `json:",inline"`

	// links is a list of attachment points for the FEntry program. Each entry in
	// the list includes a linkStatus, which indicates if the attachment was
	// successful or not on this node, a linkId, which is the kernel ID for the
	// link if successfully attached, and other attachment specific data.
	// +optional
	// +kubebuilder:validation:MaxItems=1
	Links []ClFentryAttachInfoState `json:"links,omitempty"`
}

type ClFentryAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`
}
