package ovs

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
)

type FlowsConfigOVNKController struct {
	namespace string
	config    flowsv1alpha1.OVNKubernetesConfig
	client    reconcilers.ClientHelper
	lookupIP  func(string) ([]net.IP, error)
}

func NewFlowsConfigOVNKController(client reconcilers.ClientHelper, namespace string, config flowsv1alpha1.OVNKubernetesConfig, lookupIP func(string) ([]net.IP, error)) *FlowsConfigOVNKController {
	return &FlowsConfigOVNKController{
		client:    client,
		namespace: namespace,
		config:    config,
		lookupIP:  lookupIP,
	}
}

// Reconcile reconciles the status of the ovs-flows-config configmap with
// the target FlowCollector ipfix section map
func (c *FlowsConfigOVNKController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {

	desiredEnv, err := c.desiredEnv(ctx, target)
	if err != nil {
		return err
	}

	return c.updateEnv(ctx, target, desiredEnv)
}

func (c *FlowsConfigOVNKController) updateEnv(ctx context.Context, target *flowsv1alpha1.FlowCollector, desiredEnv map[string]string) error {
	rlog := log.FromContext(ctx, "component", "FlowsConfigOVNKController")

	ds, err := c.getDaemonSet(ctx)
	if err != nil {
		return err
	}

	ovnkubeNode := reconcilers.FindContainer(&ds.Spec.Template.Spec, target.Spec.OVNKubernetes.ContainerName)
	if ovnkubeNode == nil {
		return errors.New("could not find container ovnkube-node")
	}

	anyUpdate := false
	for k, v := range desiredEnv {
		if checkUpdateEnv(k, v, ovnkubeNode) {
			anyUpdate = true
		}
	}
	if anyUpdate {
		rlog.Info("Provided IPFIX configuration differs current configuration. Updating")
		return c.client.Update(ctx, ds)
	}

	rlog.Info("No changes needed")
	return nil
}

func (c *FlowsConfigOVNKController) getDaemonSet(ctx context.Context) (*appsv1.DaemonSet, error) {
	curr := &appsv1.DaemonSet{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      c.config.DaemonSetName,
		Namespace: c.config.Namespace,
	}, curr); err != nil {
		return nil, fmt.Errorf("retrieving %s/%s daemonset: %w", c.config.Namespace, c.config.DaemonSetName, err)
	}
	return curr, nil
}

func (c *FlowsConfigOVNKController) desiredEnv(ctx context.Context, coll *flowsv1alpha1.FlowCollector) (map[string]string, error) {
	cacheTimeout, err := time.ParseDuration(coll.Spec.IPFIX.CacheActiveTimeout)
	if err != nil {
		return nil, err
	}
	// Adapt sampling if necessary. See https://bugzilla.redhat.com/show_bug.cgi?id=2103136 , https://bugzilla.redhat.com/show_bug.cgi?id=2104943
	sampling := correctSampling(ctx, &coll.Spec.IPFIX)

	envs := map[string]string{
		"OVN_IPFIX_TARGETS":              "",
		"OVN_IPFIX_CACHE_ACTIVE_TIMEOUT": strconv.Itoa(int(cacheTimeout.Seconds())),
		"OVN_IPFIX_CACHE_MAX_FLOWS":      strconv.Itoa(int(coll.Spec.IPFIX.CacheMaxFlows)),
		"OVN_IPFIX_SAMPLING":             strconv.Itoa(int(sampling)),
	}

	if coll.Spec.Agent != flowsv1alpha1.AgentIPFIX {
		// No IPFIX => leave target empty and return
		return envs, nil
	}

	// According to the "OVS flow export configuration" RFE:
	// nodePort be set by the NOO when the collector is deployed as a DaemonSet
	// sharedTarget set when deployed as Deployment + Service
	switch coll.Spec.FlowlogsPipeline.Kind {
	case constants.DaemonSetKind:
		envs["OVN_IPFIX_TARGETS"] = fmt.Sprintf(":%d", coll.Spec.FlowlogsPipeline.Port)
	case constants.DeploymentKind:
		svc := corev1.Service{}
		if err := c.client.Get(ctx, types.NamespacedName{
			Namespace: c.namespace,
			Name:      constants.FLPName,
		}, &svc); err != nil {
			return nil, fmt.Errorf("can't get service %s in %s: %w", constants.FLPName, c.namespace, err)
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
		envs["OVN_IPFIX_TARGETS"] = net.JoinHostPort(ip, strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)))
	}
	return envs, nil
}

func checkUpdateEnv(name, value string, container *corev1.Container) bool {
	for i, env := range container.Env {
		if env.Name == name {
			if env.Value == value {
				return false
			}
			container.Env[i].Value = value
			return true
		}
	}
	container.Env = append(container.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return true
}

// Finalize will remove IPFIX config from ovn pods env
func (c *FlowsConfigOVNKController) Finalize(ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	// Remove all env
	desiredEnv := map[string]string{
		"OVN_IPFIX_TARGETS":              "",
		"OVN_IPFIX_CACHE_ACTIVE_TIMEOUT": "",
		"OVN_IPFIX_CACHE_MAX_FLOWS":      "",
		"OVN_IPFIX_SAMPLING":             "",
	}
	return c.updateEnv(ctx, target, desiredEnv)
}
