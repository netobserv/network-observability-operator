package helper

import (
	"crypto/tls"
	"fmt"

	"github.com/netobserv/netobserv-operator/internal/pkg/tlsconfig"
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

// AppendTLSEnvVars appends TLS configuration environment variables from the cluster's
// composed TLS config. This allows components (FLP, eBPF agent, console plugin) to
// inherit TLS settings from the cluster. Returns the input envs unchanged if tlsCfg is nil.
func AppendTLSEnvVars(envs []corev1.EnvVar, tlsCfg *tls.Config) []corev1.EnvVar {
	if tlsCfg == nil {
		return envs
	}

	return append(envs, tlsconfig.ConfigToEnvVars(tlsCfg)...)
}
