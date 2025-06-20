package loki

import (
	_ "embed"
	"encoding/json"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
)

//go:embed loki-labels.json
var rawLokiLabels []byte
var lokiLabels *map[LabelsType][]string

type LabelsType string

const (
	Default      LabelsType = "default"
	Conntrack    LabelsType = "conntrack"
	MultiCluster LabelsType = "multiCluster"
	Zones        LabelsType = "zones"
)

func GetLabelsPerType() (map[LabelsType][]string, error) {
	if lokiLabels == nil {
		cfg := make(map[LabelsType][]string)
		err := json.Unmarshal(rawLokiLabels, &cfg)
		if err != nil {
			return cfg, err
		}
		lokiLabels = &cfg
	}
	return *lokiLabels, nil
}

func GetLabels(desired *flowslatest.FlowCollectorFLP) ([]string, error) {
	labelsPerType, err := GetLabelsPerType()
	if err != nil {
		return []string{}, err
	}
	labels := labelsPerType[Default]

	if helper.IsConntrack(desired) {
		labels = append(labels, labelsPerType[Conntrack]...)
	}

	if helper.IsMultiClusterEnabled(desired) {
		labels = append(labels, labelsPerType[MultiCluster]...)
	}

	if helper.IsZoneEnabled(desired) {
		labels = append(labels, labelsPerType[Zones]...)
	}

	return labels, nil
}
