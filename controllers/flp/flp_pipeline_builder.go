package flp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/conversion"
	"github.com/netobserv/network-observability-operator/pkg/filters"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/loki"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

type PipelineBuilder struct {
	*config.PipelineBuilderStage
	desired     *flowslatest.FlowCollectorSpec
	flowMetrics metricslatest.FlowMetricList
	volumes     *volumes.Builder
	loki        *helper.LokiConfig
	clusterID   string
}

func newPipelineBuilder(
	desired *flowslatest.FlowCollectorSpec,
	flowMetrics *metricslatest.FlowMetricList,
	loki *helper.LokiConfig,
	clusterID string,
	volumes *volumes.Builder,
	pipeline *config.PipelineBuilderStage,
) PipelineBuilder {
	return PipelineBuilder{
		PipelineBuilderStage: pipeline,
		desired:              desired,
		flowMetrics:          *flowMetrics,
		loki:                 loki,
		clusterID:            clusterID,
		volumes:              volumes,
	}
}

const openshiftNamespacesPrefixes = "openshift"

func (b *PipelineBuilder) AddProcessorStages() error {
	lastStage := *b.PipelineBuilderStage
	lastStage = b.addTransformFilter(lastStage)
	lastStage = b.addConnectionTracking(lastStage)

	addZone := helper.IsZoneEnabled(&b.desired.Processor)

	// enrich stage (transform) configuration
	enrichedStage := lastStage.TransformNetwork("enrich", api.TransformNetwork{
		Rules: api.NetworkTransformRules{{
			Type: api.AddKubernetesRuleType,
			Kubernetes: &api.K8sRule{
				Input:   "SrcAddr",
				Output:  "SrcK8S",
				AddZone: addZone,
			},
		}, {
			Type: api.AddKubernetesRuleType,
			Kubernetes: &api.K8sRule{
				Input:   "DstAddr",
				Output:  "DstK8S",
				AddZone: addZone,
			},
		}, {
			Type: api.ReinterpretDirectionRuleType,
		}, {
			Type: api.OpAddKubernetesInfra,
			KubernetesInfra: &api.K8sInfraRule{
				Inputs: []string{
					"SrcAddr",
					"DstAddr",
				},
				Output:        "K8S_FlowLayer",
				InfraPrefixes: []string{b.desired.Namespace, openshiftNamespacesPrefixes},
				InfraRefs: []api.K8sReference{
					{
						Name:      "kubernetes",
						Namespace: "default",
					},
					{
						Name:      "openshift",
						Namespace: "default",
					},
				},
			},
		}},
		DirectionInfo: api.NetworkTransformDirectionInfo{
			ReporterIPField:    "AgentIP",
			SrcHostField:       "SrcK8S_HostIP",
			DstHostField:       "DstK8S_HostIP",
			FlowDirectionField: "FlowDirection",
			IfDirectionField:   "IfDirection",
		},
	})

	// loki stage (write) configuration
	debugConfig := helper.GetAdvancedLokiConfig(b.desired.Loki.Advanced)
	if helper.UseLoki(b.desired) {
		lokiWrite := api.WriteLoki{
			Labels:         loki.GetLokiLabels(b.desired),
			BatchSize:      int(b.desired.Loki.WriteBatchSize),
			BatchWait:      helper.UnstructuredDuration(b.desired.Loki.WriteBatchWait),
			MaxBackoff:     helper.UnstructuredDuration(debugConfig.WriteMaxBackoff),
			MaxRetries:     int(helper.PtrInt32(debugConfig.WriteMaxRetries)),
			MinBackoff:     helper.UnstructuredDuration(debugConfig.WriteMinBackoff),
			StaticLabels:   model.LabelSet{},
			Timeout:        helper.UnstructuredDuration(b.desired.Loki.WriteTimeout),
			URL:            b.loki.IngesterURL,
			TimestampLabel: "TimeFlowEndMs",
			TimestampScale: "1ms",
			TenantID:       b.loki.TenantID,
		}

		for k, v := range debugConfig.StaticLabels {
			lokiWrite.StaticLabels[model.LabelName(k)] = model.LabelValue(v)
		}

		var authorization *promConfig.Authorization
		if b.loki.UseHostToken() || b.loki.UseForwardToken() {
			b.volumes.AddToken(constants.FLPName)
			authorization = &promConfig.Authorization{
				Type:            "Bearer",
				CredentialsFile: constants.TokensPath + constants.FLPName,
			}
		}

		if b.loki.TLS.Enable {
			if b.loki.TLS.InsecureSkipVerify {
				lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
					Authorization: authorization,
					TLSConfig: promConfig.TLSConfig{
						InsecureSkipVerify: true,
					},
				}
			} else {
				caPath := b.volumes.AddCACertificate(&b.loki.TLS, "loki-certs")
				lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
					Authorization: authorization,
					TLSConfig: promConfig.TLSConfig{
						CAFile: caPath,
					},
				}
			}
		} else {
			lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
				Authorization: authorization,
			}
		}
		enrichedStage.WriteLoki("loki", lokiWrite)
	}

	// write on Stdout if logging trace enabled
	if b.desired.Processor.LogLevel == "trace" {
		enrichedStage.WriteStdout("stdout", api.WriteStdout{Format: "json"})
	}

	// obtain encode_prometheus stage from metrics_definitions
	names := metrics.GetIncludeList(b.desired)
	promMetrics := metrics.GetDefinitions(names)

	for i := range b.flowMetrics.Items {
		fm := &b.flowMetrics.Items[i]
		m, err := flowMetricToFLP(&fm.Spec)
		if err != nil {
			return fmt.Errorf("error reading FlowMetric definition '%s': %w", fm.Name, err)
		}
		promMetrics = append(promMetrics, *m)
	}

	if len(promMetrics) > 0 {
		// prometheus stage (encode) configuration
		promEncode := api.PromEncode{
			Prefix:  "netobserv_",
			Metrics: promMetrics,
		}
		enrichedStage.EncodePrometheus("prometheus", promEncode)
	}

	b.addCustomExportStages(&enrichedStage)
	return nil
}

func flowMetricToFLP(flowMetric *metricslatest.FlowMetricSpec) (*api.MetricsItem, error) {
	m := &api.MetricsItem{
		Name:     flowMetric.MetricName,
		Type:     strings.ToLower(string(flowMetric.Type)),
		Filters:  []api.MetricsFilter{},
		Labels:   flowMetric.Labels,
		ValueKey: flowMetric.ValueField,
	}
	for _, f := range flowMetric.Filters {
		m.Filters = append(m.Filters, api.MetricsFilter{Key: f.Field, Value: f.Value, Type: conversion.PascalToLower(string(f.MatchType), '_')})
	}
	if !flowMetric.IncludeDuplicates {
		m.Filters = append(m.Filters, api.MetricsFilter{Key: "Duplicate", Value: "true", Type: api.PromFilterNotEqual})
	}
	if flowMetric.Direction == metricslatest.Egress {
		m.Filters = append(m.Filters, api.MetricsFilter{Key: "FlowDirection", Value: "1|2", Type: api.PromFilterRegex})
	} else if flowMetric.Direction == metricslatest.Ingress {
		m.Filters = append(m.Filters, api.MetricsFilter{Key: "FlowDirection", Value: "0|2", Type: api.PromFilterRegex})
	}
	for _, b := range flowMetric.Buckets {
		f, err := strconv.ParseFloat(b, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse metric buckets as floats: '%s'; error was: %w", b, err)
		}
		m.Buckets = append(m.Buckets, f)
	}
	return m, nil
}

func (b *PipelineBuilder) addConnectionTracking(lastStage config.PipelineBuilderStage) config.PipelineBuilderStage {
	outputFields := []api.OutputField{
		{
			Name:      "Bytes",
			Operation: "sum",
		},
		{
			Name:      "Bytes",
			Operation: "sum",
			SplitAB:   true,
		},
		{
			Name:      "Packets",
			Operation: "sum",
		},
		{
			Name:      "Packets",
			Operation: "sum",
			SplitAB:   true,
		},
		{
			Name:      "numFlowLogs",
			Operation: "count",
		},
		{
			Name:          "TimeFlowStartMs",
			Operation:     "min",
			ReportMissing: true,
		},
		{
			Name:          "TimeFlowEndMs",
			Operation:     "max",
			ReportMissing: true,
		},
		{
			Name:          "FlowDirection",
			Operation:     "first",
			ReportMissing: true,
		},
		{
			Name:          "IfDirection",
			Operation:     "first",
			ReportMissing: true,
		},
		{
			Name:          "AgentIP",
			Operation:     "first",
			ReportMissing: true,
		},
	}

	if helper.IsPktDropEnabled(&b.desired.Agent.EBPF) {
		outputPktDropFields := []api.OutputField{
			{
				Name:      "PktDropBytes",
				Operation: "sum",
			},
			{
				Name:      "PktDropBytes",
				Operation: "sum",
				SplitAB:   true,
			},
			{
				Name:      "PktDropPackets",
				Operation: "sum",
			},
			{
				Name:      "PktDropPackets",
				Operation: "sum",
				SplitAB:   true,
			},
			{
				Name:      "PktDropLatestState",
				Operation: "last",
			},
			{
				Name:      "PktDropLatestDropCause",
				Operation: "last",
			},
		}
		outputFields = append(outputFields, outputPktDropFields...)
	}

	if helper.IsDNSTrackingEnabled(&b.desired.Agent.EBPF) {
		outDNSTrackingFields := []api.OutputField{
			{
				Name:      "DnsFlagsResponseCode",
				Operation: "last",
			},
			{
				Name:      "DnsLatencyMs",
				Operation: "max",
			},
		}
		outputFields = append(outputFields, outDNSTrackingFields...)
	}

	if helper.IsFlowRTTEnabled(&b.desired.Agent.EBPF) {
		outputFields = append(outputFields, api.OutputField{
			Name:      "MaxTimeFlowRttNs",
			Operation: "max",
			Input:     "TimeFlowRttNs",
		})
	}

	// Connection tracking stage (only if LogTypes is not FLOWS)
	if b.desired.Processor.LogTypes != nil && *b.desired.Processor.LogTypes != flowslatest.LogTypeFlows {
		outputRecordTypes := helper.GetRecordTypes(&b.desired.Processor)
		debugConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
		lastStage = lastStage.ConnTrack("extract_conntrack", api.ConnTrack{
			KeyDefinition: api.KeyDefinition{
				FieldGroups: []api.FieldGroup{
					{Name: "src", Fields: []string{"SrcAddr", "SrcPort"}},
					{Name: "dst", Fields: []string{"DstAddr", "DstPort"}},
					{Name: "common", Fields: []string{"Proto"}},
				},
				Hash: api.ConnTrackHash{
					FieldGroupRefs: []string{
						"common",
					},
					FieldGroupARef: "src",
					FieldGroupBRef: "dst",
				},
			},
			OutputRecordTypes: outputRecordTypes,
			OutputFields:      outputFields,
			Scheduling: []api.ConnTrackSchedulingGroup{
				{
					Selector:             nil, // Default group. Match all flowlogs
					HeartbeatInterval:    api.Duration{Duration: debugConfig.ConversationHeartbeatInterval.Duration},
					EndConnectionTimeout: api.Duration{Duration: debugConfig.ConversationEndTimeout.Duration},
					TerminatingTimeout:   api.Duration{Duration: debugConfig.ConversationTerminatingTimeout.Duration},
				},
			},
			TCPFlags: api.ConnTrackTCPFlags{
				FieldName:           "Flags",
				DetectEndConnection: true,
				SwapAB:              true,
			},
		})
	}
	return lastStage
}

func (b *PipelineBuilder) addTransformFilter(lastStage config.PipelineBuilderStage) config.PipelineBuilderStage {
	var clusterName string
	transformFilterRules := []api.TransformFilterRule{}

	if helper.IsMultiClusterEnabled(&b.desired.Processor) {
		if b.desired.Processor.ClusterName != "" {
			clusterName = b.desired.Processor.ClusterName
		} else {
			//take clustername from openshift
			clusterName = string(b.clusterID)
		}
		if clusterName != "" {
			transformFilterRules = []api.TransformFilterRule{
				{
					Input: constants.ClusterNameLabelName,
					Type:  "add_field_if_doesnt_exist",
					Value: clusterName,
				},
			}
		}
	}

	// Filter-out unused fields?
	if *helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced).DropUnusedFields {
		if helper.UseIPFIX(b.desired) {
			rules := filters.GetOVSGoflowUnusedRules()
			transformFilterRules = append(transformFilterRules, rules...)
		}
		// Else: nothing for eBPF at the moment
	}
	if len(transformFilterRules) > 0 {
		lastStage = lastStage.TransformFilter("filter", api.TransformFilter{
			Rules: transformFilterRules,
		})
	}
	return lastStage
}

func (b *PipelineBuilder) addCustomExportStages(enrichedStage *config.PipelineBuilderStage) {
	for i, exporter := range b.desired.Exporters {
		if exporter.Type == flowslatest.KafkaExporter {
			b.createKafkaWriteStage(fmt.Sprintf("kafka-export-%d", i), &exporter.Kafka, enrichedStage)
		}
		if exporter.Type == flowslatest.IpfixExporter {
			createIPFIXWriteStage(fmt.Sprintf("IPFIX-export-%d", i), &exporter.IPFIX, enrichedStage)
		}
	}
}

func (b *PipelineBuilder) createKafkaWriteStage(name string, spec *flowslatest.FlowCollectorKafka, fromStage *config.PipelineBuilderStage) config.PipelineBuilderStage {
	return fromStage.EncodeKafka(name, api.EncodeKafka{
		Address: spec.Address,
		Topic:   spec.Topic,
		TLS:     getKafkaTLS(&spec.TLS, name, b.volumes),
		SASL:    getKafkaSASL(&spec.SASL, name, b.volumes),
	})
}

func (b *PipelineBuilder) AddKafkaWriteStage(name string, spec *flowslatest.FlowCollectorKafka) config.PipelineBuilderStage {
	return b.createKafkaWriteStage(name, spec, b.PipelineBuilderStage)
}

func createIPFIXWriteStage(name string, spec *flowslatest.FlowCollectorIPFIXReceiver, fromStage *config.PipelineBuilderStage) config.PipelineBuilderStage {
	return fromStage.WriteIpfix(name, api.WriteIpfix{
		TargetHost:   spec.TargetHost,
		TargetPort:   spec.TargetPort,
		Transport:    getIPFIXTransport(spec.Transport),
		EnterpriseID: 2,
	})
}

func getIPFIXTransport(transport string) string {
	switch transport {
	case "UDP":
		return "udp"
	default:
		return "tcp" //always fallback on tcp
	}
}

func getKafkaTLS(tls *flowslatest.ClientTLS, volumeName string, volumes *volumes.Builder) *api.ClientTLS {
	if tls.Enable {
		caPath, userCertPath, userKeyPath := volumes.AddMutualTLSCertificates(tls, volumeName)
		return &api.ClientTLS{
			InsecureSkipVerify: tls.InsecureSkipVerify,
			CACertPath:         caPath,
			UserCertPath:       userCertPath,
			UserKeyPath:        userKeyPath,
		}
	}
	return nil
}

func getKafkaSASL(sasl *flowslatest.SASLConfig, volumePrefix string, volumes *volumes.Builder) *api.SASLConfig {
	if !helper.UseSASL(sasl) {
		return nil
	}
	t := "plain"
	if sasl.Type == flowslatest.SASLScramSHA512 {
		t = "scramSHA512"
	}
	idPath := volumes.AddVolume(&sasl.ClientIDReference, volumePrefix+"-sasl-id")
	secretPath := volumes.AddVolume(&sasl.ClientSecretReference, volumePrefix+"-sasl-secret")
	return &api.SASLConfig{
		Type:             t,
		ClientIDPath:     idPath,
		ClientSecretPath: secretPath,
	}
}
