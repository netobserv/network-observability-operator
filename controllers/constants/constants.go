// Package constants defines some values that are shared across multiple packages
package constants

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
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

	// PodConfigurationDigest is an annotation name to facilitate pod restart after
	// any external configuration change
	AnnotationDomain        = "flows.netobserv.io"
	PodConfigurationDigest  = AnnotationDomain + "/config-digest"
	PodWatchedSuffix        = AnnotationDomain + "/watched-"
	ConversionAnnotation    = AnnotationDomain + "/conversion-data"
	NamespaceCopyAnnotation = AnnotationDomain + "/copied-from"

	TokensPath = "/var/run/secrets/tokens/"

	FlowLogType       = "flowLog"
	NewConnectionType = "newConnection"
	HeartbeatType     = "heartbeat"
	EndConnectionType = "endConnection"

	MonitoringNamespace      = "openshift-monitoring"
	MonitoringServiceAccount = "prometheus-k8s"
)

var LokiIndexFields = []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "SrcK8S_Type", "DstK8S_Namespace", "DstK8S_OwnerName", "DstK8S_Type", "FlowDirection", "Duplicate"}
var LokiConnectionIndexFields = []string{"_RecordType"}
var FlowCollectorName = types.NamespacedName{Name: "cluster"}
var EnvNoHTTP2 = corev1.EnvVar{
	Name:  "GODEBUG",
	Value: "http2server=0",
}
