package ovs

import (
	"context"
	"fmt"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
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
func (c *FlowsConfigCNOController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	rlog := log.FromContext(ctx, "component", "FlowsConfigCNOController")

	current, err := c.current(ctx)
	if err != nil {
		return err
	}
	if target.Spec.Agent.Type != flowsv1alpha1.AgentIPFIX {
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

	desired, err := c.desired(ctx, target)
	// compare current and desired
	if err != nil {
		return err
	}

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
	ctx context.Context, coll *flowsv1alpha1.FlowCollector) (*flowsConfig, error) {

	conf := flowsConfig{FlowCollectorIPFIX: coll.Spec.Agent.IPFIX}

	// According to the "OVS flow export configuration" RFE:
	// nodePort be set by the NOO when the collector is deployed as a DaemonSet
	// sharedTarget set when deployed as Deployment + Service
	switch coll.Spec.FlowlogsPipeline.Kind {
	case constants.DaemonSetKind:
		conf.NodePort = coll.Spec.FlowlogsPipeline.Port
		return &conf, nil
	case constants.DeploymentKind:
		svc := corev1.Service{}
		if err := c.client.Get(ctx, types.NamespacedName{
			Namespace: c.collectorNamespace,
			Name:      constants.FLPName,
		}, &svc); err != nil {
			return nil, fmt.Errorf("can't get service %s in %s: %w", constants.FLPName, c.collectorNamespace, err)
		}
		// service IP resolution
		svcHost := svc.Name + "." + svc.Namespace
		addrs, err := c.lookupIP(svcHost)
		if err != nil {
			return nil, fmt.Errorf("can't resolve IP address for service %v: %w", svcHost, err)
		}
		var ip string
		for _, addr := range addrs {
			if len(addr) > 0 {
				ip = addr.String()
				break
			}
		}
		if ip == "" {
			return nil, fmt.Errorf("can't find any suitable IP for host %s", svcHost)
		}
		// TODO: if spec/flowlogsPipeline is empty or port is empty, fetch first UDP port in the service spec
		conf.SharedTarget = net.JoinHostPort(ip, strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)))
		return &conf, nil
	}
	return nil, fmt.Errorf("unexpected flowlogsPipeline kind: %s", coll.Spec.FlowlogsPipeline.Kind)
}

func (c *FlowsConfigCNOController) flowsConfigMap(fc *flowsConfig) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
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
	if err := c.client.SetControllerReference(cm); err != nil {
		return nil, err
	}
	return cm, nil
}
