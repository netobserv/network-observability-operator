
#!/usr/bin/env bash

echo "Updating container file"

# : "${COMMIT:=$(git rev-list --abbrev-commit --tags --max-count=1)}"
# : "${CONTAINER_FILE:=./Dockerfile}"
# : "${BUNDLE_CONTAINER_FILE:=./bundle.Dockerfile}"
# : "${CATALOG_CONTAINER_FILE:=./catalog.Dockerfile}"
: "${TARGET_VERSION:=1.9.0}"
: "${REPLACE_VERSION:=1.8.0}"

# supported_ocp_versions="v4.13"
manifests_dir="./bundle/manifests"
# metadata_dir="./bundle/metadata"
crd_name="flows.netobserv.io_flowcollectors.yaml"
crd_file="${manifests_dir}/${crd_name}"
csv_name="netobserv-operator.clusterserviceversion.yaml"
csv_file="${manifests_dir}/${csv_name}"
new_bundle_file="./catalog/rc.yaml"

source ./hack/nudging/container_digest.sh
source ./hack/nudging/bundle_digest.sh

[ ! -f "${crd_file}" ] && { echo "CustomResourceDefinition file not found, the version or name might have changed on us!"; exit 5; }

sed -i 's/\<NetObserv\>/network observability/g' "${crd_file}"

export EPOC_TIMESTAMP=$(date +%s)
export IN_CSV_DESC="./config/descriptions/ocp.md"

REPLACES="${REPLACE_VERSION}" VERSION="${TARGET_VERSION}" TARGET_CSV_FILE="${csv_file}" python3 ./hack/patch_csv.py
NEW_BUNDLE_FILE="${new_bundle_file}" python3 ./hack/patch_catalog.py

sed -i 's/operators.operatorframework.io.bundle.channels.v1: latest,community/operators.operatorframework.io.bundle.channels.v1: stable/g' ./bundle/metadata/annotations.yaml
sed -i 's/operators.operatorframework.io.bundle.channel.default.v1: community/operators.operatorframework.io.bundle.channel.default.v1: stable/g' ./bundle/metadata/annotations.yaml

#Using downstream base image
echo "Container file updated"
