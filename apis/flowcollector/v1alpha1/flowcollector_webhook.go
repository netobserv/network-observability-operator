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

	"github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	utilconversion "github.com/netobserv/network-observability-operator/pkg/conversion"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
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

	// Agent
	if restored.Spec.Agent.EBPF.Features != nil {
		dst.Spec.Agent.EBPF.Features = make([]v1beta2.AgentFeature, len(restored.Spec.Agent.EBPF.Features))
		copy(dst.Spec.Agent.EBPF.Features, restored.Spec.Agent.EBPF.Features)
	}

	// Processor
	dst.Spec.Processor.LogTypes = restored.Spec.Processor.LogTypes
	if restored.Spec.Processor.Advanced.ConversationHeartbeatInterval != nil {
		dst.Spec.Processor.Advanced.ConversationHeartbeatInterval = restored.Spec.Processor.Advanced.ConversationHeartbeatInterval
	}
	if restored.Spec.Processor.Advanced.ConversationEndTimeout != nil {
		dst.Spec.Processor.Advanced.ConversationEndTimeout = restored.Spec.Processor.Advanced.ConversationEndTimeout
	}
	if restored.Spec.Processor.Advanced.ConversationTerminatingTimeout != nil {
		dst.Spec.Processor.Advanced.ConversationTerminatingTimeout = restored.Spec.Processor.Advanced.ConversationTerminatingTimeout
	}
	if restored.Spec.Processor.Metrics.DisableAlerts != nil {
		dst.Spec.Processor.Metrics.DisableAlerts = restored.Spec.Processor.Metrics.DisableAlerts
	}
	if restored.Spec.Processor.ClusterName != "" {
		dst.Spec.Processor.ClusterName = restored.Spec.Processor.ClusterName
	}
	if restored.Spec.Processor.MultiClusterDeployment != nil {
		dst.Spec.Processor.MultiClusterDeployment = restored.Spec.Processor.MultiClusterDeployment
	}

	dst.Spec.Processor.Metrics.Server.TLS.InsecureSkipVerify = restored.Spec.Processor.Metrics.Server.TLS.InsecureSkipVerify
	dst.Spec.Processor.Metrics.Server.TLS.ProvidedCaFile = restored.Spec.Processor.Metrics.Server.TLS.ProvidedCaFile

	// Kafka
	dst.Spec.Kafka.SASL = restored.Spec.Kafka.SASL

	// Loki
	dst.Spec.Loki.Enable = restored.Spec.Loki.Enable

	if restored.Spec.Processor.Metrics.IncludeList != nil {
		list := make([]v1beta2.FLPMetric, len(*restored.Spec.Processor.Metrics.IncludeList))
		copy(list, *restored.Spec.Processor.Metrics.IncludeList)
		dst.Spec.Processor.Metrics.IncludeList = &list
	}

	return nil
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
	// Note that, despite we loose namespace info here, this isn't an issue because it's going to be restored from annotations
	manual := helper.NewLokiConfig(in, "")
	out.URL = manual.IngesterURL
	out.QuerierURL = manual.QuerierURL
	out.StatusURL = manual.StatusURL
	out.TenantID = manual.TenantID
	out.AuthToken = utilconversion.PascalToUpper(string(manual.AuthToken), '_')
	if err := Convert_v1beta2_ClientTLS_To_v1alpha1_ClientTLS(&manual.TLS, &out.TLS, nil); err != nil {
		return fmt.Errorf("copying v1beta2.Loki.TLS into v1alpha1.Loki.TLS: %w", err)
	}
	return autoConvert_v1beta2_FlowCollectorLoki_To_v1alpha1_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in *FlowCollectorLoki, out *v1beta2.FlowCollectorLoki, s apiconversion.Scope) error {
	out.Mode = v1beta2.LokiModeManual
	out.Manual = v1beta2.LokiManualParams{
		IngesterURL: in.URL,
		QuerierURL:  in.QuerierURL,
		StatusURL:   in.StatusURL,
		TenantID:    in.TenantID,
		AuthToken:   v1beta2.LokiAuthToken(utilconversion.UpperToPascal(in.AuthToken)),
	}
	// fallback on ingester url if querier is not set
	if len(out.Manual.QuerierURL) == 0 {
		out.Manual.QuerierURL = out.Manual.IngesterURL
	}
	if err := Convert_v1alpha1_ClientTLS_To_v1beta2_ClientTLS(&in.TLS, &out.Manual.TLS, nil); err != nil {
		return fmt.Errorf("copying v1alpha1.Loki.TLS into v1beta2.Loki.Manual.TLS: %w", err)
	}
	return autoConvert_v1alpha1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorConsolePlugin_To_v1alpha1_FlowCollectorConsolePlugin(in *v1beta2.FlowCollectorConsolePlugin, out *FlowCollectorConsolePlugin, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorConsolePlugin_To_v1alpha1_FlowCollectorConsolePlugin(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorEBPF_To_v1alpha1_FlowCollectorEBPF(in *v1beta2.FlowCollectorEBPF, out *FlowCollectorEBPF, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorEBPF_To_v1alpha1_FlowCollectorEBPF(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorSpec_To_v1beta2_FlowCollectorSpec(in *FlowCollectorSpec, out *v1beta2.FlowCollectorSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_FlowCollectorSpec_To_v1beta2_FlowCollectorSpec(in, out, s); err != nil {
		return err
	}
	out.DeploymentModel = v1beta2.FlowCollectorDeploymentModel(utilconversion.UpperToPascal(in.DeploymentModel))
	out.Exporters = []*v1beta2.FlowCollectorExporter{}
	for _, inExporter := range in.Exporters {
		outExporter := &v1beta2.FlowCollectorExporter{}
		if err := Convert_v1alpha1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(inExporter, outExporter, s); err != nil {
			return err
		}
		out.Exporters = append(out.Exporters, outExporter)
	}
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorSpec_To_v1alpha1_FlowCollectorSpec(in *v1beta2.FlowCollectorSpec, out *FlowCollectorSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorSpec_To_v1alpha1_FlowCollectorSpec(in, out, s); err != nil {
		return err
	}
	out.DeploymentModel = utilconversion.PascalToUpper(string(in.DeploymentModel), '_')
	out.Exporters = []*FlowCollectorExporter{}
	for _, inExporter := range in.Exporters {
		outExporter := &FlowCollectorExporter{}
		if err := Convert_v1beta2_FlowCollectorExporter_To_v1alpha1_FlowCollectorExporter(inExporter, outExporter, s); err != nil {
			return err
		}
		out.Exporters = append(out.Exporters, outExporter)
	}
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorAgent_To_v1beta2_FlowCollectorAgent(in *FlowCollectorAgent, out *v1beta2.FlowCollectorAgent, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_FlowCollectorAgent_To_v1beta2_FlowCollectorAgent(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.FlowCollectorAgentType(utilconversion.UpperToPascal(in.Type))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorAgent_To_v1alpha1_FlowCollectorAgent(in *v1beta2.FlowCollectorAgent, out *FlowCollectorAgent, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorAgent_To_v1alpha1_FlowCollectorAgent(in, out, s); err != nil {
		return err
	}
	out.Type = utilconversion.PascalToUpper(string(in.Type), '_')
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_ServerTLS_To_v1beta2_ServerTLS(in *ServerTLS, out *v1beta2.ServerTLS, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_ServerTLS_To_v1beta2_ServerTLS(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.ServerTLSConfigType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_ServerTLS_To_v1alpha1_ServerTLS(in *v1beta2.ServerTLS, out *ServerTLS, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_ServerTLS_To_v1alpha1_ServerTLS(in, out, s); err != nil {
		return err
	}
	out.Type = ServerTLSConfigType(utilconversion.PascalToUpper(string(in.Type), '_'))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorHPA_To_v1beta2_FlowCollectorHPA(in *FlowCollectorHPA, out *v1beta2.FlowCollectorHPA, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_FlowCollectorHPA_To_v1beta2_FlowCollectorHPA(in, out, s); err != nil {
		return err
	}
	out.Status = v1beta2.HPAStatus(utilconversion.UpperToPascal(in.Status))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorHPA_To_v1alpha1_FlowCollectorHPA(in *v1beta2.FlowCollectorHPA, out *FlowCollectorHPA, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorHPA_To_v1alpha1_FlowCollectorHPA(in, out, s); err != nil {
		return err
	}
	out.Status = utilconversion.PascalToUpper(string(in.Status), '_')
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_SASLConfig_To_v1beta2_SASLConfig(in *SASLConfig, out *v1beta2.SASLConfig, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_SASLConfig_To_v1beta2_SASLConfig(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.SASLType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_SASLConfig_To_v1alpha1_SASLConfig(in *v1beta2.SASLConfig, out *SASLConfig, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_SASLConfig_To_v1alpha1_SASLConfig(in, out, s); err != nil {
		return err
	}
	out.Type = SASLType(utilconversion.PascalToUpper(string(in.Type), '_'))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(in *FlowCollectorExporter, out *v1beta2.FlowCollectorExporter, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.ExporterType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorExporter_To_v1alpha1_FlowCollectorExporter(in *v1beta2.FlowCollectorExporter, out *FlowCollectorExporter, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorExporter_To_v1alpha1_FlowCollectorExporter(in, out, s); err != nil {
		return err
	}
	out.Type = ExporterType(utilconversion.PascalToUpper(string(in.Type), '_'))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FLPMetrics_To_v1beta2_FLPMetrics(in *FLPMetrics, out *v1beta2.FLPMetrics, s apiconversion.Scope) error {
	out.IncludeList = metrics.GetAsIncludeList(in.IgnoreTags, nil)
	return autoConvert_v1alpha1_FLPMetrics_To_v1beta2_FLPMetrics(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorConsolePlugin_To_v1beta2_FlowCollectorConsolePlugin(in *FlowCollectorConsolePlugin, out *v1beta2.FlowCollectorConsolePlugin, s apiconversion.Scope) error {
	return autoConvert_v1alpha1_FlowCollectorConsolePlugin_To_v1beta2_FlowCollectorConsolePlugin(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorFLP_To_v1beta2_FlowCollectorFLP(in *FlowCollectorFLP, out *v1beta2.FlowCollectorFLP, s apiconversion.Scope) error {
	return autoConvert_v1alpha1_FlowCollectorFLP_To_v1beta2_FlowCollectorFLP(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_DebugConfig_To_v1beta2_AdvancedAgentConfig(in *DebugConfig, out *v1beta2.AdvancedAgentConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_AdvancedAgentConfig_To_v1alpha1_DebugConfig(in *v1beta2.AdvancedAgentConfig, out *DebugConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_DebugConfig_To_v1beta2_AdvancedProcessorConfig(in *DebugConfig, out *v1beta2.AdvancedProcessorConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_AdvancedProcessorConfig_To_v1alpha1_DebugConfig(in *v1beta2.AdvancedProcessorConfig, out *DebugConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1alpha1
// nolint:golint,stylecheck,revive
func Convert_v1alpha1_FlowCollectorEBPF_To_v1beta2_FlowCollectorEBPF(in *FlowCollectorEBPF, out *v1beta2.FlowCollectorEBPF, s apiconversion.Scope) error {
	out.Advanced = &v1beta2.AdvancedAgentConfig{
		Env: in.Debug.Env,
	}
	return autoConvert_v1alpha1_FlowCollectorEBPF_To_v1beta2_FlowCollectorEBPF(in, out, s)
}
