package ovs

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/controllers/constants"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

type FlowsConfigController struct {
	ovsConfigMapName  string
	operatorNamespace string
	cnoNamespace      string
	client            client.Client
}

func NewFlowsConfigController(client client.Client,
	operatorNamespace, cnoNamespace, ovsConfigMapName string) FlowsConfigController {
	return FlowsConfigController{
		client:            client,
		operatorNamespace: operatorNamespace,
		cnoNamespace:      cnoNamespace,
		ovsConfigMapName:  ovsConfigMapName,
	}
}

// Reconcile reconciles the status of the ovs-flows-config configmap with
// the target FlowCollector ipfix section map
func (c *FlowsConfigController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	rlog := log.FromContext(ctx, "component", "FlowsConfigController")

	if !target.ObjectMeta.DeletionTimestamp.IsZero() {
		rlog.Info("no need to reconcile status of a FlowCollector that is being deleted. Ignoring")
		return nil
	}
	current, err := c.current(ctx)
	if err != nil {
		return err
	}
	desired, err := c.desired(ctx, target)
	// compare current and desired
	if err != nil {
		return err
	}

	if current == nil {
		rlog.Info("Provided IPFIX configuration. Creating " + c.ovsConfigMapName + " ConfigMap")
		return c.client.Create(ctx, c.flowsConfigMap(desired))
	}

	if desired != nil && *desired != *current {
		rlog.Info("Provided IPFIX configuration differs current configuration. Updating")
		return c.client.Update(ctx, c.flowsConfigMap(desired))
	}

	rlog.Info("No changes needed")
	return nil
}

func (c *FlowsConfigController) Finalize(ctx context.Context) error {
	return c.client.Delete(ctx, &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.ovsConfigMapName,
			Namespace: c.cnoNamespace,
		},
	})
}

func (c *FlowsConfigController) current(ctx context.Context) (*flowsConfig, error) {
	curr := &corev1.ConfigMap{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      c.ovsConfigMapName,
		Namespace: c.cnoNamespace,
	}, curr); err != nil {
		if errors.IsNotFound(err) {
			// the map is not yet created. As it is associated to a flowCollector that already
			// exists (premise to invoke this controller). We will handle accordingly this "nil"
			// as an expected value
			return nil, nil
		}
		return nil, fmt.Errorf("retrieving %s/%s configmap: %w",
			c.cnoNamespace, c.ovsConfigMapName, err)
	}

	return configFromMap(curr.Data)
}

func (c *FlowsConfigController) desired(
	ctx context.Context, coll *flowsv1alpha1.FlowCollector) (*flowsConfig, error) {

	conf := flowsConfig{FlowCollectorIPFIX: coll.Spec.IPFIX}

	// According to the "OVS flow export configuration" RFE:
	// nodePort be set by the NOO when the collector is deployed as a DaemonSet
	// sharedTarget set when deployed as Deployment + Service
	switch coll.Spec.GoflowKube.Kind {
	case constants.DaemonSetKind:
		conf.NodePort = coll.Spec.GoflowKube.Port
	case constants.DeploymentKind:
		svc := corev1.Service{}
		if err := c.client.Get(ctx, types.NamespacedName{
			Namespace: c.operatorNamespace,
			Name:      constants.GoflowKubeName,
		}, &svc); err != nil {
			return nil, err
		}
		// TODO: if spec/goflowkube is empty or port is empty, fetch first UDP port in the service spec
		conf.SharedTarget = fmt.Sprintf("%s.%s:%d", svc.Name, svc.Namespace, coll.Spec.GoflowKube.Port)
		return &conf, nil
	}
	return nil, fmt.Errorf("unexpected GoflowKube kind: %s", coll.Spec.GoflowKube.Kind)
}

func (c *FlowsConfigController) flowsConfigMap(fc *flowsConfig) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.ovsConfigMapName,
			Namespace: c.cnoNamespace,
		},
		Data: fc.asStringMap(),
	}
}
