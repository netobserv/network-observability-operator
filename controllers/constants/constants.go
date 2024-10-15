// Package constants defines some values that are shared across multiple packages
package constants

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DefaultOperatorNamespace = "netobserv"
	OperatorName             = "netobserv-operator"
	WebhookPort              = 9443
	FLPName                  = "flowlogs-pipeline"
	FLPPortName              = "flp" // must be <15 chars
	FLPMetricsPort           = 9401
	PluginName               = "netobserv-plugin"

	// EBPFAgentName and other constants for it
	EBPFAgentName                     = "netobserv-ebpf-agent"
	EBPFAgentMetricsSvcName           = "ebpf-agent-svc-prom"
	EBPFAgentMetricsSvcMonitoringName = "ebpf-agent-svc-monitor"
	EBPFAgentPromoAlertRule           = "ebpf-agent-prom-alert"
	EBPFPrivilegedNSSuffix            = "-privileged"
	EBPFServiceAccount                = EBPFAgentName
	EBPFSecurityContext               = EBPFAgentName
	EBPFMetricPort                    = 9400

	OpenShiftCertificateAnnotation = "service.beta.openshift.io/serving-cert-secret-name"

	// PodConfigurationDigest is an annotation name to facilitate pod restart after
	// any external configuration change
	AnnotationDomain        = "flows.netobserv.io"
	PodConfigurationDigest  = AnnotationDomain + "/config-digest"
	PodWatchedSuffix        = AnnotationDomain + "/watched-"
	ConversionAnnotation    = AnnotationDomain + "/conversion-data"
	NamespaceCopyAnnotation = AnnotationDomain + "/copied-from"

	TokensPath = "/var/run/secrets/tokens/"

	ClusterNameLabelName = "K8S_ClusterName"

	MonitoringNamespace      = "openshift-monitoring"
	MonitoringServiceAccount = "prometheus-k8s"
	UWMonitoringNamespace    = "openshift-user-workload-monitoring"
	ConsoleNamespace         = "openshift-console"

	// Roles
	LokiCRWriter  = "netobserv-writer"
	LokiCRBWriter = "netobserv-writer-flp"
	LokiCRReader  = "netobserv-reader"
	PromCRReader  = "netobserv-metrics-reader"

	EnvTestConsole = "TEST_CONSOLE"
)

var LokiIndexFields = []string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "SrcK8S_Type", "DstK8S_Namespace", "DstK8S_OwnerName", "DstK8S_Type", "K8S_FlowLayer", "FlowDirection"}
var LokiConnectionIndexFields = []string{"_RecordType"}
var LokiZoneIndexFields = []string{"SrcK8S_Zone", "DstK8S_Zone"}
var FlowCollectorName = types.NamespacedName{Name: "cluster"}
var EnvNoHTTP2 = corev1.EnvVar{
	Name:  "GODEBUG",
	Value: "http2server=0",
}
