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

// ClFexitProgramInfo defines the Fexit program details
type ClFexitProgramInfo struct {
	ClFexitLoadInfo `json:",inline"`
	// Whether the program should be attached to the function.
	// +optional
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:default:={}
	Links []ClFexitAttachInfo `json:"links"`
}

// ClFexitLoadInfo contains the program-specific load information for Fexit
// programs
type ClFexitLoadInfo struct {
	// function is the name of the function to attach the Fexit program to.
	// +kubebuilder:validation:Pattern="^[a-zA-Z][a-zA-Z0-9_]+."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Function string `json:"function"`
}

// ClFexitAttachInfo indicates that the Fentry program should be attached to
// the function identified in ClFentryLoadInfo. The only valid value for Attach
// is true.
type ClFexitAttachInfo struct {
	// +kubebuilder:validation:Enum=Attach;Dettach;
	Mode AttachTypeAttach `json:"mode"`
}

type ClFexitProgramInfoState struct {
	ClFexitLoadInfo `json:",inline"`
	// +optional
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:default:={}
	Links []ClFexitAttachInfoState `json:"links"`
}

type ClFexitAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`
}
