package flowlogspipeline

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

type ingestBuilder struct {
	generic builder
}

func newIngestBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec) (ingestBuilder, error) {
	gen, err := newBuilder(info, desired, ConfKafkaIngester)
	return ingestBuilder{
		generic: gen,
	}, err
}

func (b *ingestBuilder) daemonSet(annotations map[string]string) *appsv1.DaemonSet {
	pod := b.generic.podTemplate(true /*listens*/, !b.generic.info.UseOpenShiftSCC, annotations)
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.generic.name(),
			Namespace: b.generic.info.Namespace,
			Labels:    b.generic.labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.generic.selector,
			},
			Template: pod,
		},
	}
}

func (b *ingestBuilder) configMap() (*corev1.ConfigMap, string, error) {
	stages, params, err := b.buildPipelineConfig()
	if err != nil {
		return nil, "", err
	}
	return b.generic.configMap(stages, params)
}

func (b *ingestBuilder) buildPipelineConfig() ([]config.Stage, []config.StageParam, error) {
	var pipeline config.PipelineBuilderStage
	if helper.UseIpfix(b.generic.desired) {
		// IPFIX collector
		pipeline = config.NewCollectorPipeline("ipfix", api.IngestCollector{
			Port:     int(b.generic.desired.Processor.Port),
			HostName: "0.0.0.0",
		})
	} else {
		// GRPC collector (eBPF agent)
		pipeline = config.NewGRPCPipeline("grpc", api.IngestGRPCProto{
			Port: int(b.generic.desired.Processor.Port),
		})
	}

	pipeline = b.generic.createKafkaWriteStage("kafka-write", &b.generic.desired.Kafka, &pipeline)

	return pipeline.GetStages(), pipeline.GetStageParams(), nil
}

func (b *ingestBuilder) promService() *corev1.Service {
	return b.generic.promService()
}

func buildClusterRoleIngester(useOpenShiftSCC bool) *rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name(ConfKafkaIngester),
		},
		Rules: []rbacv1.PolicyRule{},
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

func (b *ingestBuilder) serviceAccount() *corev1.ServiceAccount {
	return b.generic.serviceAccount()
}

func (b *ingestBuilder) clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return b.generic.clusterRoleBinding(ConfKafkaIngester, false)
}
