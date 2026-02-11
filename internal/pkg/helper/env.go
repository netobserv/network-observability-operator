package helper

import (
	"fmt"

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

func EnvFromReqsLimits(envs []corev1.EnvVar, reqs *corev1.ResourceRequirements) []corev1.EnvVar {
	// set GOMEMLIMIT which allows specifying a soft memory cap to force GC when resource limit is reached to prevent OOM
	if reqs.Limits.Memory() != nil {
		if memLimit, ok := reqs.Limits.Memory().AsInt64(); ok && memLimit > 0 {
			// we will set the GOMEMLIMIT to current memlimit - 10% as a headroom to account for
			// memory sources the Go runtime is unaware of
			memLimit -= int64(float64(memLimit) * 0.1)
			envs = append(envs, corev1.EnvVar{Name: "GOMEMLIMIT", Value: fmt.Sprint(memLimit)})
		}
	}
	return envs
}
