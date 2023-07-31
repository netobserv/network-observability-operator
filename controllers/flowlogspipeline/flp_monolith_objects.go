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

type monolithBuilder struct {
	generic builder
}

func newMonolithBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec) monolithBuilder {
	gen := newBuilder(info, desired, ConfMonolith)
	return monolithBuilder{
		generic: gen,
	}
}

func (b *monolithBuilder) daemonSet(annotations map[string]string) *appsv1.DaemonSet {
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

func (b *monolithBuilder) configMap() (*corev1.ConfigMap, string, error) {
	stages, params, err := b.buildPipelineConfig()
	if err != nil {
		return nil, "", err
	}
	pipelineConfigMap, digest, err := b.generic.configMap(stages, params)
	return pipelineConfigMap, digest, err
}

func (b *monolithBuilder) buildPipelineConfig() ([]config.Stage, []config.StageParam, error) {
	var pipeline config.PipelineBuilderStage
	if helper.UseIPFIX(b.generic.desired) {
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

	err := b.generic.addTransformStages(&pipeline)
	if err != nil {
		return nil, nil, err
	}
	return pipeline.GetStages(), pipeline.GetStageParams(), nil
}

func (b *monolithBuilder) promService() *corev1.Service {
	return b.generic.promService()
}

func (b *monolithBuilder) serviceAccount() *corev1.ServiceAccount {
	return b.generic.serviceAccount()
}

func (b *monolithBuilder) clusterRoleBinding(ck ConfKind) *rbacv1.ClusterRoleBinding {
	return b.generic.clusterRoleBinding(ck, true)
}
