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
	ControllerName           = "netobserv-controller-manager"
	WebhookPort              = 9443
	K8sAPIServerPort         = 6443
	FLPName                  = "flowlogs-pipeline"
	FLPShortName             = "flp"
	FLPPortName              = "flp" // must be <15 chars
	FLPMetricsSvcName        = FLPName + "-prom"
	FLPTransfoName           = FLPName + "-transformer"
	FLPTransfoMetricsSvcName = FLPTransfoName + "-prom"
	FLPMetricsPort           = 9401
	PluginName               = "netobserv-plugin"
	StaticPluginName         = "netobserv-plugin-static"
	PluginShortName          = "plugin"
	LokiDev                  = "loki"

	// EBPFAgentName and other constants for it
	EBPFAgentName                     = "netobserv-ebpf-agent"
	EBPFAgentMetricsSvcName           = "ebpf-agent-svc-prom"
	EBPFAgentMetricsSvcMonitoringName = "ebpf-agent-svc-monitor"
	EBPFAgentPromAlertRule            = "ebpf-agent-prom-alert"
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

	KubeSystemNamespace             = "kube-system"
	OpenShiftAPIServerNamespace     = "openshift-apiserver"
	OpenShiftKubeAPIServerNamespace = "openshift-kube-apiserver"
	MonitoringNamespace             = "openshift-monitoring"
	MonitoringServiceAccount        = "prometheus-k8s"
	UWMonitoringNamespace           = "openshift-user-workload-monitoring"
	ConsoleNamespace                = "openshift-console"
	DNSNamespace                    = "openshift-dns"

	// [Cluster]Roles, must match names in config/rbac/component_roles.yaml (without netobserv- prefix)
	LokiWriterRole         ClusterRoleName = "netobserv-loki-writer"
	LokiReaderRole         ClusterRoleName = "netobserv-loki-reader"
	PromReaderRole         ClusterRoleName = "netobserv-metrics-reader"
	ExposeMetricsRole      RoleName        = "netobserv-expose-metrics"
	FLPInformersRole       ClusterRoleName = "netobserv-informers"
	HostNetworkRole        ClusterRoleName = "netobserv-hostnetwork"
	ConsoleTokenReviewRole ClusterRoleName = "netobserv-token-review"
	ConfigWatcherRole      RoleName        = "netobserv-config-watcher"
)

var FlowCollectorName = types.NamespacedName{Name: "cluster"}
var EnvNoHTTP2 = corev1.EnvVar{
	Name:  "GODEBUG",
	Value: "http2server=0",
}
