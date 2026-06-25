package tlsconfig

import (
	"crypto/tls"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Environment variable names for TLS configuration
const (
	EnvTLSMinVersion       = "TLS_MIN_VERSION"
	EnvTLSCipherSuites     = "TLS_CIPHER_SUITES"
	EnvTLSCurvePreferences = "TLS_CURVE_PREFERENCES"
)

// ConfigToEnvVars converts a tls.Config to environment variables suitable for
// injection into component containers.
// Values are encoded as decimal uint16 strings so that new TLS versions, cipher
// suites, and curves are automatically supported without code changes on either side.
// Returns nil if tlsCfg is nil (components will use Go defaults).
func ConfigToEnvVars(tlsCfg *tls.Config) []corev1.EnvVar {
	if tlsCfg == nil {
		return nil
	}

	var envVars []corev1.EnvVar

	if tlsCfg.MinVersion != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvTLSMinVersion,
			Value: fmt.Sprintf("%d", tlsCfg.MinVersion),
		})
	}

	if len(tlsCfg.CipherSuites) > 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvTLSCipherSuites,
			Value: uint16SliceToString(tlsCfg.CipherSuites),
		})
	}

	if len(tlsCfg.CurvePreferences) > 0 {
		ids := make([]uint16, len(tlsCfg.CurvePreferences))
		for i, c := range tlsCfg.CurvePreferences {
			ids[i] = uint16(c)
		}
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvTLSCurvePreferences,
			Value: uint16SliceToString(ids),
		})
	}

	return envVars
}

// uint16SliceToString encodes a slice of uint16 values as a comma-separated decimal string.
func uint16SliceToString(ids []uint16) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}
