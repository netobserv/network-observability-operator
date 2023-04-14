package ovs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

type FlowsConfigOVNKController struct {
	*reconcilers.Common
	config flowslatest.OVNKubernetesConfig
}

func NewFlowsConfigOVNKController(common *reconcilers.Common, config flowslatest.OVNKubernetesConfig) *FlowsConfigOVNKController {
	return &FlowsConfigOVNKController{
		Common: common,
		config: config,
	}
}

// Reconcile reconciles the status of the ovs-flows-config configmap with
// the target FlowCollector ipfix section map
func (c *FlowsConfigOVNKController) Reconcile(
	ctx context.Context, target *flowslatest.FlowCollector) error {

	desiredEnv, err := c.desiredEnv(ctx, target)
	if err != nil {
		return err
	}

	return c.updateEnv(ctx, target, desiredEnv)
}

func (c *FlowsConfigOVNKController) updateEnv(ctx context.Context, target *flowslatest.FlowCollector, desiredEnv map[string]string) error {
	rlog := log.FromContext(ctx, "component", "FlowsConfigOVNKController")

	ds := &appsv1.DaemonSet{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      c.config.DaemonSetName,
		Namespace: c.config.Namespace,
	}, ds); err != nil {
		if kerr.IsNotFound(err) && !helper.UseIPFIX(&target.Spec) {
			// If we don't want IPFIX and ovn-k daemonset is not found, assume there no ovn-k, just succeed
			rlog.Info("Skip reconciling OVN: OVN DaemonSet not found")
			return nil
		}
		return fmt.Errorf("retrieving %s/%s daemonset: %w", c.config.Namespace, c.config.DaemonSetName, err)
	}

	ovnkubeNode := helper.FindContainer(&ds.Spec.Template.Spec, target.Spec.Agent.IPFIX.OVNKubernetes.ContainerName)
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
		return c.Update(ctx, ds)
	}

	rlog.Info("No changes needed")
	return nil
}

func (c *FlowsConfigOVNKController) desiredEnv(ctx context.Context, coll *flowslatest.FlowCollector) (map[string]string, error) {
	cacheTimeout, err := time.ParseDuration(coll.Spec.Agent.IPFIX.CacheActiveTimeout)
	if err != nil {
		return nil, err
	}
	sampling := getSampling(ctx, &coll.Spec.Agent.IPFIX)

	envs := map[string]string{
		"OVN_IPFIX_TARGETS":              "",
		"OVN_IPFIX_CACHE_ACTIVE_TIMEOUT": strconv.Itoa(int(cacheTimeout.Seconds())),
		"OVN_IPFIX_CACHE_MAX_FLOWS":      strconv.Itoa(int(coll.Spec.Agent.IPFIX.CacheMaxFlows)),
		"OVN_IPFIX_SAMPLING":             strconv.Itoa(int(sampling)),
	}

	if !helper.UseIPFIX(&coll.Spec) {
		// No IPFIX => leave target empty and return
		return envs, nil
	}

	envs["OVN_IPFIX_TARGETS"] = fmt.Sprintf(":%d", coll.Spec.Processor.Port)
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
func (c *FlowsConfigOVNKController) Finalize(ctx context.Context, target *flowslatest.FlowCollector) error {
	// Remove all env
	desiredEnv := map[string]string{
		"OVN_IPFIX_TARGETS":              "",
		"OVN_IPFIX_CACHE_ACTIVE_TIMEOUT": "",
		"OVN_IPFIX_CACHE_MAX_FLOWS":      "",
		"OVN_IPFIX_SAMPLING":             "",
	}
	return c.updateEnv(ctx, target, desiredEnv)
}
