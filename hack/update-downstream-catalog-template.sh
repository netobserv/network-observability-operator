
#!/usr/bin/env bash

echo "Updating downstream catalog template"

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

mkdir -p _tmp
rm -rf _tmp/bundle
cp -r bundle _tmp/

new_bundle_file="./catalog/unreleased/downstream-test-fbc/bundle.yaml"

BUNDLE_PATH=_tmp/bundle ./hack/update-build.sh
# NOTE/HACK: --alpha-image-ref-template=placeholder allows to render a catalog in csv-metadata format without having the bundle image built in the first place
# There's no guarantee that the OLM team keep this flag and/or fix the rendering in csv-metadata format for this use case unfortunately
# If that had to happen, it's possible to fork opm and just remove `b.Image == ""` here https://github.com/operator-framework/operator-registry/blob/5e4172fdb25ac92ff498184a0ff16e5b3b782e6b/alpha/action/migrations/000_bundle_object_to_csv_metadata.go#L14
# The alternative being to maintain our catalog template up-to-date manually.
${OPM} render _tmp/bundle --output=yaml --migrate-level=bundle-object-to-csv-metadata --alpha-image-ref-template=placeholder > ${new_bundle_file}

# Remove information that will be provided at catalog build time
${YQ} -i '.image = ""' ${new_bundle_file}
${YQ} -i '.relatedImages = [{"image": "", "name": "console_plugin"}, {"image": "", "name": "ebpf_agent"}, {"image": "", "name": "flowlogs_pipeline"}, {"image": "", "name": "bundle"}, {"image": "", "name": "manager"}]' ${new_bundle_file}
${YQ} -i '(.properties.[] | select(.type=="olm.csv.metadata") | .value.annotations.containerImage) = ""' ${new_bundle_file}
${YQ} -i '(.properties.[] | select(.type=="olm.csv.metadata") | .value.annotations.createdAt) = ""' ${new_bundle_file}

echo "Downstream catalog template updated"
