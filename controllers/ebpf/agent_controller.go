package ebpf

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const (
	flowsTargetHostEnvVar = "FLOWS_TARGET_HOST"
	flowsTargetPortEnvVar = "FLOWS_TARGET_PORT"
)

type reconcileAction int

const (
	actionNone = iota
	actionCreate
	actionUpdate
	actionDelete
)

type AgentController struct {
	client    reconcilers.ClientHelper
	namespace string
}

func NewAgentController(client reconcilers.ClientHelper, namespace string) *AgentController {
	return &AgentController{
		client:    client,
		namespace: namespace,
	}
}

// Reconcile reconciles the status of the ovs-flows-config configmap with
// the target FlowCollector ipfix section map
func (c *AgentController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	rlog := log.FromContext(ctx).WithName("AgentController")
	ctx = log.IntoContext(ctx, rlog)

	current, err := c.current(ctx)
	if err != nil {
		rlog.Info("can't fetch current Agent. Assuming as non-existing", "error", err)
	}
	desired := c.desired(target)
	switch c.requiredAction(current, desired) {
	case actionNone:
		rlog.Info("action: none")
		return nil
	case actionCreate:
		rlog.Info("action: create agent")
		if err := permissions.Apply(ctx, c.client, c.namespace); err != nil {
			return err
		}
		return c.client.Create(ctx, desired)
	case actionDelete:
		rlog.Info("action: delete agent")
		return c.client.Delete(ctx, current)
	case actionUpdate:
		if err := permissions.Apply(ctx, c.client, c.namespace); err != nil {
			return err
		}
		rlog.Info("action: update agent")
		return c.client.Update(ctx, current)
	}
	rlog.Info("unexpected action. Doing nothing")
	return nil
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.namespace,
	}, &agentDS); err != nil {
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w",
			c.namespace, constants.EBPFAgentName, err)
	}
	return &agentDS, nil
}

func (c *AgentController) desired(coll *flowsv1alpha1.FlowCollector) *v1.DaemonSet {
	if coll == nil || coll.Spec.EBPF == nil {
		return nil
	}
	trueVal := true
	version := helper.ExtractVersion(coll.Spec.EBPF.Image)
	return &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentName,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":     constants.EBPFAgentName,
				"version": version,
			},
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": constants.EBPFAgentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": constants.EBPFAgentName},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: constants.EBPFServiceAccount,
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					Containers: []corev1.Container{{
						Name:            constants.EBPFAgentName,
						Image:           coll.Spec.EBPF.Image,
						ImagePullPolicy: corev1.PullPolicy(coll.Spec.EBPF.ImagePullPolicy),
						Resources:       coll.Spec.EBPF.Resources,
						// TODO: other parameters when NETOBSERV-201 is implemented
						SecurityContext: &corev1.SecurityContext{
							Privileged: &trueVal,
						},
						Env: c.flpEndpoint(coll),
					}},
				},
			},
		},
	}
}

func (c *AgentController) flpEndpoint(coll *flowsv1alpha1.FlowCollector) []corev1.EnvVar {
	switch coll.Spec.FlowlogsPipeline.Kind {
	case constants.DaemonSetKind:
		// When flowlogs-pipeline is deployed as a daemonset, each agent must send
		// data to the pod that is deployed in the same host
		return []corev1.EnvVar{{
			Name: flowsTargetHostEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		}, {
			Name:  flowsTargetPortEnvVar,
			Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
		}}
	case constants.DeploymentKind:
		return []corev1.EnvVar{{
			Name:  flowsTargetHostEnvVar,
			Value: constants.FLPName + "." + c.namespace,
		}, {
			Name:  flowsTargetPortEnvVar,
			Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
		}}
	}
	return nil
}

func (c *AgentController) requiredAction(current, desired *v1.DaemonSet) reconcileAction {
	if current == nil && desired == nil {
		return actionNone
	}
	if current != nil && desired == nil {
		return actionDelete
	}
	if current == nil && desired != nil {
		return actionCreate
	}
	same := reflect.DeepEqual(current.Spec.Template.Spec, desired.Spec.Template.Spec) &&
		reflect.DeepEqual(current.Spec.Template.ObjectMeta, desired.Spec.Template.ObjectMeta) &&
		reflect.DeepEqual(current.Spec.Selector, desired.Spec.Selector) &&
		reflect.DeepEqual(current.ObjectMeta.Labels, desired.ObjectMeta.Labels)
	if same {
		return actionNone
	}
	return actionUpdate
}
