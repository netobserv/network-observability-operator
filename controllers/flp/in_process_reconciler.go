package flp

import (
	"context"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type inProcessReconciler struct {
	monolith *monolithReconciler
}

type InProcessInfo struct {
	JSONConfig  string
	Annotations map[string]string
	Volumes     volumes.Builder
}

func ReconcileInProcess(ctx context.Context, parent *reconcilers.Instance, desired *flowslatest.FlowCollector) (*InProcessInfo, error) {
	i := newInProcessReconciler(parent)
	return i.reconcileInProcess(ctx, desired)
}

func newInProcessReconciler(parent *reconcilers.Instance) *inProcessReconciler {
	cloneInfo := *parent.Common
	cloneInfo.Namespace = parent.PrivilegedNamespace()
	inst := cloneInfo.NewInstance(parent.Image, parent.Status)
	m := newMonolithReconciler(inst)
	return &inProcessReconciler{monolith: m}
}

func (i *inProcessReconciler) reconcileInProcess(ctx context.Context, desired *flowslatest.FlowCollector) (*InProcessInfo, error) {
	result := InProcessInfo{}

	// Retrieve current owned objects
	err := i.monolith.Managed.FetchAll(ctx)
	if err != nil {
		return nil, err
	}

	fm := metricslatest.FlowMetricList{}
	if err := i.monolith.List(ctx, &fm, &client.ListOptions{Namespace: desired.Namespace}); err != nil {
		return nil, i.monolith.Status.Error("CantListFlowMetrics", err)
	}

	builder, err := newMonolithBuilder(i.monolith.Instance, &desired.Spec, &fm)
	if err != nil {
		return nil, err
	}

	// Override target app
	builder.generic.overrideApp(constants.EBPFAgentName)
	// Build pipeline
	pipeline := builder.generic.NewInProcessPipeline()
	err = pipeline.AddProcessorStages()
	if err != nil {
		return nil, err
	}
	cfg, err := builder.generic.GetJSONConfig()
	if err != nil {
		return nil, err
	}
	result.JSONConfig = cfg

	err = i.reconcileRoles(ctx, &builder)
	if err != nil {
		return nil, err
	}

	err = i.monolith.reconcilePrometheusService(ctx, &builder)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{}
	err = i.monolith.reconcileCertificates(ctx, desired, annotations)
	if err != nil {
		return nil, err
	}
	result.Annotations = annotations
	result.Volumes = builder.generic.volumes
	return &result, nil
}

func (i *inProcessReconciler) reconcileRoles(ctx context.Context, builder *monolithBuilder) error {
	cr := BuildClusterRoleTransformer()
	if err := i.monolith.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name + "-agent",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     cr.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      constants.EBPFServiceAccount,
			Namespace: i.monolith.Namespace,
		}},
	}
	if err := i.monolith.ReconcileClusterRoleBinding(ctx, crb); err != nil {
		return err
	}

	return ReconcileLokiRoles(
		ctx,
		i.monolith.Common,
		builder.generic.desired,
		constants.EBPFAgentName,
		constants.EBPFServiceAccount,
		i.monolith.Namespace,
	)
}
