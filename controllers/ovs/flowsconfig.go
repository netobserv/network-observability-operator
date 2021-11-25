package ovs

import (
	"context"
	"fmt"
	"net"
	"strconv"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type FlowsConfigController struct {
	ovsConfigMapName    string
	goflowkubeNamespace string
	cnoNamespace        string
	client              client.Client
	lookupIP            func(string) ([]net.IP, error)
}

func NewFlowsConfigController(client client.Client,
	goflowkubeNamespace, cnoNamespace, ovsConfigMapName string) *FlowsConfigController {
	return &FlowsConfigController{
		client:              client,
		goflowkubeNamespace: goflowkubeNamespace,
		cnoNamespace:        cnoNamespace,
		ovsConfigMapName:    ovsConfigMapName,
		lookupIP:            net.LookupIP,
	}
}

// NewTestFlowsConfigController allows creating a FlowsConfigController instance that with an
// injected IP resolver for testing.
func NewTestFlowsConfigController(client client.Client,
	goflowkubeNamespace, cnoNamespace, ovsConfigMapName string,
	lookupIP func(string) ([]net.IP, error),
) *FlowsConfigController {
	fc := NewFlowsConfigController(client, goflowkubeNamespace, cnoNamespace, ovsConfigMapName)
	fc.lookupIP = lookupIP
	return fc
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
	err := c.client.Delete(ctx, &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.ovsConfigMapName,
			Namespace: c.cnoNamespace,
		},
	})
	if errors.IsNotFound(err) {
		rlog := log.FromContext(ctx, "component", "FlowsConfigController")
		rlog.Error(err, "can't delete non-existing configmap. Ignoring",
			"name", c.ovsConfigMapName, "namespace", c.cnoNamespace)
		return nil
	}
	return err
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
		return &conf, nil
	case constants.DeploymentKind:
		svc := corev1.Service{}
		if err := c.client.Get(ctx, types.NamespacedName{
			Namespace: c.goflowkubeNamespace,
			Name:      constants.GoflowKubeName,
		}, &svc); err != nil {
			return nil, err
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
		// TODO: if spec/goflowkube is empty or port is empty, fetch first UDP port in the service spec
		conf.SharedTarget = net.JoinHostPort(ip, strconv.Itoa(int(coll.Spec.GoflowKube.Port)))
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
