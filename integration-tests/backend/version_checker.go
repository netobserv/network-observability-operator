package e2etests

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"golang.org/x/mod/semver"
)

var clusterVersion string

func GetOCPVersion() (string, error) {
	if clusterVersion != "" {
		return clusterVersion, nil
	}

	obj, err := getDynamicResource("clusterversion", "version", "")
	if err != nil {
		return "", err
	}
	version, found := getNestedField(obj.Object, ".status.desired.version")
	if !found {
		return "", fmt.Errorf("desired version not found in clusterversion")
	}
	clusterVersion = semver.Canonical("v" + version)
	clusterVersion = semver.MajorMinor(clusterVersion)
	fmt.Printf("Detected OCP version: %s\n", clusterVersion)
	return clusterVersion, nil
}

// validateRequiredVersion validates and canonicalizes the required version string
func validateRequiredVersion(requiredVersion string) string {
	if clusterVersion == "" {
		ginkgo.Fail("Cluster version not initialized")
	}

	requiredVersion = semver.Canonical(requiredVersion)
	if !semver.IsValid(requiredVersion) {
		ginkgo.Fail("Requested cluster version is invalid")
	}

	return requiredVersion
}

// SkipIfOCPBelow skips test if cluster version is below requirement
// expects "v4.19" format
func SkipIfOCPBelow(requiredVersion string) {
	requiredVersion = validateRequiredVersion(requiredVersion)

	if semver.Compare(clusterVersion, requiredVersion) == -1 {
		ginkgo.Skip(fmt.Sprintf("Requires at least OCP %s+, cluster is %s", requiredVersion, clusterVersion))
	}
}

// IsOCPVersionAtLeast returns true if cluster version is at or above the required version
// expects "v4.15" format
func IsOCPVersionAtLeast(requiredVersion string) bool {
	requiredVersion = validateRequiredVersion(requiredVersion)

	return semver.Compare(clusterVersion, requiredVersion) >= 0
}
