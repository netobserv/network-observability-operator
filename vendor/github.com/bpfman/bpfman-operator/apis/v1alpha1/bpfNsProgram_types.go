/*
Copyright 2024.

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
// +kubebuilder:validation:Required
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BpfNsProgram is the Schema for the Bpfnsprograms API
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BpfNsProgram struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BpfProgramSpec `json:"spec"`
	// +optional
	Status BpfProgramStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BpfNsProgramList contains a list of BpfProgram
type BpfNsProgramList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BpfNsProgram `json:"items"`
}

func (bp BpfNsProgram) GetName() string {
	return bp.Name
}

func (bp BpfNsProgram) GetUID() metav1types.UID {
	return bp.UID
}

func (bp BpfNsProgram) GetAnnotations() map[string]string {
	return bp.Annotations
}

func (bp BpfNsProgram) GetLabels() map[string]string {
	return bp.Labels
}

func (bp BpfNsProgram) GetStatus() *BpfProgramStatus {
	return &bp.Status
}

func (bp BpfNsProgram) GetClientObject() client.Object {
	return &bp
}

func (bpl BpfNsProgramList) GetItems() []BpfNsProgram {
	return bpl.Items
}
