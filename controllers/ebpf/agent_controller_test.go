package ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestRequiredAction(t *testing.T) {
	assert := assert.New(t)

	action := requiredAction(nil, nil)
	assert.Equal(actionNone, int(action))

	current := sampleDS()
	current.Labels["injected"] = "injected"
	current.Annotations["injected"] = "injected"
	current.Spec.Template.Labels["injected"] = "injected"
	current.Spec.Template.Annotations["injected"] = "injected"

	action = requiredAction(&current, nil)
	assert.Equal(actionNone, int(action))

	action = requiredAction(nil, &current)
	assert.Equal(actionCreate, int(action))

	desired := sampleDS()

	// Check derivatives
	action = requiredAction(&current, &desired)
	assert.Equal(actionNone, int(action))

	desired.Labels = map[string]string{
		"app": "bar",
	}
	action = requiredAction(&current, &desired)
	assert.Equal(actionUpdate, int(action))

	desired = sampleDS()
	desired.Spec.Template.Spec.Containers[0].Env[0].Value = "B"
	action = requiredAction(&current, &desired)
	assert.Equal(actionUpdate, int(action))

	// Make sure we don't use derivative for Env, which would ignore empty fields in "desired"
	desired = sampleDS()
	desired.Spec.Template.Spec.Containers[0].Env[0] = corev1.EnvVar{}
	action = requiredAction(&current, &desired)
	assert.Equal(actionUpdate, int(action))
}
