package ebpf

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	"github.com/netobserv/network-observability-operator/controllers/flp"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	"github.com/netobserv/network-observability-operator/pkg/watchers"

	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	envCacheActiveTimeout         = "CACHE_ACTIVE_TIMEOUT"
	envCacheMaxFlows              = "CACHE_MAX_FLOWS"
	envExcludeInterfaces          = "EXCLUDE_INTERFACES"
	envInterfaces                 = "INTERFACES"
	envAgentIP                    = "AGENT_IP"
	envFlowsTargetHost            = "FLOWS_TARGET_HOST"
	envFlowsTargetPort            = "FLOWS_TARGET_PORT"
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
	envDedupeJustMark             = "DEDUPER_JUST_MARK"
	envFLPConfig                  = "FLP_CONFIG"
	envGoMemLimit                 = "GOMEMLIMIT"
	envEnablePktDrop              = "ENABLE_PKT_DROPS"
	envEnableDNSTracking          = "ENABLE_DNS_TRACKING"
	envEnableFlowRTT              = "ENABLE_RTT"
	envEnableMetrics              = "METRICS_ENABLE"
	envMetricsPort                = "METRICS_SERVER_PORT"
	envMetricPrefix               = "METRICS_PREFIX"
	envListSeparator              = ","
)

const (
	exportKafka        = "kafka"
	exportFLP          = "direct-flp"
	kafkaCerts         = "kafka-certs"
	averageMessageSize = 100
	bpfTraceMountName  = "bpf-kernel-debug"
	bpfTraceMountPath  = "/sys/kernel/debug"
	bpfNetNSMountName  = "var-run-netns"
	bpfNetNSMountPath  = "/var/run/netns"
)

type reconcileAction int

const (
	actionNone = iota
	actionCreate
	actionUpdate
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
}

func NewAgentController(common *reconcilers.Instance) *AgentController {
	common.Managed.Namespace = common.PrivilegedNamespace()
	agent := AgentController{
		Instance:    common,
		permissions: permissions.NewReconciler(common),
		promSvc:     common.Managed.NewService(constants.EBPFAgentMetricsSvcName),
	}
	if common.AvailableAPIs.HasSvcMonitor() {
		agent.serviceMonitor = common.Managed.NewServiceMonitor(constants.EBPFAgentMetricsSvcMonitoringName)
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

	if c.PreviousPrivilegedNamespace() != c.PrivilegedNamespace() {
		c.Managed.TryDeleteAll(ctx)

		if current == nil {
			rlog.Info("nothing to do, namespace already cleaned up", "currentAgent", target.Spec.Agent)
			return nil
		}
		rlog.Info("namespace cleanup: deleting eBPF agent", "currentAgent", target.Spec.Agent)
		if err := c.Delete(ctx, current); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("deleting eBPF agent: %w", err)
		}
		// Current now has been deleted. Set it to nil to that it triggers actionCreate if we are changing namespace
		current = nil
	}

	if err := c.permissions.Reconcile(ctx, &target.Spec.Agent.EBPF); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}
	desired, err := c.desired(ctx, target, rlog)
	if err != nil {
		return err
	}

	err = c.reconcileMetricsService(ctx, &target.Spec.Agent.EBPF)
	if err != nil {
		return fmt.Errorf("reconciling prometheus service: %w", err)
	}

	switch requiredAction(current, desired) {
	case actionCreate:
		rlog.Info("action: create agent")
		c.Status.SetCreatingDaemonSet(desired)
		return c.CreateOwned(ctx, desired)
	case actionUpdate:
		rlog.Info("action: update agent")
		return c.UpdateIfOwned(ctx, current, desired)
	default:
		rlog.Info("action: nothing to do")
		c.Status.CheckDaemonSetProgress(current)
		return nil
	}
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.PreviousPrivilegedNamespace(),
	}, &agentDS); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w", c.PreviousPrivilegedNamespace(), constants.EBPFAgentName, err)
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
	version := helper.ExtractVersion(c.Image)
	annotations := make(map[string]string)

	fm := metricslatest.FlowMetricList{}
	if !helper.UseKafka(&coll.Spec) {
		// Direct-FLP mode => list custom metrics
		if err := c.List(ctx, &fm, &client.ListOptions{Namespace: coll.Spec.Namespace}); err != nil {
			return nil, c.Status.Error("CantListFlowMetrics", err)
		}
	}

	env, err := c.envConfig(ctx, coll, annotations, &fm)
	if err != nil {
		return nil, err
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
					Tolerations:        []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
					ServiceAccountName: constants.EBPFServiceAccount,
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					Volumes:            volumes,
					Containers: []corev1.Container{{
						Name:            constants.EBPFAgentName,
						Image:           c.Image,
						ImagePullPolicy: corev1.PullPolicy(coll.Spec.Agent.EBPF.ImagePullPolicy),
						Resources:       coll.Spec.Agent.EBPF.Resources,
						SecurityContext: c.securityContext(coll),
						Env:             env,
						VolumeMounts:    volumeMounts,
					}},
				},
			},
		},
	}, nil
}

func (c *AgentController) envConfig(ctx context.Context, coll *flowslatest.FlowCollector, annots map[string]string, metrics *metricslatest.FlowMetricList) ([]corev1.EnvVar, error) {
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
		flpConfig, err := c.buildFLPConfig(&coll.Spec, metrics)
		if err != nil {
			return nil, err
		}
		debugConfig := helper.GetAdvancedProcessorConfig(coll.Spec.Processor.Advanced)
		config = append(config,
			corev1.EnvVar{
				Name:  envExport,
				Value: exportFLP,
			},
			corev1.EnvVar{
				Name:  envFLPConfig,
				Value: flpConfig,
			},
			corev1.EnvVar{
				Name: envFlowsTargetHost,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.hostIP",
					},
				},
			},
			corev1.EnvVar{
				Name:  envFlowsTargetPort,
				Value: strconv.Itoa(int(*debugConfig.Port)),
			},
		)
	}
	return config, nil
}

func (c *AgentController) buildFLPConfig(desired *flowslatest.FlowCollectorSpec, metrics *metricslatest.FlowMetricList) (string, error) {
	flpBuilder, err := flp.NewBuilder(c.NewInstance(c.Image, c.Status), desired, metrics, flp.ConfMonolith)
	if err != nil {
		return "", err
	}
	pipeline := flpBuilder.NewInProcessPipeline()
	err = pipeline.AddProcessorStages()
	if err != nil {
		return "", err
	}
	return flpBuilder.GetJSONConfig()
}

func requiredAction(current, desired *v1.DaemonSet) reconcileAction {
	if desired == nil {
		return actionNone
	}
	if current == nil {
		return actionCreate
	}
	cSpec, dSpec := current.Spec, desired.Spec
	eq := equality.Semantic.DeepDerivative
	if !helper.IsSubSet(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) ||
		!eq(dSpec.Selector, cSpec.Selector) ||
		!eq(dSpec.Template, cSpec.Template) {

		return actionUpdate
	}

	// Env vars aren't covered by DeepDerivative when they are removed: deep-compare them
	dConts := dSpec.Template.Spec.Containers
	cConts := cSpec.Template.Spec.Containers
	if len(dConts) > 0 && len(cConts) > 0 && !equality.Semantic.DeepEqual(dConts[0].Env, cConts[0].Env) {
		return actionUpdate
	}

	return actionNone
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
	if sampling != nil && *sampling > 1 {
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
			Value: strconv.Itoa(int(coll.Spec.Agent.EBPF.Metrics.Server.Port)),
		})
		config = append(config, corev1.EnvVar{
			Name:  envMetricPrefix,
			Value: "netobserv_agent_",
		})
	}

	// Init with defaults
	envs := map[string]string{
		envDedupe:         "firstCome",
		envDedupeJustMark: "true",
	}
	debugConfig := helper.GetAdvancedAgentConfig(coll.Spec.Agent.EBPF.Advanced)
	// Merge configured
	for k, v := range debugConfig.Env {
		envs[k] = v
	}
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(envs) {
		k, v := pair[0], pair[1]
		config = append(config, corev1.EnvVar{Name: k, Value: v})
	}

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

	return config
}
