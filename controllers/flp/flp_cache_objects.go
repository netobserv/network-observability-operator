package flp

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	flpCacheName      = constants.FLPName + "-cache"
	flpCacheConfigMap = "flp-cache-config"
	flpCacheTopic     = "informers"
)

func (b *builder) cachePodTemplate(annotations map[string]string) corev1.PodTemplateSpec {
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	var ports []corev1.ContainerPort

	if advancedConfig.ProfilePort != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: *advancedConfig.ProfilePort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	volumeMounts := b.volumes.AppendMounts([]corev1.VolumeMount{{
		MountPath: configPath + "-cache",
		Name:      configVolume + "-cache",
	}})
	volumes := b.volumes.AppendVolumes([]corev1.Volume{{
		Name: configVolume + "-cache",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: flpCacheConfigMap,
				},
			},
		},
	}})

	var envs []corev1.EnvVar
	envs = append(envs, constants.EnvNoHTTP2)
	imageName := strings.Replace(b.info.Image, "flowlogs-pipeline", "flowlogs-pipeline-cache", 1)

	container := corev1.Container{
		Name:            flpCacheName,
		Image:           imageName,
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath+"-cache", configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
		SecurityContext: helper.ContainerDefaultSecurityContext(),
	}
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.cacheLabels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: b.name(),
			NodeSelector:       advancedConfig.Scheduling.NodeSelector,
			Tolerations:        advancedConfig.Scheduling.Tolerations,
			Affinity:           advancedConfig.Scheduling.Affinity,
			PriorityClassName:  advancedConfig.Scheduling.PriorityClassName,
		},
	}
}

type CacheConfig struct {
	KubeConfigPath string          `yaml:"kubeConfigPath"`
	KafkaConfig    api.EncodeKafka `yaml:"kafkaConfig"`
	PProfPort      int32           `yaml:"pprofPort"`
	LogLevel       string          `yaml:"logLevel"`
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) cacheConfigMap() (*corev1.ConfigMap, string, error) {
	// Re-use the initial stage (which should be Kafka ingester), with a different topic
	// TODO: that's ugly and deserves more refactoring
	params := b.pipeline.GetStageParams()[0]

	kafkaSpec := b.desired.Kafka
	cc := CacheConfig{
		LogLevel: b.desired.Processor.LogLevel,
		KafkaConfig: api.EncodeKafka{
			Address: kafkaSpec.Address,
			Topic:   flpCacheTopic,
			TLS:     params.Ingest.Kafka.TLS,
			SASL:    params.Ingest.Kafka.SASL,
		},
	}
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	if advancedConfig.ProfilePort != nil {
		cc.PProfPort = *advancedConfig.ProfilePort
	}

	bs, err := yaml.Marshal(cc)
	if err != nil {
		return nil, "", err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flpCacheConfigMap,
			Namespace: b.info.Namespace,
			Labels:    b.cacheLabels,
		},
		Data: map[string]string{
			configFile: string(bs),
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write(bs)
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest, nil
}
