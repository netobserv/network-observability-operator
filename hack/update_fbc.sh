# Helper tool to update the catalog artifacts and bundle artifacts
# These genereated artifacts are used to build the catalog image
# Pre-requisites: opm, make
# Usage: ./hack/update_fbc.sh

#!/bin/bash

set -euo pipefail

: ${IS_DOWNSTREAM:="false"}

: ${OPM:=$(command -v opm)}
echo "using opm from ${OPM}"
# check if opm version is v1.39.0 or exit
if [ -z "${OPM}" ]; then
  echo "opm is required"
  exit 1
fi

: ${YQ:=$(command -v yq)}
echo "using yq from ${YQ}"
# check if yq exists
if [ -z "${YQ}" ]; then
  echo "yq is required"
  exit 1
fi

echo "Creating new bundle using image ${BUNDLE_IMAGE}..."

dir_catalog="catalog/unreleased/${BUNDLE_TAG}"
dir_catalog_legacy="catalog/unreleased-legacy/${BUNDLE_TAG}"
mkdir -p "${dir_catalog}"
mkdir -p "${dir_catalog_legacy}"
cp -f catalog/released/other.yaml ${dir_catalog}
cp -f catalog/released-legacy/other.yaml ${dir_catalog_legacy}

${OPM} render "${BUNDLE_IMAGE}" --output=yaml --migrate-level=bundle-object-to-csv-metadata > "${dir_catalog}/bundle.yaml"
${OPM} render "${BUNDLE_IMAGE}" --output=yaml > "${dir_catalog_legacy}/bundle.yaml"

if [[ "${IS_DOWNSTREAM}" == "true" ]]; then
  echo "Adding to index..."
  previous=$(${YQ} 'select(.schema=="olm.channel") | select(.name=="stable") | .entries[-1].name' catalog/released/index.yaml)
  ${YQ} "(select(.name == \"stable\") | .entries) += {\"name\": \"network-observability-operator.${BUNDLE_TAG}\", \"replaces\": \"${previous}\"}" catalog/released/index.yaml >  "${dir_catalog}/index.yaml"
  echo "Fixing bundle name..."
  sed -i 's/name: netobserv-operator/name: network-observability-operator/' ${dir_catalog}/bundle.yaml
  sed -i 's/name: netobserv-operator/name: network-observability-operator/' ${dir_catalog_legacy}/bundle.yaml
else
  echo "Generating single index..."
  cat <<EOF > "${dir_catalog}/index.yaml"
---
entries:
  - name: netobserv-operator.${BUNDLE_TAG}
name: latest
package: netobserv-operator
schema: olm.channel
EOF
  sed -i 's/defaultChannel: stable/defaultChannel: latest/' ${dir_catalog}/other.yaml
  sed -i 's/defaultChannel: stable/defaultChannel: latest/' ${dir_catalog_legacy}/other.yaml
fi

cp -f "${dir_catalog}/index.yaml" "${dir_catalog_legacy}"

echo "Finished running $(basename "$0")"
