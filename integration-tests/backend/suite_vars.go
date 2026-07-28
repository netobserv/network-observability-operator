package e2etests

import (
	"os"
	filePath "path/filepath"
)

var (
	// NetObserv Operator variables
	NOcatSrc = Resource{"catsrc", "netobserv-konflux-fbc", netobservNS}
	NOSource = CatalogSourceObjects{"stable", NOcatSrc.Name, NOcatSrc.Namespace}

	// Template directories
	baseDir, _      = filePath.Abs("testdata")
	subscriptionDir = filePath.Join(baseDir, "subscription")
	flowFixturePath = filePath.Join(baseDir, "flowcollector_v1beta2_template.yaml")

	// Operator namespace object
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
