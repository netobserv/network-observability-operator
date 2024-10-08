
#!/usr/bin/env bash

echo "Updating container file"

: "${COMMIT:=$(git rev-list --abbrev-commit --tags --max-count=1)}"
: "${CONTAINER_FILE:=./Dockerfile}"
: "${BUNDLE_CONTAINER_FILE:=./bundle.Dockerfile}"
: "${CATALOG_CONTAINER_FILE:=./catalog.Dockerfile}"
: "${TARGET_VERSION:=1.7.0}"
: "${REPLACE_VERSION:=1.6.1}"


supported_ocp_versions="v4.13"
manifests_dir="./bundle/manifests"
metadata_dir="./bundle/metadata"
crd_name="flows.netobserv.io_flowcollectors.yaml"
crd_file="${manifests_dir}/${crd_name}"
csv_name="netobserv-operator.clusterserviceversion.yaml"
csv_file="${manifests_dir}/${csv_name}"
index_file="./catalog/index.yaml"

source ./hack/container_digest.sh
source ./hack/bundle_digest.sh

cat <<EOF >>"${CONTAINER_FILE}"
LABEL com.redhat.component="network-observability-operator-container"
LABEL name="network-observability-operator"
LABEL io.k8s.display-name="Network Observability Operator"
LABEL io.k8s.description="Network Observability Operator"
LABEL summary="Network Observability Operator"
LABEL maintainer="support@redhat.com"
LABEL io.openshift.tags="network-observability-operator"
LABEL upstream-vcs-ref="${COMMIT}"
LABEL upstream-vcs-type="git"
LABEL description="NetObserv Operator is a Kubernetes / OpenShift operator for network observability."
EOF

echo "Updating fbc bundle file"


cat <<EOF >>"${BUNDLE_CONTAINER_FILE}"
LABEL com.redhat.component="network-observability-operator-bundle-container"
LABEL name="network-observability-operator-bundle"
LABEL io.k8s.display-name="Network Observability Operator Bundle"
LABEL io.k8s.description="Network Observability Operator Bundle"
LABEL summary="Network Observability Operator Bundle"
LABEL maintainer="support@redhat.com"
LABEL io.openshift.tags="network-observability-operator-bundle"
LABEL upstream-vcs-ref="${COMMIT}"
LABEL upstream-vcs-type="git"
LABEL description="NetObserv Operator is a Kubernetes / OpenShift operator for network observability."
EOF

echo "Updating catalog container file"

cat <<EOF >>"${CATALOG_CONTAINER_FILE}"
LABEL com.redhat.component="network-observability-operator-catalog-container"
LABEL name="network-observability-operator-catalog"
LABEL io.k8s.display-name="Network Observability Operator Catalog"
LABEL io.k8s.description="Network Observability Operator Catalog"
LABEL summary="Network Observability Operator Catalog"
LABEL maintainer="support@redhat.com"
LABEL io.openshift.tags="network-observability-operator-catalog"
LABEL upstream-vcs-ref="${COMMIT}"
LABEL upstream-vcs-type="git"
LABEL description="NetObserv Operator is a Kubernetes / OpenShift operator for network observability."
EOF


[ ! -f "${crd_file}" ] && { echo "CustomResourceDefinition file not found, the version or name might have changed on us!"; exit 5; }

sed -i 's/\<NetObserv\>/network observability/g' "${crd_file}"

export EPOC_TIMESTAMP=$(date +%s)
export IN_CSV_DESC="./config/descriptions/ocp.md"

REPLACES="${REPLACE_VERSION}" VERSION="${TARGET_VERSION}" TARGET_CSV_FILE="${csv_file}" python3 ./hack/patch_csv.py
REPLACES="${REPLACE_VERSION}" VERSION="${TARGET_VERSION}" TARGET_INDEX_FILE="${index_file}" python3 ./hack/patch_catalog.py

sed -i 's/operators.operatorframework.io.bundle.channels.v1: latest,community/operators.operatorframework.io.bundle.channels.v1: stable/g' ./bundle/metadata/annotations.yaml
sed -i 's/operators.operatorframework.io.bundle.channel.default.v1: community/operators.operatorframework.io.bundle.channel.default.v1: stable/g' ./bundle/metadata/annotations.yaml

#Using downstream base image
sed -i 's/\(FROM.*\)docker.io\/library\/golang:1.22\(.*\)/\1brew.registry.redhat.io\/rh-osbs\/openshift-golang-builder:v1.22.5-202407301806.g4c8b32d.el9\2/g' ./Dockerfile
