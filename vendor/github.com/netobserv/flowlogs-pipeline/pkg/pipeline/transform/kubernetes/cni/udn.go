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
	ovnAnnotation = "k8s.ovn.org/pod-networks"
)

type UDNHandler struct {
}

func UDNKey(label, ip string) SecondaryNetKey {
	key := label + "~" + ip
	return SecondaryNetKey{NetworkName: label, Key: key}
}

func (m *UDNHandler) Manages(indexKey string) bool {
	return indexKey == "udn"
}

func (m *UDNHandler) BuildKeys(flow config.GenericMap, rule *api.K8sRule) []SecondaryNetKey {
	var keys []SecondaryNetKey

	var ip string
	var udns []string
	var ok bool
	if len(rule.IPField) > 0 {
		ip, ok = flow.LookupString(rule.IPField)
		if !ok {
			return nil
		}
	}
	if len(rule.UDNsField) > 0 {
		v, ok := flow[rule.UDNsField]
		if !ok {
			return nil
		}
		udns, ok = v.([]string)
		if !ok || len(udns) == 0 {
			return nil
		}
	}

	for _, udn := range udns {
		if udn != "" && udn != "default" {
			keys = append(keys, UDNKey(udn, ip))
		}
	}

	return keys
}

func (m *UDNHandler) GetPodUniqueKeys(pod *v1.Pod) ([]string, error) {
	// Example:
	// k8s.ovn.org/pod-networks: '{"default":{"ip_addresses":["10.128.2.20/23"],"mac_address":"0a:58:0a:80:02:14","routes":[{"dest":"10.128.0.0/14","nextHop":"10.128.2.1"},{"dest":"100.64.0.0/16","nextHop":"10.128.2.1"}],"ip_address":"10.128.2.20/23","role":"infrastructure-locked"},"mesh-arena/primary-udn":{"ip_addresses":["10.200.200.12/24"],"mac_address":"0a:58:0a:c8:c8:0c","gateway_ips":["10.200.200.1"],"routes":[{"dest":"172.30.0.0/16","nextHop":"10.200.200.1"},{"dest":"100.65.0.0/16","nextHop":"10.200.200.1"}],"ip_address":"10.200.200.12/24","gateway_ip":"10.200.200.1","tunnel_id":16,"role":"primary"}}'
	if statusAnnotationJSON, ok := pod.Annotations[ovnAnnotation]; ok {
		var annot map[string]map[string]any
		if err := json.Unmarshal([]byte(statusAnnotationJSON), &annot); err != nil {
			return nil, fmt.Errorf("failed to index from OVN annotation, cannot read annotation %s: %w", ovnAnnotation, err)
		}
		var keys []string
		for label, info := range annot {
			if label != "default" {
				if rawip, ok := info["ip_address"]; ok {
					if ip, ok := rawip.(string); ok {
						// IP has a CIDR prefix (bug??)
						parts := strings.SplitN(ip, "/", 2)
						if len(parts) > 0 {
							key := UDNKey(label, parts[0])
							keys = append(keys, key.Key)
						}
					}
				}
			}
		}
		return keys, nil
	}
	// Annotation not present => just ignore, no error
	return nil, nil
}
