package tlsconfig

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
)

func TestConfigToEnvVars_NilConfig(t *testing.T) {
	envVars := ConfigToEnvVars(nil)
	if envVars != nil {
		t.Error("expected nil env vars for nil config")
	}
}

func TestConfigToEnvVars_IntermediateProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileIntermediateType,
	}
	tlsCfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envVars := ConfigToEnvVars(tlsCfg)
	if len(envVars) == 0 {
		t.Fatal("expected env vars to be set")
	}

	var hasMinVersion bool
	var minVersionValue string
	for _, env := range envVars {
		if env.Name == EnvTLSMinVersion {
			hasMinVersion = true
			minVersionValue = env.Value
		}
	}

	if !hasMinVersion {
		t.Error("expected TLS_MIN_VERSION env var to be set")
	}

	// Intermediate profile uses TLS 1.2 (0x0303 = 771)
	expected := fmt.Sprintf("%d", tls.VersionTLS12)
	if minVersionValue != expected {
		t.Errorf("expected TLS_MIN_VERSION=%s, got %s", expected, minVersionValue)
	}
}

func TestConfigToEnvVars_ModernProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileModernType,
	}
	tlsCfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envVars := ConfigToEnvVars(tlsCfg)

	var minVersionValue string
	for _, env := range envVars {
		if env.Name == EnvTLSMinVersion {
			minVersionValue = env.Value
		}
	}

	// Modern profile uses TLS 1.3 (0x0304 = 772)
	expected := fmt.Sprintf("%d", tls.VersionTLS13)
	if minVersionValue != expected {
		t.Errorf("expected TLS_MIN_VERSION=%s, got %s", expected, minVersionValue)
	}
}

func TestConfigToEnvVars_OldProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileOldType,
	}
	tlsCfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envVars := ConfigToEnvVars(tlsCfg)

	var minVersionValue string
	var cipherSuitesValue string
	for _, env := range envVars {
		if env.Name == EnvTLSMinVersion {
			minVersionValue = env.Value
		}
		if env.Name == EnvTLSCipherSuites {
			cipherSuitesValue = env.Value
		}
	}

	// Old profile uses TLS 1.0 (0x0301 = 769)
	expected := fmt.Sprintf("%d", tls.VersionTLS10)
	if minVersionValue != expected {
		t.Errorf("expected TLS_MIN_VERSION=%s, got %s", expected, minVersionValue)
	}

	if cipherSuitesValue == "" {
		t.Error("expected TLS_CIPHER_SUITES to be set for Old profile")
	}

	// Values must be comma-separated decimal uint16s
	assertUint16List(t, "TLS_CIPHER_SUITES", cipherSuitesValue)
}

func TestConfigToEnvVars_CustomProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers: []string{
					"ECDHE-ECDSA-AES128-GCM-SHA256",
				},
			},
		},
	}
	tlsCfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envVars := ConfigToEnvVars(tlsCfg)
	if len(envVars) == 0 {
		t.Fatal("expected env vars to be set")
	}

	expected := fmt.Sprintf("%d", tls.VersionTLS12)
	var hasMinVersion, hasCipherSuites bool
	for _, env := range envVars {
		if env.Name == EnvTLSMinVersion {
			hasMinVersion = true
			if env.Value != expected {
				t.Errorf("expected TLS_MIN_VERSION=%s, got %s", expected, env.Value)
			}
		}
		if env.Name == EnvTLSCipherSuites {
			hasCipherSuites = true
			assertUint16List(t, "TLS_CIPHER_SUITES", env.Value)
		}
	}

	if !hasMinVersion {
		t.Error("expected TLS_MIN_VERSION to be set")
	}
	if !hasCipherSuites {
		t.Error("expected TLS_CIPHER_SUITES to be set")
	}
}

func TestConfigToEnvVars_CurvePreferences(t *testing.T) {
	// Intermediate profile includes curve preferences
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileIntermediateType,
	}
	tlsCfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	envVars := ConfigToEnvVars(tlsCfg)

	for _, env := range envVars {
		if env.Name == EnvTLSCurvePreferences {
			assertUint16List(t, "TLS_CURVE_PREFERENCES", env.Value)
			return
		}
	}
	// Curve preferences may not be set if the profile doesn't include them — that's OK
}

func TestUint16SliceToString(t *testing.T) {
	result := uint16SliceToString([]uint16{769, 771, 772})
	if result != "769,771,772" {
		t.Errorf("expected '769,771,772', got %q", result)
	}

	if uint16SliceToString(nil) != "" {
		t.Error("expected empty string for nil slice")
	}

	if uint16SliceToString([]uint16{}) != "" {
		t.Error("expected empty string for empty slice")
	}
}

// assertUint16List verifies that s is a non-empty comma-separated list of decimal uint16 values.
func assertUint16List(t *testing.T, name, s string) {
	t.Helper()
	if s == "" {
		t.Errorf("%s: expected non-empty value", name)
		return
	}
	for _, part := range strings.Split(s, ",") {
		if _, err := strconv.ParseUint(strings.TrimSpace(part), 10, 16); err != nil {
			t.Errorf("%s: %q is not a valid uint16 decimal: %v", name, part, err)
		}
	}
}
