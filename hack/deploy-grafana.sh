#!/bin/bash

namespace=${1-netobserv}

create_grafana_deployment() {
  echo "--> Creating grafana deployment (ConfigMap, Pod and Service) "
  kubectl create -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  labels:
    app: grafana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
        - name: grafana
          image: grafana/grafana:latest
          ports:
            - containerPort: 3000
          imagePullPolicy: Always
          volumeMounts:
            - name: datasources
              mountPath: "/etc/grafana/provisioning/datasources/"
      volumes:
        - name: datasources
          configMap:
            name: grafana-datasources
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  labels:
    app: grafana
spec:
  ports:
    - port: 3000
      targetPort: 3000
      name: ui
  selector:
    app: grafana
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  labels:
    name: grafana-datasources
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - access: proxy
      isDefault: true
      name: loki
      type: loki
      url: http://loki.netobserv.svc.cluster.local:3100
      version: 1
---
EOF
  echo -e "\nWaiting for Grafana pod to be ready.\n"
  kubectl wait --timeout=180s --for=condition=ready pod -l app=grafana
  kubectl get pod -l app=grafana
}

cleanup() {
  echo "--> Cleaning up old deployment"
  kubectl delete --ignore-not-found=true route grafana
  kubectl delete --ignore-not-found=true service grafana
  kubectl delete --ignore-not-found=true deployment grafana
  kubectl delete --ignore-not-found=true configMap grafana-datasources
}

main() {
  echo -e "\n====> Deploying Grafana into namespace $namespace\n"
  kubectl config set-context --current --namespace="$namespace"
  cleanup
  create_grafana_deployment
  echo -e "\n====> Done."
}

main




