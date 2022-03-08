// Package constants defines some values that are shared across multiple packages
package constants

const (
	FLPName        = "flowlogs-pipeline"
	FLPPortName    = "flp" // must be <15 chars
	PluginName     = "network-observability-plugin"
	DeploymentKind = "Deployment"
	DaemonSetKind  = "DaemonSet"
)

var LokiIndexFields = []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection"}
