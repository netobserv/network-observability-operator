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

package v1alpha1

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	v1 "github.com/netobserv/network-observability-operator/api/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

//following values need to be synced with the corresponding
//.spec.agent.ebpf.sampling values of each version
const (
	DefaultV1AgentSampling       = 25
	DefaultV1Alpha1AgentSampling = 50
)

// log is for logging in this package.
var fclog = logf.Log.WithName("flowcollector-resource")

func (afc *FlowCollector) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(afc).
		Complete()
}

// ConvertTo converts this v1alpha1 FlowCollector to its v1 equivalent (the conversion Hub)
// https://book.kubebuilder.io/multiversion-tutorial/conversion.html
func (afc *FlowCollector) ConvertTo(dstRaw conversion.Hub) error {
	fclog.Info("converting v1alpha1.FlowCollector into v1.FlowCollector")
	dst := dstRaw.(*v1.FlowCollector)

	// The whole spec is so large that we adopt the following strategy to avoid
	// copying field by field:
	// 1. Marshall: this --> mapstructure --> dst
	// 2. Manually copy the fields that have changed
	err := mapstructure.Decode(afc, dst)
	if err != nil {
		return fmt.Errorf("copying v1alpha1.FlowCollector into v1.FlowCollector: %w", err)
	}
	// TODO: remove, as these are just example placeholders
	dst.Spec.Agent.EBPF.Debugging.Env = afc.Spec.Agent.EBPF.Debug.Env
	if dst.Spec.Agent.EBPF.Sampling == nil {
		// To decide: when converting from a default v1alpha1 into v1,
		// should we set the v1alpha1 defaults (e.g. sampling 50)
		// or the v1 defaults (sampling 25)?
		dst.Spec.Agent.EBPF.Sampling = pointer.Int32Ptr(DefaultV1AgentSampling)
	}
	return nil
}

func (afc *FlowCollector) ConvertFrom(srcRaw conversion.Hub) error {
	fclog.Info("converting v1.FlowCollector into v1alpha1.FlowCollector")
	src := srcRaw.(*v1.FlowCollector)

	// The whole spec is so large that we adopt the following strategy to avoid
	// copying field by field:
	// 1. Marshall: src --> mapstructure --> this
	// 2. Manually copy the fields that have changed
	err := mapstructure.Decode(src, afc)
	if err != nil {
		return fmt.Errorf("copying v1.FlowCollector into v1alpha1.FlowCollector: %w", err)
	}
	// TODO: remove, as these are just example placeholders
	afc.Spec.Agent.EBPF.Debug.Env = src.Spec.Agent.EBPF.Debugging.Env
	if afc.Spec.Agent.EBPF.Sampling == nil {
		// To decide: when converting from a default v1 into v1alpha1,
		// should we set the v1alpha1 defaults (e.g. sampling 50)
		// or the v1 defaults (sampling 25)?
		afc.Spec.Agent.EBPF.Sampling = pointer.Int32Ptr(DefaultV1Alpha1AgentSampling)
	}
	return nil
}

func (afc *FlowCollector) Default() {
	fclog.Info("defaulting v1alpha1.FlowCollector values")
	if afc.Spec.Agent.EBPF.Sampling == nil {
		afc.Spec.Agent.EBPF.Sampling = pointer.Int32Ptr(DefaultV1Alpha1AgentSampling)
	}
}
