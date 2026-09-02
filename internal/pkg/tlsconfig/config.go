package tlsconfig

import (
	"crypto/tls"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/crypto"
)

// ComposeTLSConfig creates a tls.Config from an OpenShift TLS security profile.
// It always returns a valid config using secure defaults as a base, so callers can
// safely use it even when an error is returned. The error should be surfaced (e.g.
// logged or reported in status conditions) but must not prevent TLS from being applied.
func ComposeTLSConfig(profile *configv1.TLSSecurityProfile) (*tls.Config, error) {
	base := crypto.SecureTLSConfig(&tls.Config{})
	if profile == nil {
		return base, nil
	}

	minVersion, cipherSuites, curves, err := profileToTLSConfig(profile)
	if minVersion != 0 {
		base.MinVersion = minVersion
	}
	if len(cipherSuites) > 0 {
		base.CipherSuites = cipherSuites
	}
	if len(curves) > 0 {
		base.CurvePreferences = curves
	}
	return base, err
}

// profileToTLSConfig converts an OpenShift TLSSecurityProfile to TLS version, cipher suites,
// and curve preferences. Returns partial results alongside any error: callers should apply
// non-zero/non-empty values even when an error is returned.
func profileToTLSConfig(profile *configv1.TLSSecurityProfile) (uint16, []uint16, []tls.CurveID, error) {
	if profile == nil {
		return 0, nil, nil, nil
	}

	var minVersionStr configv1.TLSProtocolVersion
	var cipherNames []string
	var groups []configv1.TLSGroup

	switch profile.Type {
	case configv1.TLSProfileOldType, configv1.TLSProfileIntermediateType, configv1.TLSProfileModernType:
		spec := configv1.TLSProfiles[profile.Type]
		minVersionStr = spec.MinTLSVersion
		cipherNames = spec.Ciphers
		groups = spec.Groups
	case configv1.TLSProfileCustomType:
		if profile.Custom == nil {
			return 0, nil, nil, fmt.Errorf("custom TLS profile specified but Custom field is nil")
		}
		minVersionStr = profile.Custom.MinTLSVersion
		cipherNames = profile.Custom.Ciphers
		groups = profile.Custom.Groups
	default:
		return 0, nil, nil, fmt.Errorf("unknown TLS profile type %q", profile.Type)
	}

	minVersion, versionErr := tlsVersionFromString(minVersionStr)

	ianaCipherNames := crypto.OpenSSLToIANACipherSuites(cipherNames)
	cipherSuites := make([]uint16, 0, len(ianaCipherNames))
	for _, name := range ianaCipherNames {
		if id := cipherSuiteByName(name); id != 0 {
			cipherSuites = append(cipherSuites, id)
		}
	}

	curves := make([]tls.CurveID, 0, len(groups))
	for _, group := range groups {
		if id, ok := curveIDByGroup(group); ok {
			curves = append(curves, id)
		}
	}

	return minVersion, cipherSuites, curves, versionErr
}

// curveIDByGroup maps an OpenShift TLSGroup to its Go crypto/tls.CurveID equivalent.
// There is a one-to-one mapping between these names and Go's group IDs (per the TLSGroup
// API doc), so unlike ciphers this doesn't need OpenSSL-to-IANA name translation.
// Returns false for groups Go's crypto/tls doesn't support (e.g. built with an older Go
// toolchain), so callers can silently skip them rather than fail the whole config.
func curveIDByGroup(group configv1.TLSGroup) (tls.CurveID, bool) {
	switch group {
	case configv1.TLSGroupX25519:
		return tls.X25519, true
	case configv1.TLSGroupSecP256r1:
		return tls.CurveP256, true
	case configv1.TLSGroupSecP384r1:
		return tls.CurveP384, true
	case configv1.TLSGroupSecP521r1:
		return tls.CurveP521, true
	case configv1.TLSGroupX25519MLKEM768:
		return tls.X25519MLKEM768, true
	case configv1.TLSGroupSecP256r1MLKEM768:
		return tls.SecP256r1MLKEM768, true
	case configv1.TLSGroupSecP384r1MLKEM1024:
		return tls.SecP384r1MLKEM1024, true
	default:
		return 0, false
	}
}

// tlsVersionFromString converts a TLS version string to its uint16 constant.
// Returns an error for unknown versions instead of silently defaulting, so that
// future TLS versions are not silently ignored.
func tlsVersionFromString(version configv1.TLSProtocolVersion) (uint16, error) {
	switch version {
	case configv1.VersionTLS10:
		return tls.VersionTLS10, nil
	case configv1.VersionTLS11:
		return tls.VersionTLS11, nil
	case configv1.VersionTLS12:
		return tls.VersionTLS12, nil
	case configv1.VersionTLS13:
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unknown TLS version %q", version)
	}
}

// cipherSuiteByName returns the cipher suite ID for a given IANA name.
// Returns 0 if not found.
func cipherSuiteByName(name string) uint16 {
	for _, suite := range tls.CipherSuites() {
		if suite.Name == name {
			return suite.ID
		}
	}
	// Insecure cipher suites (for compatibility with Old profile)
	for _, suite := range tls.InsecureCipherSuites() {
		if suite.Name == name {
			return suite.ID
		}
	}
	return 0
}

// ExtractTLSProfileSpec extracts the TLSProfileSpec from a TLSSecurityProfile.
// This is needed for SecurityProfileWatcher from controller-runtime-common.
func ExtractTLSProfileSpec(profile *configv1.TLSSecurityProfile) *configv1.TLSProfileSpec {
	if profile == nil {
		return nil
	}

	switch profile.Type {
	case configv1.TLSProfileOldType:
		return configv1.TLSProfiles[configv1.TLSProfileOldType]
	case configv1.TLSProfileIntermediateType:
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	case configv1.TLSProfileModernType:
		return configv1.TLSProfiles[configv1.TLSProfileModernType]
	case configv1.TLSProfileCustomType:
		if profile.Custom != nil {
			return &profile.Custom.TLSProfileSpec
		}
	}

	return nil
}
