package flp

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// This function validate that each stage has its matching parameter
func validatePipelineConfig(t *testing.T, staticCm *corev1.ConfigMap, dynamicCm *corev1.ConfigMap) (*config.ConfigFileStruct, string) {
	var cfs config.ConfigFileStruct
	err := json.Unmarshal([]byte(staticCm.Data[configFile]), &cfs)
	assert.NoError(t, err)

	var dynCfs config.HotReloadStruct
	err = json.Unmarshal([]byte(dynamicCm.Data[configFile]), &dynCfs)
	assert.NoError(t, err)

	cfs.Parameters = append(cfs.Parameters, dynCfs.Parameters...)

	for _, stage := range cfs.Pipeline {
		assert.NotEmpty(t, stage.Name)
		exist := false
		for _, parameter := range cfs.Parameters {
			if stage.Name == parameter.Name {
				exist = true
				break
			}
		}
		assert.True(t, exist, "stage params not found", stage.Name)
	}
	b, err := json.Marshal(cfs.Pipeline)
	assert.NoError(t, err)
	return &cfs, string(b)
}

func TestPipelineConfig(t *testing.T) {
	assert := assert.New(t)

	// Single config
	ns := "namespace"
	cfg := getConfig()
	cfg.Processor.LogLevel = "info"
	b := monoBuilder(ns, &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"extract_conntrack","follows":"grpc"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)

	// Kafka Transformer
	cfg.DeploymentModel = flowslatest.DeploymentModelKafka
	bt := transfBuilder(ns, &cfg)
	scm, _, err = bt.staticConfigMap()
	assert.NoError(err)
	dcm, err = bt.dynamicConfigMap()
	assert.NoError(err)
	_, pipeline = validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"kafka-read"},{"name":"extract_conntrack","follows":"kafka-read"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}

func TestPipelineTraceStage(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"extract_conntrack","follows":"grpc"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}

func getSortedMetricsNames(m []api.MetricsItem) []string {
	ret := []string{}
	for i := range m {
		ret = append(ret, m[i].Name)
	}
	sort.Strings(ret)
	return ret
}

func TestMergeMetricsConfiguration_Default(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, scm, dcm)
	names := getSortedMetricsNames(cfs.Parameters[5].Encode.Prom.Metrics)
	assert.Equal([]string{
		"namespace_flows_total",
		"node_egress_bytes_total",
		"node_ingress_bytes_total",
		"workload_egress_bytes_total",
		"workload_ingress_bytes_total",
	}, names)
	assert.Equal("netobserv_", cfs.Parameters[5].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_DefaultWithFeatures(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Agent.EBPF.Privileged = true
	cfg.Agent.EBPF.Features = []flowslatest.AgentFeature{flowslatest.DNSTracking, flowslatest.FlowRTT, flowslatest.PacketDrop}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, scm, dcm)
	names := getSortedMetricsNames(cfs.Parameters[5].Encode.Prom.Metrics)
	assert.Equal([]string{
		"namespace_dns_latency_seconds",
		"namespace_drop_packets_total",
		"namespace_flows_total",
		"namespace_rtt_seconds",
		"node_egress_bytes_total",
		"node_ingress_bytes_total",
		"workload_egress_bytes_total",
		"workload_ingress_bytes_total",
	}, names)
	assert.Equal("netobserv_", cfs.Parameters[5].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_WithList(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.Metrics.IncludeList = &[]flowslatest.FLPMetric{"namespace_egress_bytes_total", "namespace_ingress_bytes_total"}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, scm, dcm)
	names := getSortedMetricsNames(cfs.Parameters[5].Encode.Prom.Metrics)
	assert.Len(names, 2)
	assert.Equal("namespace_egress_bytes_total", names[0])
	assert.Equal("namespace_ingress_bytes_total", names[1])
	assert.Equal("netobserv_", cfs.Parameters[5].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_EmptyList(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.Metrics.IncludeList = &[]flowslatest.FLPMetric{}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, scm, dcm)
	assert.Len(cfs.Parameters, 5)
}

func TestPipelineWithExporter(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Exporters = append(cfg.Exporters, &flowslatest.FlowCollectorExporter{
		Type:  flowslatest.KafkaExporter,
		Kafka: flowslatest.FlowCollectorKafka{Address: "kafka-test", Topic: "topic-test"},
	})

	cfg.Exporters = append(cfg.Exporters, &flowslatest.FlowCollectorExporter{
		Type: flowslatest.IpfixExporter,
		IPFIX: flowslatest.FlowCollectorIPFIXReceiver{
			TargetHost: "ipfix-receiver-test",
			TargetPort: 9999,
			Transport:  "TCP",
		},
	})

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"extract_conntrack","follows":"grpc"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"},{"name":"kafka-export-0","follows":"enrich"},{"name":"IPFIX-export-1","follows":"enrich"}]`,
		pipeline,
	)

	assert.Equal("kafka-test", cfs.Parameters[5].Encode.Kafka.Address)
	assert.Equal("topic-test", cfs.Parameters[5].Encode.Kafka.Topic)

	assert.Equal("ipfix-receiver-test", cfs.Parameters[6].Write.Ipfix.TargetHost)
	assert.Equal(9999, cfs.Parameters[6].Write.Ipfix.TargetPort)
	assert.Equal("tcp", cfs.Parameters[6].Write.Ipfix.Transport)
}

func TestPipelineWithoutLoki(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Loki.Enable = ptr.To(false)

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"extract_conntrack","follows":"grpc"},{"name":"enrich","follows":"extract_conntrack"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}

func TestReadMachineNetworks(t *testing.T) {
	cm := corev1.ConfigMap{
		Data: map[string]string{
			"install-config": `
additionalTrustBundlePolicy: Proxyonly
apiVersion: v1
baseDomain: my.openshift.com
compute:
- architecture: amd64
  hyperthreading: Enabled
  name: worker
  platform: {}
  replicas: 3
controlPlane:
  architecture: amd64
  hyperthreading: Enabled
  name: master
  platform: {}
  replicas: 3
metadata:
  creationTimestamp: null
  name: my-cluster
networking:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineNetwork:
  - cidr: 10.0.0.0/16
  networkType: OVNKubernetes
  serviceNetwork:
  - 172.30.0.0/16
platform:
  aws:
    region: eu-west-3
publish: External`,
		},
	}

	machines, err := readMachineNetworks(&cm)
	assert.NoError(t, err)

	assert.Equal(t,
		[]flowslatest.SubnetLabel{
			{
				Name:  "Machines",
				CIDRs: []string{"10.0.0.0/16"},
			},
		}, machines)
}

func TestPipelineWithSubnetLabels(t *testing.T) {
	assert := assert.New(t)

	cfg := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			SubnetLabels: flowslatest.SubnetLabels{
				OpenShiftAutoDetect: ptr.To(true),
				CustomLabels: []flowslatest.SubnetLabel{
					{
						Name:  "Foo",
						CIDRs: []string{"8.8.8.8/32"},
					},
				},
			},
		},
		Loki: flowslatest.FlowCollectorLoki{Enable: ptr.To(false)},
	}

	b := monoBuilder("namespace", &cfg)
	b.generic.detectedSubnets = []flowslatest.SubnetLabel{
		{
			Name:  "Pods",
			CIDRs: []string{"10.128.0.0/14"},
		},
	}
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"enrich","follows":"grpc"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
	assert.Equal(
		[]api.NetworkTransformSubnetLabel{
			{
				Name:  "Foo",
				CIDRs: []string{"8.8.8.8/32"},
			},
			{
				Name:  "Pods",
				CIDRs: []string{"10.128.0.0/14"},
			},
		},
		cfs.Parameters[1].Transform.Network.SubnetLabels,
	)
}

func TestPipelineWithFilters_WantNamespacesABC(t *testing.T) {
	assert := assert.New(t)

	cfg := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Filters: []flowslatest.FLPFilterSet{
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterEqual,
							Field:     "namespace",
							Value:     "A",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
				},
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterEqual,
							Field:     "namespace",
							Value:     "B",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
				},
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterEqual,
							Field:     "namespace",
							Value:     "C",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
				},
			},
		},
	}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"enrich","follows":"grpc"},{"name":"filters","follows":"enrich"},{"name":"loki","follows":"filters"},{"name":"prometheus","follows":"filters"}]`,
		pipeline,
	)
	assert.Equal(
		[]api.TransformFilterRule{
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "A"},
					},
				},
			},
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "B"},
					},
				},
			},
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "C"},
					},
				},
			},
		},
		cfs.Parameters[2].Transform.Filter.Rules,
	)
}

func TestPipelineWithFilters_DontWantNamespacesABC_LokiOnly(t *testing.T) {
	assert := assert.New(t)

	cfg := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Filters: []flowslatest.FLPFilterSet{
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterNotEqual,
							Field:     "namespace",
							Value:     "A",
						},
						{
							MatchType: flowslatest.FLPFilterNotEqual,
							Field:     "namespace",
							Value:     "B",
						},
						{
							MatchType: flowslatest.FLPFilterNotEqual,
							Field:     "namespace",
							Value:     "C",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetLoki,
				},
			},
		},
	}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"enrich","follows":"grpc"},{"name":"filters-loki","follows":"enrich"},{"name":"loki","follows":"filters-loki"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
	assert.Equal(
		[]api.TransformFilterRule{
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfNotEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "A"},
					},
					{
						Type:      api.KeepEntryIfNotEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "B"},
					},
					{
						Type:      api.KeepEntryIfNotEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "C"},
					},
				},
			},
		},
		cfs.Parameters[2].Transform.Filter.Rules,
	)
}

func TestPipelineWithFilters_ComplexFilter(t *testing.T) {
	assert := assert.New(t)

	// Keep flow when: (namespace!=dont_want_1 AND namespace!=dont_want_2) OR (keep_if_exist is present) OR (throw_if_not_exist is absent)

	cfg := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Filters: []flowslatest.FLPFilterSet{
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterNotEqual,
							Field:     "namespace",
							Value:     "dont_want_1",
						},
						{
							MatchType: flowslatest.FLPFilterNotEqual,
							Field:     "namespace",
							Value:     "dont_want_2",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
				},
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterPresence,
							Field:     "sample_if_exist",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
					Sampling:     50,
				},
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterAbsence,
							Field:     "keep_if_not_exist",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetAll,
				},
				{
					AllOf: []flowslatest.FLPSingleFilter{
						{
							MatchType: flowslatest.FLPFilterEqual,
							Field:     "namespace",
							Value:     "C",
						},
						{
							MatchType: flowslatest.FLPFilterEqual,
							Field:     "workload",
							Value:     "C1",
						},
					},
					OutputTarget: flowslatest.FLPFilterTargetLoki,
				},
			},
		},
	}

	b := monoBuilder("namespace", &cfg)
	scm, _, err := b.staticConfigMap()
	assert.NoError(err)
	dcm, err := b.dynamicConfigMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, scm, dcm)
	assert.Equal(
		`[{"name":"grpc"},{"name":"enrich","follows":"grpc"},{"name":"filters","follows":"enrich"},{"name":"filters-loki","follows":"filters"},{"name":"loki","follows":"filters-loki"},{"name":"prometheus","follows":"filters"}]`,
		pipeline,
	)

	// Keep flow when: (namespace!=dont_want_1 AND namespace!=dont_want_2) OR (keep_if_exist is present) OR (throw_if_not_exist is absent)
	assert.Equal(
		[]api.TransformFilterRule{
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfNotEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "dont_want_1"},
					},
					{
						Type:      api.KeepEntryIfNotEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "dont_want_2"},
					},
				},
			},
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfExists,
						KeepEntry: &api.TransformFilterGenericRule{Input: "sample_if_exist", Value: ""},
					},
				},
				KeepEntrySampling: 50,
			},
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfDoesntExist,
						KeepEntry: &api.TransformFilterGenericRule{Input: "keep_if_not_exist", Value: ""},
					},
				},
			},
		},
		cfs.Parameters[2].Transform.Filter.Rules,
	)
	assert.Equal(
		[]api.TransformFilterRule{
			{
				Type: api.KeepEntryAllSatisfied,
				KeepEntryAllSatisfied: []*api.KeepEntryRule{
					{
						Type:      api.KeepEntryIfEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "namespace", Value: "C"},
					},
					{
						Type:      api.KeepEntryIfEqual,
						KeepEntry: &api.TransformFilterGenericRule{Input: "workload", Value: "C1"},
					},
				},
			},
		},
		cfs.Parameters[3].Transform.Filter.Rules,
	)
}
