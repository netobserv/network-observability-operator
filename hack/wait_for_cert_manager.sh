#!/bin/bash

NAMESPACE=${NAMESPACE:-"cert-manager"}

ATTEMPTS=0
MAX_ATTEMPTS=60
cert_manager_ready=false
until $cert_manager_ready || [ $ATTEMPTS -eq $MAX_ATTEMPTS ]
do
    echo "waiting for cert-manager to be ready attempt:${ATTEMPTS}"
    kubectl apply -f hack/self_signed_cert.yaml
    if [[ $? != 0 ]]; then
        echo "failed, retrying"
        sleep 5
    else
        echo "cert-manager is ready"
        cert_manager_ready=true
        kubectl delete -f hack/self_signed_cert.yaml
    fi
    (( ATTEMPTS++ ))
done

if ! $cert_manager_ready; then
    echo "Timed out waiting for cert-manage to be ready"
    exit 1
fi
