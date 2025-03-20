# Helper tool to refresh the FBC already published from the prod red hat catalog
# Pre-requisites: opm, make

#!/bin/bash

set -euo pipefail

dir_catalog_tmp=_tmp/catalog
dir_catalog_legacy_tmp=_tmp/catalog-legacy

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

mkdir -p ${dir_catalog_tmp}
mkdir -p ${dir_catalog_legacy_tmp}

# echo "Fetching catalog from 4.18"
${OPM} migrate "registry.redhat.io/redhat/redhat-operator-index:v4.18" --migrate-level=bundle-object-to-csv-metadata ${dir_catalog_tmp} -oyaml

# echo "Fetching legacy catalog from 4.16"
${OPM} migrate "registry.redhat.io/redhat/redhat-operator-index:v4.16" ${dir_catalog_legacy_tmp} -oyaml

# echo "Extracting netobserv info"
${YQ} 'select(.schema == "olm.channel")' "${dir_catalog_tmp}/netobserv-operator/catalog.yaml" > catalog/released/index.yaml
${YQ} 'select(.schema == "olm.bundle")' "${dir_catalog_tmp}/netobserv-operator/catalog.yaml" > catalog/released/bundles.yaml
${YQ} 'select(.schema != "olm.channel") | select(.schema != "olm.bundle")' "${dir_catalog_tmp}/netobserv-operator/catalog.yaml" > catalog/released/other.yaml
${YQ} 'select(.schema == "olm.channel")' "${dir_catalog_legacy_tmp}/netobserv-operator/catalog.yaml" > catalog/released-legacy/index.yaml
${YQ} 'select(.schema == "olm.bundle")' "${dir_catalog_legacy_tmp}/netobserv-operator/catalog.yaml" > catalog/released-legacy/bundles.yaml
${YQ} 'select(.schema != "olm.channel") | select(.schema != "olm.bundle")' "${dir_catalog_legacy_tmp}/netobserv-operator/catalog.yaml" > catalog/released-legacy/other.yaml

echo "Validating..."

${OPM} validate catalog/released
if [ $? -ne 0 ]; then
  echo "Validation failed for catalog"
  exit 1
else
  echo "Validation passed for catalog"
fi

echo "Validating legacy..."

${OPM} validate catalog/released-legacy
if [ $? -ne 0 ]; then
  echo "Validation failed for catalog-legacy"
  exit 1
else
  echo "Validation passed for catalog-legacy"
fi

echo "Finished running $(basename "$0")"
