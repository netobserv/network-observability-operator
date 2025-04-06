/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespacedPeerApplyConfiguration represents an declarative configuration of the NamespacedPeer type for use
// with apply.
type NamespacedPeerApplyConfiguration struct {
	NamespaceSelector *v1.LabelSelector `json:"namespaceSelector,omitempty"`
	SameLabels        []string          `json:"sameLabels,omitempty"`
	NotSameLabels     []string          `json:"notSameLabels,omitempty"`
}

// NamespacedPeerApplyConfiguration constructs an declarative configuration of the NamespacedPeer type for use with
// apply.
func NamespacedPeer() *NamespacedPeerApplyConfiguration {
	return &NamespacedPeerApplyConfiguration{}
}

// WithNamespaceSelector sets the NamespaceSelector field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the NamespaceSelector field is set to the value of the last call.
func (b *NamespacedPeerApplyConfiguration) WithNamespaceSelector(value v1.LabelSelector) *NamespacedPeerApplyConfiguration {
	b.NamespaceSelector = &value
	return b
}

// WithSameLabels adds the given value to the SameLabels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the SameLabels field.
func (b *NamespacedPeerApplyConfiguration) WithSameLabels(values ...string) *NamespacedPeerApplyConfiguration {
	for i := range values {
		b.SameLabels = append(b.SameLabels, values[i])
	}
	return b
}

// WithNotSameLabels adds the given value to the NotSameLabels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the NotSameLabels field.
func (b *NamespacedPeerApplyConfiguration) WithNotSameLabels(values ...string) *NamespacedPeerApplyConfiguration {
	for i := range values {
		b.NotSameLabels = append(b.NotSameLabels, values[i])
	}
	return b
}
