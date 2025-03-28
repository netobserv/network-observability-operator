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

// +kubebuilder:validation:Required
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InterfaceDiscovery struct {
	// interfaceAutoDiscovery when enabled, the agent process monitors the creation and deletion of interfaces,
	// automatically attaching eBPF hooks to newly discovered interfaces in both directions.
	//+kubebuilder:default:=false
	// +optional
	InterfaceAutoDiscovery *bool `json:"interfaceAutoDiscovery,omitempty"`

	// excludeInterfaces contains the interface names that are excluded from interface discovery
	// it is matched as a case-sensitive string.
	//+kubebuilder:default:={"lo"}
	//+optional
	ExcludeInterfaces []string `json:"excludeInterfaces,omitempty"`
}

// InterfaceSelector defines interface to attach to.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type InterfaceSelector struct {
	// discoveryConfig allow configuring interface discovery functionality,
	// +optional
	InterfacesDiscoveryConfig *InterfaceDiscovery `json:"interfacesDiscoveryConfig,omitempty"`

	// interfaces refers to a list of network interfaces to attach the BPF
	// program to.
	// +optional
	Interfaces []string `json:"interfaces,omitempty"`

	// primaryNodeInterface to attach BPF program to the primary interface on the node. Only 'true' accepted.
	// +optional
	PrimaryNodeInterface *bool `json:"primaryNodeInterface,omitempty"`
}

// ClContainerSelector identifies a set of containers. For example, this can be
// used to identify a set of containers in which to attach uprobes.
type ClContainerSelector struct {
	// namespaces indicate the target namespaces.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// pods indicate the target pods. This field must be specified, to select all pods use
	// standard metav1.LabelSelector semantics and make it empty.
	Pods metav1.LabelSelector `json:"pods"`

	// containerNames indicate the Name(s) of container(s).  If none are specified, all containers in the
	// pod are selected.
	// +optional
	ContainerNames []string `json:"containerNames,omitempty"`
}

// ContainerSelector identifies a set of containers. It is different from ContainerSelector
// in that "Namespace" was removed. Namespace scoped programs can only attach to the namespace
// they are created in, so namespace at this level doesn't apply.
type ContainerSelector struct {
	// pods indicate the target pods. This field must be specified, to select all pods use
	// standard metav1.LabelSelector semantics and make it empty.
	Pods metav1.LabelSelector `json:"pods"`

	// containerNames indicate the name(s) of container(s).  If none are specified, all containers in the
	// pod are selected.
	// +optional
	ContainerNames []string `json:"containerNames,omitempty"`
}

// ClNetworkNamespaceSelector identifies a network namespace for network-related
// program types in the cluster-scoped ClusterBpfApplication object.
type ClNetworkNamespaceSelector struct {
	// Target namespace.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Target pods. This field must be specified, to select all pods use
	// standard metav1.LabelSelector semantics and make it empty.
	Pods metav1.LabelSelector `json:"pods"`
}

// NetworkNamespaceSelector identifies a network namespace for network-related
// program types in the namespace-scoped BpfApplication object.
type NetworkNamespaceSelector struct {
	// Target pods. This field must be specified, to select all pods use
	// standard metav1.LabelSelector semantics and make it empty.
	Pods metav1.LabelSelector `json:"pods"`
}

// BpfAppCommon defines the common attributes for all BpfApp programs
type BpfAppCommon struct {
	// nodeSelector allows the user to specify which nodes to deploy the
	// bpf program to. This field must be specified, to select all nodes
	// use standard metav1.LabelSelector semantics and make it empty.
	NodeSelector metav1.LabelSelector `json:"nodeSelector"`

	// globalData allows the user to set global variables when the program is loaded
	// with an array of raw bytes. This is a very low level primitive. The caller
	// is responsible for formatting the byte string appropriately considering
	// such things as size, endianness, alignment and packing of data structures.
	// +optional
	GlobalData map[string][]byte `json:"globalData,omitempty"`

	// bytecode configures where the bpf program's bytecode should be loaded
	// from.
	ByteCode ByteCodeSelector `json:"byteCode"`

	// TODO: need to work out how MapOwnerSelector will work after load-attach-split
	// mapOwnerSelector is used to select the loaded eBPF program this eBPF program
	// will share a map with.
	// +optional
	MapOwnerSelector *metav1.LabelSelector `json:"mapOwnerSelector,omitempty"`
}

// BpfAppStatus reflects the status of a BpfApplication or BpfApplicationState object
type BpfAppStatus struct {
	// For a BpfApplication object, Conditions contains the global cluster state
	// for the object. For a BpfApplicationState object, Conditions contains the
	// state of the BpfApplication object on the given node.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// AttachInfoStateCommon reflects the status for one attach point for a given bpf
// application program
type AttachInfoStateCommon struct {
	// shouldAttach reflects whether the attachment should exist.
	ShouldAttach bool `json:"shouldAttach"`
	// uuid is an Unique identifier for the attach point assigned by bpfman agent.
	UUID string `json:"uuid"`
	// linkId is an identifier for the link assigned by bpfman. This field is
	// empty until the program is successfully attached and bpfman returns the
	// id.
	// +optional
	LinkId *uint32 `json:"linkId,omitempty"`
	// linkStatus reflects whether the attachment has been reconciled
	// successfully, and if not, why.
	LinkStatus LinkStatus `json:"linkStatus"`
}

type BpfProgramStateCommon struct {
	// name is the name of the function that is the entry point for the BPF
	// program
	Name string `json:"name"`
	// programLinkStatus reflects whether all links requested for the program
	// are in the correct state.
	ProgramLinkStatus ProgramLinkStatus `json:"programLinkStatus"`
	// programId is the id of the program in the kernel.  Not set until the
	// program is loaded.
	// +optional
	ProgramId *uint32 `json:"programId,omitempty"`
}

// PullPolicy describes a policy for if/when to pull a container image
// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
type PullPolicy string

const (
	// PullAlways means that bpfman always attempts to pull the latest bytecode image. Container will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that bpfman never pulls an image, but only uses a local image. Container will fail if the image isn't present
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that bpfman pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

// ByteCodeSelector defines the various ways to reference bpf bytecode objects.
type ByteCodeSelector struct {
	// image used to specify a bytecode container image.
	Image *ByteCodeImage `json:"image,omitempty"`

	// path is used to specify a bytecode object via filepath.
	Path *string `json:"path,omitempty"`
}

// ByteCodeImage defines how to specify a bytecode container image.
type ByteCodeImage struct {
	// url is a valid container image URL used to reference a remote bytecode image.
	Url string `json:"url"`

	// pullPolicy describes a policy for if/when to pull a bytecode image. Defaults to IfNotPresent.
	// +kubebuilder:default:=IfNotPresent
	// +optional
	ImagePullPolicy PullPolicy `json:"imagePullPolicy"`

	// imagePullSecret is the name of the secret bpfman should use to get remote image
	// repository secrets.
	// +optional
	ImagePullSecret *ImagePullSecretSelector `json:"imagePullSecret,omitempty"`
}

// ImagePullSecretSelector defines the name and namespace of an image pull secret.
type ImagePullSecretSelector struct {
	// name of the secret which contains the credentials to access the image repository.
	Name string `json:"name"`

	// namespace of the secret which contains the credentials to access the image repository.
	Namespace string `json:"namespace"`
}

// -----------------------------------------------------------------------------
// Status Conditions - BPF Programs
// -----------------------------------------------------------------------------

// BpfApplicationConditionType is a condition type to indicate the status of a BPF
// application at the cluster level.
type BpfApplicationConditionType string

const (
	// BpfAppCondPending indicates that bpfman has not yet completed reconciling
	// the Bpf Application on all nodes in the cluster.
	BpfAppCondPending BpfApplicationConditionType = "Pending"

	// BpfAppCondSuccess indicates that the BPF application has been
	// successfully loaded and attached as requested on all nodes in the
	// cluster.
	BpfAppCondSuccess BpfApplicationConditionType = "Success"

	// BpfAppCondError indicates that an error has occurred on one or more nodes
	// while attempting to apply the configuration described in the CRD.
	BpfAppCondError BpfApplicationConditionType = "Error"

	// BpfAppCondDeleteError indicates that the BPF Application was marked for
	// deletion, but deletion was unsuccessful on one or more nodes.
	BpfAppCondDeleteError BpfApplicationConditionType = "DeleteError"
)

// Condition is a helper method to promote any given BpfApplicationConditionType
// to a full metav1.Condition in an opinionated fashion.
//
// TODO: this was created in the early days to provide at least SOME status
// information to the user, but the hardcoded messages need to be replaced in
// the future with dynamic and situation-aware messages later.
//
// See: https://github.com/bpfman/bpfman/issues/430
func (b BpfApplicationConditionType) Condition(message string) metav1.Condition {
	cond := metav1.Condition{}

	switch b {
	case BpfAppCondPending:
		if len(message) == 0 {
			message = "Waiting for Bpf Application Object to be reconciled on all nodes"
		}
		condType := string(BpfAppCondPending)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Pending",
			Message: message,
		}
	case BpfAppCondError:
		if len(message) == 0 {
			message = "An error has occurred on one or more nodes"
		}
		condType := string(BpfAppCondError)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Error",
			Message: message,
		}
	case BpfAppCondSuccess:
		if len(message) == 0 {
			message = "BPF application configuration successfully applied on all nodes"
		}
		condType := string(BpfAppCondSuccess)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Success",
			Message: message,
		}
	case BpfAppCondDeleteError:
		if len(message) == 0 {
			message = "Deletion failed on one or more nodes"
		}
		condType := string(BpfAppCondDeleteError)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "DeleteError",
			Message: message,
		}
	}

	return cond
}

// BpfApplicationStateConditionType is used to indicate the status of a BPF
// application on a given node.
type BpfApplicationStateConditionType string

const (
	// BpfAppStateCondPending indicates that bpfman has not yet completed
	// reconciling the Bpf Application on the given node.
	BpfAppStateCondPending BpfApplicationStateConditionType = "Pending"

	// BpfAppStateCondSuccess indicates that the BPF application has been
	// successfully loaded and attached as requested on the given node.
	BpfAppStateCondSuccess BpfApplicationStateConditionType = "Success"

	// BpfAppStateCondError indicates that an error has occurred on the given
	// node while attempting to apply the configuration described in the CRD.
	BpfAppStateCondError BpfApplicationStateConditionType = "Error"

	// BpfAppStateCondError indicates that an error has occurred on the given
	// node while attempting to apply the configuration described in the CRD.
	BpfAppStateCondProgramListChangedError BpfApplicationStateConditionType = "ProgramListChangedError"

	// BpfAppStateCondDeleteError indicates that the BPF Application was marked
	// for deletion, but deletion was unsuccessful on the given node.
	BpfAppStateCondDeleteError BpfApplicationStateConditionType = "DeleteError"
)

// Condition is a helper method to promote any given
// BpfApplicationStateConditionType to a full metav1.Condition in an opinionated
// fashion.
func (b BpfApplicationStateConditionType) Condition() metav1.Condition {
	cond := metav1.Condition{}

	switch b {
	case BpfAppStateCondPending:
		condType := string(BpfAppStateCondPending)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Pending",
			Message: "Not yet complete",
		}
	case BpfAppStateCondSuccess:
		condType := string(BpfAppStateCondSuccess)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Success",
			Message: "The BPF application has been successfully loaded and attached",
		}
	case BpfAppStateCondError:
		condType := string(BpfAppStateCondError)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Error",
			Message: "An error has occurred",
		}
	case BpfAppStateCondDeleteError:
		condType := string(BpfAppStateCondDeleteError)
		cond = metav1.Condition{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "Delete Error",
			Message: "Deletion failed on one or more nodes",
		}
	}
	return cond
}

type AppLoadStatus string

const (
	// The initial load condition
	AppLoadNotLoaded AppLoadStatus = "NotLoaded"
	// All programs for app have been loaded
	AppLoadSuccess AppLoadStatus = "LoadSuccess"
	// One or more programs for app has not been loaded
	AppLoadError AppLoadStatus = "LoadError"
	// All programs for app have been unloaded
	AppUnLoadSuccess AppLoadStatus = "UnloadSuccess"
	// One or more programs for app has not been unloaded
	AppUnloadError AppLoadStatus = "UnloadError"
	// The app is not selected to run on the node
	NotSelected AppLoadStatus = "NotSelected"
	// The app is not selected to run on the node
	ProgListChangedError AppLoadStatus = "ProgramListChangedError"
)

type ProgramLinkStatus string

const (
	// The initial program attach state
	ProgAttachPending ProgramLinkStatus = "Pending"
	// All attachments for program are in the correct state
	ProgAttachSuccess ProgramLinkStatus = "Success"
	// One or more attachments for program are not in the correct state
	ProgAttachError ProgramLinkStatus = "Error"
	// There was an error updating the attach info
	UpdateAttachInfoError ProgramLinkStatus = "UpdateAttachInfoError"
)

type LinkStatus string

const (
	// Attach point is attached
	ApAttachAttached LinkStatus = "Attached"
	// Attach point is not attached
	ApAttachNotAttached LinkStatus = "NotAttached"
	// An attach was attempted, but there was an error
	ApAttachError LinkStatus = "AttachError"
	// A detach was attempted, but there was an error
	ApDetachError LinkStatus = "DetachError"
)
