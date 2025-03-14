package ebpf

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	ebpfconfig "github.com/netobserv/netobserv-ebpf-agent/pkg/agent"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	"github.com/netobserv/network-observability-operator/pkg/watchers"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	envCacheActiveTimeout         = "CACHE_ACTIVE_TIMEOUT"
	envCacheMaxFlows              = "CACHE_MAX_FLOWS"
	envExcludeInterfaces          = "EXCLUDE_INTERFACES"
	envInterfaces                 = "INTERFACES"
	envAgentIP                    = "AGENT_IP"
	envFlowsTargetHost            = "TARGET_HOST"
	envFlowsTargetPort            = "TARGET_PORT"
	envSampling                   = "SAMPLING"
	envExport                     = "EXPORT"
	envKafkaBrokers               = "KAFKA_BROKERS"
	envKafkaTopic                 = "KAFKA_TOPIC"
	envKafkaBatchSize             = "KAFKA_BATCH_SIZE"
	envKafkaBatchMessages         = "KAFKA_BATCH_MESSAGES"
	envKafkaEnableTLS             = "KAFKA_ENABLE_TLS"
	envKafkaTLSInsecureSkipVerify = "KAFKA_TLS_INSECURE_SKIP_VERIFY"
	envKafkaTLSCACertPath         = "KAFKA_TLS_CA_CERT_PATH"
	envKafkaTLSUserCertPath       = "KAFKA_TLS_USER_CERT_PATH"
	envKafkaTLSUserKeyPath        = "KAFKA_TLS_USER_KEY_PATH"
	envKafkaEnableSASL            = "KAFKA_ENABLE_SASL"
	envKafkaSASLType              = "KAFKA_SASL_TYPE"
	envKafkaSASLIDPath            = "KAFKA_SASL_CLIENT_ID_PATH"
	envKafkaSASLSecretPath        = "KAFKA_SASL_CLIENT_SECRET_PATH"
	envLogLevel                   = "LOG_LEVEL"
	envDedupe                     = "DEDUPER"
	dedupeDefault                 = "firstCome"
	envGoMemLimit                 = "GOMEMLIMIT"
	envEnablePktDrop              = "ENABLE_PKT_DROPS"
	envEnableDNSTracking          = "ENABLE_DNS_TRACKING"
	envEnableFlowRTT              = "ENABLE_RTT"
	envEnableNetworkEvents        = "ENABLE_NETWORK_EVENTS_MONITORING"
	envNetworkEventsGroupID       = "NETWORK_EVENTS_MONITORING_GROUP_ID"
	envEnableMetrics              = "METRICS_ENABLE"
	envMetricsPort                = "METRICS_SERVER_PORT"
	envMetricPrefix               = "METRICS_PREFIX"
	envMetricsTLSCertPath         = "METRICS_TLS_CERT_PATH"
	envMetricsTLSKeyPath          = "METRICS_TLS_KEY_PATH"
	envEnableFlowFilter           = "ENABLE_FLOW_FILTER"
	envFilterRules                = "FLOW_FILTER_RULES"
	envEnablePacketTranslation    = "ENABLE_PKT_TRANSLATION"
	envEnableEbpfMgr              = "EBPF_PROGRAM_MANAGER_MODE"
	envEnableUDNMapping           = "ENABLE_UDN_MAPPING"
	envListSeparator              = ","
)

const (
	exportKafka                 = "kafka"
	exportGRPC                  = "grpc"
	kafkaCerts                  = "kafka-certs"
	averageMessageSize          = 100
	bpfTraceMountName           = "bpf-kernel-debug"
	bpfTraceMountPath           = "/sys/kernel/debug"
	bpfNetNSMountName           = "var-run-netns"
	bpfNetNSMountPath           = "/var/run/netns"
	droppedFlowsAlertThreshold  = 100
	ovnObservMountName          = "var-run-ovn"
	ovnObservMountPath          = "/var/run/ovn"
	ovnObservHostMountPath      = "/var/run/ovn-ic"
	ovsMountPath                = "/var/run/openvswitch"
	ovsHostMountPath            = "/var/run/openvswitch"
	ovsMountName                = "var-run-ovs"
	defaultNetworkEventsGroupID = "10"
)

const (
	EnvDedupeJustMark      = "DEDUPER_JUST_MARK"
	EnvDedupeMerge         = "DEDUPER_MERGE"
	envDNSTrackingPort     = "DNS_TRACKING_PORT"
	DedupeJustMarkDefault  = "false"
	DedupeMergeDefault     = "true"
	defaultDNSTrackingPort = "53"
	bpfmanMapsVolumeName   = "bpfman-maps"
	bpfManBpfFSPath        = "/run/netobserv/maps"
)

// AgentController reconciles the status of the eBPF agent Daemonset, as well as the
// associated objects that are required to bind the proper permissions: namespace, service
// accounts, SecurityContextConstraints...
type AgentController struct {
	*reconcilers.Instance
	permissions    permissions.Reconciler
	volumes        volumes.Builder
	promSvc        *corev1.Service
	serviceMonitor *monitoringv1.ServiceMonitor
	prometheusRule *monitoringv1.PrometheusRule
}

func NewAgentController(common *reconcilers.Instance) *AgentController {
	common.Managed.Namespace = common.PrivilegedNamespace()
	agent := AgentController{
		Instance:    common,
		permissions: permissions.NewReconciler(common),
		promSvc:     common.Managed.NewService(constants.EBPFAgentMetricsSvcName),
	}
	if common.ClusterInfo.HasSvcMonitor() {
		agent.serviceMonitor = common.Managed.NewServiceMonitor(constants.EBPFAgentMetricsSvcMonitoringName)
	}
	if common.ClusterInfo.HasPromRule() {
		agent.prometheusRule = common.Managed.NewPrometheusRule(constants.EBPFAgentPromoAlertRule)
	}
	return &agent
}

func (c *AgentController) Reconcile(ctx context.Context, target *flowslatest.FlowCollector) error {
	rlog := log.FromContext(ctx).WithName("ebpf")
	ctx = log.IntoContext(ctx, rlog)
	current, err := c.current(ctx)
	if err != nil {
		return fmt.Errorf("fetching current eBPF agent: %w", err)
	}

	// Retrieve other owned objects
	err = c.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if err := c.permissions.Reconcile(ctx, &target.Spec.Agent.EBPF); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}

	err = c.reconcileMetricsService(ctx, &target.Spec.Agent.EBPF)
	if err != nil {
		return fmt.Errorf("reconciling prometheus service: %w", err)
	}

	desired, err := c.desired(ctx, target, rlog)
	if err != nil {
		return err
	}

	switch helper.DaemonSetChanged(current, desired) {
	case helper.ActionCreate:
		rlog.Info("action: create agent")
		c.Status.SetCreatingDaemonSet(desired)
		err = c.CreateOwned(ctx, desired)
	case helper.ActionUpdate:
		rlog.Info("action: update agent")
		err = c.UpdateIfOwned(ctx, current, desired)
	default:
		rlog.Info("action: nothing to do")
		c.Status.CheckDaemonSetProgress(current)
	}

	if err != nil {
		return err
	}

	if helper.IsAgentFeatureEnabled(&target.Spec.Agent.EBPF, flowslatest.EbpfManager) {
		if err := c.bpfmanAttachNetobserv(ctx, target); err != nil {
			return fmt.Errorf("failed to attach netobserv: %w", err)
		}
	}
	return nil
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.PrivilegedNamespace(),
	}, &agentDS); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w", c.PrivilegedNamespace(), constants.EBPFAgentName, err)
	}
	return &agentDS, nil
}

func newHostPathType(pathType corev1.HostPathType) *corev1.HostPathType {
	hostPathType := new(corev1.HostPathType)
	*hostPathType = corev1.HostPathType(pathType)
	return hostPathType
}

func newMountPropagationMode(m corev1.MountPropagationMode) *corev1.MountPropagationMode {
	mode := new(corev1.MountPropagationMode)
	*mode = corev1.MountPropagationMode(m)
	return mode
}

func (c *AgentController) desired(ctx context.Context, coll *flowslatest.FlowCollector, rlog logr.Logger) (*v1.DaemonSet, error) {
	if coll == nil {
		return nil, nil
	}
	version := helper.ExtractVersion(c.Images[constants.ControllerBaseImageIndex])
	annotations := make(map[string]string)
	env, err := c.envConfig(ctx, coll, annotations)
	if err != nil {
		return nil, err
	}

	if coll.Spec.Agent.EBPF.Metrics.Server.TLS.Type != flowslatest.ServerTLSDisabled {
		var promTLS *flowslatest.CertificateReference
		switch coll.Spec.Agent.EBPF.Metrics.Server.TLS.Type {
		case flowslatest.ServerTLSProvided:
			promTLS = coll.Spec.Agent.EBPF.Metrics.Server.TLS.Provided
			if promTLS == nil {
				rlog.Info("EBPF agent metric tls configuration set to provided but none is provided")
			}
		case flowslatest.ServerTLSAuto:
			promTLS = &flowslatest.CertificateReference{
				Type:     "secret",
				Name:     constants.EBPFAgentMetricsSvcName,
				CertFile: "tls.crt",
				CertKey:  "tls.key",
			}
		case flowslatest.ServerTLSDisabled:
			// show never happens added for linting purposes
		}
		cert, key := c.volumes.AddCertificate(promTLS, "prom-certs")
		if cert != "" && key != "" {
			env = append(env, corev1.EnvVar{Name: envMetricsTLSKeyPath,
				Value: key,
			})
			env = append(env, corev1.EnvVar{
				Name:  envMetricsTLSCertPath,
				Value: cert,
			})
		}
	}

	volumeMounts := c.volumes.GetMounts()
	volumes := c.volumes.GetVolumes()

	if helper.IsPrivileged(&coll.Spec.Agent.EBPF) {
		volume := corev1.Volume{
			Name: bpfNetNSMountName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Type: newHostPathType(corev1.HostPathDirectory),
					Path: bpfNetNSMountPath,
				},
			},
		}
		volumes = append(volumes, volume)
		volumeMount := corev1.VolumeMount{
			Name:             bpfNetNSMountName,
			MountPath:        bpfNetNSMountPath,
			MountPropagation: newMountPropagationMode(corev1.MountPropagationBidirectional),
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	if helper.IsAgentFeatureEnabled(&coll.Spec.Agent.EBPF, flowslatest.PacketDrop) {
		if !coll.Spec.Agent.EBPF.Privileged {
			rlog.Error(fmt.Errorf("invalid configuration"), "To use PacketsDrop feature privileged mode needs to be enabled")
		} else {
			volume := corev1.Volume{
				Name: bpfTraceMountName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Type: newHostPathType(corev1.HostPathDirectory),
						Path: bpfTraceMountPath,
					},
				},
			}
			volumes = append(volumes, volume)
			volumeMount := corev1.VolumeMount{
				Name:             bpfTraceMountName,
				MountPath:        bpfTraceMountPath,
				MountPropagation: newMountPropagationMode(corev1.MountPropagationBidirectional),
			}
			volumeMounts = append(volumeMounts, volumeMount)
		}
	}

	if helper.IsAgentFeatureEnabled(&coll.Spec.Agent.EBPF, flowslatest.NetworkEvents) ||
		helper.IsAgentFeatureEnabled(&coll.Spec.Agent.EBPF, flowslatest.UDNMapping) {
		if !coll.Spec.Agent.EBPF.Privileged {
			rlog.Error(fmt.Errorf("invalid configuration"), "To use Network Events Monitor"+
				"features privileged mode needs to be enabled")
		} else {
			volume := corev1.Volume{
				Name: ovnObservMountName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Type: newHostPathType(corev1.HostPathDirectory),
						Path: ovnObservHostMountPath,
					},
				},
			}
			volumes = append(volumes, volume)
			volumeMount := corev1.VolumeMount{
				Name:             ovnObservMountName,
				MountPath:        ovnObservMountPath,
				MountPropagation: newMountPropagationMode(corev1.MountPropagationBidirectional),
			}
			volumeMounts = append(volumeMounts, volumeMount)

			volume = corev1.Volume{
				Name: ovsMountName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Type: newHostPathType(corev1.HostPathDirectory),
						Path: ovsHostMountPath,
					},
				},
			}
			volumes = append(volumes, volume)
			volumeMount = corev1.VolumeMount{
				Name:             ovsMountName,
				MountPath:        ovsMountPath,
				MountPropagation: newMountPropagationMode(corev1.MountPropagationBidirectional),
			}
			volumeMounts = append(volumeMounts, volumeMount)
		}
	}

	if helper.IsAgentFeatureEnabled(&coll.Spec.Agent.EBPF, flowslatest.EbpfManager) {
		volume := corev1.Volume{
			Name: bpfmanMapsVolumeName,
			VolumeSource: corev1.VolumeSource{
				CSI: &corev1.CSIVolumeSource{
					Driver: "csi.bpfman.io",
					VolumeAttributes: map[string]string{
						"csi.bpfman.io/program": "netobserv",
						"csi.bpfman.io/maps":    "aggregated_flows,additional_flow_metrics,direct_flows,dns_flows,filter_map,peer_filter_map,global_counters,packet_record",
					},
				},
			},
		}
		volumes = append(volumes, volume)
		volumeMount := corev1.VolumeMount{
			Name:             bpfmanMapsVolumeName,
			MountPath:        bpfManBpfFSPath,
			MountPropagation: newMountPropagationMode(corev1.MountPropagationBidirectional),
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	advancedConfig := helper.GetAdvancedAgentConfig(coll.Spec.Agent.EBPF.Advanced)

	return &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentName,
			Namespace: c.PrivilegedNamespace(),
			Labels: map[string]string{
				"app":     constants.EBPFAgentName,
				"version": helper.MaxLabelLength(version),
			},
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": constants.EBPFAgentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": constants.EBPFAgentName},
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					// Allows deploying an instance in the master node
					ServiceAccountName: constants.EBPFServiceAccount,
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					Volumes:            volumes,
					Containers: []corev1.Container{{
						Name:            constants.EBPFAgentName,
						Image:           c.Images[constants.ControllerBaseImageIndex],
						ImagePullPolicy: corev1.PullPolicy(coll.Spec.Agent.EBPF.ImagePullPolicy),
						Resources:       coll.Spec.Agent.EBPF.Resources,
						SecurityContext: c.securityContext(coll),
						Env:             env,
						VolumeMounts:    volumeMounts,
					}},
					NodeSelector:      advancedConfig.Scheduling.NodeSelector,
					Tolerations:       advancedConfig.Scheduling.Tolerations,
					Affinity:          advancedConfig.Scheduling.Affinity,
					PriorityClassName: advancedConfig.Scheduling.PriorityClassName,
				},
			},
		},
	}, nil
}

func (c *AgentController) envConfig(ctx context.Context, coll *flowslatest.FlowCollector, annots map[string]string) ([]corev1.EnvVar, error) {
	config := c.setEnvConfig(coll)

	if helper.UseKafka(&coll.Spec) {
		config = append(config,
			corev1.EnvVar{Name: envExport, Value: exportKafka},
			corev1.EnvVar{Name: envKafkaBrokers, Value: coll.Spec.Kafka.Address},
			corev1.EnvVar{Name: envKafkaTopic, Value: coll.Spec.Kafka.Topic},
			corev1.EnvVar{Name: envKafkaBatchSize, Value: strconv.Itoa(coll.Spec.Agent.EBPF.KafkaBatchSize)},
			// For easier user configuration, we can assume a constant message size per flow (~100B in protobuf)
			corev1.EnvVar{Name: envKafkaBatchMessages, Value: strconv.Itoa(coll.Spec.Agent.EBPF.KafkaBatchSize / averageMessageSize)},
		)
		if coll.Spec.Kafka.TLS.Enable {
			// Annotate pod with certificate reference so that it is reloaded if modified
			// If user cert is provided, it will use mTLS. Else, simple TLS (the userDigest and paths will be empty)
			caDigest, userDigest, err := c.Watcher.ProcessMTLSCerts(ctx, c.Client, &coll.Spec.Kafka.TLS, c.PrivilegedNamespace())
			if err != nil {
				return nil, err
			}
			annots[watchers.Annotation("kafka-ca")] = caDigest
			annots[watchers.Annotation("kafka-user")] = userDigest

			caPath, userCertPath, userKeyPath := c.volumes.AddMutualTLSCertificates(&coll.Spec.Kafka.TLS, "kafka-certs")
			config = append(config,
				corev1.EnvVar{Name: envKafkaEnableTLS, Value: "true"},
				corev1.EnvVar{Name: envKafkaTLSInsecureSkipVerify, Value: strconv.FormatBool(coll.Spec.Kafka.TLS.InsecureSkipVerify)},
				corev1.EnvVar{Name: envKafkaTLSCACertPath, Value: caPath},
				corev1.EnvVar{Name: envKafkaTLSUserCertPath, Value: userCertPath},
				corev1.EnvVar{Name: envKafkaTLSUserKeyPath, Value: userKeyPath},
			)
		}
		if helper.UseSASL(&coll.Spec.Kafka.SASL) {
			sasl := &coll.Spec.Kafka.SASL
			// Annotate pod with secret reference so that it is reloaded if modified
			d1, d2, err := c.Watcher.ProcessSASL(ctx, c.Client, sasl, c.PrivilegedNamespace())
			if err != nil {
				return nil, err
			}
			annots[watchers.Annotation("kafka-sd1")] = d1
			annots[watchers.Annotation("kafka-sd2")] = d2

			t := "plain"
			if coll.Spec.Kafka.SASL.Type == flowslatest.SASLScramSHA512 {
				t = "scramSHA512"
			}
			idPath := c.volumes.AddVolume(&sasl.ClientIDReference, "kafka-sasl-id")
			secretPath := c.volumes.AddVolume(&sasl.ClientSecretReference, "kafka-sasl-secret")
			config = append(config,
				corev1.EnvVar{Name: envKafkaEnableSASL, Value: "true"},
				corev1.EnvVar{Name: envKafkaSASLType, Value: t},
				corev1.EnvVar{Name: envKafkaSASLIDPath, Value: idPath},
				corev1.EnvVar{Name: envKafkaSASLSecretPath, Value: secretPath},
			)
		}
	} else {
		config = append(config, corev1.EnvVar{Name: envExport, Value: exportGRPC})
		advancedConfig := helper.GetAdvancedProcessorConfig(coll.Spec.Processor.Advanced)
		// When flowlogs-pipeline is deployed as a daemonset, each agent must send
		// data to the pod that is deployed in the same host
		config = append(config, corev1.EnvVar{
			Name: envFlowsTargetHost,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		}, corev1.EnvVar{
			Name:  envFlowsTargetPort,
			Value: strconv.Itoa(int(*advancedConfig.Port)),
		})
	}

	if helper.IsEBFPFlowFilterEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{Name: envEnableFlowFilter, Value: "true"})
		if len(coll.Spec.Agent.EBPF.FlowFilter.Rules) != 0 {
			if filterRules := c.configureFlowFiltersRules(coll.Spec.Agent.EBPF.FlowFilter.Rules); filterRules != nil {
				config = append(config, filterRules...)
			}
		} else {
			if filter := c.configureFlowFilter(coll.Spec.Agent.EBPF.FlowFilter); filter != nil {
				config = append(config, filter...)
			}
		}
	}

	return config, nil
}

func mapFlowFilterRuleToFilter(rule *flowslatest.EBPFFlowFilterRule) ebpfconfig.FlowFilter {
	f := ebpfconfig.FlowFilter{
		FilterIPCIDR:    rule.CIDR,
		FilterAction:    rule.Action,
		FilterDirection: rule.Direction,
		FilterProtocol:  rule.Protocol,
	}

	if rule.ICMPType != nil && *rule.ICMPType != 0 {
		f.FilterICMPType = *rule.ICMPType
	}
	if rule.ICMPCode != nil && *rule.ICMPCode != 0 {
		f.FilterICMPCode = *rule.ICMPCode
	}

	processPorts(rule.SourcePorts, &f.FilterSourcePort, &f.FilterSourcePorts, &f.FilterSourcePortRange)
	processPorts(rule.DestPorts, &f.FilterDestinationPort, &f.FilterDestinationPorts, &f.FilterDestinationPortRange)
	processPorts(rule.Ports, &f.FilterPort, &f.FilterPorts, &f.FilterPortRange)

	if rule.PeerIP != "" {
		f.FilterPeerIP = rule.PeerIP
	}
	if rule.PeerCIDR != "" {
		f.FilterPeerCIDR = rule.PeerCIDR
	}
	if rule.TCPFlags != "" {
		f.FilterTCPFlags = rule.TCPFlags
	}
	if rule.PktDrops != nil && *rule.PktDrops {
		f.FilterDrops = *rule.PktDrops
	}
	if rule.Sampling != nil && *rule.Sampling != 0 {
		f.FilterSample = *rule.Sampling
	}

	return f
}

func mapFlowFilterToFilter(filter *flowslatest.EBPFFlowFilter) ebpfconfig.FlowFilter {
	f := ebpfconfig.FlowFilter{
		FilterIPCIDR:    filter.CIDR,
		FilterAction:    filter.Action,
		FilterDirection: filter.Direction,
		FilterProtocol:  filter.Protocol,
	}

	if filter.ICMPType != nil && *filter.ICMPType != 0 {
		f.FilterICMPType = *filter.ICMPType
	}
	if filter.ICMPCode != nil && *filter.ICMPCode != 0 {
		f.FilterICMPCode = *filter.ICMPCode
	}

	processPorts(filter.SourcePorts, &f.FilterSourcePort, &f.FilterSourcePorts, &f.FilterSourcePortRange)
	processPorts(filter.DestPorts, &f.FilterDestinationPort, &f.FilterDestinationPorts, &f.FilterDestinationPortRange)
	processPorts(filter.Ports, &f.FilterPort, &f.FilterPorts, &f.FilterPortRange)

	if filter.PeerIP != "" {
		f.FilterPeerIP = filter.PeerIP
	}
	if filter.PeerCIDR != "" {
		f.FilterPeerCIDR = filter.PeerCIDR
	}
	if filter.TCPFlags != "" {
		f.FilterTCPFlags = filter.TCPFlags
	}
	if filter.PktDrops != nil && *filter.PktDrops {
		f.FilterDrops = *filter.PktDrops
	}
	if filter.Sampling != nil && *filter.Sampling != 0 {
		f.FilterSample = *filter.Sampling
	}

	return f
}

func processPorts(ports intstr.IntOrString, single *int32, list *string, rangeField *string) {
	if ports.Type == intstr.String {
		portStr := ports.String()
		if strings.Contains(portStr, "-") {
			*rangeField = portStr
		}
		if strings.Contains(portStr, ",") {
			*list = portStr
		}
	} else if ports.Type == intstr.Int {
		*single = int32(ports.IntValue())
	}
}

func (c *AgentController) configureFlowFiltersRules(rules []flowslatest.EBPFFlowFilterRule) []corev1.EnvVar {
	filters := make([]ebpfconfig.FlowFilter, 0)
	for i := range rules {
		filters = append(filters, mapFlowFilterRuleToFilter(&rules[i]))
	}

	jsonData, err := json.Marshal(filters)
	if err != nil {
		return nil
	}

	return []corev1.EnvVar{{Name: envFilterRules, Value: string(jsonData)}}
}

func (c *AgentController) configureFlowFilter(filter *flowslatest.EBPFFlowFilter) []corev1.EnvVar {
	f := mapFlowFilterToFilter(filter)
	jsonData, err := json.Marshal([]ebpfconfig.FlowFilter{f})
	if err != nil {
		return nil
	}

	return []corev1.EnvVar{{Name: envFilterRules, Value: string(jsonData)}}
}

func (c *AgentController) securityContext(coll *flowslatest.FlowCollector) *corev1.SecurityContext {
	if coll.Spec.Agent.EBPF.Privileged {
		return &corev1.SecurityContext{
			RunAsUser:  ptr.To(int64(0)),
			Privileged: &coll.Spec.Agent.EBPF.Privileged,
		}
	}

	sc := helper.ContainerDefaultSecurityContext()
	sc.RunAsUser = ptr.To(int64(0))
	sc.Capabilities.Add = permissions.AllowedCapabilities
	return sc
}

// nolint:golint,cyclop
func (c *AgentController) setEnvConfig(coll *flowslatest.FlowCollector) []corev1.EnvVar {
	var config []corev1.EnvVar

	if coll.Spec.Agent.EBPF.CacheActiveTimeout != "" {
		config = append(config, corev1.EnvVar{
			Name:  envCacheActiveTimeout,
			Value: coll.Spec.Agent.EBPF.CacheActiveTimeout,
		})
	}

	if coll.Spec.Agent.EBPF.CacheMaxFlows != 0 {
		config = append(config, corev1.EnvVar{
			Name:  envCacheMaxFlows,
			Value: strconv.Itoa(int(coll.Spec.Agent.EBPF.CacheMaxFlows)),
		})
	}

	if coll.Spec.Agent.EBPF.LogLevel != "" {
		config = append(config, corev1.EnvVar{
			Name:  envLogLevel,
			Value: coll.Spec.Agent.EBPF.LogLevel,
		})
	}

	if len(coll.Spec.Agent.EBPF.Interfaces) > 0 {
		config = append(config, corev1.EnvVar{
			Name:  envInterfaces,
			Value: strings.Join(coll.Spec.Agent.EBPF.Interfaces, envListSeparator),
		})
	}

	if len(coll.Spec.Agent.EBPF.ExcludeInterfaces) > 0 {
		config = append(config, corev1.EnvVar{
			Name:  envExcludeInterfaces,
			Value: strings.Join(coll.Spec.Agent.EBPF.ExcludeInterfaces, envListSeparator),
		})
	}

	sampling := coll.Spec.Agent.EBPF.Sampling
	if sampling != nil && *sampling > 0 {
		config = append(config, corev1.EnvVar{
			Name:  envSampling,
			Value: strconv.Itoa(int(*sampling)),
		})
	}

	if helper.IsFlowRTTEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableFlowRTT,
			Value: "true",
		})
	}

	if helper.IsNetworkEventsEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableNetworkEvents,
			Value: "true",
		})
	}

	if helper.IsUDNMappingEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableUDNMapping,
			Value: "true",
		})
	}

	if helper.IsPacketTranslationEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnablePacketTranslation,
			Value: "true",
		})
	}
	if helper.IsEbpfManagerEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableEbpfMgr,
			Value: "true",
		})
	}

	// set GOMEMLIMIT which allows specifying a soft memory cap to force GC when resource limit is reached
	// to prevent OOM
	if coll.Spec.Agent.EBPF.Resources.Limits.Memory() != nil {
		if memLimit, ok := coll.Spec.Agent.EBPF.Resources.Limits.Memory().AsInt64(); ok {
			// we will set the GOMEMLIMIT to current memlimit - 10% as a headroom to account for
			// memory sources the Go runtime is unaware of
			memLimit -= int64(float64(memLimit) * 0.1)
			config = append(config, corev1.EnvVar{Name: envGoMemLimit, Value: fmt.Sprint(memLimit)})
		}
	}

	if helper.IsPktDropEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnablePktDrop,
			Value: "true",
		})
	}

	if helper.IsDNSTrackingEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableDNSTracking,
			Value: "true",
		})
	}

	if helper.IsEBPFMetricsEnabled(&coll.Spec.Agent.EBPF) {
		config = append(config, corev1.EnvVar{
			Name:  envEnableMetrics,
			Value: "true",
		})
		config = append(config, corev1.EnvVar{
			Name:  envMetricsPort,
			Value: strconv.Itoa(int(helper.GetEBPFMetricsPort(&coll.Spec.Agent.EBPF))),
		})
		config = append(config, corev1.EnvVar{
			Name:  envMetricPrefix,
			Value: "netobserv_agent_",
		})
	}

	dedup := dedupeDefault
	dedupJustMark := DedupeJustMarkDefault
	dedupMerge := DedupeMergeDefault
	dnsTrackingPort := defaultDNSTrackingPort
	networkEventsGroupID := defaultNetworkEventsGroupID
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	advancedConfig := helper.GetAdvancedAgentConfig(coll.Spec.Agent.EBPF.Advanced)
	for _, pair := range helper.KeySorted(advancedConfig.Env) {
		k, v := pair[0], pair[1]
		switch k {
		case envDedupe:
			dedup = v
		case EnvDedupeJustMark:
			dedupJustMark = v
		case EnvDedupeMerge:
			dedupMerge = v
		case envDNSTrackingPort:
			dnsTrackingPort = v
		case envNetworkEventsGroupID:
			networkEventsGroupID = v
		default:
			config = append(config, corev1.EnvVar{Name: k, Value: v})
		}
	}

	config = append(config, corev1.EnvVar{Name: envDedupe, Value: dedup})
	config = append(config, corev1.EnvVar{Name: EnvDedupeJustMark, Value: dedupJustMark})
	config = append(config, corev1.EnvVar{Name: envDNSTrackingPort, Value: dnsTrackingPort})
	config = append(config, corev1.EnvVar{Name: envNetworkEventsGroupID, Value: networkEventsGroupID})
	config = append(config, corev1.EnvVar{
		Name: envAgentIP,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.hostIP",
			},
		},
	},
	)
	config = append(config, corev1.EnvVar{Name: EnvDedupeMerge, Value: dedupMerge})

	return config
}
