package e2etests

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	exutil "github.com/openshift/origin/test/extended/util"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"

	"golang.org/x/mod/semver"
)

var clusterVersion string

func GetOCPVersion(oc *exutil.CLI) (string, error) {
	if clusterVersion != "" {
		return clusterVersion, nil
	}

	var err error
	_, clusterVersion, err = compat_otp.GetClusterVersion(oc)
	clusterVersion = semver.Canonical("v" + clusterVersion)
	clusterVersion = semver.MajorMinor(clusterVersion)
	fmt.Printf("Detected OCP version: %s\n", clusterVersion)
	return clusterVersion, err
}

// SkipIfOCPBelow skips test if cluster version is below requirement
// expects "v4.19" format
func SkipIfOCPBelow(requiredVersion string) {
	if clusterVersion == "" {
		ginkgo.Fail("Cluster version not initialized")
	}

	requiredVersion = semver.Canonical(requiredVersion)
	if !semver.IsValid(requiredVersion) {
		ginkgo.Fail("Requested cluster version is invalid")
	}

	if semver.Compare(clusterVersion, requiredVersion) == -1 {
		ginkgo.Skip(fmt.Sprintf("Requires at least OCP %s+, cluster is %s", requiredVersion, clusterVersion))
	}
}
