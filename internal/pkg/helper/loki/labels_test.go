package loki

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestDefaultLokiLabels(t *testing.T) {
	defaultLabels, err := GetLabels(nil)
	assert.Equal(t, err, nil)
	assert.Equal(t, defaultLabels, []string{
		"SrcK8S_Namespace",
		"SrcK8S_OwnerName",
		"SrcK8S_Type",
		"DstK8S_Namespace",
		"DstK8S_OwnerName",
		"DstK8S_Type",
		"K8S_FlowLayer",
		"FlowDirection",
		"UdnId",
	})
}

func TestAllLokiLabels(t *testing.T) {
	outputRecordTypes := flowslatest.LogTypeAll
	defaultLabels, err := GetLabels(&flowslatest.FlowCollectorFLP{
		LogTypes:               &outputRecordTypes,
		MultiClusterDeployment: ptr.To(true),
		AddZone:                ptr.To(true),
	})
	assert.Equal(t, err, nil)
	assert.Equal(t, defaultLabels, []string{
		"SrcK8S_Namespace",
		"SrcK8S_OwnerName",
		"SrcK8S_Type",
		"DstK8S_Namespace",
		"DstK8S_OwnerName",
		"DstK8S_Type",
		"K8S_FlowLayer",
		"FlowDirection",
		"UdnId",
		"_RecordType",
		"K8S_ClusterName",
		"SrcK8S_Zone",
		"DstK8S_Zone",
	})
}
