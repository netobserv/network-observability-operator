package cni

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	v1 "k8s.io/api/core/v1"
)

const (
	statusAnnotation = "k8s.v1.cni.cncf.io/network-status"
	// Index names
	indexIP        = "ip"
	indexMAC       = "mac"
	indexInterface = "interface"
)

type MultusHandler struct {
}

type SecondaryNetKey struct {
	NetworkName string
	Key         string
}

func (m *MultusHandler) Manages(indexKey string) bool {
	return indexKey == indexIP || indexKey == indexMAC || indexKey == indexInterface
}

func (m *MultusHandler) BuildKeys(flow config.GenericMap, rule *api.K8sRule, secNets []api.SecondaryNetwork) []SecondaryNetKey {
	if len(secNets) == 0 {
		return nil
	}
	var keys []SecondaryNetKey
	for _, sn := range secNets {
		snKeys := m.buildSNKeys(flow, rule, &sn)
		if snKeys != nil {
			keys = append(keys, snKeys...)
		}
	}
	return keys
}

func (m *MultusHandler) buildSNKeys(flow config.GenericMap, rule *api.K8sRule, sn *api.SecondaryNetwork) []SecondaryNetKey {
	var keys []SecondaryNetKey

	var ip, mac string
	var interfaces []string
	if _, ok := sn.Index[indexIP]; ok && len(rule.IPField) > 0 {
		ip, ok = flow.LookupString(rule.IPField)
		if !ok {
			return nil
		}
	}
	if _, ok := sn.Index[indexMAC]; ok && len(rule.MACField) > 0 {
		mac, ok = flow.LookupString(rule.MACField)
		if !ok {
			return nil
		}
	}
	if _, ok := sn.Index[indexInterface]; ok && len(rule.InterfacesField) > 0 {
		v, ok := flow[rule.InterfacesField]
		if !ok {
			return nil
		}
		interfaces, ok = v.([]string)
		if !ok {
			return nil
		}
	}
	if mac == "" && ip == "" && len(interfaces) == 0 {
		return nil
	}

	macIP := "~" + ip + "~" + mac
	if interfaces == nil {
		return []SecondaryNetKey{{NetworkName: sn.Name, Key: macIP}}
	}
	for _, intf := range interfaces {
		keys = append(keys, SecondaryNetKey{NetworkName: sn.Name, Key: intf + macIP})
	}

	return keys
}

func (m *MultusHandler) GetPodUniqueKeys(pod *v1.Pod, secNets []api.SecondaryNetwork) ([]string, error) {
	if len(secNets) == 0 {
		return nil, nil
	}
	// Cf https://k8snetworkplumbingwg.github.io/multus-cni/docs/quickstart.html#network-status-annotations
	if statusAnnotationJSON, ok := pod.Annotations[statusAnnotation]; ok {
		var networks []NetStatItem
		if err := json.Unmarshal([]byte(statusAnnotationJSON), &networks); err != nil {
			return nil, fmt.Errorf("failed to index from network-status annotation, cannot read annotation %s: %w", statusAnnotation, err)
		}
		var keys []string
		for _, network := range networks {
			for _, snConfig := range secNets {
				if snConfig.Name == network.Name {
					keys = append(keys, network.Keys(snConfig)...)
				}
			}
		}
		return keys, nil
	}
	// Annotation not present => just ignore, no error
	return nil, nil
}

type NetStatItem struct {
	Name      string   `json:"name"`
	Interface string   `json:"interface"`
	IPs       []string `json:"ips"`
	MAC       string   `json:"mac"`
}

func (n *NetStatItem) Keys(snConfig api.SecondaryNetwork) []string {
	var mac, intf string
	if _, ok := snConfig.Index[indexMAC]; ok {
		mac = n.MAC
	}
	if _, ok := snConfig.Index[indexInterface]; ok {
		intf = n.Interface
	}
	if _, ok := snConfig.Index[indexIP]; ok {
		var keys []string
		for _, ip := range n.IPs {
			keys = append(keys, key(intf, ip, mac))
		}
		return keys
	}
	// Ignore IP
	return []string{key(intf, "", mac)}
}

func key(intf, ip, mac string) string {
	return intf + "~" + ip + "~" + strings.ToUpper(mac)
}
