package ebpf

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	"github.com/netobserv/network-observability-operator/controllers/operator"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"k8s.io/apimachinery/pkg/api/equality"

	bpfdHelpers "github.com/redhat-et/bpfd/bpfd-operator/pkg/helpers"
)

const (
	envCacheActiveTimeout         = "CACHE_ACTIVE_TIMEOUT"
	envCacheMaxFlows              = "CACHE_MAX_FLOWS"
	envExcludeInterfaces          = "EXCLUDE_INTERFACES"
	envInterfaces                 = "INTERFACES"
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
	envLogLevel                   = "LOG_LEVEL"
	envDedupe                     = "DEDUPER"
	dedupeDefault                 = "firstCome"
	envDedupeJustMark             = "DEDUPER_JUST_MARK"
	dedupeJustMarkDefault         = "true"

	envListSeparator = ","
)

const (
	exportKafka = "kafka"
	exportGRPC  = "grpc"
)

const kafkaCerts = "kafka-certs"
const averageMessageSize = 100

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
	client                      reconcilers.ClientHelper
	baseNamespace               string
	privilegedNamespace         string
	previousPrivilegedNamespace string
	permissions                 permissions.Reconciler
	config                      *operator.Config
}

func NewAgentController(
	client reconcilers.ClientHelper,
	baseNamespace string,
	previousBaseNamespace string,
	permissionsVendor *discover.Permissions,
	config *operator.Config,
) *AgentController {
	pns := baseNamespace + constants.EBPFPrivilegedNSSuffix
	opns := previousBaseNamespace + constants.EBPFPrivilegedNSSuffix
	return &AgentController{
		client:                      client,
		baseNamespace:               baseNamespace,
		privilegedNamespace:         pns,
		previousPrivilegedNamespace: opns,
		permissions:                 permissions.NewReconciler(client, pns, opns, permissionsVendor),
		config:                      config,
	}
}

func (c *AgentController) Reconcile(
	ctx context.Context, target *flowslatest.FlowCollector) error {
	rlog := log.FromContext(ctx).WithName("ebpf.AgentController")
	ctx = log.IntoContext(ctx, rlog)
	current, err := c.current(ctx)
	if err != nil {
		return fmt.Errorf("fetching current EBPF Agent: %w", err)
	}

	bpfdEnabled := bpfdHelpers.IsBpfdDeployed()

	if !helper.UseEBPF(&target.Spec) || c.previousPrivilegedNamespace != c.privilegedNamespace {
		if current == nil {
			rlog.Info("nothing to do, as the requested agent is not eBPF",
				"currentAgent", target.Spec.Agent)
			return nil
		}
		// If the user has changed the agent type or changed the target namespace, we need to manually
		// undeploy the agent
		rlog.Info("user changed the agent type, or the target namespace. Deleting eBPF agent",
			"currentAgent", target.Spec.Agent)
		if err := c.client.Delete(ctx, current); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("deleting eBPF agent: %w", err)
		}

		// Delete any bpfProgConfigs with the label app=netobserv
		labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"app": "netobserv"}}
		if bpfdEnabled {
			rlog.Info("bpfd is enabled, deploying bpfProgramConfigs")
			if err := c.desiredBpfdState(target.Spec.Agent.EBPF.Interfaces, &labelSelector, target); err != nil {
				return err
			}
		} else {
			rlog.Info("bpfd isn't enabled, not deploying bpfProgramConfigs")
		}

		// Current now has been deleted. Set it to nil to that it triggers actionCreate if we are changing namespace
		current = nil
	}

	if err := c.permissions.Reconcile(ctx, &target.Spec.Agent.EBPF); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}
	desired := c.desired(target, bpfdEnabled)

	// Annotate pod with certificate reference so that it is reloaded if modified
	if err := c.client.CertWatcher.AnnotatePod(ctx, c.client, &desired.Spec.Template, kafkaCerts); err != nil {
		return err
	}

	switch c.requiredAction(current, desired) {
	case actionCreate:
		rlog.Info("action: create agent")
		// Create bpfdProgramConfig if bpfd is enabled
		// TODO(astoycos) not enabled for exclude interfaces
		if bpfdEnabled {
			rlog.Info("bpfd is enabled, deploying bpfProgramConfigs")
			if err := c.desiredBpfdState(target.Spec.Agent.EBPF.Interfaces, nil, target); err != nil {
				return err
			}
		} else {
			rlog.Info("bpfd isn't enabled, not deploying bpfProgramConfigs")
		}
		return c.client.CreateOwned(ctx, desired)
	case actionUpdate:
		rlog.Info("action: update agent")
		if bpfdEnabled {
			rlog.Info("bpfd is enabled, deploying bpfProgramConfigs")
			if err := c.desiredBpfdState(target.Spec.Agent.EBPF.Interfaces, nil, target); err != nil {
				return err
			}
		} else {
			rlog.Info("bpfd isn't enabled, not deploying bpfProgramConfigs")
		}
		return c.client.UpdateOwned(ctx, current, desired)
	default:
		rlog.Info("action: nothing to do")
		c.client.CheckDaemonSetInProgress(current)
		return nil
	}
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.previousPrivilegedNamespace,
	}, &agentDS); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w",
			c.previousPrivilegedNamespace, constants.EBPFAgentName, err)
	}
	return &agentDS, nil
}

func (c *AgentController) desiredBpfdState(interfaces []string, deleteLabels *metav1.LabelSelector, target *flowslatest.FlowCollector) error {
	bpfdClient := bpfdHelpers.GetClientOrDie()
	flowCollectorScheme := k8sruntime.NewScheme()
	if err := flowslatest.AddToScheme(flowCollectorScheme); err != nil {
		return fmt.Errorf("failed to build flowcollector Scheme: %w", err)
	}

	if deleteLabels != nil {
		return bpfdHelpers.DeleteBpfProgConfLabels(bpfdClient, deleteLabels)
	}

	for _, intf := range interfaces {
		progConfEgr := bpfdHelpers.NewBpfProgramConfig(fmt.Sprintf("netobserv-flow-monitor-egress-%s", intf), bpfdHelpers.Tc)

		// Add app label to make deletion easier
		progConfEgr.Labels = map[string]string{"app": "netobserv"}
		// Must match section name
		progConfEgr.Spec.Name = "egress_flow_parse"
		// Empty label selector selects all nodes
		progConfEgr.Spec.NodeSelector = metav1.LabelSelector{}
		// Attatch to the specified interface
		progConfEgr.Spec.AttachPoint.NetworkMultiAttach.Interface = intf
		// Set bytecode container image
		progConfEgr.Spec.ByteCode = "image://quay.io/bpfd-bytecode/netobserv_egress:latest"
		// Load on int Egress
		progConfEgr.Spec.AttachPoint.NetworkMultiAttach.Direction = "EGRESS"
		// Set Priority
		progConfEgr.Spec.AttachPoint.NetworkMultiAttach.Priority = 50

		err := bpfdHelpers.CreateOrUpdateOwnedBpfProgConf(bpfdClient, progConfEgr, target, flowCollectorScheme)
		if err != nil {
			return fmt.Errorf("failed to create bpfProgramConfig: %w", err)
		}

		if err := bpfdHelpers.WaitForBpfProgConfLoad(bpfdClient, progConfEgr.Name, 10*time.Second); err != nil {
			return fmt.Errorf("failed to wait bpfProgramConfig readiness: %w", err)
		}

		progConfIng := bpfdHelpers.NewBpfProgramConfig(fmt.Sprintf("netobserv-flow-monitor-ingress-%s", intf), bpfdHelpers.Tc)

		// add app label to make deletion easier
		progConfIng.Labels = map[string]string{"app": "netobserv"}
		// Must match section name
		progConfIng.Spec.Name = "ingress_flow_parse"
		// Empty label selector selects all nodes
		progConfIng.Spec.NodeSelector = metav1.LabelSelector{}
		// Attatch to the specified interface
		progConfIng.Spec.AttachPoint.NetworkMultiAttach.Interface = intf
		// Set bytecode container image
		progConfIng.Spec.ByteCode = "image://quay.io/bpfd-bytecode/netobserv_ingress:latest"
		// Load on int Ingress
		progConfIng.Spec.AttachPoint.NetworkMultiAttach.Direction = "INGRESS"
		// Set Priority
		progConfIng.Spec.AttachPoint.NetworkMultiAttach.Priority = 50

		err = bpfdHelpers.CreateOrUpdateOwnedBpfProgConf(bpfdClient, progConfIng, target, flowCollectorScheme)
		if err != nil {
			return fmt.Errorf("failed to create bpfProgramConfig: %w", err)
		}

		if err := bpfdHelpers.WaitForBpfProgConfLoad(bpfdClient, progConfIng.Name, 10*time.Second); err != nil {
			return fmt.Errorf("failed to wait bpfProgramConfig readiness: %w", err)
		}
	}

	return nil
}

func (c *AgentController) desired(coll *flowslatest.FlowCollector, bpfdEnabled bool) *v1.DaemonSet {
	if coll == nil || !helper.UseEBPF(&coll.Spec) {
		return nil
	}
	version := helper.ExtractVersion(c.config.EBPFAgentImage)
	volumeMounts := []corev1.VolumeMount{}
	volumes := []corev1.Volume{}
	if bpfdEnabled {
		volumes = append(volumes, corev1.Volume{
			Name: "bpfd-fs",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/run/bpfd/fs/maps"},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "bpfd-fs",
			MountPath: "/run/bpfd/fs/maps",
		})
	}
	if helper.UseKafka(&coll.Spec) && coll.Spec.Kafka.TLS.Enable {
		// NOTE: secrets need to be copied from the base netobserv namespace to the privileged one.
		// This operation must currently be performed manually (run "make fix-ebpf-kafka-tls"). It could be automated here.
		volumes, volumeMounts = helper.AppendCertVolumes(volumes, volumeMounts, &coll.Spec.Kafka.TLS, kafkaCerts, c.client.CertWatcher)
	}

	return &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentName,
			Namespace: c.privilegedNamespace,
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
					Labels: map[string]string{"app": constants.EBPFAgentName},
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
						Image:           c.config.EBPFAgentImage,
						ImagePullPolicy: corev1.PullPolicy(coll.Spec.Agent.EBPF.ImagePullPolicy),
						Resources:       coll.Spec.Agent.EBPF.Resources,
						SecurityContext: c.securityContext(coll),
						Env:             c.envConfig(coll),
						VolumeMounts:    volumeMounts,
					}},
				},
			},
		},
	}
}

func (c *AgentController) envConfig(coll *flowslatest.FlowCollector) []corev1.EnvVar {
	var config []corev1.EnvVar
	config = append(config, corev1.EnvVar{
		Name:      "NODENAME",
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
	})
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
	dedup := dedupeDefault
	dedupJustMark := dedupeJustMarkDefault
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(coll.Spec.Agent.EBPF.Debug.Env) {
		k, v := pair[0], pair[1]
		if k == envDedupe {
			dedup = v
		} else if k == envDedupeJustMark {
			dedupJustMark = v
		} else {
			config = append(config, corev1.EnvVar{Name: k, Value: v})
		}
	}
	config = append(config, corev1.EnvVar{Name: envDedupe, Value: dedup})
	config = append(config, corev1.EnvVar{Name: envDedupeJustMark, Value: dedupJustMark})

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
			config = append(config,
				corev1.EnvVar{Name: envKafkaEnableTLS, Value: "true"},
				corev1.EnvVar{Name: envKafkaTLSInsecureSkipVerify, Value: strconv.FormatBool(coll.Spec.Kafka.TLS.InsecureSkipVerify)},
				corev1.EnvVar{Name: envKafkaTLSCACertPath, Value: helper.GetCACertPath(&coll.Spec.Kafka.TLS, kafkaCerts)},
				corev1.EnvVar{Name: envKafkaTLSUserCertPath, Value: helper.GetUserCertPath(&coll.Spec.Kafka.TLS, kafkaCerts)},
				corev1.EnvVar{Name: envKafkaTLSUserKeyPath, Value: helper.GetUserKeyPath(&coll.Spec.Kafka.TLS, kafkaCerts)},
			)
		}
	} else {
		config = append(config, corev1.EnvVar{Name: envExport, Value: exportGRPC})
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
			Value: strconv.Itoa(int(coll.Spec.Processor.Port)),
		})
	}
	return config
}

func (c *AgentController) requiredAction(current, desired *v1.DaemonSet) reconcileAction {
	if desired == nil {
		return actionNone
	}
	if current == nil && desired != nil {
		return actionCreate
	}
	cSpec, dSpec := current.Spec, desired.Spec
	eq := equality.Semantic.DeepDerivative
	if !helper.IsSubSet(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) ||
		!eq(dSpec.Selector, cSpec.Selector) ||
		!eq(dSpec.Template, cSpec.Template) {

		return actionUpdate
	}

	return actionNone
}

func (c *AgentController) securityContext(coll *flowslatest.FlowCollector) *corev1.SecurityContext {
	sc := corev1.SecurityContext{
		RunAsUser: pointer.Int64(0),
	}

	if coll.Spec.Agent.EBPF.Privileged {
		sc.Privileged = &coll.Spec.Agent.EBPF.Privileged
	} else {
		sc.Capabilities = &corev1.Capabilities{Add: permissions.AllowedCapabilities}
	}

	return &sc
}
