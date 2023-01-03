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

cleanup() {
  echo "--> Cleaning up old deployment"
  kubectl delete --ignore-not-found=true configMap cluster-monitoring-config
}

main() {
  echo -e "\n====> Enabling flp metrics in console \n"
  cleanup
  enable_flp_metrics
  echo -e "\n====> Done."
}

main




