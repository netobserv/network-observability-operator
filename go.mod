module github.com/netobserv/network-observability-operator

go 1.16

require (
	github.com/mitchellh/mapstructure v1.4.3
	github.com/netobserv/flowlogs-pipeline v0.1.2-0.20220602063928-b7e7bdbdfff0
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.19.0
	github.com/openshift/api v0.0.0-20220112145620-704957ce4980
	github.com/prometheus/common v0.32.1
	github.com/stretchr/testify v1.7.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/kube-aggregator v0.23.5
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.11.0
)
