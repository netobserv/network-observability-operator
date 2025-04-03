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

// +kubebuilder:validation:Enum:=Aborted;Drop;Pass;TX;ReDirect;DispatcherReturn;
type XdpProceedOnValue string

// ClXdpProgramInfo contains the xdp program details
type ClXdpProgramInfo struct {
	// links is the list of points to which the program should be attached.  The list items
	// are optional and may be updated after the bpf program has been loaded
	// +optional
	Links []ClXdpAttachInfo `json:"links,omitempty"`
}

type ClXdpAttachInfo struct {
	// interfaceSelector to determine the network interface (or interfaces)
	InterfaceSelector InterfaceSelector `json:"interfaceSelector"`

	// networkNamespaces identifies the set of network namespaces in which to
	// attach the eBPF program. If networkNamespaces is not specified, the BPF
	// program will be attached in the root network namespace.
	// +optional
	NetworkNamespaces *ClNetworkNamespaceSelector `json:"networkNamespaces,omitempty"`

	// priority specifies the priority of the bpf program in relation to
	// other programs of the same type with the same attach point. It is a value
	// from 0 to 1000 where lower values have higher precedence.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	// +optional
	Priority int32 `json:"priority,omitempty"`

	// proceedOn allows the user to call other xdp programs in chain on this exit code.
	// Multiple values are supported by repeating the parameter.
	// +optional
	// +kubebuilder:default:={Pass,DispatcherReturn}
	ProceedOn []XdpProceedOnValue `json:"proceedOn,omitempty"`
}

type ClXdpProgramInfoState struct {
	// links is the list of attach points for the BPF program on the given node. Each entry
	// in *AttachInfoState represents a specific, unique attach point that is
	// derived from *AttachInfo by fully expanding any selectors.  Each entry
	// also contains information about the attach point required by the
	// reconciler
	// +optional
	Links []ClXdpAttachInfoState `json:"links,omitempty"`
}

type ClXdpAttachInfoState struct {
	AttachInfoStateCommon `json:",inline"`

	// interfaceName is the interface name to attach the xdp program to.
	InterfaceName string `json:"interfaceName"`

	// netnsPath is an optional path for a network namespace to attach the xdp
	// program in.
	// +optional
	NetnsPath string `json:"netnsPath,omitempty"`

	// priority specifies the priority of the xdp program in relation to
	// other programs of the same type with the same attach point. It is a value
	// from 0 to 1000 where lower values have higher precedence.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	Priority int32 `json:"priority"`

	// proceedOn allows the user to call other xdp programs in chain on this exit code.
	// Multiple values are supported by repeating the parameter.
	ProceedOn []XdpProceedOnValue `json:"proceedOn"`
}
