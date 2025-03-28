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

// ClTracepointProgramInfo contains the Tracepoint program details
type ClTracepointProgramInfo struct {
	// links is the list of points to which the program should be attached.  The list items
	// are optional and may be updated after the bpf program has been loaded
	// +optional
	Links []ClTracepointAttachInfo `json:"links,omitempty"`
}

type ClTracepointAttachInfo struct {
	// name refers to the name of a kernel tracepoint to attach the
	// bpf program to.
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Name string `json:"name"`
}

type ClTracepointProgramInfoState struct {
	// links is the list of attach points for the BPF program on the given node. Each entry
	// in *AttachInfoState represents a specific, unique attach point that is
	// derived from *AttachInfo by fully expanding any selectors.  Each entry
	// also contains information about the attach point required by the
	// reconciler
	// +optional
	Links []ClTracepointAttachInfoState `json:"links,omitempty"`
}

type ClTracepointAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`

	// The name of a kernel tracepoint to attach the bpf program to.
	Name string `json:"name"`
}
