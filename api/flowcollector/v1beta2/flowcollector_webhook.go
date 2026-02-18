/*
Copyright 2021.

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

package v1beta2

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-flows-netobserv-io-v1beta2-flowcollector,mutating=false,failurePolicy=fail,sideEffects=None,groups=flows.netobserv.io,resources=flowcollectors,versions=v1beta2,name=flowcollectorconversionwebhook.netobserv.io,admissionReviewVersions=v1
func (r *FlowCollector) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithValidator(r).
		Complete()
}

// Hub marks this version as a conversion hub.
// All the other version need to provide converters from/to this version.
// https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html
func (*FlowCollector) Hub()     {}
func (*FlowCollectorList) Hub() {}
