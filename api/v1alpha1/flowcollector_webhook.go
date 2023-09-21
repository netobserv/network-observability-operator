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
	"reflect"

	"github.com/netobserv/network-observability-operator/api/v1beta2"
	utilconversion "github.com/netobserv/network-observability-operator/pkg/conversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1alpha1 FlowCollector to its v1beta2 equivalent (the conversion Hub)
// https://book.kubebuilder.io/multiversion-tutorial/conversion.html
func (r *FlowCollector) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1alpha1_FlowCollector_To_v1beta2_FlowCollector(r, dst, nil); err != nil {
		return fmt.Errorf("copying v1alpha1.FlowCollector into v1beta2.FlowCollector: %w", err)
	}
	dst.Status.Conditions = make([]v1.Condition, len(r.Status.Conditions))
	copy(dst.Status.Conditions, r.Status.Conditions)

	// Manually restore data.
	restored := &v1beta2.FlowCollector{}
	if ok, err := utilconversion.UnmarshalData(r, restored); err != nil || !ok {
		return err
	}

	dst.Spec.Processor.LogTypes = restored.Spec.Processor.LogTypes

	if restored.Spec.Processor.ConversationHeartbeatInterval != nil {
		dst.Spec.Processor.ConversationHeartbeatInterval = restored.Spec.Processor.ConversationHeartbeatInterval
	}

	if restored.Spec.Processor.ConversationEndTimeout != nil {
		dst.Spec.Processor.ConversationEndTimeout = restored.Spec.Processor.ConversationEndTimeout
	}

	if restored.Spec.Processor.ConversationTerminatingTimeout != nil {
		dst.Spec.Processor.ConversationTerminatingTimeout = restored.Spec.Processor.ConversationTerminatingTimeout
	}

	if restored.Spec.Processor.Metrics.DisableAlerts != nil {
		dst.Spec.Processor.Metrics.DisableAlerts = restored.Spec.Processor.Metrics.DisableAlerts
	}

	dst.Spec.Loki.Enable = restored.Spec.Loki.Enable

	if restored.Spec.Agent.EBPF.Features != nil {
		dst.Spec.Agent.EBPF.Features = make([]v1beta2.AgentFeature, len(restored.Spec.Agent.EBPF.Features))
		copy(dst.Spec.Agent.EBPF.Features, restored.Spec.Agent.EBPF.Features)
	}

	dst.Spec.Kafka.SASL = restored.Spec.Kafka.SASL

	// restore loki configuration from metadata
	if len(dst.Spec.Loki.Mode) > 0 {
		dst.Spec.Loki.Mode = restored.Spec.Loki.Mode
		dst.Spec.Loki.Manual = restored.Spec.Loki.Manual
		dst.Spec.Loki.LokiStack = restored.Spec.Loki.LokiStack
	} else {
		// fallback on previous Manual mode
		dst.Spec.Loki.Mode = v1beta2.LokiModeManual
		dst.Spec.Loki.Manual.IngesterURL = r.Spec.Loki.URL
		dst.Spec.Loki.Manual.QuerierURL = r.Spec.Loki.QuerierURL
		dst.Spec.Loki.Manual.StatusURL = r.Spec.Loki.StatusURL
		dst.Spec.Loki.Manual.TenantID = r.Spec.Loki.TenantID
		dst.Spec.Loki.Manual.AuthToken = r.Spec.Loki.AuthToken
		if err := Convert_v1alpha1_ClientTLS_To_v1beta2_ClientTLS(&r.Spec.Loki.TLS, &dst.Spec.Loki.Manual.TLS, nil); err != nil {
			return fmt.Errorf("copying v1alplha1.Loki.TLS into v1beta2.Loki.Manual.TLS: %w", err)
		}
	}
	dst.Spec.Processor.Metrics.Server.TLS.InsecureSkipVerify = restored.Spec.Processor.Metrics.Server.TLS.InsecureSkipVerify
	dst.Spec.Processor.Metrics.Server.TLS.ProvidedCaFile = restored.Spec.Processor.Metrics.Server.TLS.ProvidedCaFile

	if restored.Spec.Exporters != nil {
		for _, restoredExp := range restored.Spec.Exporters {
			if !isExporterIn(restoredExp, dst.Spec.Exporters) {
				dst.Spec.Exporters = append(dst.Spec.Exporters, restoredExp)
			}
		}
	}

	return nil
}

func isExporterIn(restoredExporter *v1beta2.FlowCollectorExporter, dstExporters []*v1beta2.FlowCollectorExporter) bool {

	for _, dstExp := range dstExporters {
		if reflect.DeepEqual(restoredExporter, dstExp) {
			return true
		}
	}
	return false
}

// ConvertFrom converts the hub version v1beta2 FlowCollector object to v1alpha1
func (r *FlowCollector) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1beta2_FlowCollector_To_v1alpha1_FlowCollector(src, r, nil); err != nil {
		return fmt.Errorf("copying v1beta2.FlowCollector into v1alpha1.FlowCollector: %w", err)
	}
	r.Status.Conditions = make([]v1.Condition, len(src.Status.Conditions))
	copy(r.Status.Conditions, src.Status.Conditions)

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, r)
}

func (r *FlowCollectorList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1alpha1_FlowCollectorList_To_v1beta2_FlowCollectorList(r, dst, nil)
}

func (r *FlowCollectorList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1beta2_FlowCollectorList_To_v1alpha1_FlowCollectorList(src, r, nil)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorFLP_To_v1alpha1_FlowCollectorFLP(in *v1beta2.FlowCollectorFLP, out *FlowCollectorFLP, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorFLP_To_v1alpha1_FlowCollectorFLP(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FLPMetrics_To_v1alpha1_FLPMetrics(in *v1beta2.FLPMetrics, out *FLPMetrics, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FLPMetrics_To_v1alpha1_FLPMetrics(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorLoki_To_v1alpha1_FlowCollectorLoki(in *v1beta2.FlowCollectorLoki, out *FlowCollectorLoki, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorLoki_To_v1alpha1_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in *FlowCollectorLoki, out *v1beta2.FlowCollectorLoki, s apiconversion.Scope) error {
	return autoConvert_v1alpha1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta1 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorConsolePlugin_To_v1alpha1_FlowCollectorConsolePlugin(in *v1beta2.FlowCollectorConsolePlugin, out *FlowCollectorConsolePlugin, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorConsolePlugin_To_v1alpha1_FlowCollectorConsolePlugin(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta1 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorExporter_To_v1alpha1_FlowCollectorExporter(in *v1beta2.FlowCollectorExporter, out *FlowCollectorExporter, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorExporter_To_v1alpha1_FlowCollectorExporter(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta1 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorEBPF_To_v1alpha1_FlowCollectorEBPF(in *v1beta2.FlowCollectorEBPF, out *FlowCollectorEBPF, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorEBPF_To_v1alpha1_FlowCollectorEBPF(in, out, s)
}

// // This function need to be manually created because conversion-gen not able to create it intentionally because
// // we have new defined fields in v1beta2 not in v1alpha1
// // nolint:golint,stylecheck,revive
func Convert_v1beta2_ServerTLS_To_v1alpha1_ServerTLS(in *v1beta2.ServerTLS, out *ServerTLS, s apiconversion.Scope) error {
	return autoConvert_v1beta2_ServerTLS_To_v1alpha1_ServerTLS(in, out, s)
}
