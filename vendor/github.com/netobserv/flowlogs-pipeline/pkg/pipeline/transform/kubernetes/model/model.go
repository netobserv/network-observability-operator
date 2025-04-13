package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KindNode    = "Node"
	KindPod     = "Pod"
	KindService = "Service"
)

// ResourceMetaData contains precollected metadata for Pods, Nodes and Services.
// Not all the fields are populated for all the above types. To save
// memory, we just keep in memory the necessary data for each Type.
// For more information about which fields are set for each type, please
// refer to the instantiation function of the respective informers.
type ResourceMetaData struct {
	// Informers need that internal object is an ObjectMeta instance
	metav1.ObjectMeta
	Kind             string
	OwnerName        string
	OwnerKind        string
	HostName         string
	HostIP           string
	NetworkName      string
	IPs              []string
	SecondaryNetKeys []string
}
