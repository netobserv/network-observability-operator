/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cni

import (
	"encoding/json"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	ovnSubnetAnnotation = "k8s.ovn.org/node-subnets"
)

type OVNPlugin struct {
	Plugin
}

func (o *OVNPlugin) GetNodeIPs(node *v1.Node) []string {
	// Add IP that is used in OVN for some traffic on mp0 interface
	// (no IP / error returned when not using ovn-k)
	ip, err := findOvnMp0IP(node.Annotations)
	if err != nil {
		// Log the error as Info, do not block other ips indexing
		log.Infof("failed to index OVN mp0 IP: %v", err)
	} else if ip != "" {
		return []string{ip}
	}
	return nil
}

func unmarshalOVNAnnotation(annot []byte) (string, error) {
	// Depending on OVN (OCP) version, the annotation might be JSON-encoded as a string (legacy), or an array of strings
	var subnetsAsArray map[string][]string
	err := json.Unmarshal(annot, &subnetsAsArray)
	if err == nil {
		if subnets, ok := subnetsAsArray["default"]; ok {
			if len(subnets) > 0 {
				return subnets[0], nil
			}
		}
		return "", fmt.Errorf("unexpected content for annotation %s: %s", ovnSubnetAnnotation, annot)
	}

	var subnetsAsString map[string]string
	err = json.Unmarshal(annot, &subnetsAsString)
	if err == nil {
		if subnet, ok := subnetsAsString["default"]; ok {
			return subnet, nil
		}
		return "", fmt.Errorf("unexpected content for annotation %s: %s", ovnSubnetAnnotation, annot)
	}

	return "", fmt.Errorf("cannot read annotation %s: %w", ovnSubnetAnnotation, err)
}

func findOvnMp0IP(annotations map[string]string) (string, error) {
	if subnetsJSON, ok := annotations[ovnSubnetAnnotation]; ok {
		subnet, err := unmarshalOVNAnnotation([]byte(subnetsJSON))
		if err != nil {
			return "", err
		}
		// From subnet like 10.128.0.0/23, we want to index IP 10.128.0.2
		ip0, _, err := net.ParseCIDR(subnet)
		if err != nil {
			return "", err
		}
		ip4 := ip0.To4()
		if ip4 == nil {
			// TODO: what's the rule with ipv6?
			return "", nil
		}
		return fmt.Sprintf("%d.%d.%d.%d", ip4[0], ip4[1], ip4[2], ip4[3]+2), nil
	}
	// Annotation not present (expected if not using ovn-kubernetes) => just ignore, no error
	return "", nil
}
