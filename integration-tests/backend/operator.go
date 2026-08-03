package e2etests

import (
	"context"
	"fmt"
	filePath "path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	exutil "github.com/openshift/origin/test/extended/util"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

const (
	netobservNS   = "openshift-netobserv-operator"
	NOPackageName = "netobserv-operator"
)

// SubscriptionObjects objects are used to create operators via OLM
type SubscriptionObjects struct {
	OperatorName     string
	Namespace        string
	OperatorGroup    string // the file used to create operator group
	Subscription     string // the file used to create subscription
	PackageName      string
	CatalogSource    *CatalogSourceObjects `json:",omitempty"`
	OperatorPodLabel string
}

// CatalogSourceObjects defines the source used to subscribe an operator
type CatalogSourceObjects struct {
	Channel         string `json:",omitempty"`
	SourceName      string `json:",omitempty"`
	SourceNamespace string `json:",omitempty"`
}

// OperatorNamespace struct to handle creation of namespace
type OperatorNamespace struct {
	Name              string
	NamespaceTemplate string
}

// waitForPackagemanifestAppear waits for the packagemanifest to appear in the cluster
// chSource: bool value, true means the packagemanifests' source name must match the so.CatalogSource.SourceName, e.g.: oc get packagemanifests xxxx -l catalog=$source-name
func (so *SubscriptionObjects) waitForPackagemanifestAppear(oc *exutil.CLI, chSource bool) {
	args := []string{"-n", so.CatalogSource.SourceNamespace, "packagemanifests"}
	if chSource {
		args = append(args, "-l", "catalog="+so.CatalogSource.SourceName)
	} else {
		args = append(args, so.PackageName)
	}
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, false, func(context.Context) (done bool, err error) {
		packages, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			if strings.Contains(msg, "No resources found") || strings.Contains(msg, "NotFound") {
				return false, nil
			}
			return false, err
		}
		if strings.Contains(packages, so.PackageName) {
			return true, nil
		}
		e2e.Logf("Waiting for packagemanifest/%s to appear", so.PackageName)
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Packagemanifest %s is not availabile", so.PackageName))
}

// setCatalogSourceObjects set the default values of channel, source namespace and source name if they're not specified
func (so *SubscriptionObjects) setCatalogSourceObjects(oc *exutil.CLI) {
	// set channel
	if so.CatalogSource.Channel == "" {
		so.CatalogSource.Channel = "stable"
	}

	// set source namespace
	if so.CatalogSource.SourceNamespace == "" {
		so.CatalogSource.SourceNamespace = "openshift-marketplace"
	}

	// set source and check if the packagemanifest exists or not
	if so.CatalogSource.SourceName != "" {
		so.waitForPackagemanifestAppear(oc, true)
	} else {
		catsrc, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("catsrc", "-n", so.CatalogSource.SourceNamespace, "qe-app-registry").Output()
		if catsrc != "" && !(strings.Contains(catsrc, "NotFound")) {
			so.CatalogSource.SourceName = "qe-app-registry"
			so.waitForPackagemanifestAppear(oc, true)
		} else {
			so.waitForPackagemanifestAppear(oc, false)
			source, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests", so.PackageName, "-o", "jsonpath={.status.catalogSource}").Output()
			if err != nil {
				e2e.Logf("error getting catalog source name: %v", err)
			}
			so.CatalogSource.SourceName = source
		}
	}
}

// SubscribeOperator is used to subcribe the CLO and EO
func (so *SubscriptionObjects) SubscribeOperator(oc *exutil.CLI) {
	// check if the namespace exists, if it doesn't exist, create the namespace
	_, err := oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(), so.Namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			e2e.Logf("The project %s is not found, create it now...", so.Namespace)
			namespaceTemplate, _ := filePath.Abs("testdata/logging/subscription/namespace.yaml")
			namespaceFile, err := processTemplate(oc, "-f", namespaceTemplate, "-p", "NAMESPACE_NAME="+so.Namespace)
			o.Expect(err).NotTo(o.HaveOccurred())
			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 120*time.Second, false, func(context.Context) (done bool, err error) {
				output, err := oc.AsAdmin().Run("apply").Args("-f", namespaceFile).Output()
				if err != nil {
					if strings.Contains(output, "AlreadyExists") {
						return true, nil
					}
					return false, err
				}
				return true, nil
			})
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create project %s", so.Namespace))
		}
	}

	// check the operator group, if no object found, then create an operator group in the project
	og, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("-n", so.Namespace, "og").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	msg := fmt.Sprintf("%v", og)
	if strings.Contains(msg, "No resources found") {
		// create operator group
		ogFile, err := processTemplate(oc, "-n", so.Namespace, "-f", so.OperatorGroup, "-p", "OG_NAME="+so.Namespace, "NAMESPACE="+so.Namespace)
		o.Expect(err).NotTo(o.HaveOccurred())
		err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 120*time.Second, false, func(context.Context) (done bool, err error) {
			output, err := oc.AsAdmin().Run("apply").Args("-f", ogFile, "-n", so.Namespace).Output()
			if err != nil {
				if strings.Contains(output, "AlreadyExists") {
					return true, nil
				}
				return false, err
			}
			return true, nil
		})
		compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create operatorgroup %s in %s project", so.Namespace, so.Namespace))
	}

	// check subscription, if there is no subscription objets, then create one
	sub, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", so.Namespace, so.PackageName).Output()
	if err != nil {
		msg := fmt.Sprint("v%", sub)
		if strings.Contains(msg, "NotFound") {
			so.setCatalogSourceObjects(oc)
			// create subscription object
			subscriptionFile, err := processTemplate(oc, "-n", so.Namespace, "-f", so.Subscription, "-p", "PACKAGE_NAME="+so.PackageName, "NAMESPACE="+so.Namespace, "CHANNEL="+so.CatalogSource.Channel, "SOURCE="+so.CatalogSource.SourceName, "SOURCE_NAMESPACE="+so.CatalogSource.SourceNamespace)
			o.Expect(err).NotTo(o.HaveOccurred())
			err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 120*time.Second, false, func(context.Context) (done bool, err error) {
				output, err := oc.AsAdmin().Run("apply").Args("-f", subscriptionFile, "-n", so.Namespace).Output()
				if err != nil {
					if strings.Contains(output, "AlreadyExists") {
						return true, nil
					}
					return false, err
				}
				return true, nil
			})
			compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("can't create subscription %s in %s project", so.PackageName, so.Namespace))
		}
	}
	//WaitForDeploymentPodsToBeReady(oc, so.Namespace, so.OperatorName)
}

func deleteNamespace(oc *exutil.CLI, ns string) {
	err := oc.AdminKubeClient().CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		}
	}
	o.Expect(err).NotTo(o.HaveOccurred())
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 180*time.Second, false, func(context.Context) (bool, error) {
		_, err := oc.AdminKubeClient().CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
	compat_otp.AssertWaitPollNoErr(err, fmt.Sprintf("Namespace %s is not deleted in 3 minutes", ns))
}

func (so *SubscriptionObjects) uninstallOperator(oc *exutil.CLI) {
	_ = Resource{"subscription", so.PackageName, so.Namespace}.clear(oc)
	_ = oc.AsAdmin().WithoutNamespace().Run("delete").Args("-n", so.Namespace, "csv", "-l", "operators.coreos.com/"+so.PackageName+"."+so.Namespace+"=").Execute()
	// do not remove namespace openshift-logging and openshift-operators-redhat, and preserve the operatorgroup as there may have several operators deployed in one namespace
	// for example: loki-operator and elasticsearch-operator
	if so.Namespace != "openshift-logging" && so.Namespace != "openshift-operators-redhat" && so.Namespace != "openshift-operators" && so.Namespace != "openshift-netobserv-operator" && !strings.HasPrefix(so.Namespace, "e2e-test-") {
		deleteNamespace(oc, so.Namespace)
	}
}

func checkOperatorChannel(oc *exutil.CLI, operatorNamespace string, operatorName string) (string, error) {
	channelName, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", operatorName, "-n", operatorNamespace, "-o=jsonpath={.spec.channel}").Output()
	if err != nil {
		return "", err
	}
	return channelName, nil
}

func CheckOperatorStatus(oc *exutil.CLI, operatorNamespace string, operatorName string) (bool, error) {
	err := oc.AsAdmin().WithoutNamespace().Run("get").Args("namespace", operatorNamespace).Execute()
	if err != nil {
		e2e.Logf("%s operator will be created by tests", operatorName)
		return false, nil
	}

	// Check for subscription by exact name first, then fall back to finding any
	// subscription for this package (e.g. operator-sdk run bundle creates subs
	// with a different naming convention like <name>-v1-11-4-community-sub).
	csvName := ""
	err1 := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", operatorName, "-n", operatorNamespace).Execute()
	if err1 == nil {
		csvName, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", operatorName, "-n", operatorNamespace, "-o=jsonpath={.status.installedCSV}").Output()
	}
	if csvName == "" {
		// Look for any CSV owned by a subscription for this package in the namespace
		// operator-sdk run bundle creates subs with names like <name>-v0-0-0-sha-main-sub
		allSubs, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", "-n", operatorNamespace,
			"-o=jsonpath={range .items[?(@.spec.name==\""+operatorName+"\")]}{.status.installedCSV}{end}").Output()
		if allSubs != "" {
			csvName = allSubs
		}
	}
	if csvName == "" {
		e2e.Logf("%s operator will be created by tests", operatorName)
		return false, nil
	}

	e2e.Logf("Found CSV %s for operator %s", csvName, operatorName)
	err = wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 360*time.Second, false, func(context.Context) (bool, error) {
		csvState, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", csvName, "-n", operatorNamespace, "-o=jsonpath={.status.phase}").Output()
		if err != nil {
			return false, err
		}
		e2e.Logf("CSV %s state: %s", csvName, csvState)
		return csvState == "Succeeded", nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ns *OperatorNamespace) DeployOperatorNamespace(oc *exutil.CLI) {
	e2e.Logf("Creating %s operator namespace", ns.Name)
	// -n default: oc process requires a valid namespace context; in BeforeSuite the CLI may not have one yet
	nsParameters := []string{"-n", "default", "--ignore-unknown-parameters=true", "-f", ns.NamespaceTemplate, "-p", "NAMESPACE_NAME=" + ns.Name}
	compat_otp.ApplyClusterResourceFromTemplate(oc, nsParameters...)
}

// setupCatalogSource deploys the catalog source and image digest mirror set
func setupCatalogSource(oc *exutil.CLI, catSrc Resource, catSrcTemplate, imageDigest, catalogSource string, isHypershift bool, NOSource *CatalogSourceObjects, NO *SubscriptionObjects) error {
	g.By("Deploy konflux FBC and ImageDigestMirrorSet")
	upstreamCatalogSource := "quay.io/netobserv/network-observability-operator-catalog:v0.0.0-sha-main"
	var catsrcErr error

	if catalogSource != "" {
		e2e.Logf("Using %s catalog", catalogSource)
		catsrcErr = catSrc.applyFromTemplate(oc, "-n", catSrc.Namespace, "-f", catSrcTemplate, "-p", "NAMESPACE="+catSrc.Namespace, "IMAGE="+catalogSource)
	} else if isHypershift {
		e2e.Logf("Using v0.0.0-sha-main catalog for hypershift")
		catsrcErr = catSrc.applyFromTemplate(oc, "-n", catSrc.Namespace, "-f", catSrcTemplate, "-p", "NAMESPACE="+catSrc.Namespace, "IMAGE="+upstreamCatalogSource)
		NOSource.Channel = "latest"
		NO.CatalogSource = NOSource
	} else {
		e2e.Logf("Using default ystream catalog")
		catsrcErr = catSrc.applyFromTemplate(oc, "-n", catSrc.Namespace, "-f", catSrcTemplate, "-p", "NAMESPACE="+catSrc.Namespace)
	}
	catSrc.WaitUntilCatSrcReady(oc)

	if !isHypershift {
		ApplyResourceFromFile(oc, catSrc.Namespace, imageDigest)
	}
	return catsrcErr
}

// ensureOperatorDeployed checks and deploys an operator if not already present
func ensureOperatorDeployed(oc *exutil.CLI, operator SubscriptionObjects, operatorSource CatalogSourceObjects, podLabel string) {
	g.By(fmt.Sprintf("Subscribe %s operator to %s channel", operator.OperatorName, operatorSource.Channel))
	operatorExisting, err := CheckOperatorStatus(oc, operator.Namespace, operator.PackageName)
	o.Expect(err).NotTo(o.HaveOccurred())

	if !operatorExisting {
		e2e.Logf("%s operator not found, subscribing to operator", operator.OperatorName)
		operator.SubscribeOperator(oc)

		// Wait for operator pods to be ready
		if podLabel != "" {
			WaitForPodsReadyWithLabel(oc, operator.Namespace, podLabel)
		}

		// Verify operator status
		operatorStatus, err := CheckOperatorStatus(oc, operator.Namespace, operator.PackageName)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(operatorStatus).To(o.BeTrue())

		e2e.Logf("%s operator deployed successfully", operator.OperatorName)
	} else {
		e2e.Logf("%s operator already exists, skipping deployment", operator.OperatorName)
	}
}

// ensureNetObservOperatorDeployed checks and deploys the NetObserv operator with specific configurations
func ensureNetObservOperatorDeployed(oc *exutil.CLI, NO SubscriptionObjects, NOSource CatalogSourceObjects) {
	ensureOperatorDeployed(oc, NO, NOSource, "app="+NO.OperatorName)

	// NetObserv-specific checks only if operator was just deployed
	NOexisting, err := CheckOperatorStatus(oc, NO.Namespace, NO.PackageName)
	o.Expect(err).NotTo(o.HaveOccurred())

	if NOexisting {
		// Verify FlowCollector API exists
		flowcollectorAPIExists, err := isFlowCollectorAPIExists(oc)
		o.Expect(flowcollectorAPIExists).To(o.BeTrue())
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}

func getOperatorChannel(oc *exutil.CLI, catalog string, packageName string) (operatorChannel string, err error) {
	channels, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests", "-l", "catalog="+catalog, "-n", "openshift-marketplace", "-o=jsonpath={.items[?(@.metadata.name==\""+packageName+"\")].status.channels[*].name}").Output()
	channelArr := strings.Split(channels, " ")
	return channelArr[len(channelArr)-1], err
}

// getCSVVersion returns the version of the installed NetObserv Operator CSV in the given namespace.
// It uses the dynamic client to list CSVs and finds the one whose name starts with the operator name prefix.
func getCSVVersion(dynamicClient dynamic.Interface, operatorNamespace string) (string, error) {
	csvGVR := schema.GroupVersionResource{
		Group:    "operators.coreos.com",
		Version:  "v1alpha1",
		Resource: "clusterserviceversions",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	csvList, err := dynamicClient.Resource(csvGVR).Namespace(operatorNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list CSVs in namespace %s: %v", operatorNamespace, err)
	}

	var matched []string
	for _, csv := range csvList.Items {
		if !strings.HasPrefix(csv.GetName(), NOPackageName+".") {
			continue
		}
		// Only consider the active CSV (phase "Succeeded"); during upgrades,
		// replaced CSVs have phase "Replacing" or "Deleting".
		phase, _, _ := unstructured.NestedString(csv.Object, "status", "phase")
		if phase != "Succeeded" {
			continue
		}
		version, found, err := unstructured.NestedString(csv.Object, "spec", "version")
		if err != nil || !found {
			return "", fmt.Errorf("version field not found in CSV %s", csv.GetName())
		}
		matched = append(matched, version)
	}
	switch len(matched) {
	case 0:
		return "", fmt.Errorf("no active CSV found for %s in namespace %s", NOPackageName, operatorNamespace)
	case 1:
		return matched[0], nil
	default:
		return "", fmt.Errorf("multiple active CSVs found for %s in namespace %s: %v", NOPackageName, operatorNamespace, matched)
	}
}
