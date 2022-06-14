package flowlogspipeline

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/prometheus/common/model"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const configMapName = "flowlogs-pipeline-config"
const configVolume = "config-volume"
const configPath = "/etc/flowlogs-pipeline"
const configFile = "config.json"

const (
	healthServiceName       = "health"
	prometheusServiceName   = "prometheus"
	healthTimeoutSeconds    = 5
	livenessPeriodSeconds   = 10
	startupFailureThreshold = 5
	startupPeriodSeconds    = 10
)

const (
	ConfSingle           = "allInOne"
	ConfKafkaIngester    = "kafkaIngester"
	ConfKafkaTransformer = "kafkaTransformer"
)

var FlpConfSuffix = map[string]string{
	ConfSingle:           "",
	ConfKafkaIngester:    "-ingester",
	ConfKafkaTransformer: "-transformer",
}

// PodConfigurationDigest is an annotation name to facilitate pod restart after
// any external configuration change
const PodConfigurationDigest = "flows.netobserv.io/" + configMapName

type builder struct {
	namespace       string
	labels          map[string]string
	selector        map[string]string
	portProtocol    corev1.Protocol
	desired         *flowsv1alpha1.FlowCollectorFLP
	desiredLoki     *flowsv1alpha1.FlowCollectorLoki
	desiredKafka    *flowsv1alpha1.FlowCollectorKafka
	confKind        string
	confKindSuffix  string
	useOpenShiftSCC bool
}

func newBuilder(ns string, portProtocol corev1.Protocol, desired *flowsv1alpha1.FlowCollectorFLP, desiredLoki *flowsv1alpha1.FlowCollectorLoki, desiredKafka *flowsv1alpha1.FlowCollectorKafka, confKind string, useOpenShiftSCC bool) builder {
	version := helper.ExtractVersion(desired.Image)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.FLPName + FlpConfSuffix[confKind],
			"version": version,
		},
		selector: map[string]string{
			"app": constants.FLPName + FlpConfSuffix[confKind],
		},
		desired:         desired,
		desiredLoki:     desiredLoki,
		desiredKafka:    desiredKafka,
		portProtocol:    portProtocol,
		confKind:        confKind,
		confKindSuffix:  FlpConfSuffix[confKind],
		useOpenShiftSCC: useOpenShiftSCC,
	}
}

func (b *builder) deployment(configDigest string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName + b.confKindSuffix,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &b.desired.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: b.podTemplate(false, configDigest),
		},
	}
}

func (b *builder) daemonSet(configDigest string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName + b.confKindSuffix,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: b.podTemplate(!b.useOpenShiftSCC, configDigest),
		},
	}
}

func (b *builder) podTemplate(hostNetwork bool, configDigest string) corev1.PodTemplateSpec {
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if b.desired.Kind == constants.DaemonSetKind && b.confKind != ConfKafkaTransformer {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName + b.confKindSuffix,
			HostPort:      b.desired.Port,
			ContainerPort: b.desired.Port,
			Protocol:      b.portProtocol,
		}}
		// This allows deploying an instance in the master node, the same technique used in the
		// companion ovnkube-node daemonset definition
		tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: b.desired.HealthPort,
	})

	ports = append(ports, corev1.ContainerPort{
		Name:          prometheusServiceName,
		ContainerPort: b.desired.PrometheusPort,
	})

	container := corev1.Container{
		Name:            constants.FLPName + b.confKindSuffix,
		Image:           b.desired.Image,
		ImagePullPolicy: corev1.PullPolicy(b.desired.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Resources.DeepCopy(),
		VolumeMounts: []corev1.VolumeMount{{
			MountPath: configPath,
			Name:      configVolume,
		}},
		Ports: ports,
	}
	if b.desired.EnableKubeProbes {
		container.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/live",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds: healthTimeoutSeconds,
			PeriodSeconds:  livenessPeriodSeconds,
		}
		container.StartupProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds:   healthTimeoutSeconds,
			PeriodSeconds:    startupPeriodSeconds,
			FailureThreshold: startupFailureThreshold,
		}
	}
	dnsPolicy := corev1.DNSClusterFirst
	if hostNetwork {
		dnsPolicy = corev1.DNSClusterFirstWithHostNet
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.labels,
			Annotations: map[string]string{
				PodConfigurationDigest: configDigest,
			},
		},
		Spec: corev1.PodSpec{
			Tolerations: tolerations,
			Volumes: []corev1.Volume{{
				Name: configVolume,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName + b.confKindSuffix,
						},
					},
				},
			}},
			Containers:         []corev1.Container{container},
			ServiceAccountName: constants.FLPName + b.confKindSuffix,
			HostNetwork:        hostNetwork,
			DNSPolicy:          dnsPolicy,
		},
	}
}

func (b *builder) addTransformStages(lastStage *config.PipelineBuilderStage) {
	// enrich stage (transform) configuration
	enrichedStage := lastStage.TransformNetwork("enrich", api.TransformNetwork{
		Rules: api.NetworkTransformRules{{
			Input:  "SrcAddr",
			Output: "SrcK8S",
			Type:   api.AddKubernetesRuleType,
		}, {
			Input:  "DstAddr",
			Output: "DstK8S",
			Type:   api.AddKubernetesRuleType,
		}},
	})

	// loki stage (write) configuration
	lokiWrite := api.WriteLoki{
		Labels: constants.LokiIndexFields,
	}

	if b.desiredLoki != nil {
		lokiWrite.BatchSize = int(b.desiredLoki.BatchSize)
		lokiWrite.BatchWait = b.desiredLoki.BatchWait.ToUnstructured().(string)
		lokiWrite.MaxBackoff = b.desiredLoki.MaxBackoff.ToUnstructured().(string)
		lokiWrite.MaxRetries = int(b.desiredLoki.MaxRetries)
		lokiWrite.MinBackoff = b.desiredLoki.MinBackoff.ToUnstructured().(string)
		lokiWrite.StaticLabels = model.LabelSet{}
		for k, v := range b.desiredLoki.StaticLabels {
			lokiWrite.StaticLabels[model.LabelName(k)] = model.LabelValue(v)
		}
		lokiWrite.Timeout = b.desiredLoki.Timeout.ToUnstructured().(string)
		lokiWrite.URL = b.desiredLoki.URL
		lokiWrite.TimestampLabel = "TimeFlowEndMs"
		lokiWrite.TimestampScale = "1ms"
	}
	enrichedStage.WriteLoki("loki", lokiWrite)

	// write on Stdout if logging trace enabled
	if b.desired.LogLevel == "trace" {
		enrichedStage.WriteStdout("stdout", api.WriteStdout{Format: "json"})
	}

	// prometheus stage (encode) configuration
	agg := enrichedStage.Aggregate("aggregate", []api.AggregateDefinition{})
	agg.EncodePrometheus("prometheus", api.PromEncode{
		Port:   int(b.desired.PrometheusPort),
		Prefix: "flp_",
	})
}

func (b *builder) buildPipelineConfig() ([]config.Stage, []config.StageParam) {
	var pipeline config.PipelineBuilderStage
	if b.confKind == ConfKafkaTransformer {
		pipeline = config.NewKafkaPipeline("ingest", api.IngestKafka{
			Brokers: []string{b.desiredKafka.Address},
			Topic:   b.desiredKafka.Topic,
			GroupId: b.confKind, // Without groupid, each message is delivered to each consumers
		})
		pipeline = pipeline.DecodeJSON("decode")
	} else if b.portProtocol == corev1.ProtocolUDP {
		// UDP Port: IPFIX collector with JSON decoder
		pipeline = config.NewCollectorPipeline("ingest", api.IngestCollector{
			Port:     int(b.desired.Port),
			HostName: "0.0.0.0",
		})
		pipeline = pipeline.DecodeJSON("decode")
	} else {
		// TCP Port: GRPC collector (eBPF agent) with Protobuf decoder
		pipeline = config.NewGRPCPipeline("ingest", api.IngestGRPCProto{
			Port: int(b.desired.Port),
		})
		pipeline = pipeline.DecodeProtobuf("decode")
	}

	if b.confKind == ConfKafkaIngester {
		pipeline = pipeline.EncodeKafka("kafka", api.EncodeKafka{
			Address: b.desiredKafka.Address,
			Topic:   b.desiredKafka.Topic,
		})
	} else {
		b.addTransformStages(&pipeline)
	}
	return pipeline.GetStages(), pipeline.GetStageParams()
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string) {
	stages, parameters := b.buildPipelineConfig()

	config := map[string]interface{}{
		"log-level": b.desired.LogLevel,
		"health": map[string]interface{}{
			"port": b.desired.HealthPort,
		},
		"pipeline":   stages,
		"parameters": parameters,
	}

	configStr := "{}"
	bs, err := json.Marshal(config)
	if err == nil {
		configStr = string(bs)
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName + b.confKindSuffix,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Data: map[string]string{
			configFile: configStr,
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(configStr))
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest
}

func (b *builder) service(old *corev1.Service) *corev1.Service {
	if old == nil {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.FLPName, //We don't add suffix here so we always use the same service
				Namespace: b.namespace,
				Labels:    b.labels,
			},
			Spec: corev1.ServiceSpec{
				Selector:        b.selector,
				SessionAffinity: corev1.ServiceAffinityClientIP,
				Ports: []corev1.ServicePort{{
					Port:     b.desired.Port,
					Protocol: b.portProtocol,
				}},
			},
		}
	}
	// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
	newService := old.DeepCopy()
	newService.Spec.Selector = b.selector
	newService.Spec.SessionAffinity = corev1.ServiceAffinityClientIP
	newService.Spec.Ports = []corev1.ServicePort{{
		Port:     b.desired.Port,
		Protocol: b.portProtocol,
	}}
	newService.ObjectMeta.Labels = b.labels
	return newService
}

func (b *builder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName + b.confKindSuffix,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind:       constants.DeploymentKind,
				Name:       constants.FLPName + b.confKindSuffix,
				APIVersion: "apps/v1",
			},
			MinReplicas: b.desired.HPA.MinReplicas,
			MaxReplicas: b.desired.HPA.MaxReplicas,
			Metrics:     b.desired.HPA.Metrics,
		},
	}
}

// The operator needs to have at least the same permissions as flowlogs-pipeline in order to grant them
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=core,resources=pods;services;nodes,verbs=get;list;watch

func buildAppLabel(confKind string) map[string]string {
	return map[string]string{
		"app": constants.FLPName + FlpConfSuffix[confKind],
	}
}

func buildClusterRoleIngester(useOpenShiftSCC bool) *rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.FLPName + FlpConfSuffix[ConfKafkaIngester],
			Labels: buildAppLabel(""),
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"pods", "services", "nodes"},
		}, {
			APIGroups: []string{"apps"},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"replicasets"},
		}, {
			APIGroups: []string{"autoscaling"},
			Verbs:     []string{"create", "delete", "patch", "update", "get", "watch", "list"},
			Resources: []string{"horizontalpodautoscalers"},
		}},
	}
	if useOpenShiftSCC {
		cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"security.openshift.io"},
			Verbs:         []string{"use"},
			Resources:     []string{"securitycontextconstraints"},
			ResourceNames: []string{"hostnetwork"},
		})
	}
	return &cr
}

func buildClusterRoleTransformer(useOpenShiftSCC bool) *rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.FLPName + FlpConfSuffix[ConfKafkaTransformer],
			Labels: buildAppLabel(""),
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"pods", "services", "nodes"},
		}, {
			APIGroups: []string{"apps"},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"replicasets"},
		}, {
			APIGroups: []string{"autoscaling"},
			Verbs:     []string{"create", "delete", "patch", "update", "get", "watch", "list"},
			Resources: []string{"horizontalpodautoscalers"},
		}, {
			APIGroups:     []string{"security.openshift.io"},
			Verbs:         []string{"use"},
			Resources:     []string{"securitycontextconstraints"},
			ResourceNames: []string{"hostnetwork"},
		}},
	}
	if useOpenShiftSCC {
		cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"security.openshift.io"},
			Verbs:         []string{"use"},
			Resources:     []string{"securitycontextconstraints"},
			ResourceNames: []string{"hostnetwork"},
		})
	}
	return &cr
}

func (b *builder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName + b.confKindSuffix,
			Namespace: b.namespace,
			Labels:    buildAppLabel(""),
		},
	}
}

func (b *builder) clusterRoleBinding(roleKind string) *rbacv1.ClusterRoleBinding {
	//Adding role here to disembiguate between the deployment kind and the role binded
	name := constants.FLPName + b.confKindSuffix + FlpConfSuffix[roleKind] + "role"
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: buildAppLabel(""),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.FLPName + FlpConfSuffix[roleKind],
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.FLPName + b.confKindSuffix,
			Namespace: b.namespace,
		}},
	}
}
