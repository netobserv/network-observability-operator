package cni

import (
	v1 "k8s.io/api/core/v1"
)

type Plugin interface {
	GetNodeIPs(node *v1.Node) []string
}
