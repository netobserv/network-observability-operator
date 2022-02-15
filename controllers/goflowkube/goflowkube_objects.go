package goflowkube

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const configMapName = "goflow-kube-config"
const configVolume = "config-volume"
const configPath = "/etc/goflow-kube"
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
const PodConfigurationDigest = "flows.netobserv.io/goflow-kube-config"

type ConfigMap struct {
	Listen      string        `json:"listen,omitempty"`
	Loki        LokiConfigMap `json:"loki,omitempty"`
	PrintInput  bool          `json:"printInput"`
	PrintOutput bool          `json:"printOutput"`
}

type LokiConfigMap struct {
	URL            string            `json:"url,omitempty"`
	BatchWait      metav1.Duration   `json:"batchWait,omitempty"`
	BatchSize      int64             `json:"batchSize,omitempty"`
	Timeout        metav1.Duration   `json:"timeout,omitempty"`
	MinBackoff     metav1.Duration   `json:"minBackoff,omitempty"`
	MaxBackoff     metav1.Duration   `json:"maxBackoff,omitempty"`
	MaxRetries     int32             `json:"maxRetries,omitempty"`
	Labels         []string          `json:"labels,omitempty"`
	StaticLabels   map[string]string `json:"staticLabels,omitempty"`
	TimestampLabel string            `json:"timestampLabel,omitempty"`
}

type builder struct {
	namespace   string
	labels      map[string]string
	selector    map[string]string
	desired     *flowsv1alpha1.FlowCollectorGoflowKube
	desiredLoki *flowsv1alpha1.FlowCollectorLoki
}

func newBuilder(ns string, desired *flowsv1alpha1.FlowCollectorGoflowKube, desiredLoki *flowsv1alpha1.FlowCollectorLoki) builder {
	version := helper.ExtractVersion(desired.Image)
	return builder{
		namespace: ns,
		labels: map[string]string{
			"app":     constants.GoflowKubeName,
			"version": version,
		},
		selector: map[string]string{
			"app": constants.GoflowKubeName,
		},
		desired:     desired,
		desiredLoki: desiredLoki,
	}
}

func (b *builder) deployment(configDigest string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.GoflowKubeName,
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
			Name:      constants.GoflowKubeName,
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
	cmd := buildMainCommand(b.desired)
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if b.desired.Kind == constants.DaemonSetKind {
		ports = []corev1.ContainerPort{{
			Name:          constants.GoflowKubeName,
			HostPort:      b.desired.Port,
			ContainerPort: b.desired.Port,
			Protocol:      corev1.ProtocolUDP,
		}}
		// This allows deploying an instance in the master node, the same technique used in the
		// companion ovnkube-node daemonset definition
		tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: b.desired.HealthPort,
	})

	healthProbe := corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/health/live",
			Port: intstr.FromString(healthServiceName),
		},
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
			Containers: []corev1.Container{{
				Name:            constants.GoflowKubeName,
				Image:           b.desired.Image,
				ImagePullPolicy: corev1.PullPolicy(b.desired.ImagePullPolicy),
				Command:         []string{"/bin/sh", "-c", cmd},
				Resources:       *b.desired.Resources.DeepCopy(),
				VolumeMounts: []corev1.VolumeMount{{
					MountPath: configPath,
					Name:      configVolume,
				}},
				Ports: ports,
				LivenessProbe: &corev1.Probe{
					ProbeHandler:   healthProbe,
					TimeoutSeconds: healthTimeoutSeconds,
					PeriodSeconds:  livenessPeriodSeconds,
				},
				StartupProbe: &corev1.Probe{
					ProbeHandler:     healthProbe,
					TimeoutSeconds:   healthTimeoutSeconds,
					PeriodSeconds:    startupPeriodSeconds,
					FailureThreshold: startupFailureThreshold,
				},
			}},
			ServiceAccountName: constants.GoflowKubeName,
		},
	}
}

func buildMainCommand(desired *flowsv1alpha1.FlowCollectorGoflowKube) string {
	return fmt.Sprintf(`/goflow-kube -loglevel "%s" -config %s/%s -healthport %d`,
		desired.LogLevel, configPath, configFile, desired.HealthPort)
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap() (*corev1.ConfigMap, string) {
	configStr := `{}`
	config := &ConfigMap{
		Listen:      fmt.Sprintf("netflow://:%d", b.desired.Port),
		Loki:        LokiConfigMap{},
		PrintInput:  false,
		PrintOutput: b.desired.PrintOutput,
	}
	if b.desiredLoki != nil {
		config.Loki.BatchSize = b.desiredLoki.BatchSize
		config.Loki.BatchWait = b.desiredLoki.BatchWait
		config.Loki.MaxBackoff = b.desiredLoki.MaxBackoff
		config.Loki.MaxRetries = b.desiredLoki.MaxRetries
		config.Loki.MinBackoff = b.desiredLoki.MinBackoff
		config.Loki.StaticLabels = b.desiredLoki.StaticLabels
		config.Loki.Timeout = b.desiredLoki.Timeout
		config.Loki.URL = b.desiredLoki.URL
		config.Loki.TimestampLabel = b.desiredLoki.TimestampLabel
	}
	config.Loki.Labels = constants.Labels

	bs, err := json.Marshal(config)
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
				Name:      constants.GoflowKubeName,
				Namespace: b.namespace,
				Labels:    b.labels,
			},
			Spec: corev1.ServiceSpec{
				Selector:        b.selector,
				SessionAffinity: corev1.ServiceAffinityClientIP,
				Ports: []corev1.ServicePort{{
					Port:     b.desired.Port,
					Protocol: corev1.ProtocolUDP,
				}},
			},
		}
	}
	// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
	newService := old.DeepCopy()
	newService.Spec.Ports = []corev1.ServicePort{{
		Port:     b.desired.Port,
		Protocol: corev1.ProtocolUDP,
	}}
	return newService
}

func (b *builder) autoScaler() *ascv1.HorizontalPodAutoscaler {
	return &ascv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.GoflowKubeName,
			Namespace: b.namespace,
			Labels:    b.labels,
		},
		Spec: ascv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv1.CrossVersionObjectReference{
				Kind:       constants.DeploymentKind,
				Name:       constants.GoflowKubeName,
				APIVersion: "apps/v1",
			},
			MinReplicas:                    b.desired.HPA.MinReplicas,
			MaxReplicas:                    b.desired.HPA.MaxReplicas,
			TargetCPUUtilizationPercentage: b.desired.HPA.TargetCPUUtilizationPercentage,
		},
	}
}

// The operator needs to have at least the same permissions as goflow-kube in order to grant them
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;delete;patch;update;get;watch;list
//+kubebuilder:rbac:groups=core,resources=pods;services,verbs=get;list;watch

func buildAppLabel() map[string]string {
	return map[string]string{
		"app": constants.GoflowKubeName,
	}
}

func buildClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.GoflowKubeName,
			Labels: buildAppLabel(),
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Verbs:     []string{"list", "get", "watch"},
			Resources: []string{"pods", "services"},
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
			Name:      constants.GoflowKubeName,
			Namespace: ns,
			Labels:    buildAppLabel(),
		},
	}
}

func buildClusterRoleBinding(ns string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.GoflowKubeName,
			Labels: buildAppLabel(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.GoflowKubeName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.GoflowKubeName,
			Namespace: ns,
		}},
	}
}
