package ebpf

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"k8s.io/apimachinery/pkg/api/equality"
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
	envKafkaEnableTLS             = "KAFKA_ENABLE_TLS"
	envKafkaTLSInsecureSkipVerify = "KAFKA_TLS_INSECURE_SKIP_VERIFY"
	envKafkaTLSCACertPath         = "KAFKA_TLS_CA_CERT_PATH"
	envKafkaTLSUserCertPath       = "KAFKA_TLS_USER_CERT_PATH"
	envKafkaTLSUserKeyPath        = "KAFKA_TLS_USER_KEY_PATH"
	envLogLevel                   = "LOG_LEVEL"

	envListSeparator = ","
)

const (
	exportKafka = "kafka"
	exportGRPC  = "grpc"
)

const kafkaCerts = "kafka-certs"

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
	client              reconcilers.ClientHelper
	baseNamespace       string
	privilegedNamespace string
	permissions         permissions.Reconciler
}

func NewAgentController(
	client reconcilers.ClientHelper,
	baseNamespace string,
	permissionsVendor *discover.Permissions,
) *AgentController {
	pns := baseNamespace + constants.EBPFPrivilegedNSSuffix
	return &AgentController{
		client:              client,
		baseNamespace:       baseNamespace,
		privilegedNamespace: pns,
		permissions:         permissions.NewReconciler(client, pns, permissionsVendor),
	}
}

func (c *AgentController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	rlog := log.FromContext(ctx).WithName("ebpf.AgentController")
	ctx = log.IntoContext(ctx, rlog)
	current, err := c.current(ctx)
	if err != nil {
		return fmt.Errorf("fetching current EBPF Agent: %w", err)
	}
	if target.Spec.Agent.Type != flowsv1alpha1.AgentEBPF {
		if current == nil {
			rlog.Info("nothing to do, as the requested agent is not eBPF",
				"currentAgent", target.Spec.Agent)
			return nil
		}
		// If the user has changed the agent type, we need to manually
		// undeploy the agent
		rlog.Info("user changed the agent type. Deleting eBPF agent",
			"currentAgent", target.Spec.Agent)
		if err := c.client.Delete(ctx, current); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("deleting eBPF agent: %w", err)
		}
	}

	if err := c.permissions.Reconcile(ctx, &target.Spec.Agent.EBPF); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}
	desired := c.desired(target)
	switch c.requiredAction(current, desired) {
	case actionCreate:
		rlog.Info("action: create agent")
		return c.client.CreateOwned(ctx, desired)
	case actionUpdate:
		rlog.Info("action: update agent")
		return c.client.UpdateOwned(ctx, current, desired)
	default:
		rlog.Info("action: nonthing to do")
		c.client.CheckDaemonSetInProgress(current)
		return nil
	}
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.privilegedNamespace,
	}, &agentDS); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w",
			c.privilegedNamespace, constants.EBPFAgentName, err)
	}
	return &agentDS, nil
}

func (c *AgentController) desired(coll *flowsv1alpha1.FlowCollector) *v1.DaemonSet {
	if coll == nil || coll.Spec.Agent.Type != flowsv1alpha1.AgentEBPF {
		return nil
	}
	version := helper.ExtractVersion(coll.Spec.Agent.EBPF.Image)
	volumeMounts := []corev1.VolumeMount{}
	volumes := []corev1.Volume{}
	if coll.Spec.Kafka.Enable && coll.Spec.Kafka.TLS.Enable {
		// NOTE: secrets need to be copied from the base network-observability namespace to the privileged one.
		// This operation must currently be performed manually (run "make fix-ebpf-kafka-tls"). It could be automated here.
		volumes, volumeMounts = helper.AppendCertVolumes(volumes, volumeMounts, &coll.Spec.Kafka.TLS, kafkaCerts)
	}
	return &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentName,
			Namespace: c.privilegedNamespace,
			Labels: map[string]string{
				"app":     constants.EBPFAgentName,
				"version": version,
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
						Image:           coll.Spec.Agent.EBPF.Image,
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

func (c *AgentController) envConfig(coll *flowsv1alpha1.FlowCollector) []corev1.EnvVar {
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
	if coll.Spec.Agent.EBPF.Sampling > 1 {
		config = append(config, corev1.EnvVar{
			Name:  envSampling,
			Value: strconv.Itoa(int(coll.Spec.Agent.EBPF.Sampling)),
		})
	}
	for k, v := range coll.Spec.Agent.EBPF.Env {
		config = append(config, corev1.EnvVar{Name: k, Value: v})
	}
	if coll.Spec.Kafka.Enable {
		config = append(config,
			corev1.EnvVar{Name: envExport, Value: exportKafka},
			corev1.EnvVar{Name: envKafkaBrokers, Value: coll.Spec.Kafka.Address},
			corev1.EnvVar{Name: envKafkaTopic, Value: coll.Spec.Kafka.Topic},
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
		switch coll.Spec.FlowlogsPipeline.Kind {
		case constants.DaemonSetKind:
			// When flowlogs-pipeline is deployed as a daemonset, each agent must send
			// data to the pod that is deployed in the same host
			config = append(config, corev1.EnvVar{
				Name: envFlowsTargetHost,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			}, corev1.EnvVar{
				Name:  envFlowsTargetPort,
				Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
			})
		case constants.DeploymentKind:
			config = append(config, corev1.EnvVar{
				Name:  envFlowsTargetHost,
				Value: constants.FLPName + "." + c.baseNamespace,
			}, corev1.EnvVar{
				Name:  envFlowsTargetPort,
				Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
			})
		}
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
	if equality.Semantic.DeepDerivative(&desired.Spec, current.Spec) {
		return actionNone
	}
	return actionUpdate
}

func (c *AgentController) securityContext(coll *flowsv1alpha1.FlowCollector) *corev1.SecurityContext {
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
