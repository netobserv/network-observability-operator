module github.com/netobserv/network-observability-operator

go 1.16

require (
	github.com/RedHatInsights/strimzi-client-go v0.26.0
	github.com/go-logr/logr v1.2.3
	github.com/grafana-operator/grafana-operator/v4 v4.1.1
	github.com/grafana/loki/operator v0.0.0-20220503111539-93de7a7061f4
	github.com/mitchellh/mapstructure v1.4.1
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.19.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/api v0.13.0
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.54.1
	github.com/stretchr/testify v1.7.1
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver v0.23.0
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/kube-aggregator v0.23.5
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.11.0
)

//replace openshift/api version 3.9.0+incompatible from current + grafana-operator to specific one
replace github.com/openshift/api v3.9.0+incompatible => github.com/openshift/api v0.0.0-20220112145620-704957ce4980
