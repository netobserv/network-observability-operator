// Package constants defines some values that are shared across multiple packages
package constants

const (
	DefaultOperatorNamespace = "network-observability"
	FLPName                  = "flowlogs-pipeline"
	FLPPortName              = "flp" // must be <15 chars
	PluginName               = "network-observability-plugin"
	DeploymentKind           = "Deployment"
	DaemonSetKind            = "DaemonSet"

	// EBPFAgentName and other constants for it
	EBPFAgentName          = "netobserv-ebpf-agent"
	EBPFPrivilegedNSSuffix = "-privileged"
	EBPFServiceAccount     = EBPFAgentName
	EBPFSecurityContext    = EBPFAgentName

	// Operators group and subscriptions
	OperatorGroup      = "OperatorGroup"
	GrafanaOperator    = "grafana-operator"
	StrimziOperator    = "strimzi-kafka-operator"
	LokiOperator       = "loki-operator"
	PrometheusOperator = "prometheus"
	SubscriptionKind   = "Subscription"

	// Operators instances
	KafkaName      = "kafka"
	GrafanaName    = "grafana"
	LokiName       = "loki"
	PrometheusName = "prometheus"

	// parf-of label
	ObservabilityName = "netobserv"
)

var LokiIndexFields = []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection"}
