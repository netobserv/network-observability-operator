# VERSION defines the project version for the deploy scripts, not for bundles
VERSION ?= main
BUILD_DATE := $(shell date +%Y-%m-%d\ %H:%M)
TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
BUILD_SHA := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
	BUILD_VERSION := $(BUILD_VERSION)-$(BUILD_SHA)
endif
ifneq ($(shell git status --porcelain),)
	BUILD_VERSION := $(BUILD_VERSION)-dirty
endif

# Go architecture and targets images to build
GOARCH ?= amd64
MULTIARCH_TARGETS ?= amd64

# In CI, to be replaced by `netobserv`
IMAGE_ORG ?= $(USER)

# Default image repo
REPO ?= quay.io/$(IMAGE_ORG)

# Component versions to use in bundle / release (do not use $VERSION for that)
PREVIOUS_VERSION ?= v1.0.4
BUNDLE_VERSION ?= 1.0.5
# console plugin
export PLG_VERSION ?= v0.1.12
# flowlogs-pipeline
export FLP_VERSION ?= v0.1.11
# eBPF agent
export BPF_VERSION ?= v0.3.3

# Allows building bundles in Mac replacing BSD 'sed' command by GNU-compatible 'gsed'
ifeq (,$(shell which gsed 2>/dev/null))
SED ?= sed
else
SED ?= gsed
endif

# Port-forward (for loki/grafana deployments)
PORT_FWD ?= true

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
CHANNELS := latest,v1.0.x
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL := latest
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-push catalog-push' will build and push both
# netobserv/network-observability-operator-bundle:$BUNDLE_VERSION and netobserv/network-observability-operator-catalog:$BUNDLE_VERSION.
IMAGE_TAG_BASE ?= $(REPO)/network-observability-operator

# BUNDLE_IMAGE defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMAGE=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMAGE ?= $(IMAGE_TAG_BASE)-bundle:v$(BUNDLE_VERSION)

# BUNDLE_CONFIG is the config sources to use for OLM bundle - "config/openshift-olm" for OpenShift, or "config/k8s-olm" for upstream Kubernetes.
BUNDLE_CONFIG ?= config/openshift-olm

# Image URL to use all building/pushing image targets
IMAGE ?= $(IMAGE_TAG_BASE):$(VERSION)
OCI_BUILD_OPTS ?=
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23
GOLANGCI_LINT_VERSION = v1.53.3

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Image building tool (docker / podman) - docker is preferred in CI
OCI_BIN_PATH := $(shell which docker 2>/dev/null || which podman)
OCI_BIN ?= $(shell basename ${OCI_BIN_PATH})

DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

NAMESPACE ?= netobserv

# Local paths from preparing upstream release to OperatorHub
ifeq ("$(BUNDLE_CONFIG)", "config/openshift-olm")
	OPERATORHUB_PATH ?= "../community-operators-prod"
else
	OPERATORHUB_PATH ?= "../community-operators"
endif

all: help

include .bingo/Variables.mk

# build a single arch target provided as argument
define build_target
	echo 'building image for arch $(1)'; \
	DOCKER_BUILDKIT=1 $(OCI_BIN) buildx build --load --build-arg TARGETPLATFORM=linux/$(1) --build-arg TARGETARCH=$(1) --build-arg BUILDPLATFORM=linux/amd64 ${OCI_BUILD_OPTS} -t ${IMAGE}-$(1) -f Dockerfile .;
endef

# push a single arch target image
define push_target
	echo 'pushing image ${IMAGE}-$(1)'; \
	DOCKER_BUILDKIT=1 $(OCI_BIN) push ${IMAGE}-$(1);
endef

# manifest create a single arch target provided as argument
define manifest_add_target
	echo 'manifest add target $(1)'; \
	DOCKER_BUILDKIT=1 $(OCI_BIN) manifest add ${IMAGE} ${IMAGE}-$(1);
endef

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

# Directories.

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(abspath bin)
GO_INSTALL := ./scripts/go_install.sh
CONVERSION_GEN_VER := v0.25.0
CONVERSION_GEN_BIN := conversion-gen
# We are intentionally using the binary without version suffix, to avoid the version
# in generated files.
CONVERSION_GEN := $(BIN_DIR)/$(CONVERSION_GEN_BIN)
CONVERSION_GEN_PKG := k8s.io/code-generator/cmd/conversion-gen
# Set --output-base for conversion-gen if we are not within GOPATH
ifneq ($(findstring $(shell go env GOPATH), $(ROOT_DIR)), $(shell go env GOPATH))
	CONVERSION_GEN_OUTPUT_BASE := --output-base=$(ROOT_DIR)
else
	export GOPATH := $(shell go env GOPATH)
endif

##@ Tools
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

.PHONY: $(CONVERSION_GEN_BIN)
$(CONVERSION_GEN_BIN): $(CONVERSION_GEN) ## Build a local copy of conversion-gen.

## We are forcing a rebuilt of conversion-gen via PHONY so that we're always using an up-to-date version.
## We can't use a versioned name for the binary, because that would be reflected in generated files.
.PHONY: $(CONVERSION_GEN)
$(CONVERSION_GEN): ## Build conversion-gen from tools folder.
	GOBIN=$(BIN_DIR) $(GO_INSTALL) $(CONVERSION_GEN_PKG) $(CONVERSION_GEN_BIN) $(CONVERSION_GEN_VER)

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
OPSDK: ## Download operator-sdk locally if necessary.
ifeq (,$(shell which $(OPSDK) 2>/dev/null))
	@{ \
	echo "### Downloading operator-sdk"; \
	set -e ;\
	mkdir -p $(dir $(OPSDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPSDK) https://github.com/operator-framework/operator-sdk/releases/download/v1.25.3/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPSDK) ;\
	}
endif

.PHONY: YQ
YQ = ./bin/yq
YQ: ## Download yq locally if necessary.
ifeq (,$(shell which $(YQ) 2>/dev/null))
	@{ \
	echo "### Downloading yq"; \
	set -e ;\
	mkdir -p $(dir $(YQ)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/v4.35.2/yq_$${OS}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}
endif

##@ Code / files generation
manifests: YQ controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(YQ) -i 'del(.spec.versions[].schema.openAPIV3Schema.properties.spec.properties.processor.properties.kafkaConsumerAutoscaler.properties.metrics.items | .. | select(has("description")) | .description)' config/crd/bases/flows.netobserv.io_flowcollectors.yaml
	$(YQ) -i 'del(.spec.versions[].schema.openAPIV3Schema.properties.spec.properties.consolePlugin.properties.autoscaler.properties.metrics.items | .. | select(has("description")) | .description)' config/crd/bases/flows.netobserv.io_flowcollectors.yaml
	$(YQ) -i 'del(.spec.versions[].schema.openAPIV3Schema.properties.spec.properties.agent.properties.ebpf.properties.advanced.properties.affinity.properties | .. | select(has("description")) | .description)' config/crd/bases/flows.netobserv.io_flowcollectors.yaml
	$(YQ) -i 'del(.spec.versions[].schema.openAPIV3Schema.properties.spec.properties.processor.properties.advanced.properties.affinity.properties | .. | select(has("description")) | .description)' config/crd/bases/flows.netobserv.io_flowcollectors.yaml
	$(YQ) -i 'del(.spec.versions[].schema.openAPIV3Schema.properties.spec.properties.consolePlugin.properties.advanced.properties.affinity.properties | .. | select(has("description")) | .description)' config/crd/bases/flows.netobserv.io_flowcollectors.yaml

gencode: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

doc: crdoc ## Generate markdown documentation
	$(CRDOC) --resources config/crd/bases/flows.netobserv.io_flowcollectors.yaml --output docs/FlowCollector.md

generate-go-conversions: $(CONVERSION_GEN) ## Run all generate-go-conversions
		$(MAKE) clean-generated-conversions SRC_DIRS="./apis/flowcollector/v1beta1"
		$(CONVERSION_GEN) \
		--input-dirs=./apis/flowcollector/v1beta1 \
		--build-tag=ignore_autogenerated_core \
		--output-file-base=zz_generated.conversion \
		$(CONVERSION_GEN_OUTPUT_BASE) \
		--go-header-file=./hack/boilerplate/boilerplate.generatego.txt

# Hack to reintroduce when the API stored version != latest version; see also envtest.go (CRD path config)
# .PHONY: hack-crd-for-test
# hack-crd-for-test: YQ
# 	cat ./config/crd/bases/flows.netobserv.io_flowcollectors.yaml \
# 		| $(YQ) eval-all \
# 		'(.spec.versions.[]|select(.name != "v1beta2").storage) = false,(.spec.versions.[]|select(.name == "v1beta2").storage) = true' \
# 		> ./hack/cloned.flows.netobserv.io_flowcollectors.yaml
# 	cp ./config/crd/bases/flows.netobserv.io_flowmetrics.yaml ./hack/cloned.flows.netobserv.io_flowmetrics.yaml

generate: gencode manifests doc generate-go-conversions ## Run all code/file generators

.PHONY: clean-generated-conversions
clean-generated-conversions: ## Remove files generated by conversion-gen from the mentioned dirs
	(IFS=','; for i in $(SRC_DIRS); do find $$i -type f -name 'zz_generated.conversion*' -exec rm -f {} \;; done)

##@ Development
.PHONY: prereqs
prereqs:
	@echo "### Test if prerequisites are met, and installing missing dependencies"
	GOFLAGS="" go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

.PHONY: vendors
vendors: ## Refresh vendors directory.
	@echo "### Checking vendors"
	go mod tidy && go mod vendor

fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: lint
lint: prereqs ## Run linter (golangci-lint).
	@echo "### Linting code"
	golangci-lint run --timeout 5m ./...

test: envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverpkg=./... -coverprofile cover.out

coverage-report: ## Generate coverage report
	go tool cover --func=./cover.out

coverage-report-html: ## Generate HTML coverage report
	go tool cover --html=./cover.out

build: fmt lint ## Build manager binary.
	GOARCH=${GOARCH} go build -ldflags "-X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -mod vendor -o bin/manager main.go

##@ Images

# note: to build and push custom image tag use: IMAGE_ORG=myuser VERSION=dev make images
.PHONY: image-build
image-build: ## Build MULTIARCH_TARGETS images
	trap 'exit' INT; \
	$(foreach target,$(MULTIARCH_TARGETS),$(call build_target,$(target)))

.PHONY: image-push
image-push: ## Push MULTIARCH_TARGETS images
	trap 'exit' INT; \
	$(foreach target,$(MULTIARCH_TARGETS),$(call push_target,$(target)))

.PHONY: manifest-build
manifest-build: ## Build MULTIARCH_TARGETS manifest
	@echo 'building manifest $(IMAGE)'
	DOCKER_BUILDKIT=1 $(OCI_BIN) rmi ${IMAGE} -f
	DOCKER_BUILDKIT=1 $(OCI_BIN) manifest create ${IMAGE} $(foreach target,$(MULTIARCH_TARGETS), --amend ${IMAGE}-$(target));

.PHONY: manifest-push
manifest-push: ## Push MULTIARCH_TARGETS manifest
	@echo 'publish manifest $(IMAGE)'
ifeq (${OCI_BIN}, docker)
	DOCKER_BUILDKIT=1 $(OCI_BIN) manifest push ${IMAGE};
else
	DOCKER_BUILDKIT=1 $(OCI_BIN) manifest push ${IMAGE} docker://${IMAGE};
endif

##@ Deployment

install: kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl --ignore-not-found=true delete -f - || true

deploy: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMAGE}
	$(SED) -i -r 's~ebpf-agent:.+~ebpf-agent:main~' ./config/manager/manager.yaml
	$(SED) -i -r 's~flowlogs-pipeline:.+~flowlogs-pipeline:main~' ./config/manager/manager.yaml
	$(SED) -i -r 's~console-plugin:.+~console-plugin:main~' ./config/manager/manager.yaml
	$(KUSTOMIZE) build config/openshift | sed -r "s/openshift-netobserv-operator\.svc/${NAMESPACE}.svc/" | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/openshift | kubectl --ignore-not-found=true delete -f - || true

run: fmt lint ## Run a controller from your host.
	go run ./main.go

##@ OLM

.PHONY: bundle-prepare
bundle-prepare: OPSDK generate kustomize ## Generate bundle manifests and metadata, then validate generated files.
# $(OPSDK) generate kustomize manifests -q --input-dir $(BUNDLE_CONFIG) --output-dir $(BUNDLE_CONFIG)
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE)
	$(SED) -i -r 's~ebpf-agent:.+~ebpf-agent:$(BPF_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~flowlogs-pipeline:.+~flowlogs-pipeline:$(FLP_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~console-plugin:.+~console-plugin:$(PLG_VERSION)~' ./config/manager/manager.yaml
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(VERSION)/~g' ./config/csv/bases/netobserv-operator.clusterserviceversion.yaml
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(VERSION)/~g' ./config/descriptions/upstream.md
	$(SED) -i -r 's~network-observability-operator/blob/[^/]+/~network-observability-operator/blob/$(VERSION)/~g' ./config/descriptions/ocp.md
	$(SED) -i -r 's~replaces: netobserv-operator\.v.*~replaces: netobserv-operator\.$(PREVIOUS_VERSION)~' ./config/csv/bases/netobserv-operator.clusterserviceversion.yaml

.PHONY: bundle
bundle: bundle-prepare ## Generate final bundle files.
	rm -r bundle/manifests
	rm -r bundle/metadata
	cp ./config/csv/bases/netobserv-operator.clusterserviceversion.yaml tmp-csv
	hack/crd2csvSpecDesc.sh v1beta2
	$(SED) -e 's/^/    /' config/descriptions/upstream.md > tmp-desc
	$(KUSTOMIZE) build $(BUNDLE_CONFIG) \
		| $(SED) -e 's~:container-image:~$(IMAGE)~' \
		| $(SED) -e "/':full-description:'/r tmp-desc" \
		| $(SED) -e "s/':full-description:'/|\-/" \
		| $(OPSDK) generate bundle -q --overwrite --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	mv tmp-csv ./config/csv/bases/netobserv-operator.clusterserviceversion.yaml
	rm tmp-desc
	sh -c '\
	VALIDATION_OUTPUT=$$($(OPSDK) bundle validate ./bundle --select-optional suite=operatorframework); \
	echo $${VALIDATION_OUTPUT}; \
	if [ $$(echo $${VALIDATION_OUTPUT} | grep -i 'warning' | wc -c) -gt 0 ]; then echo "please correct warnings and errors first"; exit -1 ; fi \
	'

.PHONY: update-bundle
update-bundle: VERSION=$(BUNDLE_VERSION)
update-bundle: IMAGE_ORG=netobserv
update-bundle: bundle ## Prepare a clean bundle to be commited

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	cp ./bundle/manifests/netobserv-operator.clusterserviceversion.yaml tmp-bundle
	$(SED) -i -r 's~:created-at:~$(DATE)~' ./bundle/manifests/netobserv-operator.clusterserviceversion.yaml
	-$(OCI_BIN) build $(OCI_BUILD_OPTS) -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .
	mv tmp-bundle ./bundle/manifests/netobserv-operator.clusterserviceversion.yaml

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(OCI_BIN) push ${BUNDLE_IMAGE};

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMAGES=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMAGES ?= $(BUNDLE_IMAGE)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMAGE=example.com/operator-catalog:v0.2.0).
CATALOG_IMAGE ?= $(IMAGE_TAG_BASE)-catalog:v$(BUNDLE_VERSION)

# Set CATALOG_BASE_IMAGE to an existing catalog image tag to add $BUNDLE_IMAGES to that image.
ifneq ($(origin CATALOG_BASE_IMAGE), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMAGE)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool ${OCI_BIN} --mode semver --tag $(CATALOG_IMAGE) --bundles $(BUNDLE_IMAGES) $(FROM_INDEX_OPT) $(OPM_OPTS)

shortlived-catalog-build: ## Build a temporary catalog image, expiring after 2 weeks on quay
	$(MAKE) catalog-build CATALOG_IMAGE=temp-catalog
	echo "FROM temp-catalog" | $(OCI_BIN) build --label quay.expires-after=2w -t $(CATALOG_IMAGE) -

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(OCI_BIN) push ${CATALOG_IMAGE};

# Deploy the catalog.
.PHONY: catalog-deploy
catalog-deploy: ## Deploy a catalog image.
	$(SED) -e 's~<IMAGE>~$(CATALOG_IMAGE)~' ./config/samples/catalog/catalog.yaml | kubectl apply -f -

# Undeploy the catalog.
.PHONY: catalog-undeploy
catalog-undeploy: ## Undeploy a catalog image.
	kubectl delete -f ./config/samples/catalog/catalog.yaml

##@ Misc

.PHONY: test-workflow
test-workflow: ## Run some tests on this Makefile and the github workflow
	hack/test-workflow.sh

.PHONY: related-release-notes
related-release-notes: ## Grab release notes for related components (to be inserted in operator's release note upstream, cf RELEASE.md)
	echo -e "## Related components\n\n" > /tmp/related.md
	echo -e "### eBPF Agent\n\n" >> /tmp/related.md
	curl -s  https://api.github.com/repos/netobserv/netobserv-ebpf-agent/releases/tags/$(BPF_VERSION) | jq -r .body | xargs -0 printf "%b" | sed -r "s/##/####/" >> /tmp/related.md
	echo -e "### Flowlogs-Pipeline\n\n" >> /tmp/related.md
	curl -s  https://api.github.com/repos/netobserv/flowlogs-pipeline/releases/tags/$(FLP_VERSION) | jq -r .body | xargs -0 printf "%b" | sed -r "s/##/####/" >> /tmp/related.md
	echo -e "### Console Plugin\n\n" >> /tmp/related.md
	curl -s  https://api.github.com/repos/netobserv/network-observability-console-plugin/releases/tags/$(PLG_VERSION) | jq -r .body | xargs -0 printf "%b" | sed -r "s/##/####/" >> /tmp/related.md
	wl-copy < /tmp/related.md
	cat /tmp/related.md
	echo -e "\nText has been copied to the clipboard.\n"

.PHONY: prepare-operatorhub
prepare-operatorhub: ## Copy bundle for an upstream release on OperatorHub
	$(SED) -i '/scorecard/d' ./bundle.Dockerfile
	$(SED) -i '/scorecard/d' ./bundle/metadata/annotations.yaml
	$(SED) -i '/Annotations for testing/d' ./bundle/metadata/annotations.yaml
	$(SED) -i -r 's~:created-at:~$(DATE)~' ./bundle/manifests/netobserv-operator.clusterserviceversion.yaml
	@read -p "Going to hard-reset git's $(OPERATORHUB_PATH) - type y to proceed: " -n 1 -r; \
	if [[ $$REPLY =~ ^[^Yy] ]]; \
	then \
			exit 1; \
	fi
	cd $(OPERATORHUB_PATH) && git fetch upstream && git reset --hard upstream/main && cd -
	mkdir -p "$(OPERATORHUB_PATH)/operators/netobserv-operator/$(BUNDLE_VERSION)"
	cp -r bundle/manifests "$(OPERATORHUB_PATH)/operators/netobserv-operator/$(BUNDLE_VERSION)"
	cp -r bundle/metadata "$(OPERATORHUB_PATH)/operators/netobserv-operator/$(BUNDLE_VERSION)"
ifeq ($(BUNDLE_CONFIG), "config/openshift-olm")
	echo "  com.redhat.openshift.versions: \"v4.10-v4.13\"" >> $(OPERATORHUB_PATH)/operators/netobserv-operator/$(BUNDLE_VERSION)/metadata/annotations.yaml
endif
	cd $(OPERATORHUB_PATH) && git add -A

	@echo ""
	@echo "Everything is ready to be pushed. Before that, you should compare the content of $(BUNDLE_VERSION) with $(PREVIOUS_VERSION) to make sure it looks correct."

include .mk/sample.mk
include .mk/development.mk
include .mk/local.mk
include .mk/ocp.mk
include .mk/shortcuts.mk
