package ovs

import (
	"context"
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

type FlowsConfigCNOController struct {
	ovsConfigMapName   string
	collectorNamespace string
	cnoNamespace       string
	client             reconcilers.ClientHelper
	lookupIP           func(string) ([]net.IP, error)
}

func NewFlowsConfigCNOController(client reconcilers.ClientHelper,
	collectorNamespace, cnoNamespace, ovsConfigMapName string,
	lookupIP func(string) ([]net.IP, error)) *FlowsConfigCNOController {
	return &FlowsConfigCNOController{
		client:             client,
		collectorNamespace: collectorNamespace,
		cnoNamespace:       cnoNamespace,
		ovsConfigMapName:   ovsConfigMapName,
		lookupIP:           lookupIP,
	}
}

// Reconcile reconciles the status of the ovs-flows-config configmap with
// the target FlowCollector ipfix section map
func (c *FlowsConfigCNOController) Reconcile(ctx context.Context, target *flowslatest.FlowCollector) error {
	rlog := log.FromContext(ctx, "component", "FlowsConfigCNOController")

	current, err := c.current(ctx)
	if err != nil {
		return err
	}
	if !target.Spec.UseIPFIX() {
		if current == nil {
			return nil
		}
		// If the user has changed the agent type, we need to manually undeploy the configmap
		if current != nil {
			return c.client.Delete(ctx, &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      c.ovsConfigMapName,
					Namespace: c.cnoNamespace,
				},
			})
		}
		return nil
	}

	desired := c.desired(ctx, target)

	// compare current and desired
	if current == nil {
		rlog.Info("Provided IPFIX configuration. Creating " + c.ovsConfigMapName + " ConfigMap")
		cm, err := c.flowsConfigMap(desired)
		if err != nil {
			return err
		}
		return c.client.Create(ctx, cm)
	}

	if desired != nil && *desired != *current {
		rlog.Info("Provided IPFIX configuration differs current configuration. Updating")
		cm, err := c.flowsConfigMap(desired)
		if err != nil {
			return err
		}
		return c.client.Update(ctx, cm)
	}

	rlog.Info("No changes needed")
	return nil
}

func (c *FlowsConfigCNOController) current(ctx context.Context) (*flowsConfig, error) {
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

func (c *FlowsConfigCNOController) desired(
	ctx context.Context, coll *flowslatest.FlowCollector) *flowsConfig {

	corrected := coll.Spec.Agent.IPFIX.DeepCopy()
	corrected.Sampling = getSampling(ctx, corrected)

	return &flowsConfig{
		FlowCollectorIPFIX: *corrected,
		NodePort:           coll.Spec.Processor.Port,
	}
}

func (c *FlowsConfigCNOController) flowsConfigMap(fc *flowsConfig) (*corev1.ConfigMap, error) {
	data, err := fc.asStringMap()
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.ovsConfigMapName,
			Namespace: c.cnoNamespace,
		},
		Data: data,
	}
	if err := c.client.SetControllerReference(cm); err != nil {
		return nil, err
	}
	return cm, nil
}
