#!/usr/bin/env bash

namespaced_res="daemonsets,deployments,horizontalpodautoscalers,configmaps,secrets,serviceaccounts,services,pods,services,prometheusrules,servicemonitors"

kubectl get ${namespaced_res} --show-kind --ignore-not-found -n netobserv
echo ""
kubectl get ${namespaced_res} --show-kind --ignore-not-found -n netobserv-privileged
echo ""
echo "CLUSTER ROLES AND BINDINGS"
kubectl get clusterroles,clusterrolebindings --show-kind --ignore-not-found | grep netobserv
kubectl get clusterroles,clusterrolebindings --show-kind --ignore-not-found | grep flowlogs
