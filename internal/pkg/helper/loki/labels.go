package loki

import (
	_ "embed"
	"encoding/json"
	"slices"

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
	UDN          LabelsType = "udn"
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

func GetLabels(desired *flowslatest.FlowCollectorSpec) ([]string, error) {
	labelsPerType, err := GetLabelsPerType()
	if err != nil {
		return []string{}, err
	}
	var labels []string
	var excluding []string
	if desired.Loki.Advanced != nil {
		excluding = desired.Loki.Advanced.ExcludeLabels
	}
	labels = addExcluding(labels, labelsPerType[Default], excluding)

	if helper.IsConntrack(&desired.Processor) {
		labels = addExcluding(labels, labelsPerType[Conntrack], excluding)
	}

	if helper.IsMultiClusterEnabled(&desired.Processor) {
		labels = addExcluding(labels, labelsPerType[MultiCluster], excluding)
	}

	if helper.IsZoneEnabled(&desired.Processor) {
		labels = addExcluding(labels, labelsPerType[Zones], excluding)
	}

	if helper.IsUDNMappingEnabled(&desired.Agent.EBPF) {
		labels = addExcluding(labels, labelsPerType[UDN], excluding)
	}

	return labels, nil
}

func addExcluding(list []string, toAdd []string, excluding []string) []string {
	for _, add := range toAdd {
		if !slices.Contains(excluding, add) {
			list = append(list, add)
		}
	}
	return list
}
