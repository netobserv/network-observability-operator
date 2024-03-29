//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta2

import (
	"k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvancedAgentConfig) DeepCopyInto(out *AdvancedAgentConfig) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvancedAgentConfig.
func (in *AdvancedAgentConfig) DeepCopy() *AdvancedAgentConfig {
	if in == nil {
		return nil
	}
	out := new(AdvancedAgentConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvancedLokiConfig) DeepCopyInto(out *AdvancedLokiConfig) {
	*out = *in
	if in.WriteMinBackoff != nil {
		in, out := &in.WriteMinBackoff, &out.WriteMinBackoff
		*out = new(v1.Duration)
		**out = **in
	}
	if in.WriteMaxBackoff != nil {
		in, out := &in.WriteMaxBackoff, &out.WriteMaxBackoff
		*out = new(v1.Duration)
		**out = **in
	}
	if in.WriteMaxRetries != nil {
		in, out := &in.WriteMaxRetries, &out.WriteMaxRetries
		*out = new(int32)
		**out = **in
	}
	if in.StaticLabels != nil {
		in, out := &in.StaticLabels, &out.StaticLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvancedLokiConfig.
func (in *AdvancedLokiConfig) DeepCopy() *AdvancedLokiConfig {
	if in == nil {
		return nil
	}
	out := new(AdvancedLokiConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvancedPluginConfig) DeepCopyInto(out *AdvancedPluginConfig) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Register != nil {
		in, out := &in.Register, &out.Register
		*out = new(bool)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvancedPluginConfig.
func (in *AdvancedPluginConfig) DeepCopy() *AdvancedPluginConfig {
	if in == nil {
		return nil
	}
	out := new(AdvancedPluginConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvancedProcessorConfig) DeepCopyInto(out *AdvancedProcessorConfig) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	if in.HealthPort != nil {
		in, out := &in.HealthPort, &out.HealthPort
		*out = new(int32)
		**out = **in
	}
	if in.ProfilePort != nil {
		in, out := &in.ProfilePort, &out.ProfilePort
		*out = new(int32)
		**out = **in
	}
	if in.EnableKubeProbes != nil {
		in, out := &in.EnableKubeProbes, &out.EnableKubeProbes
		*out = new(bool)
		**out = **in
	}
	if in.DropUnusedFields != nil {
		in, out := &in.DropUnusedFields, &out.DropUnusedFields
		*out = new(bool)
		**out = **in
	}
	if in.ConversationHeartbeatInterval != nil {
		in, out := &in.ConversationHeartbeatInterval, &out.ConversationHeartbeatInterval
		*out = new(v1.Duration)
		**out = **in
	}
	if in.ConversationEndTimeout != nil {
		in, out := &in.ConversationEndTimeout, &out.ConversationEndTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.ConversationTerminatingTimeout != nil {
		in, out := &in.ConversationTerminatingTimeout, &out.ConversationTerminatingTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvancedProcessorConfig.
func (in *AdvancedProcessorConfig) DeepCopy() *AdvancedProcessorConfig {
	if in == nil {
		return nil
	}
	out := new(AdvancedProcessorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateReference) DeepCopyInto(out *CertificateReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateReference.
func (in *CertificateReference) DeepCopy() *CertificateReference {
	if in == nil {
		return nil
	}
	out := new(CertificateReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientTLS) DeepCopyInto(out *ClientTLS) {
	*out = *in
	out.CACert = in.CACert
	out.UserCert = in.UserCert
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientTLS.
func (in *ClientTLS) DeepCopy() *ClientTLS {
	if in == nil {
		return nil
	}
	out := new(ClientTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkOperatorConfig) DeepCopyInto(out *ClusterNetworkOperatorConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkOperatorConfig.
func (in *ClusterNetworkOperatorConfig) DeepCopy() *ClusterNetworkOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsolePluginPortConfig) DeepCopyInto(out *ConsolePluginPortConfig) {
	*out = *in
	if in.Enable != nil {
		in, out := &in.Enable, &out.Enable
		*out = new(bool)
		**out = **in
	}
	if in.PortNames != nil {
		in, out := &in.PortNames, &out.PortNames
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsolePluginPortConfig.
func (in *ConsolePluginPortConfig) DeepCopy() *ConsolePluginPortConfig {
	if in == nil {
		return nil
	}
	out := new(ConsolePluginPortConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EBPFMetrics) DeepCopyInto(out *EBPFMetrics) {
	*out = *in
	in.Server.DeepCopyInto(&out.Server)
	if in.Enable != nil {
		in, out := &in.Enable, &out.Enable
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EBPFMetrics.
func (in *EBPFMetrics) DeepCopy() *EBPFMetrics {
	if in == nil {
		return nil
	}
	out := new(EBPFMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FLPMetrics) DeepCopyInto(out *FLPMetrics) {
	*out = *in
	in.Server.DeepCopyInto(&out.Server)
	if in.IncludeList != nil {
		in, out := &in.IncludeList, &out.IncludeList
		*out = new([]FLPMetric)
		if **in != nil {
			in, out := *in, *out
			*out = make([]FLPMetric, len(*in))
			copy(*out, *in)
		}
	}
	if in.DisableAlerts != nil {
		in, out := &in.DisableAlerts, &out.DisableAlerts
		*out = make([]FLPAlert, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FLPMetrics.
func (in *FLPMetrics) DeepCopy() *FLPMetrics {
	if in == nil {
		return nil
	}
	out := new(FLPMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileReference) DeepCopyInto(out *FileReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileReference.
func (in *FileReference) DeepCopy() *FileReference {
	if in == nil {
		return nil
	}
	out := new(FileReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollector) DeepCopyInto(out *FlowCollector) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollector.
func (in *FlowCollector) DeepCopy() *FlowCollector {
	if in == nil {
		return nil
	}
	out := new(FlowCollector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowCollector) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorAgent) DeepCopyInto(out *FlowCollectorAgent) {
	*out = *in
	out.IPFIX = in.IPFIX
	in.EBPF.DeepCopyInto(&out.EBPF)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorAgent.
func (in *FlowCollectorAgent) DeepCopy() *FlowCollectorAgent {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorAgent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorConsolePlugin) DeepCopyInto(out *FlowCollectorConsolePlugin) {
	*out = *in
	if in.Enable != nil {
		in, out := &in.Enable, &out.Enable
		*out = new(bool)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	in.Autoscaler.DeepCopyInto(&out.Autoscaler)
	in.PortNaming.DeepCopyInto(&out.PortNaming)
	if in.QuickFilters != nil {
		in, out := &in.QuickFilters, &out.QuickFilters
		*out = make([]QuickFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Advanced != nil {
		in, out := &in.Advanced, &out.Advanced
		*out = new(AdvancedPluginConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorConsolePlugin.
func (in *FlowCollectorConsolePlugin) DeepCopy() *FlowCollectorConsolePlugin {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorConsolePlugin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorEBPF) DeepCopyInto(out *FlowCollectorEBPF) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Sampling != nil {
		in, out := &in.Sampling, &out.Sampling
		*out = new(int32)
		**out = **in
	}
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExcludeInterfaces != nil {
		in, out := &in.ExcludeInterfaces, &out.ExcludeInterfaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Advanced != nil {
		in, out := &in.Advanced, &out.Advanced
		*out = new(AdvancedAgentConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Features != nil {
		in, out := &in.Features, &out.Features
		*out = make([]AgentFeature, len(*in))
		copy(*out, *in)
	}
	in.Metrics.DeepCopyInto(&out.Metrics)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorEBPF.
func (in *FlowCollectorEBPF) DeepCopy() *FlowCollectorEBPF {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorEBPF)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorExporter) DeepCopyInto(out *FlowCollectorExporter) {
	*out = *in
	out.Kafka = in.Kafka
	out.IPFIX = in.IPFIX
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorExporter.
func (in *FlowCollectorExporter) DeepCopy() *FlowCollectorExporter {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorExporter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorFLP) DeepCopyInto(out *FlowCollectorFLP) {
	*out = *in
	in.Metrics.DeepCopyInto(&out.Metrics)
	in.Resources.DeepCopyInto(&out.Resources)
	if in.KafkaConsumerReplicas != nil {
		in, out := &in.KafkaConsumerReplicas, &out.KafkaConsumerReplicas
		*out = new(int32)
		**out = **in
	}
	in.KafkaConsumerAutoscaler.DeepCopyInto(&out.KafkaConsumerAutoscaler)
	if in.LogTypes != nil {
		in, out := &in.LogTypes, &out.LogTypes
		*out = new(FLPLogTypes)
		**out = **in
	}
	if in.MultiClusterDeployment != nil {
		in, out := &in.MultiClusterDeployment, &out.MultiClusterDeployment
		*out = new(bool)
		**out = **in
	}
	if in.AddZone != nil {
		in, out := &in.AddZone, &out.AddZone
		*out = new(bool)
		**out = **in
	}
	if in.Advanced != nil {
		in, out := &in.Advanced, &out.Advanced
		*out = new(AdvancedProcessorConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorFLP.
func (in *FlowCollectorFLP) DeepCopy() *FlowCollectorFLP {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorFLP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorHPA) DeepCopyInto(out *FlowCollectorHPA) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]v2.MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorHPA.
func (in *FlowCollectorHPA) DeepCopy() *FlowCollectorHPA {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorHPA)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorIPFIX) DeepCopyInto(out *FlowCollectorIPFIX) {
	*out = *in
	out.ClusterNetworkOperator = in.ClusterNetworkOperator
	out.OVNKubernetes = in.OVNKubernetes
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorIPFIX.
func (in *FlowCollectorIPFIX) DeepCopy() *FlowCollectorIPFIX {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorIPFIX)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorIPFIXReceiver) DeepCopyInto(out *FlowCollectorIPFIXReceiver) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorIPFIXReceiver.
func (in *FlowCollectorIPFIXReceiver) DeepCopy() *FlowCollectorIPFIXReceiver {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorIPFIXReceiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorKafka) DeepCopyInto(out *FlowCollectorKafka) {
	*out = *in
	out.TLS = in.TLS
	out.SASL = in.SASL
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorKafka.
func (in *FlowCollectorKafka) DeepCopy() *FlowCollectorKafka {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorKafka)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorList) DeepCopyInto(out *FlowCollectorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FlowCollector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorList.
func (in *FlowCollectorList) DeepCopy() *FlowCollectorList {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowCollectorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorLoki) DeepCopyInto(out *FlowCollectorLoki) {
	*out = *in
	if in.Enable != nil {
		in, out := &in.Enable, &out.Enable
		*out = new(bool)
		**out = **in
	}
	out.Manual = in.Manual
	out.Microservices = in.Microservices
	out.Monolithic = in.Monolithic
	out.LokiStack = in.LokiStack
	if in.ReadTimeout != nil {
		in, out := &in.ReadTimeout, &out.ReadTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.WriteTimeout != nil {
		in, out := &in.WriteTimeout, &out.WriteTimeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.WriteBatchWait != nil {
		in, out := &in.WriteBatchWait, &out.WriteBatchWait
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Advanced != nil {
		in, out := &in.Advanced, &out.Advanced
		*out = new(AdvancedLokiConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorLoki.
func (in *FlowCollectorLoki) DeepCopy() *FlowCollectorLoki {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorLoki)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorSpec) DeepCopyInto(out *FlowCollectorSpec) {
	*out = *in
	in.Agent.DeepCopyInto(&out.Agent)
	in.Processor.DeepCopyInto(&out.Processor)
	in.Loki.DeepCopyInto(&out.Loki)
	in.ConsolePlugin.DeepCopyInto(&out.ConsolePlugin)
	out.Kafka = in.Kafka
	if in.Exporters != nil {
		in, out := &in.Exporters, &out.Exporters
		*out = make([]*FlowCollectorExporter, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(FlowCollectorExporter)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorSpec.
func (in *FlowCollectorSpec) DeepCopy() *FlowCollectorSpec {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorStatus) DeepCopyInto(out *FlowCollectorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorStatus.
func (in *FlowCollectorStatus) DeepCopy() *FlowCollectorStatus {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LokiManualParams) DeepCopyInto(out *LokiManualParams) {
	*out = *in
	out.TLS = in.TLS
	out.StatusTLS = in.StatusTLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LokiManualParams.
func (in *LokiManualParams) DeepCopy() *LokiManualParams {
	if in == nil {
		return nil
	}
	out := new(LokiManualParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LokiMicroservicesParams) DeepCopyInto(out *LokiMicroservicesParams) {
	*out = *in
	out.TLS = in.TLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LokiMicroservicesParams.
func (in *LokiMicroservicesParams) DeepCopy() *LokiMicroservicesParams {
	if in == nil {
		return nil
	}
	out := new(LokiMicroservicesParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LokiMonolithParams) DeepCopyInto(out *LokiMonolithParams) {
	*out = *in
	out.TLS = in.TLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LokiMonolithParams.
func (in *LokiMonolithParams) DeepCopy() *LokiMonolithParams {
	if in == nil {
		return nil
	}
	out := new(LokiMonolithParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LokiStackRef) DeepCopyInto(out *LokiStackRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LokiStackRef.
func (in *LokiStackRef) DeepCopy() *LokiStackRef {
	if in == nil {
		return nil
	}
	out := new(LokiStackRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsServerConfig) DeepCopyInto(out *MetricsServerConfig) {
	*out = *in
	in.TLS.DeepCopyInto(&out.TLS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsServerConfig.
func (in *MetricsServerConfig) DeepCopy() *MetricsServerConfig {
	if in == nil {
		return nil
	}
	out := new(MetricsServerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVNKubernetesConfig) DeepCopyInto(out *OVNKubernetesConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVNKubernetesConfig.
func (in *OVNKubernetesConfig) DeepCopy() *OVNKubernetesConfig {
	if in == nil {
		return nil
	}
	out := new(OVNKubernetesConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QuickFilter) DeepCopyInto(out *QuickFilter) {
	*out = *in
	if in.Filter != nil {
		in, out := &in.Filter, &out.Filter
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QuickFilter.
func (in *QuickFilter) DeepCopy() *QuickFilter {
	if in == nil {
		return nil
	}
	out := new(QuickFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SASLConfig) DeepCopyInto(out *SASLConfig) {
	*out = *in
	out.ClientIDReference = in.ClientIDReference
	out.ClientSecretReference = in.ClientSecretReference
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SASLConfig.
func (in *SASLConfig) DeepCopy() *SASLConfig {
	if in == nil {
		return nil
	}
	out := new(SASLConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerTLS) DeepCopyInto(out *ServerTLS) {
	*out = *in
	if in.Provided != nil {
		in, out := &in.Provided, &out.Provided
		*out = new(CertificateReference)
		**out = **in
	}
	if in.ProvidedCaFile != nil {
		in, out := &in.ProvidedCaFile, &out.ProvidedCaFile
		*out = new(FileReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerTLS.
func (in *ServerTLS) DeepCopy() *ServerTLS {
	if in == nil {
		return nil
	}
	out := new(ServerTLS)
	in.DeepCopyInto(out)
	return out
}
