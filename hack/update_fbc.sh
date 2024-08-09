# Helper tool to update the catalog artifacts and bundle artifacts
# These genereated artifacts are used to build the catalog image
# Pre-requisites: opm, make
# Usage: ./hack/update_bundle_catalog.sh

#!/bin/bash

set -euo pipefail

cleanup() {
  # remove temporary bundle file
  if [ -n "${TEMP_BUNDLE_FILE}" ]; then
    rm -f "${TEMP_BUNDLE_FILE}"
  fi

}

trap cleanup EXIT

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

: "${CATALOG_FILE:=catalog/index.yaml}"

cat <<EOF >"${CATALOG_FILE}"
---
defaultChannel: latest
icon:
  base64data: PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDI2LjAuMiwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IkxheWVyXzEiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiIHg9IjBweCIgeT0iMHB4IgoJIHZpZXdCb3g9IjAgMCAxMDAgMTAwIiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAxMDAgMTAwOyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+CjxzdHlsZSB0eXBlPSJ0ZXh0L2NzcyI+Cgkuc3Qwe2ZpbGw6dXJsKCNTVkdJRF8xXyk7fQoJLnN0MXtmaWxsOiNGRkZGRkY7fQoJLnN0MntvcGFjaXR5OjAuNjt9Cgkuc3Qze29wYWNpdHk6MC41O30KCS5zdDR7b3BhY2l0eTowLjQ7fQo8L3N0eWxlPgo8Zz4KCTxnPgoJCTxnPgoJCQk8cmFkaWFsR3JhZGllbnQgaWQ9IlNWR0lEXzFfIiBjeD0iMTQuNzc1OCIgY3k9Ii0yLjk3NzEiIHI9IjkxLjYyNyIgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiPgoJCQkJPHN0b3AgIG9mZnNldD0iMCIgc3R5bGU9InN0b3AtY29sb3I6IzNDM0ZBNiIvPgoJCQkJPHN0b3AgIG9mZnNldD0iMSIgc3R5bGU9InN0b3AtY29sb3I6IzNCMDM0MCIvPgoJCQk8L3JhZGlhbEdyYWRpZW50PgoJCQk8cGF0aCBjbGFzcz0ic3QwIiBkPSJNNTAsOTljLTEzLjMsMC0yNS40LTUuMy0zNC4yLTEzLjlDNi43LDc2LjIsMSw2My43LDEsNTBDMSwyMi45LDIyLjksMSw1MCwxYzEzLjcsMCwyNi4yLDUuNywzNS4xLDE0LjgKCQkJCUM5My43LDI0LjYsOTksMzYuNyw5OSw1MEM5OSw3Ny4xLDc3LjEsOTksNTAsOTl6Ii8+CgkJPC9nPgoJCTxnPgoJCQk8Y2lyY2xlIGNsYXNzPSJzdDEiIGN4PSIzNy41IiBjeT0iODEuOSIgcj0iNSIvPgoJCTwvZz4KCQk8cGF0aCBjbGFzcz0ic3QxIiBkPSJNNDguNiw5MS45bDE4LjgtNDMuM2MtMi41LTAuMS01LTAuNy03LjItMkwzMy4yLDY4LjJsMS40LTEuOGwyMC0yNS4xYy0xLjUtMi40LTIuMy01LjEtMi4zLTcuOUw5LDUyLjIKCQkJbDQ3LjYtMjkuOWwwLDBjMC4xLTAuMSwwLjItMC4yLDAuMi0wLjJjNi4xLTYuMSwxNS45LTYuMSwyMiwwbDAuMSwwLjFjNiw2LjEsNiwxNS45LTAuMSwyMS45Yy0wLjEsMC4xLTAuMiwwLjItMC4yLDAuMmwwLDAKCQkJTDQ4LjYsOTEuOXoiLz4KCQk8ZyBjbGFzcz0ic3QyIj4KCQkJPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iNTAuMyIgY3k9IjE0LjciIHI9IjMuMSIvPgoJCTwvZz4KCQk8ZyBjbGFzcz0ic3QzIj4KCQkJPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iMjcuNyIgY3k9IjU4IiByPSIxLjciLz4KCQk8L2c+CgkJPGc+CgkJCTxjaXJjbGUgY2xhc3M9InN0MSIgY3g9Ijc3LjQiIGN5PSI2OS4zIiByPSIxLjciLz4KCQk8L2c+CgkJPGc+CgkJCTxjaXJjbGUgY2xhc3M9InN0MSIgY3g9IjE2LjMiIGN5PSIzNi42IiByPSIxLjciLz4KCQk8L2c+CgkJPGcgY2xhc3M9InN0NCI+CgkJCTxjaXJjbGUgY2xhc3M9InN0MSIgY3g9IjYzLjciIGN5PSI4NS45IiByPSIyLjIiLz4KCQk8L2c+CgkJPGc+CgkJCTxjaXJjbGUgY2xhc3M9InN0MSIgY3g9IjI5LjQiIGN5PSIxOS42IiByPSI0LjgiLz4KCQk8L2c+CgkJPGcgY2xhc3M9InN0MyI+CgkJCTxjaXJjbGUgY2xhc3M9InN0MSIgY3g9Ijg4IiBjeT0iNTAiIHI9IjQuOCIvPgoJCTwvZz4KCTwvZz4KPC9nPgo8L3N2Zz4K
  mediatype: image/svg+xml
name: netobserv-operator
schema: olm.package
EOF


echo "Adding bundle image to FBC using image ${BUNDLE_IMAGE}"

# # create a temporary file for the bundle part of the catalog
TEMP_BUNDLE_FILE=$(mktemp)

${OPM} render "${BUNDLE_IMAGE}" --output=yaml >"${TEMP_BUNDLE_FILE}"

cat ${TEMP_BUNDLE_FILE} >>${CATALOG_FILE}

cat <<EOF >>"${CATALOG_FILE}"
---
schema: olm.channel
package: netobserv-operator
name: latest
entries:
  - name: netobserv-operator.${BUNDLE_TAG}
EOF

${OPM} validate $(dirname "${CATALOG_FILE}")
if [ $? -ne 0 ]; then
  echo "Validation failed for ${CATALOG_FILE}"
  exit 1
else
  echo "Validation passed for ${CATALOG_FILE}"
fi

echo "Finished running $(basename "$0")"
