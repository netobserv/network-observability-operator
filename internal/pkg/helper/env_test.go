package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestBuildEnvFromDefaults(t *testing.T) {
	envs := BuildEnvFromDefaults(
		// config:
		map[string]string{
			"a": "1",
			"z": "99",
			"b": "11",
		},
		// default:
		map[string]string{
			"a": "2",
			"y": "88",
			"c": "18",
		},
	)
	assert.Equal(t, []corev1.EnvVar{
		// configs present in defaults come first, ordered
		{Name: "a", Value: "1"},
		{Name: "c", Value: "18"},
		{Name: "y", Value: "88"},
		// then configs only in the override, ordered as well
		{Name: "b", Value: "11"},
		{Name: "z", Value: "99"},
	}, envs)
}
