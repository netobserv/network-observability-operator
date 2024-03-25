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

package v1beta1

import (
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	utilconversion "github.com/netobserv/network-observability-operator/pkg/conversion"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 FlowCollector to its v1beta2 equivalent (the conversion Hub)
// https://book.kubebuilder.io/multiversion-tutorial/conversion.html
func (r *FlowCollector) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1beta1_FlowCollector_To_v1beta2_FlowCollector(r, dst, nil); err != nil {
		return fmt.Errorf("copying v1beta1.FlowCollector into v1beta2.FlowCollector: %w", err)
	}
	dst.Status.Conditions = make([]v1.Condition, len(r.Status.Conditions))
	copy(dst.Status.Conditions, r.Status.Conditions)

	// Manually restore data.
	restored := &v1beta2.FlowCollector{}
	if ok, err := utilconversion.UnmarshalData(r, restored); err != nil || !ok {
		return err
	}
	// Restore elements that can't be guessed from v1beta1
	dst.Spec.Loki.Mode = restored.Spec.Loki.Mode
	dst.Spec.Loki.LokiStack = restored.Spec.Loki.LokiStack
	dst.Spec.Loki.Monolithic = restored.Spec.Loki.Monolithic
	dst.Spec.Loki.Microservices = restored.Spec.Loki.Microservices
	dst.Spec.Loki.Manual = restored.Spec.Loki.Manual

	if restored.Spec.Agent.EBPF.Advanced != nil {
		if dst.Spec.Agent.EBPF.Advanced == nil {
			dst.Spec.Agent.EBPF.Advanced = &v1beta2.AdvancedAgentConfig{}
		}
		dst.Spec.Agent.EBPF.Advanced.NodeSelector = restored.Spec.Agent.EBPF.Advanced.NodeSelector
		dst.Spec.Agent.EBPF.Advanced.Affinity = restored.Spec.Agent.EBPF.Advanced.Affinity
		dst.Spec.Agent.EBPF.Advanced.PriorityClassName = restored.Spec.Agent.EBPF.Advanced.PriorityClassName
	}
	if restored.Spec.Processor.Advanced != nil {
		if dst.Spec.Processor.Advanced == nil {
			dst.Spec.Processor.Advanced = &v1beta2.AdvancedProcessorConfig{}
		}
		dst.Spec.Processor.Advanced.NodeSelector = restored.Spec.Processor.Advanced.NodeSelector
		dst.Spec.Processor.Advanced.Affinity = restored.Spec.Processor.Advanced.Affinity
		dst.Spec.Processor.Advanced.PriorityClassName = restored.Spec.Processor.Advanced.PriorityClassName
	}
	if restored.Spec.ConsolePlugin.Advanced != nil {
		if dst.Spec.ConsolePlugin.Advanced == nil {
			dst.Spec.ConsolePlugin.Advanced = &v1beta2.AdvancedPluginConfig{}
		}
		dst.Spec.ConsolePlugin.Advanced.NodeSelector = restored.Spec.ConsolePlugin.Advanced.NodeSelector
		dst.Spec.ConsolePlugin.Advanced.Affinity = restored.Spec.ConsolePlugin.Advanced.Affinity
		dst.Spec.ConsolePlugin.Advanced.PriorityClassName = restored.Spec.ConsolePlugin.Advanced.PriorityClassName
	}
	ClearDefaultAdvancedConfig(dst)

	return nil
}

// ConvertFrom converts the hub version v1beta2 FlowCollector object to v1beta1
func (r *FlowCollector) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1beta2_FlowCollector_To_v1beta1_FlowCollector(src, r, nil); err != nil {
		return fmt.Errorf("copying v1beta2.FlowCollector into v1beta1.FlowCollector: %w", err)
	}
	r.Status.Conditions = make([]v1.Condition, len(src.Status.Conditions))
	copy(r.Status.Conditions, src.Status.Conditions)

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, r)
}

func (r *FlowCollectorList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1beta1_FlowCollectorList_To_v1beta2_FlowCollectorList(r, dst, nil)
}

func (r *FlowCollectorList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1beta2_FlowCollectorList_To_v1beta1_FlowCollectorList(src, r, nil)
}

func ClearDefaultAdvancedConfig(fc *v1beta2.FlowCollector) {
	// clear Processor advanced config if default
	if reflect.DeepEqual(helper.GetAdvancedProcessorConfig(nil), helper.GetAdvancedProcessorConfig(fc.Spec.Processor.Advanced)) {
		fc.Spec.Processor.Advanced = nil
	}
	// clear Agent advanced config if default
	if reflect.DeepEqual(helper.GetAdvancedAgentConfig(nil), helper.GetAdvancedAgentConfig(fc.Spec.Agent.EBPF.Advanced)) {
		fc.Spec.Agent.EBPF.Advanced = nil
	}
	// clear Plugin advanced config if default
	if reflect.DeepEqual(helper.GetAdvancedPluginConfig(nil), helper.GetAdvancedPluginConfig(fc.Spec.ConsolePlugin.Advanced)) {
		fc.Spec.ConsolePlugin.Advanced = nil
	}
	// clear Loki advanced config if default
	if reflect.DeepEqual(helper.GetAdvancedLokiConfig(nil), helper.GetAdvancedLokiConfig(fc.Spec.Loki.Advanced)) {
		fc.Spec.Loki.Advanced = nil
	}
}

// This function need to be manually created because we moved fields between v1beta2 and v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollector_To_v1beta2_FlowCollector(in *FlowCollector, out *v1beta2.FlowCollector, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollector_To_v1beta2_FlowCollector(in, out, s); err != nil {
		return fmt.Errorf("auto convert FlowCollector v1beta1 to v1beta2: %w", err)
	}
	ClearDefaultAdvancedConfig(out)
	return nil
}

// This function need to be manually created because we moved fields between v1beta2 and v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollector_To_v1beta1_FlowCollector(in *v1beta2.FlowCollector, out *FlowCollector, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollector_To_v1beta1_FlowCollector(in, out, s); err != nil {
		return fmt.Errorf("auto convert FlowCollector v1beta1 to v1beta2: %w", err)
	}
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in *v1beta2.FlowCollectorLoki, out *FlowCollectorLoki, s apiconversion.Scope) error {
	// Note that, despite we loose namespace info here, this isn't an issue because it's going to be restored from annotations
	manual := helper.NewLokiConfig(in, "")
	out.URL = manual.IngesterURL
	out.QuerierURL = manual.QuerierURL
	out.StatusURL = manual.StatusURL
	out.TenantID = manual.TenantID
	out.AuthToken = utilconversion.PascalToUpper(string(manual.AuthToken), '_')
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(&manual.TLS, &out.TLS, nil); err != nil {
		return fmt.Errorf("copying Loki v1beta2 TLS into v1beta1 TLS: %w", err)
	}
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(&manual.StatusTLS, &out.StatusTLS, nil); err != nil {
		return fmt.Errorf("copying Loki v1beta2 StatusTLS into v1beta1 StatusTLS: %w", err)
	}
	out.Timeout = in.WriteTimeout
	out.BatchWait = in.WriteBatchWait
	out.BatchSize = in.WriteBatchSize
	if in.Advanced != nil {
		out.MinBackoff = in.Advanced.WriteMinBackoff
		out.MaxBackoff = in.Advanced.WriteMaxBackoff
		out.MaxRetries = in.Advanced.WriteMaxRetries
		if in.Advanced.StaticLabels != nil {
			out.StaticLabels = in.Advanced.StaticLabels
		}
	}
	return autoConvert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in *FlowCollectorLoki, out *v1beta2.FlowCollectorLoki, s apiconversion.Scope) error {
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
	if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&in.TLS, &out.Manual.TLS, nil); err != nil {
		return fmt.Errorf("copying v1beta1.Loki.TLS into v1beta2.Loki.Manual.TLS: %w", err)
	}
	if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&in.StatusTLS, &out.Manual.StatusTLS, nil); err != nil {
		return fmt.Errorf("copying v1beta1.Loki.StatusTLS into v1beta2.Loki.Manual.StatusTLS: %w", err)
	}
	out.WriteTimeout = in.Timeout
	out.WriteBatchWait = in.BatchWait
	out.WriteBatchSize = in.BatchSize

	debugPath := helper.LokiAdvancedPath
	out.Advanced = &v1beta2.AdvancedLokiConfig{
		WriteMinBackoff: helper.GetAdvancedDurationValue(debugPath, "writeMinBackoff", in.MinBackoff),
		WriteMaxBackoff: helper.GetAdvancedDurationValue(debugPath, "writeMaxBackoff", in.MaxBackoff),
		WriteMaxRetries: helper.GetAdvancedInt32Value(debugPath, "writeMaxRetries", in.MaxRetries),
		StaticLabels:    helper.GetAdvancedMapValue(debugPath, "staticLabels", in.StaticLabels),
	}
	return autoConvert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorConsolePlugin_To_v1beta2_FlowCollectorConsolePlugin(in *FlowCollectorConsolePlugin, out *v1beta2.FlowCollectorConsolePlugin, s apiconversion.Scope) error {
	debugPath := helper.PluginAdvancedPath
	out.Advanced = &v1beta2.AdvancedPluginConfig{
		Env:      map[string]string{},
		Args:     []string{},
		Register: helper.GetAdvancedBoolValue(debugPath, "register", in.Register),
		Port:     helper.GetAdvancedInt32Value(debugPath, "port", &in.Port),
	}
	return autoConvert_v1beta1_FlowCollectorConsolePlugin_To_v1beta2_FlowCollectorConsolePlugin(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FLPMetrics_To_v1beta2_FLPMetrics(in *FLPMetrics, out *v1beta2.FLPMetrics, s apiconversion.Scope) error {
	err := autoConvert_v1beta1_FLPMetrics_To_v1beta2_FLPMetrics(in, out, s)
	if err != nil {
		return err
	}
	out.IncludeList = metrics.GetAsIncludeList(in.IgnoreTags, out.IncludeList)
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorSpec_To_v1beta2_FlowCollectorSpec(in *FlowCollectorSpec, out *v1beta2.FlowCollectorSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollectorSpec_To_v1beta2_FlowCollectorSpec(in, out, s); err != nil {
		return err
	}
	out.DeploymentModel = v1beta2.FlowCollectorDeploymentModel(utilconversion.UpperToPascal(in.DeploymentModel))
	out.Exporters = []*v1beta2.FlowCollectorExporter{}
	for _, inExporter := range in.Exporters {
		outExporter := &v1beta2.FlowCollectorExporter{}
		if err := Convert_v1beta1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(inExporter, outExporter, s); err != nil {
			return err
		}
		out.Exporters = append(out.Exporters, outExporter)
	}
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FLPMetrics_To_v1beta1_FLPMetrics(in *v1beta2.FLPMetrics, out *FLPMetrics, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FLPMetrics_To_v1beta1_FLPMetrics(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorSpec_To_v1beta1_FlowCollectorSpec(in *v1beta2.FlowCollectorSpec, out *FlowCollectorSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorSpec_To_v1beta1_FlowCollectorSpec(in, out, s); err != nil {
		return err
	}
	out.DeploymentModel = utilconversion.PascalToUpper(string(in.DeploymentModel), '_')
	out.Exporters = []*FlowCollectorExporter{}
	for _, inExporter := range in.Exporters {
		outExporter := &FlowCollectorExporter{}
		if err := Convert_v1beta2_FlowCollectorExporter_To_v1beta1_FlowCollectorExporter(inExporter, outExporter, s); err != nil {
			return err
		}
		out.Exporters = append(out.Exporters, outExporter)
	}
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorAgent_To_v1beta2_FlowCollectorAgent(in *FlowCollectorAgent, out *v1beta2.FlowCollectorAgent, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollectorAgent_To_v1beta2_FlowCollectorAgent(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.FlowCollectorAgentType(utilconversion.UpperToPascal(in.Type))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorAgent_To_v1beta1_FlowCollectorAgent(in *v1beta2.FlowCollectorAgent, out *FlowCollectorAgent, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorAgent_To_v1beta1_FlowCollectorAgent(in, out, s); err != nil {
		return err
	}
	out.Type = utilconversion.PascalToUpper(string(in.Type), '_')
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// and new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorFLP_To_v1beta2_FlowCollectorFLP(in *FlowCollectorFLP, out *v1beta2.FlowCollectorFLP, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollectorFLP_To_v1beta2_FlowCollectorFLP(in, out, s); err != nil {
		return err
	}
	if in.LogTypes != nil {
		logTypes := v1beta2.FLPLogTypes(utilconversion.UpperToPascal(*in.LogTypes))
		out.LogTypes = &logTypes
	}
	debugPath := helper.ProcessorAdvancedPath
	out.Advanced = &v1beta2.AdvancedProcessorConfig{
		Env:                            map[string]string{},
		Port:                           helper.GetAdvancedInt32Value(debugPath, "port", &in.Port),
		HealthPort:                     helper.GetAdvancedInt32Value(debugPath, "healthPort", &in.HealthPort),
		ProfilePort:                    helper.GetAdvancedInt32Value(debugPath, "profilePort", &in.ProfilePort),
		EnableKubeProbes:               helper.GetAdvancedBoolValue(debugPath, "enableKubeProbes", in.EnableKubeProbes),
		DropUnusedFields:               helper.GetAdvancedBoolValue(debugPath, "dropUnusedFields", in.DropUnusedFields),
		ConversationHeartbeatInterval:  helper.GetAdvancedDurationValue(debugPath, "conversationHeartbeatInterval", in.ConversationHeartbeatInterval),
		ConversationEndTimeout:         helper.GetAdvancedDurationValue(debugPath, "conversationEndTimeout", in.ConversationEndTimeout),
		ConversationTerminatingTimeout: helper.GetAdvancedDurationValue(debugPath, "conversationTerminatingTimeout", in.ConversationTerminatingTimeout),
	}
	return nil
}

// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorConsolePlugin_To_v1beta1_FlowCollectorConsolePlugin(in *v1beta2.FlowCollectorConsolePlugin, out *FlowCollectorConsolePlugin, s apiconversion.Scope) error {
	if in.Advanced != nil {
		out.Register = in.Advanced.Register
		out.Port = helper.GetValueOrDefaultInt32(helper.PluginAdvancedPath, "port", in.Advanced.Port)
	}
	return autoConvert_v1beta2_FlowCollectorConsolePlugin_To_v1beta1_FlowCollectorConsolePlugin(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_DebugConfig_To_v1beta2_AdvancedAgentConfig(in *DebugConfig, out *v1beta2.AdvancedAgentConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// and new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorFLP_To_v1beta1_FlowCollectorFLP(in *v1beta2.FlowCollectorFLP, out *FlowCollectorFLP, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorFLP_To_v1beta1_FlowCollectorFLP(in, out, s); err != nil {
		return err
	}
	if in.LogTypes != nil {
		str := utilconversion.PascalToUpper(string(*in.LogTypes), '_')
		out.LogTypes = &str
	}
	if in.Advanced != nil {
		debugPath := helper.ProcessorAdvancedPath
		out.Port = helper.GetValueOrDefaultInt32(debugPath, "port", in.Advanced.Port)
		out.HealthPort = helper.GetValueOrDefaultInt32(debugPath, "healthPort", in.Advanced.HealthPort)
		out.ProfilePort = helper.GetValueOrDefaultInt32(debugPath, "profilePort", in.Advanced.ProfilePort)
		out.EnableKubeProbes = in.Advanced.EnableKubeProbes
		out.DropUnusedFields = in.Advanced.DropUnusedFields
		out.ConversationHeartbeatInterval = in.Advanced.ConversationHeartbeatInterval
		out.ConversationEndTimeout = in.Advanced.ConversationEndTimeout
		out.ConversationTerminatingTimeout = in.Advanced.ConversationTerminatingTimeout
	}
	return nil
}

// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_AdvancedAgentConfig_To_v1beta1_DebugConfig(in *v1beta2.AdvancedAgentConfig, out *DebugConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_ServerTLS_To_v1beta2_ServerTLS(in *ServerTLS, out *v1beta2.ServerTLS, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_ServerTLS_To_v1beta2_ServerTLS(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.ServerTLSConfigType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_DebugConfig_To_v1beta2_AdvancedProcessorConfig(in *DebugConfig, out *v1beta2.AdvancedProcessorConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_ServerTLS_To_v1beta1_ServerTLS(in *v1beta2.ServerTLS, out *ServerTLS, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_ServerTLS_To_v1beta1_ServerTLS(in, out, s); err != nil {
		return err
	}
	out.Type = ServerTLSConfigType(utilconversion.PascalToUpper(string(in.Type), '_'))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorHPA_To_v1beta2_FlowCollectorHPA(in *FlowCollectorHPA, out *v1beta2.FlowCollectorHPA, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollectorHPA_To_v1beta2_FlowCollectorHPA(in, out, s); err != nil {
		return err
	}
	out.Status = v1beta2.HPAStatus(utilconversion.UpperToPascal(in.Status))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorHPA_To_v1beta1_FlowCollectorHPA(in *v1beta2.FlowCollectorHPA, out *FlowCollectorHPA, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorHPA_To_v1beta1_FlowCollectorHPA(in, out, s); err != nil {
		return err
	}
	out.Status = utilconversion.PascalToUpper(string(in.Status), '_')
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_SASLConfig_To_v1beta2_SASLConfig(in *SASLConfig, out *v1beta2.SASLConfig, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_SASLConfig_To_v1beta2_SASLConfig(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.SASLType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_SASLConfig_To_v1beta1_SASLConfig(in *v1beta2.SASLConfig, out *SASLConfig, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_SASLConfig_To_v1beta1_SASLConfig(in, out, s); err != nil {
		return err
	}
	out.Type = SASLType(utilconversion.PascalToUpper(string(in.Type), '-'))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(in *FlowCollectorExporter, out *v1beta2.FlowCollectorExporter, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_FlowCollectorExporter_To_v1beta2_FlowCollectorExporter(in, out, s); err != nil {
		return err
	}
	out.Type = v1beta2.ExporterType(utilconversion.UpperToPascal(string(in.Type)))
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have camel case enum in v1beta2 which were uppercase in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorExporter_To_v1beta1_FlowCollectorExporter(in *v1beta2.FlowCollectorExporter, out *FlowCollectorExporter, s apiconversion.Scope) error {
	if err := autoConvert_v1beta2_FlowCollectorExporter_To_v1beta1_FlowCollectorExporter(in, out, s); err != nil {
		return err
	}
	out.Type = ExporterType(utilconversion.PascalToUpper(string(in.Type), '_'))
	return nil
}

// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_AdvancedProcessorConfig_To_v1beta1_DebugConfig(in *v1beta2.AdvancedProcessorConfig, out *DebugConfig, s apiconversion.Scope) error {
	out.Env = in.Env
	return nil
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorEBPF_To_v1beta2_FlowCollectorEBPF(in *FlowCollectorEBPF, out *v1beta2.FlowCollectorEBPF, s apiconversion.Scope) error {
	out.Advanced = &v1beta2.AdvancedAgentConfig{
		Env: in.Debug.Env,
	}
	return autoConvert_v1beta1_FlowCollectorEBPF_To_v1beta2_FlowCollectorEBPF(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorEBPF_To_v1beta1_FlowCollectorEBPF(in *v1beta2.FlowCollectorEBPF, out *FlowCollectorEBPF, s apiconversion.Scope) error {
	if in.Advanced != nil {
		out.Debug.Env = in.Advanced.Env
	}
	return autoConvert_v1beta2_FlowCollectorEBPF_To_v1beta1_FlowCollectorEBPF(in, out, s)
}
