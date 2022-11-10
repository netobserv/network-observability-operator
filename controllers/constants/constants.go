// Package constants defines some values that are shared across multiple packages
package constants

const (
	Cluster = "cluster"

	DefaultOperatorNamespace = "netobserv"
	OperatorName             = "netobserv-operator"
	FLPName                  = "flowlogs-pipeline"
	FLPPortName              = "flp" // must be <15 chars
	PluginName               = "netobserv-plugin"

	// EBPFAgentName and other constants for it
	EBPFAgentName          = "netobserv-ebpf-agent"
	EBPFPrivilegedNSSuffix = "-privileged"
	EBPFServiceAccount     = EBPFAgentName
	EBPFSecurityContext    = EBPFAgentName

	OpenShiftCertificateAnnotation = "service.beta.openshift.io/serving-cert-secret-name"

	KafkaCRDName      = "kafkas.kafka.strimzi.io"
	KafkaTopicCRDName = "kafkatopics.kafka.strimzi.io"
	KafkaUserCRDName  = "kafkausers.kafka.strimzi.io"

	LokiCRDName       = "lokistacks.loki.grafana.com"
	KafkaOperator     = "kafka"
	KafkaInstanceName = "kafka-cluster"
	LokiOperator      = "loki"
	LokiInstanceName  = "lokistack"
)

var LokiIndexFields = []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection"}
