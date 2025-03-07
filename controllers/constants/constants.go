// Package constants defines some values that are shared across multiple packages
package constants

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterRoleName string
type RoleName string

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

	// [Cluster]Roles, must match names in config/rbac/component_roles.yaml (without netobserv- prefix)
	LokiWriterRole         ClusterRoleName = "netobserv-writer"
	LokiReaderRole         ClusterRoleName = "netobserv-reader"
	PromReaderRole         ClusterRoleName = "netobserv-metrics-reader"
	ExposeMetricsRole      RoleName        = "netobserv-expose-metrics"
	FLPInformersRole       ClusterRoleName = "netobserv-informers"
	HostNetworkRole        ClusterRoleName = "netobserv-hostnetwork"
	ConsoleTokenReviewRole ClusterRoleName = "netobserv-token-review"
	ConfigWatcherRole      RoleName        = "netobserv-config-watcher"

	ControllerBaseImageIndex    = 0
	EBPFAgentByteCodeImageIndex = 1
	EnvTestConsole              = "TEST_CONSOLE"
)

var FlowCollectorName = types.NamespacedName{Name: "cluster"}
var EnvNoHTTP2 = corev1.EnvVar{
	Name:  "GODEBUG",
	Value: "http2server=0",
}
