package flp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/flp/fmstatus"
	"github.com/netobserv/network-observability-operator/pkg/conversion"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/helper/loki"
	otelConfig "github.com/netobserv/network-observability-operator/pkg/helper/otel"
	"github.com/netobserv/network-observability-operator/pkg/metrics"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

const (
	ovnkSecondary = "ovn-kubernetes"
)

type PipelineBuilder struct {
	*config.PipelineBuilderStage
	desired         *flowslatest.FlowCollectorSpec
	flowMetrics     metricslatest.FlowMetricList
	detectedSubnets []flowslatest.SubnetLabel
	volumes         *volumes.Builder
	loki            *helper.LokiConfig
	clusterID       string
}

func newPipelineBuilder(
	desired *flowslatest.FlowCollectorSpec,
	flowMetrics *metricslatest.FlowMetricList,
	detectedSubnets []flowslatest.SubnetLabel,
	loki *helper.LokiConfig,
	clusterID string,
	volumes *volumes.Builder,
	pipeline *config.PipelineBuilderStage,
) PipelineBuilder {
	return PipelineBuilder{
		PipelineBuilderStage: pipeline,
		desired:              desired,
		flowMetrics:          *flowMetrics,
		detectedSubnets:      detectedSubnets,
		loki:                 loki,
		clusterID:            clusterID,
		volumes:              volumes,
	}
}

const openshiftNamespacesPrefixes = "openshift"

// nolint:cyclop
func (b *PipelineBuilder) AddProcessorStages() error {
	lastStage := *b.PipelineBuilderStage
	lastStage = b.addTransformFilter(lastStage)
	lastStage = b.addConnectionTracking(lastStage)

	addZone := helper.IsZoneEnabled(&b.desired.Processor)

	// Get all subnet labels
	allLabels := b.desired.Processor.SubnetLabels.CustomLabels
	allLabels = append(allLabels, b.detectedSubnets...)
	flpLabels := subnetLabelsToFLP(allLabels)

	rules := api.NetworkTransformRules{
		{
			Type: api.NetworkAddKubernetes,
			Kubernetes: &api.K8sRule{
				IPField:         "SrcAddr",
				MACField:        "SrcMac",
				InterfacesField: "Interfaces",
				UDNsField:       "Udns",
				Output:          "SrcK8S",
				AddZone:         addZone,
			},
		},
		{
			Type: api.NetworkAddKubernetes,
			Kubernetes: &api.K8sRule{
				IPField:         "DstAddr",
				MACField:        "DstMac",
				InterfacesField: "Interfaces",
				UDNsField:       "Udns",
				Output:          "DstK8S",
				AddZone:         addZone,
			},
		},
		{
			Type: api.NetworkAddKubernetes,
			Kubernetes: &api.K8sRule{
				IPField: "XlatSrcAddr",
				Output:  "XlatSrcK8S",
			},
		},
		{
			Type: api.NetworkAddKubernetes,
			Kubernetes: &api.K8sRule{
				IPField: "XlatDstAddr",
				Output:  "XlatDstK8S",
			},
		},
		{
			Type: api.NetworkReinterpretDirection,
		},
		{
			Type: api.NetworkDecodeTCPFlags,
			DecodeTCPFlags: &api.NetworkGenericRule{
				Input:  "Flags",
				Output: "Flags",
			},
		},
		{
			Type: api.NetworkAddKubernetesInfra,
			KubernetesInfra: &api.K8sInfraRule{
				NamespaceNameFields: []api.K8sReference{
					{Namespace: "SrcK8S_Namespace", Name: "SrcK8S_Name"},
					{Namespace: "DstK8S_Namespace", Name: "DstK8S_Name"},
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
		},
	}

	if len(flpLabels) > 0 {
		rules = append(rules, []api.NetworkTransformRule{
			{
				Type: api.NetworkAddSubnetLabel,
				AddSubnetLabel: &api.NetworkAddSubnetLabelRule{
					Input:  "SrcAddr",
					Output: "SrcSubnetLabel",
				},
			},
			{
				Type: api.NetworkAddSubnetLabel,
				AddSubnetLabel: &api.NetworkAddSubnetLabelRule{
					Input:  "DstAddr",
					Output: "DstSubnetLabel",
				},
			},
		}...)
	}

	// Propagate 2dary networks config
	var secondaryNetworks []api.SecondaryNetwork
	if b.desired.Processor.Advanced != nil && len(b.desired.Processor.Advanced.SecondaryNetworks) > 0 {
		for _, sn := range b.desired.Processor.Advanced.SecondaryNetworks {
			flpSN := api.SecondaryNetwork{
				Name:  sn.Name,
				Index: map[string]any{},
			}
			for _, index := range sn.Index {
				flpSN.Index[strings.ToLower(string(index))] = nil
			}
			secondaryNetworks = append(secondaryNetworks, flpSN)
		}
	}
	if helper.IsUDNMappingEnabled(&b.desired.Agent.EBPF) {
		secondaryNetworks = append(secondaryNetworks, api.SecondaryNetwork{
			Name:  ovnkSecondary,
			Index: map[string]any{"udn": nil},
		})
	}

	// enrich stage (transform) configuration
	nextStage := lastStage.TransformNetwork("enrich", api.TransformNetwork{
		Rules: rules,
		DirectionInfo: api.NetworkTransformDirectionInfo{
			ReporterIPField:    "AgentIP",
			SrcHostField:       "SrcK8S_HostIP",
			DstHostField:       "DstK8S_HostIP",
			FlowDirectionField: "FlowDirection",
		},
		SubnetLabels: flpLabels,
		KubeConfig: api.NetworkTransformKubeConfig{
			SecondaryNetworks: secondaryNetworks,
		},
	})

	// Custom filters
	filters := filtersToFLP(b.desired.Processor.Filters, flowslatest.FLPFilterTargetAll)
	if len(filters) > 0 {
		nextStage = nextStage.TransformFilter("filters", newTransformFilter(filters))
	}

	// Dedup stage
	if helper.HasFLPDeduper(b.desired) {
		dedupRules := []*api.RemoveEntryRule{
			{
				Type: api.RemoveEntryIfEqualD,
				RemoveEntry: &api.TransformFilterGenericRule{
					Input:   "FlowDirection",
					Value:   1,
					CastInt: true,
				},
			},
			{
				Type: api.RemoveEntryIfExistsD,
				RemoveEntry: &api.TransformFilterGenericRule{
					Input: "DstK8S_OwnerName",
				},
			},
		}
		var transformFilter api.TransformFilter
		if b.desired.Processor.Deduper.Mode == flowslatest.FLPDeduperDrop {
			transformFilter = api.TransformFilter{
				Rules: []api.TransformFilterRule{
					{
						Type:                    api.RemoveEntryAllSatisfied,
						RemoveEntryAllSatisfied: dedupRules,
					},
				},
			}
		} else {
			transformFilter = api.TransformFilter{
				Rules: []api.TransformFilterRule{
					{
						Type: api.ConditionalSampling,
						ConditionalSampling: []*api.SamplingCondition{
							{
								Rules: dedupRules,
								Value: uint16(b.desired.Processor.Deduper.Sampling),
							},
						},
					},
				},
			}
		}
		nextStage = nextStage.TransformFilter("dedup", transformFilter)
	}

	// loki stage (write) configuration
	advancedConfig := helper.GetAdvancedLokiConfig(b.desired.Loki.Advanced)
	if helper.UseLoki(b.desired) {
		lokiLabels, err := loki.GetLabels(&b.desired.Processor)
		if err != nil {
			return err
		}
		lokiStage := nextStage
		// Custom filters: Loki only
		filters := filtersToFLP(b.desired.Processor.Filters, flowslatest.FLPFilterTargetLoki)
		if len(filters) > 0 {
			lokiStage = lokiStage.TransformFilter("filters-loki", newTransformFilter(filters))
		}

		lokiWrite := api.WriteLoki{
			Labels:         lokiLabels,
			BatchSize:      int(b.desired.Loki.WriteBatchSize),
			BatchWait:      helper.UnstructuredDuration(b.desired.Loki.WriteBatchWait),
			MaxBackoff:     helper.UnstructuredDuration(advancedConfig.WriteMaxBackoff),
			MaxRetries:     int(helper.PtrInt32(advancedConfig.WriteMaxRetries)),
			MinBackoff:     helper.UnstructuredDuration(advancedConfig.WriteMinBackoff),
			StaticLabels:   model.LabelSet{},
			Timeout:        helper.UnstructuredDuration(b.desired.Loki.WriteTimeout),
			URL:            b.loki.IngesterURL,
			TimestampLabel: "TimeFlowEndMs",
			TimestampScale: "1ms",
			TenantID:       b.loki.TenantID,
		}

		for k, v := range advancedConfig.StaticLabels {
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
		lokiStage.WriteLoki("loki", lokiWrite)
	}

	// write on Stdout if logging trace enabled
	if b.desired.Processor.LogLevel == "trace" {
		nextStage.WriteStdout("stdout", api.WriteStdout{Format: "json"})
	}

	// obtain encode_prometheus stage from metrics_definitions
	allMetrics := metrics.MergePredefined(b.flowMetrics.Items, b.desired)

	var flpMetrics []api.MetricsItem
	for i := range allMetrics {
		fm := &allMetrics[i]
		m, err := flowMetricToFLP(&fm.Spec)
		if err != nil {
			// fm.Name is empty for predefined metrics; check this is a custom metric, not a predefined one
			if fm.Name != "" {
				fmstatus.SetFailure(fm, err.Error())
				continue
			}
			// Predefined metric failure => bug
			return fmt.Errorf("error reading FlowMetric definition '%s': %w", fm.Name, err)
		}
		if fm.Name != "" {
			fmstatus.CheckCardinality(fm)
		}
		flpMetrics = append(flpMetrics, *m)
	}

	if len(flpMetrics) > 0 {
		promStage := nextStage
		// Custom filters: Loki only
		filters := filtersToFLP(b.desired.Processor.Filters, flowslatest.FLPFilterTargetMetrics)
		if len(filters) > 0 {
			promStage = promStage.TransformFilter("filters-prom", newTransformFilter(filters))
		}

		// prometheus stage (encode) configuration
		promEncode := api.PromEncode{
			Prefix:  "netobserv_",
			Metrics: flpMetrics,
		}
		promStage.EncodePrometheus("prometheus", promEncode)
	}

	expStage := nextStage
	// Custom filters: Exporters only
	filters = filtersToFLP(b.desired.Processor.Filters, flowslatest.FLPFilterTargetExporters)
	if len(filters) > 0 {
		expStage = expStage.TransformFilter("filters-exp", newTransformFilter(filters))
	}
	err := b.addCustomExportStages(&expStage, flpMetrics)
	return err
}

func newTransformFilter(rules []api.TransformFilterRule) api.TransformFilter {
	return api.TransformFilter{Rules: rules, SamplingField: "Sampling"}
}

func filtersToFLP(in []flowslatest.FLPFilterSet, target flowslatest.FLPFilterTarget) []api.TransformFilterRule {
	var rules []api.TransformFilterRule
	for _, f := range in {
		if f.OutputTarget == target {
			rules = append(rules, api.TransformFilterRule{
				Type:              api.KeepEntryQuery,
				KeepEntryQuery:    f.Query,
				KeepEntrySampling: uint16(f.Sampling),
			})
		}
	}
	return rules
}

func flowMetricToFLP(flowMetric *metricslatest.FlowMetricSpec) (*api.MetricsItem, error) {
	m := &api.MetricsItem{
		Name:     flowMetric.MetricName,
		Type:     api.MetricEncodeOperationEnum(strings.ToLower(string(flowMetric.Type))),
		Filters:  []api.MetricsFilter{},
		Labels:   flowMetric.Labels,
		Remap:    flowMetric.Remap,
		Flatten:  flowMetric.Flatten,
		ValueKey: flowMetric.ValueField,
	}
	for _, f := range metrics.GetFilters(flowMetric) {
		m.Filters = append(m.Filters, api.MetricsFilter{Key: f.Field, Value: f.Value, Type: api.MetricFilterEnum(conversion.PascalToLower(string(f.MatchType), '_'))})
	}
	for _, b := range flowMetric.Buckets {
		f, err := strconv.ParseFloat(b, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse metric buckets as floats: '%s'; error was: %w", b, err)
		}
		m.Buckets = append(m.Buckets, f)
	}
	if flowMetric.Divider != "" {
		f, err := strconv.ParseFloat(flowMetric.Divider, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse metric divider as float: '%s'; error was: %w", flowMetric.Divider, err)
		}
		m.ValueScale = f
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

	if helper.IsNetworkEventsEnabled(&b.desired.Agent.EBPF) {
		outNetworkEventsFlowFields := []api.OutputField{
			{
				Name:      "NetworkEvents",
				Operation: "last",
			},
		}
		outputFields = append(outputFields, outNetworkEventsFlowFields...)
	}

	if helper.IsFlowRTTEnabled(&b.desired.Agent.EBPF) {
		outputFields = append(outputFields, api.OutputField{
			Name:      "MaxTimeFlowRttNs",
			Operation: "max",
			Input:     "TimeFlowRttNs",
		})
	}

	// Connection tracking stage (only if LogTypes is not FLOWS)
	if helper.IsConntrack(&b.desired.Processor) {
		outputRecordTypes := helper.GetRecordTypes(&b.desired.Processor)
		advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
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
					HeartbeatInterval:    api.Duration{Duration: advancedConfig.ConversationHeartbeatInterval.Duration},
					EndConnectionTimeout: api.Duration{Duration: advancedConfig.ConversationEndTimeout.Duration},
					TerminatingTimeout:   api.Duration{Duration: advancedConfig.ConversationTerminatingTimeout.Duration},
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
			// Take clustername from openshift
			clusterName = b.clusterID
		}
		if clusterName != "" {
			transformFilterRules = []api.TransformFilterRule{
				{
					Type: api.AddFieldIfDoesntExist,
					AddFieldIfDoesntExist: &api.TransformFilterGenericRule{
						Input: constants.ClusterNameLabelName,
						Value: clusterName,
					},
				},
			}
		}
	}

	if len(transformFilterRules) > 0 {
		lastStage = lastStage.TransformFilter("filter", api.TransformFilter{
			Rules: transformFilterRules,
		})
	}
	return lastStage
}

func (b *PipelineBuilder) addCustomExportStages(enrichedStage *config.PipelineBuilderStage, flpMetrics []api.MetricsItem) error {
	for i, exporter := range b.desired.Exporters {
		if exporter.Type == flowslatest.KafkaExporter {
			b.createKafkaWriteStage(fmt.Sprintf("kafka-export-%d", i), &exporter.Kafka, enrichedStage)
		}
		if exporter.Type == flowslatest.IpfixExporter {
			createIPFIXWriteStage(fmt.Sprintf("IPFIX-export-%d", i), &exporter.IPFIX, enrichedStage)
		}
		if exporter.Type == flowslatest.OpenTelemetryExporter {
			err := b.createOpenTelemetryStage(fmt.Sprintf("Otel-export-%d", i), &exporter.OpenTelemetry, enrichedStage, flpMetrics)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *PipelineBuilder) createKafkaWriteStage(name string, spec *flowslatest.FlowCollectorKafka, fromStage *config.PipelineBuilderStage) config.PipelineBuilderStage {
	return fromStage.EncodeKafka(name, api.EncodeKafka{
		Address: spec.Address,
		Topic:   spec.Topic,
		TLS:     getClientTLS(&spec.TLS, name, b.volumes),
		SASL:    getSASL(&spec.SASL, name, b.volumes),
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
		return "tcp" // Always fallback on tcp
	}
}

func (b *PipelineBuilder) createOpenTelemetryStage(name string, spec *flowslatest.FlowCollectorOpenTelemetry, fromStage *config.PipelineBuilderStage, flpMetrics []api.MetricsItem) error {
	conn := api.OtlpConnectionInfo{
		Address:        spec.TargetHost,
		Port:           spec.TargetPort,
		ConnectionType: getOtelConnType(spec.Protocol),
		TLS:            getClientTLS(&spec.TLS, name, b.volumes),
		Headers:        spec.Headers,
	}

	logsEnabled := spec.Logs.Enable != nil && *spec.Logs.Enable
	metricsEnabled := spec.Metrics.Enable != nil && *spec.Metrics.Enable

	if logsEnabled || metricsEnabled {
		// add transform stage
		transformCfg, err := otelConfig.GetOtelTransformConfig(spec.FieldsMapping)
		if err != nil {
			return err
		}
		transformStage := fromStage.TransformGeneric(fmt.Sprintf("%s-transform", name), *transformCfg)

		// otel logs config
		if logsEnabled {
			// add encode stage(s)
			transformStage.EncodeOtelLogs(fmt.Sprintf("%s-logs", name), api.EncodeOtlpLogs{
				OtlpConnectionInfo: &conn,
			})
		}

		// otel metrics config
		if metricsEnabled {
			metricsCfg, err := otelConfig.GetOtelMetrics(flpMetrics)
			if err != nil {
				return err
			}
			transformStage.EncodeOtelMetrics(fmt.Sprintf("%s-metrics", name), api.EncodeOtlpMetrics{
				OtlpConnectionInfo: &conn,
				Prefix:             "netobserv_",
				Metrics:            metricsCfg,
				PushTimeInterval:   api.Duration{Duration: spec.Metrics.PushTimeInterval.Duration},
				ExpiryTime:         api.Duration{Duration: 2 * time.Minute},
			})
		}

		// TODO: implement api.EncodeOtlpTraces
	}
	return nil
}

func getOtelConnType(connType string) string {
	switch connType {
	case "http":
		return "http"
	default:
		return "grpc"
	}
}

func getClientTLS(tls *flowslatest.ClientTLS, volumeName string, volumes *volumes.Builder) *api.ClientTLS {
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

func getSASL(sasl *flowslatest.SASLConfig, volumePrefix string, volumes *volumes.Builder) *api.SASLConfig {
	if !helper.UseSASL(sasl) {
		return nil
	}
	t := api.SASLPlain
	if sasl.Type == flowslatest.SASLScramSHA512 {
		t = api.SASLScramSHA512
	}
	idPath := volumes.AddVolume(&sasl.ClientIDReference, volumePrefix+"-sasl-id")
	secretPath := volumes.AddVolume(&sasl.ClientSecretReference, volumePrefix+"-sasl-secret")
	return &api.SASLConfig{
		Type:             t,
		ClientIDPath:     idPath,
		ClientSecretPath: secretPath,
	}
}

func subnetLabelsToFLP(labels []flowslatest.SubnetLabel) []api.NetworkTransformSubnetLabel {
	var cats []api.NetworkTransformSubnetLabel
	for _, subnetLabel := range labels {
		cats = append(cats, api.NetworkTransformSubnetLabel{
			Name:  subnetLabel.Name,
			CIDRs: subnetLabel.CIDRs,
		})
	}
	return cats
}
