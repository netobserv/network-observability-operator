// Package constants defines some values that are shared across multiple packages
package constants

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DefaultOperatorNamespace = "netobserv"
	OperatorName             = "netobserv-operator"
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

	// Loki roles
	LokiCRWriter  = "netobserv-writer"
	LokiCRBWriter = "netobserv-writer-flp"
	LokiCRReader  = "netobserv-reader"

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

// OpenTelemetryDefaultTransformRules defined the default Open Telemetry format
// See https://github.com/rhobs/observability-data-model/blob/main/network-observability.md#format-proposal
var OpenTelemetryDefaultTransformRules = []api.GenericTransformRule{{
	Input:  "SrcAddr",
	Output: "source.address",
}, {
	Input:  "SrcMac",
	Output: "source.mac",
}, {
	Input:  "SrcHostIP",
	Output: "source.host.address",
}, {
	Input:  "SrcK8S_HostName",
	Output: "source.k8s.node.name",
}, {
	Input:  "SrcPort",
	Output: "source.port",
}, {
	Input:  "SrcK8S_Name",
	Output: "source.k8s.name",
}, {
	Input:  "SrcK8S_Type",
	Output: "source.k8s.kind",
}, {
	Input:  "SrcK8S_OwnerName",
	Output: "source.k8s.owner.name",
}, {
	Input:  "SrcK8S_OwnerType",
	Output: "source.k8s.owner.kind",
}, {
	Input:  "SrcK8S_Namespace",
	Output: "source.k8s.namespace.name",
}, {
	Input:  "SrcK8S_HostIP",
	Output: "source.k8s.host.address",
}, {
	Input:  "SrcK8S_HostName",
	Output: "source.k8s.host.name",
}, {
	Input:  "SrcK8S_Zone",
	Output: "source.zone",
}, {
	Input:  "DstAddr",
	Output: "destination.address",
}, {
	Input:  "DstMac",
	Output: "destination.mac",
}, {
	Input:  "DstHostIP",
	Output: "destination.host.address",
}, {
	Input:  "DstK8S_HostName",
	Output: "destination.k8s.node.name",
}, {
	Input:  "DstPort",
	Output: "destination.port",
}, {
	Input:  "DstK8S_Name",
	Output: "destination.k8s.name",
}, {
	Input:  "DstK8S_Type",
	Output: "destination.k8s.kind",
}, {
	Input:  "DstK8S_OwnerName",
	Output: "destination.k8s.owner.name",
}, {
	Input:  "DstK8S_OwnerType",
	Output: "destination.k8s.owner.kind",
}, {
	Input:  "DstK8S_Namespace",
	Output: "destination.k8s.namespace.name",
}, {
	Input:  "DstK8S_HostIP",
	Output: "destination.k8s.host.address",
}, {
	Input:  "DstK8S_HostName",
	Output: "destination.k8s.host.name",
}, {
	Input:  "DstK8S_Zone",
	Output: "destination.zone",
}, {
	Input:  "Bytes",
	Output: "bytes",
}, {
	Input:  "Packets",
	Output: "packets",
}, {
	Input:  "Proto",
	Output: "protocol",
}, {
	Input:  "Flags",
	Output: "tcp.flags",
}, {
	Input:  "TimeFlowRttNs",
	Output: "tcp.rtt",
}, {
	Input:  "Interfaces",
	Output: "interface.names",
}, {
	Input:  "IfDirections",
	Output: "interface.directions",
}, {
	Input:  "FlowDirection",
	Output: "host.direction",
}, {
	Input:  "DnsErrno",
	Output: "dns.errno",
}, {
	Input:  "DnsFlags",
	Output: "dns.flags",
}, {
	Input:  "DnsFlagsResponseCode",
	Output: "dns.responsecode",
}, {
	Input:  "DnsId",
	Output: "dns.id",
}, {
	Input:  "DnsLatencyMs",
	Output: "dns.latency",
}, {
	Input:  "Dscp",
	Output: "dscp",
}, {
	Input:  "IcmpCode",
	Output: "icmp.code",
}, {
	Input:  "IcmpType",
	Output: "icmp.type",
}, {
	Input:  "K8S_ClusterName",
	Output: "k8s.cluster.name",
}, {
	Input:  "K8S_FlowLayer",
	Output: "k8s.layer",
}, {
	Input:  "PktDropBytes",
	Output: "drops.bytes",
}, {
	Input:  "PktDropPackets",
	Output: "drops.packets",
}, {
	Input:  "PktDropLatestDropCause",
	Output: "drops.latestcause",
}, {
	Input:  "PktDropLatestFlags",
	Output: "drops.latestflags",
}, {
	Input:  "PktDropLatestState",
	Output: "drops.lateststate",
}, {
	Input:  "TimeFlowEndMs",
	Output: "timeflowend",
}, {
	Input:  "TimeFlowStartMs",
	Output: "timeflowstart",
}, {
	Input:  "TimeReceived",
	Output: "timereceived",
}}
