package helper

import (
	corev1 "k8s.io/api/core/v1"
)

func BuildEnvFromDefaults(config, defaults map[string]string) []corev1.EnvVar {
	var result []corev1.EnvVar
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range KeySorted(defaults) {
		k, def := pair[0], pair[1]
		if override, ok := config[k]; ok {
			result = append(result, corev1.EnvVar{Name: k, Value: override})
		} else {
			result = append(result, corev1.EnvVar{Name: k, Value: def})
		}
	}
	for _, pair := range KeySorted(config) {
		k, cfg := pair[0], pair[1]
		if _, ok := defaults[k]; !ok {
			result = append(result, corev1.EnvVar{Name: k, Value: cfg})
		}
	}
	return result
}
