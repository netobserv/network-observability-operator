#!/usr/bin/env bash

mkdir -p _tmp

# Copy and edit CRDs
for crd in "flows.netobserv.io_flowcollectors.yaml" "flows.netobserv.io_flowmetrics.yaml"; do
  cp "bundle/manifests/$crd" helm/templates
  sed -i -r 's/(`[^`]*\{\{[^`]*`)/{{\1}}/g' helm/templates/$crd # escape "{{" for helm
  yq -i '.spec.conversion.webhook.clientConfig.service.namespace="{{ .Release.Namespace }}"' helm/templates/$crd
  yq -i '.metadata.annotations["cert-manager.io/inject-ca-from"]="{{ .Release.Namespace }}/serving-cert"' helm/templates/$crd
done

# Copy unchanged files
for file in "netobserv-manager-config_v1_configmap.yaml" "netobserv-metrics-service_v1_service.yaml" "netobserv-webhook-service_v1_service.yaml" ; do
  cp "bundle/manifests/$file" helm/templates
done

# Services: remove openshift annotations for certificates (and some kubeconfig labels)
yq -i 'del(.metadata.annotations)' helm/templates/netobserv-metrics-service_v1_service.yaml
yq -i 'del(.metadata.annotations)' helm/templates/netobserv-webhook-service_v1_service.yaml
yq -i 'del(.metadata.labels)' helm/templates/netobserv-webhook-service_v1_service.yaml

# Extract data from clusterserviceversion
yq '.spec.install.spec.deployments[0].spec' bundle/manifests/netobserv-operator.clusterserviceversion.yaml > _tmp/csv-deployment.yaml
yq '.spec.install.spec.clusterPermissions[0]' bundle/manifests/netobserv-operator.clusterserviceversion.yaml > _tmp/csv-clusterrole.yaml
yq '.spec.install.spec.permissions[0]' bundle/manifests/netobserv-operator.clusterserviceversion.yaml > _tmp/csv-role.yaml
 
# Create deployment
yq '{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": {"name": "netobserv-controller-manager", "labels": {"app": "netobserv-operator", "control-plane": "controller-manager"}}, "spec": .}' _tmp/csv-deployment.yaml > helm/templates/deployment.yaml

# Inject paramterized standalone console in deployment
PLUGIN_IMAGE=$(yq '(.spec.template.spec.containers[0].env[] | select(.name=="RELATED_IMAGE_CONSOLE_PLUGIN") | .value)' helm/templates/deployment.yaml)
STANDALONE_IMAGE=$(echo $PLUGIN_IMAGE | sed 's/network-observability-console-plugin/network-observability-standalone-frontend/')
yq -i "(.spec.template.spec.containers[0].env[] | select(.name==\"RELATED_IMAGE_CONSOLE_PLUGIN\") | .value) = \"{{ if .Values.standaloneConsole.enable }}$STANDALONE_IMAGE{{ else }}$PLUGIN_IMAGE{{ end }}\"" helm/templates/deployment.yaml

# Create roles
yq '{"apiVersion": "v1", "kind": "ServiceAccount", "metadata": {"name": .serviceAccountName}}' _tmp/csv-clusterrole.yaml > helm/templates/serviceaccount.yaml
yq '{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole", "metadata": {"name": "netobserv-manager-role"}, "rules": .rules}' _tmp/csv-clusterrole.yaml > helm/templates/clusterrole.yaml
yq '{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRoleBinding", "metadata": {"name": "netobserv-manager-rolebinding"}, "roleRef": {"apiGroup": "rbac.authorization.k8s.io", "kind": "ClusterRole", "name": "netobserv-manager-role"}, "subjects": [{"kind": "ServiceAccount", "name": .serviceAccountName, "namespace": "{{ .Release.Namespace }}"}]}' _tmp/csv-clusterrole.yaml > helm/templates/clusterrolebinding.yaml
yq '{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "Role", "metadata": {"name": "netobserv-leader-election-role"}, "rules": .rules}' _tmp/csv-role.yaml > helm/templates/role.yaml
yq '{"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "RoleBinding", "metadata": {"name": "netobserv-leader-election-rolebinding"}, "roleRef": {"apiGroup": "rbac.authorization.k8s.io", "kind": "Role", "name": "netobserv-leader-election-role"}, "subjects": [{"kind": "ServiceAccount", "name": .serviceAccountName, "namespace": "{{ .Release.Namespace }}"}]}' _tmp/csv-role.yaml > helm/templates/rolebinding.yaml

for f in bundle/manifests/*_rbac.authorization.k8s.io_v1_clusterrole.yaml; do cp "$f" helm/templates/ ; done
