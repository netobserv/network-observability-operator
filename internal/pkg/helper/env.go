package helper

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

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
