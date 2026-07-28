package e2etests

import (
	"os"
	filePath "path/filepath"
)

const (
	netobservNS   = "openshift-netobserv-operator"
	NOPackageName = "netobserv-operator"

	minioNS        = "minio-aosqe"
	minioSecret    = "minio-creds"
	apiPath        = "/api/logs/v1/"
	queryRangePath = "/loki/api/v1/query_range"
	loNS           = "openshift-operators-redhat"
)

var (
	oc = &CLI{}

	NOcatSrc = Resource{"catalogsource", "netobserv-konflux-fbc", netobservNS}
	NOSource = CatalogSourceObjects{"stable", NOcatSrc.Name, NOcatSrc.Namespace}

	baseDir, _      = filePath.Abs("testdata")
	subscriptionDir = filePath.Join(baseDir, "subscription")
	flowFixturePath = filePath.Join(baseDir, "flowcollector_v1beta2_template.yaml")

	OperatorNS = OperatorNamespace{
		Name:              netobservNS,
		NamespaceTemplate: filePath.Join(subscriptionDir, "namespace.yaml"),
	}
	NO = SubscriptionObjects{
		OperatorName:  "netobserv-operator",
		Namespace:     netobservNS,
		PackageName:   NOPackageName,
		Subscription:  filePath.Join(subscriptionDir, "sub-template.yaml"),
		OperatorGroup: filePath.Join(subscriptionDir, "allnamespace-og.yaml"),
		CatalogSource: &NOSource,
	}

	imageDigest    = filePath.Join(subscriptionDir, "image-digest-mirror-set.yaml")
	catSrcTemplate = filePath.Join(subscriptionDir, "catalog-source.yaml")
	catalogSource  = os.Getenv("MULTISTAGE_PARAM_OVERRIDE_NETOBSERV_CS_IMAGE")
)
