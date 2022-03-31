package flowlogspipeline

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"gopkg.in/yaml.v2"
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
const configFile = "config.yaml"

const (
	healthServiceName       = "health"
	healthTimeoutSeconds    = 5
	livenessPeriodSeconds   = 10
	startupFailureThreshold = 5
	startupPeriodSeconds    = 10
)

// PodConfigurationDigest is an annotation name to facilitate pod restart after
// any external configuration change
const PodConfigurationDigest = "flows.netobserv.io/" + configMapName

type builder struct {
	namespace    string
	labels       map[string]string
	selector     map[string]string
	portProtocol corev1.Protocol
	desired      *flowsv1alpha1.FlowCollectorFLP
	desiredLoki  *flowsv1alpha1.FlowCollectorLoki
}

func newBuilder(ns string, portProtocol corev1.Protocol, desired *flowsv1alpha1.FlowCollectorFLP, desiredLoki *flowsv1alpha1.FlowCollectorLoki) builder {
	version := helper.ExtractVersion(desired.Image)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.FLPName,
			"version": version,
		},
		selector: map[string]string{
			"app": constants.FLPName,
		},
		desired:      desired,
		desiredLoki:  desiredLoki,
		portProtocol: portProtocol,
	}
}

func (b *builder) deployment(configDigest string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &b.desired.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: b.podTemplate(configDigest),
		},
	}
}

func (b *builder) daemonSet(configDigest string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selector,
			},
			Template: b.podTemplate(configDigest),
		},
	}
}

func (b *builder) podTemplate(configDigest string) corev1.PodTemplateSpec {
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if b.desired.Kind == constants.DaemonSetKind {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName,
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

	container := corev1.Container{
		Name:            constants.FLPName,
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
							Name: configMapName,
						},
					},
				},
			}},
			Containers:         []corev1.Container{container},
			ServiceAccountName: constants.FLPName,
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string) {
	lokiWrite := map[string]interface{}{
		"type":   "loki",
		"labels": constants.LokiIndexFields,
	}
	if b.desiredLoki != nil {
		lokiWrite["batchSize"] = b.desiredLoki.BatchSize
		lokiWrite["batchWait"] = b.desiredLoki.BatchWait.ToUnstructured()
		lokiWrite["maxBackoff"] = b.desiredLoki.MaxBackoff.ToUnstructured()
		lokiWrite["maxRetries"] = b.desiredLoki.MaxRetries
		lokiWrite["minBackoff"] = b.desiredLoki.MinBackoff.ToUnstructured()
		lokiWrite["staticLabels"] = b.desiredLoki.StaticLabels
		lokiWrite["timeout"] = b.desiredLoki.Timeout.ToUnstructured()
		lokiWrite["url"] = b.desiredLoki.URL
		lokiWrite["timestampLabel"] = b.desiredLoki.TimestampLabel
	}

	var ingest, decoder map[string]interface{}
	if b.portProtocol == corev1.ProtocolUDP {
		// UDP Port: IPFIX collector with JSON decoder
		ingest = map[string]interface{}{
			"name": "ingest",
			"ingest": map[string]interface{}{
				"type": "collector",
				"collector": map[string]interface{}{
					"port":     b.desired.Port,
					"hostname": "0.0.0.0",
				},
			},
		}
		decoder = map[string]interface{}{
			"name": "decode",
			"decode": map[string]interface{}{
				"type": "json",
			},
		}
	} else {
		// TCP Port: GRPC collector (eBPF agent) with Protobuf decoder
		ingest = map[string]interface{}{
			"name": "ingest",
			"ingest": map[string]interface{}{
				"type": "grpc",
				"grpc": map[string]interface{}{
					"port": b.desired.Port,
				},
			},
		}
		decoder = map[string]interface{}{
			"name": "decode",
			"decode": map[string]interface{}{
				"type": "protobuf",
			},
		}
	}

	config := map[string]interface{}{
		"log-level": b.desired.LogLevel,
		"health": map[string]interface{}{
			"port": b.desired.HealthPort,
		},
		"pipeline": []map[string]string{
			{"name": "ingest"},
			{"name": "decode",
				"follows": "ingest",
			},
			{"name": "enrich",
				"follows": "decode",
			},
			{"name": "encode",
				"follows": "enrich",
			},
			{"name": "loki",
				"follows": "encode",
			},
		},
		"parameters": []map[string]interface{}{
			ingest, decoder,
			{"name": "enrich",
				"transform": map[string]interface{}{
					"type": "network",
					"network": map[string]interface{}{
						"rules": []map[string]interface{}{
							{
								"input":  "SrcAddr",
								"output": "SrcK8S",
								"type":   "add_kubernetes",
							},
							{
								"input":  "DstAddr",
								"output": "DstK8S",
								"type":   "add_kubernetes",
							},
						},
					},
				},
			},
			{"name": "encode",
				"encode": map[string]interface{}{
					"type": "none",
				},
			},
			{"name": "loki",
				"write": map[string]interface{}{
					"type": "loki",
					"loki": lokiWrite,
				},
			},
		},
	}

	configStr := "{}"
	bs, err := yaml.Marshal(config)
	if err == nil {
		configStr = string(bs)
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
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
				Name:      constants.FLPName,
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
	newService.Spec.Ports = []corev1.ServicePort{{
		Port:     b.desired.Port,
		Protocol: b.portProtocol,
	}}
	return newService
}

func (b *builder) autoScaler() *ascv2.HorizontalPodAutoscaler {
	return &ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind:       constants.DeploymentKind,
				Name:       constants.FLPName,
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

func buildAppLabel() map[string]string {
	return map[string]string{
		"app": constants.FLPName,
	}
}

func buildClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.FLPName,
			Labels: buildAppLabel(),
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
}

func buildServiceAccount(ns string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.FLPName,
			Namespace: ns,
			Labels:    buildAppLabel(),
		},
	}
}

func buildClusterRoleBinding(ns string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.FLPName,
			Labels: buildAppLabel(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.FLPName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.FLPName,
			Namespace: ns,
		}},
	}
}
