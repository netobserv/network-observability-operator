#!/bin/bash

enable_flp_metrics() {
  echo "--> Enabling flp metrics "
  kubectl create -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-monitoring-config
  namespace: openshift-monitoring
data:
  config.yaml: |
    enableUserWorkload: true
EOF
}

main() {
  echo -e "\n====> Enabling flp metrics in console \n"
  enable_flp_metrics
  echo -e "\n====> Done."
}

main




