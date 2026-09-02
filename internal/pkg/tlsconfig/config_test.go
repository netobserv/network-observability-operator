package tlsconfig

import (
	"crypto/tls"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
)

func TestComposeTLSConfig_NilProfile(t *testing.T) {
	cfg, err := ComposeTLSConfig(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MinVersion == 0 {
		t.Error("expected MinVersion to be set from secure defaults")
	}
}

func TestComposeTLSConfig_OldProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileOldType,
	}

	cfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MinVersion != tls.VersionTLS10 {
		t.Errorf("expected MinVersion TLS 1.0 (0x%x), got 0x%x", tls.VersionTLS10, cfg.MinVersion)
	}
	if len(cfg.CipherSuites) == 0 {
		t.Error("expected cipher suites to be set")
	}
}

func TestComposeTLSConfig_IntermediateProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileIntermediateType,
	}

	cfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2 (0x%x), got 0x%x", tls.VersionTLS12, cfg.MinVersion)
	}
	if len(cfg.CipherSuites) == 0 {
		t.Error("expected cipher suites to be set")
	}
	// TLSProfiles[Intermediate].Groups is X25519MLKEM768, X25519, secp256r1, secp384r1 (see
	// vendor/github.com/openshift/api/config/v1/types_tlssecurityprofile.go).
	expectedCurves := []tls.CurveID{tls.X25519MLKEM768, tls.X25519, tls.CurveP256, tls.CurveP384}
	if len(cfg.CurvePreferences) != len(expectedCurves) {
		t.Fatalf("expected %d curve preferences, got %d: %v", len(expectedCurves), len(cfg.CurvePreferences), cfg.CurvePreferences)
	}
	for i, want := range expectedCurves {
		if cfg.CurvePreferences[i] != want {
			t.Errorf("CurvePreferences[%d] = %v, expected %v", i, cfg.CurvePreferences[i], want)
		}
	}
}

func TestComposeTLSConfig_ModernProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileModernType,
	}

	cfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinVersion TLS 1.3 (0x%x), got 0x%x", tls.VersionTLS13, cfg.MinVersion)
	}
}

func TestComposeTLSConfig_CustomProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers: []string{
					"ECDHE-ECDSA-AES128-GCM-SHA256",
					"ECDHE-RSA-AES128-GCM-SHA256",
				},
				Groups: []configv1.TLSGroup{configv1.TLSGroupSecP384r1, configv1.TLSGroupX25519},
			},
		},
	}

	cfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2, got 0x%x", cfg.MinVersion)
	}
	if len(cfg.CipherSuites) == 0 {
		t.Error("expected cipher suites to be set")
	}
	expectedCurves := []tls.CurveID{tls.CurveP384, tls.X25519}
	if len(cfg.CurvePreferences) != len(expectedCurves) {
		t.Fatalf("expected %d curve preferences, got %d: %v", len(expectedCurves), len(cfg.CurvePreferences), cfg.CurvePreferences)
	}
	for i, want := range expectedCurves {
		if cfg.CurvePreferences[i] != want {
			t.Errorf("CurvePreferences[%d] = %v, expected %v", i, cfg.CurvePreferences[i], want)
		}
	}
}

// TestComposeTLSConfig_CustomProfileNoGroups verifies that omitting Groups leaves
// CurvePreferences unset (nil), matching the documented "no opinion" semantics.
func TestComposeTLSConfig_CustomProfileNoGroups(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers:       []string{"ECDHE-ECDSA-AES128-GCM-SHA256"},
			},
		},
	}

	cfg, err := ComposeTLSConfig(profile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.CurvePreferences) != 0 {
		t.Errorf("expected no curve preferences, got %v", cfg.CurvePreferences)
	}
}

// TestComposeTLSConfig_CustomProfileNilCustom verifies that a nil Custom field returns
// an error but still yields a valid config with secure defaults.
func TestComposeTLSConfig_CustomProfileNilCustom(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileCustomType,
		Custom: nil,
	}

	cfg, err := ComposeTLSConfig(profile)
	if err == nil {
		t.Error("expected error for custom profile with nil Custom field")
	}
	if cfg == nil {
		t.Error("expected non-nil config (secure defaults) even on error")
		return
	}
	if cfg.MinVersion == 0 {
		t.Error("expected MinVersion to be set from secure defaults on error")
	}
}

// TestComposeTLSConfig_UnknownProfileType verifies that an unknown profile type returns
// an error but still yields a valid config with secure defaults.
func TestComposeTLSConfig_UnknownProfileType(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: "UnknownType",
	}

	cfg, err := ComposeTLSConfig(profile)
	if err == nil {
		t.Error("expected error for unknown profile type")
	}
	if cfg == nil {
		t.Error("expected non-nil config (secure defaults) even on error")
		return
	}
	if cfg.MinVersion == 0 {
		t.Error("expected MinVersion to be set from secure defaults on error")
	}
}

func TestExtractTLSProfileSpec_NilProfile(t *testing.T) {
	spec := ExtractTLSProfileSpec(nil)
	if spec != nil {
		t.Error("expected nil spec for nil profile")
	}
}

func TestExtractTLSProfileSpec_OldProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileOldType,
	}

	spec := ExtractTLSProfileSpec(profile)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.MinTLSVersion != configv1.VersionTLS10 {
		t.Errorf("expected MinTLSVersion TLS 1.0, got %v", spec.MinTLSVersion)
	}
	if len(spec.Ciphers) == 0 {
		t.Error("expected ciphers to be set")
	}
}

func TestExtractTLSProfileSpec_IntermediateProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileIntermediateType,
	}

	spec := ExtractTLSProfileSpec(profile)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.MinTLSVersion != configv1.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", spec.MinTLSVersion)
	}
}

func TestExtractTLSProfileSpec_ModernProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileModernType,
	}

	spec := ExtractTLSProfileSpec(profile)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.MinTLSVersion != configv1.VersionTLS13 {
		t.Errorf("expected MinTLSVersion TLS 1.3, got %v", spec.MinTLSVersion)
	}
}

func TestExtractTLSProfileSpec_CustomProfile(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers:       []string{"ECDHE-RSA-AES256-GCM-SHA384"},
				Groups:        []configv1.TLSGroup{configv1.TLSGroupX25519},
			},
		},
	}

	spec := ExtractTLSProfileSpec(profile)
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.MinTLSVersion != configv1.VersionTLS12 {
		t.Errorf("expected MinTLSVersion TLS 1.2, got %v", spec.MinTLSVersion)
	}
	if len(spec.Ciphers) != 1 {
		t.Errorf("expected 1 cipher, got %d", len(spec.Ciphers))
	}
	// Regression test: Groups used to be silently dropped here, which meant the
	// SecurityProfileWatcher would never detect a Custom profile's Groups changing.
	if len(spec.Groups) != 1 || spec.Groups[0] != configv1.TLSGroupX25519 {
		t.Errorf("expected Groups to be preserved, got %v", spec.Groups)
	}
}

func TestExtractTLSProfileSpec_CustomProfileNilCustom(t *testing.T) {
	profile := &configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileCustomType,
		Custom: nil,
	}

	spec := ExtractTLSProfileSpec(profile)
	if spec != nil {
		t.Error("expected nil spec for custom profile with nil Custom field")
	}
}

func TestTLSVersionFromString(t *testing.T) {
	tests := []struct {
		input       configv1.TLSProtocolVersion
		expected    uint16
		expectError bool
	}{
		{configv1.VersionTLS10, tls.VersionTLS10, false},
		{configv1.VersionTLS11, tls.VersionTLS11, false},
		{configv1.VersionTLS12, tls.VersionTLS12, false},
		{configv1.VersionTLS13, tls.VersionTLS13, false},
		{"", 0, true},
		{"TLS1.4", 0, true},
	}

	for _, tt := range tests {
		result, err := tlsVersionFromString(tt.input)
		if (err != nil) != tt.expectError {
			t.Errorf("tlsVersionFromString(%q) error=%v, expectError=%v", tt.input, err, tt.expectError)
		}
		if result != tt.expected {
			t.Errorf("tlsVersionFromString(%q) = 0x%x, expected 0x%x", tt.input, result, tt.expected)
		}
	}
}

func TestCurveIDByGroup(t *testing.T) {
	tests := []struct {
		group    configv1.TLSGroup
		expected tls.CurveID
	}{
		{configv1.TLSGroupX25519, tls.X25519},
		{configv1.TLSGroupSecP256r1, tls.CurveP256},
		{configv1.TLSGroupSecP384r1, tls.CurveP384},
		{configv1.TLSGroupSecP521r1, tls.CurveP521},
		{configv1.TLSGroupX25519MLKEM768, tls.X25519MLKEM768},
		{configv1.TLSGroupSecP256r1MLKEM768, tls.SecP256r1MLKEM768},
		{configv1.TLSGroupSecP384r1MLKEM1024, tls.SecP384r1MLKEM1024},
	}

	for _, tt := range tests {
		id, ok := curveIDByGroup(tt.group)
		if !ok {
			t.Errorf("curveIDByGroup(%q): expected ok=true", tt.group)
		}
		if id != tt.expected {
			t.Errorf("curveIDByGroup(%q) = %v, expected %v", tt.group, id, tt.expected)
		}
	}

	if _, ok := curveIDByGroup("UnknownGroup"); ok {
		t.Error("expected ok=false for unknown group")
	}
}

func TestCipherSuiteByName(t *testing.T) {
	id := cipherSuiteByName("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	if id == 0 {
		t.Error("expected non-zero cipher suite ID for known cipher")
	}

	id = cipherSuiteByName("UNKNOWN_CIPHER_SUITE")
	if id != 0 {
		t.Error("expected zero for unknown cipher suite")
	}
}
