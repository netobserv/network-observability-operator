# VERSION defines the project version for the deploy scripts, not for bundles
VERSION ?= main
BUILD_VERSION := $(shell git describe --long HEAD)
BUILD_DATE := $(shell date +%Y-%m-%d\ %H:%M)
BUILD_SHA := $(shell git rev-parse --short HEAD)

# Component versions to use in bundle / release (do not use $VERSION for that)
PREVIOUS_VERSION ?= v0.2.1
BUNDLE_VERSION ?= 0.2.2
OPERATOR_VERSION ?= $(BUNDLE_VERSION)
# console plugin
export PLG_VERSION ?= v0.1.7
# flowlogs-pipeline
export FLP_VERSION ?= v0.1.6
# eBPF agent
export BPF_VERSION ?= v0.2.3

# Allows building bundles in Mac replacing BSD 'sed' command by GNU-compatible 'gsed'
SED ?= sed

# Port-forward (for loki/grafana deployments)
PORT_FWD ?= true

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
CHANNELS := v1.0.x
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL := v1.0.x
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# netobserv/network-observability-operator-bundle:$BUNDLE_VERSION and netobserv/network-observability-operator-catalog:$BUNDLE_VERSION.
IMAGE_TAG_BASE ?= quay.io/netobserv/network-observability-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(BUNDLE_VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)
IMG_SHA = $(IMAGE_TAG_BASE):$(BUILD_SHA)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23
GOLANGCI_LINT_VERSION = v1.50.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Image building tool (docker / podman)
ifndef OCI_BIN
ifeq (,$(shell which podman 2>/dev/null))
OCI_BIN=docker
else
OCI_BIN=podman
endif
endif

DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

NAMESPACE ?= netobserv

all: help

include .bingo/Variables.mk

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

gencode: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

doc: crdoc ## Generate 
	$(CRDOC) --resources config/crd/bases/flows.netobserv.io_flowcollectors.yaml --output docs/FlowCollector.md

generate: gencode manifests doc

.PHONY: prereqs
prereqs:
	@echo "### Test if prerequisites are met, and installing missing dependencies"
	test -f $(go env GOPATH)/bin/golangci-lint || GOFLAGS="" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

.PHONY: vendors
vendors:
	@echo "### Checking vendors"
	go mod tidy && go mod vendor

fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: lint
lint: prereqs
	@echo "### Linting code"
	golangci-lint run --timeout 5m ./...

test: generate envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverpkg=./... -coverprofile cover.out

coverage-report:
	go tool cover --func=./cover.out

coverage-report-html:
	go tool cover --html=./cover.out

##@ Build

build: generate fmt lint ## Build manager binary.
	go build -ldflags "-X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -mod vendor -o bin/manager main.go

image-build:
	$(OCI_BIN) build --build-arg BUILD_VERSION="${BUILD_VERSION}" -t ${IMG} .

ci-images-build: image-build
	$(OCI_BIN) build --build-arg BASE_IMAGE=$(IMG) -t $(IMG_SHA) -f ./shortlived.Dockerfile .

image-push: ## Push OCI image with the manager.
ifneq (,$(findstring quay.io/netobserv/,$(IMG)))
	$(error Do not push to quay.io/netobserv)
else
	$(OCI_BIN) push ${IMG}
endif

##@ Deployment

install: generate kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: generate kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl --ignore-not-found=true delete -f - || true

deploy: generate kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(SED) -i -r 's~ebpf-agent:.+~ebpf-agent:main~' ./config/manager/manager.yaml
	$(SED) -i -r 's~flowlogs-pipeline:.+~flowlogs-pipeline:main~' ./config/manager/manager.yaml
	$(SED) -i -r 's~console-plugin:.+~console-plugin:main~' ./config/manager/manager.yaml
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl --ignore-not-found=true delete -f - || true

run: generate fmt lint ## Run a controller from your host.
	go run ./main.go

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

CRDOC = $(shell pwd)/bin/crdoc
crdoc: ## Download crdoc locally if necessary.
	$(call go-install-tool,$(CRDOC),fybrik.io/crdoc@v0.5.2)

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(firstword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin GOFLAGS="-mod=mod" go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: operator-sdk
OPSDK = ./bin/operator-sdk
OPSDK: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPSDK)))
ifeq (,$(shell which operator-sdk 2>/dev/null))
	@{ \
	echo "### Downloading operator-sdk"; \
	set -e ;\
	mkdir -p $(dir $(OPSDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPSDK) https://github.com/operator-framework/operator-sdk/releases/download/v1.25.3/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPSDK) ;\
	}
else
OPSDK = $(shell which operator-sdk)
endif
endif

.PHONY: bundle-prepare
bundle-prepare: OPSDK generate kustomize ## Generate bundle manifests and metadata, then validate generated files.
	$(OPSDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE_TAG_BASE):$(OPERATOR_VERSION)
	$(SED) -i -r 's~ebpf-agent:.+~ebpf-agent:$(BPF_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~flowlogs-pipeline:.+~flowlogs-pipeline:$(FLP_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~console-plugin:.+~console-plugin:$(PLG_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(OPERATOR_VERSION)/~g' ./config/manifests/bases/netobserv-operator.clusterserviceversion.yaml
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(OPERATOR_VERSION)/~g' ./config/manifests/bases/description-upstream.md
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(OPERATOR_VERSION)/~g' ./config/manifests/bases/description-ocp.md
	$(SED) -i -r 's~replaces: netobserv-operator\.v.*~replaces: netobserv-operator\.$(PREVIOUS_VERSION)~' ./config/manifests/bases/netobserv-operator.clusterserviceversion.yaml
	cp config/samples/flows_v1alpha1_flowcollector.yaml config/samples/flows_v1alpha1_flowcollector_versioned.yaml

.PHONY: bundle
bundle: bundle-prepare
	$(SED) -e 's/^/    /' config/manifests/bases/description-upstream.md > tmp-desc
	$(KUSTOMIZE) build config/manifests \
		| $(SED) -e 's~:container-image:~$(IMAGE_TAG_BASE):$(OPERATOR_VERSION)~' \
		| $(SED) -e 's~:created-at:~$(DATE)~' \
		| $(SED) -e "/':full-description:'/r tmp-desc" \
		| $(SED) -e "s/':full-description:'/|\-/" \
		| $(OPSDK) generate bundle -q --overwrite --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	rm tmp-desc
	sh -c '\
	VALIDATION_OUTPUT=$$($(OPSDK) bundle validate ./bundle --select-optional suite=operatorframework); \
	echo $${VALIDATION_OUTPUT}; \
	if [ $$(echo $${VALIDATION_OUTPUT} | grep -i 'warning' | wc -c) -gt 0 ]; then echo "please correct warnings and errors first"; exit -1 ; fi \
	'


.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(OCI_BIN) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) image-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.19.5/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(BUNDLE_VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool ${OCI_BIN} --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) image-push IMG=$(CATALOG_IMG)

# Deploy the catalog.
.PHONY: catalog-deploy
catalog-deploy:
	$(SED) -e 's~<IMG>~$(CATALOG_IMG)~' ./config/samples/catalog/catalog.yaml | kubectl apply -f -

# Undeploy the catalog.
.PHONY: catalog-undeploy
catalog-undeploy:
	kubectl delete -f ./config/samples/catalog/catalog.yaml

include .mk/sample.mk
include .mk/development.mk
include .mk/local.mk
include .mk/ocp.mk
