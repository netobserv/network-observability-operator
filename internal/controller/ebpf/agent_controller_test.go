package ebpf

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/pkg/cluster"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func sampleDS() appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"app": "foo",
			},
			Annotations: map[string]string{},
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": "foo",
					},
					Annotations: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Env: []corev1.EnvVar{{
							Name:  "TEST",
							Value: "A",
						}},
					}},
				},
			},
		},
	}
}

func TestDaemonSetChanged(t *testing.T) {
	assert := assert.New(t)

	action := helper.DaemonSetChanged(nil, nil)
	assert.Equal(helper.ActionNone, int(action))

	current := sampleDS()
	current.Labels["injected"] = "injected"
	current.Annotations["injected"] = "injected"
	current.Spec.Template.Labels["injected"] = "injected"
	current.Spec.Template.Annotations["injected"] = "injected"

	action = helper.DaemonSetChanged(&current, nil)
	assert.Equal(helper.ActionNone, int(action))

	action = helper.DaemonSetChanged(nil, &current)
	assert.Equal(helper.ActionCreate, int(action))

	desired := sampleDS()

	// Check derivatives
	action = helper.DaemonSetChanged(&current, &desired)
	assert.Equal(helper.ActionNone, int(action))

	desired.Labels = map[string]string{
		"app": "bar",
	}
	action = helper.DaemonSetChanged(&current, &desired)
	assert.Equal(helper.ActionUpdate, int(action))

	desired = sampleDS()
	desired.Spec.Template.Spec.Containers[0].Env[0].Value = "B"
	action = helper.DaemonSetChanged(&current, &desired)
	assert.Equal(helper.ActionUpdate, int(action))

	// Make sure we don't use derivative for Env, which would ignore empty fields in "desired"
	desired = sampleDS()
	desired.Spec.Template.Spec.Containers[0].Env[0] = corev1.EnvVar{}
	action = helper.DaemonSetChanged(&current, &desired)
	assert.Equal(helper.ActionUpdate, int(action))
}

func TestGetEnvConfig_Default(t *testing.T) {
	fc := flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			Agent: flowslatest.FlowCollectorAgent{
				EBPF: flowslatest.FlowCollectorEBPF{},
			},
		},
	}

	env := getEnvConfig(&fc, &cluster.Info{})
	assert.Equal(t, []corev1.EnvVar{
		{Name: "GOMEMLIMIT", Value: "0"},
		{Name: "METRICS_ENABLE", Value: "true"},
		{Name: "METRICS_SERVER_PORT", Value: "9400"},
		{Name: "METRICS_PREFIX", Value: "netobserv_agent_"},
		{Name: "AGENT_IP", Value: "",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			}},
		{Name: "DNS_TRACKING_PORT", Value: "53"},
		{Name: "NETWORK_EVENTS_MONITORING_GROUP_ID", Value: "10"},
		{Name: "PREFERRED_INTERFACE_FOR_MAC_PREFIX", Value: "0a:58=eth0"},
		{Name: "TC_ATTACH_MODE", Value: "tcx"},
	}, env)
}

func TestGetEnvConfig_WithOverrides(t *testing.T) {
	fc := flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			Agent: flowslatest.FlowCollectorAgent{
				EBPF: flowslatest.FlowCollectorEBPF{
					Advanced: &flowslatest.AdvancedAgentConfig{
						Env: map[string]string{
							"PREFERRED_INTERFACE_FOR_MAC_PREFIX": "0a:58=ens5",
							"DNS_TRACKING_PORT":                  "5353",
							"NETWORK_EVENTS_MONITORING_GROUP_ID": "any",
							"TC_ATTACH_MODE":                     "any",
						},
					},
					Metrics: flowslatest.EBPFMetrics{
						Enable: ptr.To(false),
					},
					FlowFilter: &flowslatest.EBPFFlowFilter{
						Enable: ptr.To(true),
						Rules: []flowslatest.EBPFFlowFilterRule{
							{
								CIDR:   "0.0.0.0/0",
								Action: "Accept",
							},
						},
					},
				},
			},
		},
	}

	env := getEnvConfig(&fc, &cluster.Info{})
	assert.Equal(t, []corev1.EnvVar{
		{Name: "GOMEMLIMIT", Value: "0"},
		{Name: "ENABLE_FLOW_FILTER", Value: "true"},
		{Name: "FLOW_FILTER_RULES", Value: `[{"ip_cidr":"0.0.0.0/0","action":"Accept"}]`},
		{Name: "AGENT_IP", Value: "",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			}},
		{Name: "DNS_TRACKING_PORT", Value: "5353"},
		{Name: "NETWORK_EVENTS_MONITORING_GROUP_ID", Value: "any"},
		{Name: "PREFERRED_INTERFACE_FOR_MAC_PREFIX", Value: "0a:58=ens5"},
		{Name: "TC_ATTACH_MODE", Value: "any"},
	}, env)
}

func TestGetEnvConfig_OCP4_14(t *testing.T) {
	fc := flowslatest.FlowCollector{
		Spec: flowslatest.FlowCollectorSpec{
			Agent: flowslatest.FlowCollectorAgent{
				EBPF: flowslatest.FlowCollectorEBPF{},
			},
		},
	}

	info := cluster.Info{}
	info.MockOpenShiftVersion("4.14.5")
	env := getEnvConfig(&fc, &info)
	assert.Equal(t, []corev1.EnvVar{
		{Name: "GOMEMLIMIT", Value: "0"},
		{Name: "METRICS_ENABLE", Value: "true"},
		{Name: "METRICS_SERVER_PORT", Value: "9400"},
		{Name: "METRICS_PREFIX", Value: "netobserv_agent_"},
		{Name: "AGENT_IP", Value: "",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			}},
		{Name: "DNS_TRACKING_PORT", Value: "53"},
		{Name: "NETWORK_EVENTS_MONITORING_GROUP_ID", Value: "10"},
		{Name: "PREFERRED_INTERFACE_FOR_MAC_PREFIX", Value: "0a:58=eth0"},
		{Name: "TC_ATTACH_MODE", Value: "tc"},
	}, env)
}
