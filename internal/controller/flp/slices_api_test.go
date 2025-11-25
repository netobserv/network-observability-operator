package flp

import (
	"encoding/json"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	sliceslatest "github.com/netobserv/network-observability-operator/api/flowcollectorslice/v1alpha1"
	"github.com/netobserv/network-observability-operator/api/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/internal/controller/flp/fmstatus"
	"github.com/netobserv/network-observability-operator/internal/controller/reconcilers"
	"github.com/netobserv/network-observability-operator/internal/pkg/cluster"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
	"github.com/netobserv/network-observability-operator/internal/pkg/manager/status"
)

var (
	adminSubnets = []flowslatest.SubnetLabel{
		{
			Name:  "admin",
			CIDRs: []string{"10.0.0.0/16"},
		},
	}
	autoSubnets = []flowslatest.SubnetLabel{
		{
			Name:  "test",
			CIDRs: []string{"1.2.3.4/32", "10.0.0.0/8"},
		},
	}
	slicez = []sliceslatest.FlowCollectorSlice{
		{
			ObjectMeta: v1.ObjectMeta{Name: "a", Namespace: "ns-a"},
		},
		{
			ObjectMeta: v1.ObjectMeta{Name: "b1", Namespace: "ns-b"},
			Spec: sliceslatest.FlowCollectorSliceSpec{
				SubnetLabels: []sliceslatest.SubnetLabel{
					{
						Name:  "my-override",
						CIDRs: []string{"10.10.0.0/16"},
					},
					{
						Name:  "my-label",
						CIDRs: []string{"100.0.0.0/24"},
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{Name: "b2", Namespace: "ns-b"},
			Spec: sliceslatest.FlowCollectorSliceSpec{
				SubnetLabels: []sliceslatest.SubnetLabel{
					{
						Name:  "my-label-2",
						CIDRs: []string{"100.51.0.0/24"},
					},
				},
			},
		},
	}
)

func getConfiguredFiltersAndSubnets(cm *corev1.ConfigMap) ([]api.TransformFilterRule, []api.NetworkTransformSubnetLabel) {
	var cfs config.Root
	err := json.Unmarshal([]byte(cm.Data[configFile]), &cfs)
	if err != nil {
		return nil, nil
	}
	var filters []api.TransformFilterRule
	var subnetLabels []api.NetworkTransformSubnetLabel
	for _, stage := range cfs.Parameters {
		if stage.Transform != nil && stage.Name == "enrich" {
			subnetLabels = stage.Transform.Network.SubnetLabels
		}
		if stage.Transform != nil && stage.Name == "filters" {
			filters = stage.Transform.Filter.Rules
		}
	}
	return filters, subnetLabels
}

func defaultBuilderWithSlices(cfg *flowslatest.SlicesConfig) (monolithBuilder, error) {
	fc := getConfig()
	fc.Processor.SlicesConfig = cfg
	fc.Processor.SubnetLabels.CustomLabels = adminSubnets
	info := reconcilers.Common{Namespace: "namespace", Loki: &helper.LokiConfig{}, ClusterInfo: &cluster.Info{}}
	return newMonolithBuilder(info.NewInstance(image, status.Instance{}), &fc, &v1alpha1.FlowMetricList{}, slicez, autoSubnets)
}

func TestSlicesDisabled(t *testing.T) {
	fmstatus.Reset()
	b, err := defaultBuilderWithSlices(&flowslatest.SlicesConfig{Enable: false})
	assert.NoError(t, err)
	cm, _, _, err := b.configMaps()
	assert.NoError(t, err)
	filters, subnets := getConfiguredFiltersAndSubnets(cm)
	assert.Nil(t, filters)
	assert.Equal(t, []api.NetworkTransformSubnetLabel{
		{
			Name:  "admin",
			CIDRs: []string{"10.0.0.0/16"},
		},
		{
			Name:  "test",
			CIDRs: []string{"1.2.3.4/32", "10.0.0.0/8"},
		},
	}, subnets)
}

func TestSlicesEnablesCollectAll(t *testing.T) {
	fmstatus.Reset()
	b, err := defaultBuilderWithSlices(&flowslatest.SlicesConfig{
		Enable:              true,
		CollectionMode:      flowslatest.CollectionAlwaysCollect,
		NamespacesAllowList: []string{"should-be-ignored"},
	})
	assert.NoError(t, err)
	cm, _, _, err := b.configMaps()
	assert.NoError(t, err)
	filters, subnets := getConfiguredFiltersAndSubnets(cm)
	assert.Nil(t, filters)
	assert.Equal(t, []api.NetworkTransformSubnetLabel{
		{
			Name:  "admin",
			CIDRs: []string{"10.0.0.0/16"},
		},
		{
			Name:  "my-override",
			CIDRs: []string{"10.10.0.0/16"},
		},
		{
			Name:  "my-label",
			CIDRs: []string{"100.0.0.0/24"},
		},
		{
			Name:  "my-label-2",
			CIDRs: []string{"100.51.0.0/24"},
		},
		{
			Name:  "test",
			CIDRs: []string{"1.2.3.4/32", "10.0.0.0/8"},
		},
	}, subnets)
}

func TestSlicesEnablesWhitelist(t *testing.T) {
	fmstatus.Reset()
	b, err := defaultBuilderWithSlices(&flowslatest.SlicesConfig{
		Enable:              true,
		CollectionMode:      flowslatest.CollectionAllowList,
		NamespacesAllowList: []string{"should-be-filtered", "/should-.*/"},
	})
	assert.NoError(t, err)
	cm, _, _, err := b.configMaps()
	assert.NoError(t, err)
	filters, subnets := getConfiguredFiltersAndSubnets(cm)
	assert.Equal(t, []api.TransformFilterRule{
		{
			Type:           api.KeepEntryQuery,
			KeepEntryQuery: `SrcK8S_Namespace="should-be-filtered" or DstK8S_Namespace="should-be-filtered"`,
		},
		{
			Type:           api.KeepEntryQuery,
			KeepEntryQuery: `SrcK8S_Namespace=~"should-.*" or DstK8S_Namespace=~"should-.*"`,
		},
		{
			Type:           api.KeepEntryQuery,
			KeepEntryQuery: `SrcK8S_Namespace="ns-a" or DstK8S_Namespace="ns-a"`,
		},
		{
			Type:           api.KeepEntryQuery,
			KeepEntryQuery: `SrcK8S_Namespace="ns-b" or DstK8S_Namespace="ns-b"`,
		},
	}, filters)
	assert.Equal(t, []api.NetworkTransformSubnetLabel{
		{
			Name:  "admin",
			CIDRs: []string{"10.0.0.0/16"},
		},
		{
			Name:  "my-override",
			CIDRs: []string{"10.10.0.0/16"},
		},
		{
			Name:  "my-label",
			CIDRs: []string{"100.0.0.0/24"},
		},
		{
			Name:  "my-label-2",
			CIDRs: []string{"100.51.0.0/24"},
		},
		{
			Name:  "test",
			CIDRs: []string{"1.2.3.4/32", "10.0.0.0/8"},
		},
	}, subnets)
}
