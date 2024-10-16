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
	"context"
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netobserv/network-observability-operator/pkg/cluster"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-flows-netobserv-io-v1beta2-flowcollector,mutating=false,failurePolicy=fail,sideEffects=None,groups=flows.netobserv.io,resources=flowcollectors,versions=v1beta2,name=flowcollectorconversionwebhook.netobserv.io,admissionReviewVersions=v1
var (
	log                = logf.Log.WithName("flowcollector-resource")
	CurrentClusterInfo *cluster.Info
)

func (r *FlowCollector) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(r).
		Complete()
}

// Hub marks this version as a conversion hub.
// All the other version need to provide converters from/to this version.
// https://book.kubebuilder.io/multiversion-tutorial/conversion-concepts.html
func (*FlowCollector) Hub()     {}
func (*FlowCollectorList) Hub() {}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateCreate(ctx context.Context, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate create", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.validate(ctx, fc)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate update", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.validate(ctx, fc)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	log.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *FlowCollector) validate(_ context.Context, fc *FlowCollector) (admission.Warnings, error) {
	var warnings admission.Warnings
	if slices.Contains(fc.Spec.Agent.EBPF.Features, NetworkEvents) {
		// Make sure required version of ocp is installed
		if CurrentClusterInfo != nil && CurrentClusterInfo.IsOpenShift() {
			b, err := CurrentClusterInfo.OpenShiftVersionIsAtLeast("4.18.0")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Could not detect OpenShift cluster version: %s", err.Error()))
			} else if !b {
				warnings = append(warnings, fmt.Sprintf("The NetworkEvents feature requires OpenShift 4.18 or above (version detected: %s)", CurrentClusterInfo.GetOpenShiftVersion()))
			}
		} else {
			warnings = append(warnings, "The NetworkEvents feature is only supported with OpenShift")
		}
		if !fc.Spec.Agent.EBPF.Privileged {
			warnings = append(warnings, "The NetworkEvents feature requires eBPF Agent to run in privileged mode")
		}
	}
	if slices.Contains(fc.Spec.Agent.EBPF.Features, PacketDrop) && !fc.Spec.Agent.EBPF.Privileged {
		warnings = append(warnings, "The PacketDrop feature requires eBPF Agent to run in privileged mode")
	}
	return warnings, nil
}
