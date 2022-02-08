#!/bin/bash

# NOTE: this release script is provided for help purpose, but you should carefully read it and understand what it does before running it.
# Appropriate permissions will be required when pushing git tags / images
# Adapt the variables below to match your local files, desired version and remote target.
# Part of this workflow will be automated soon.

path_noo="../network-observability-operator"
path_gfk="../goflow2-kube-enricher"
path_plg="../network-observability-console-plugin"
path_hubs=("../community-operators" "../community-operators-okd")
version="0.1.0-rc0"
user=netobserv
remote=upstream

vv=v$version

cd $path_plg && \
 VERSION="$vv" USER="$user" make image push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags

cd $path_gfk && \
 VERSION="$vv" USER="$user" make image push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags

cd $path_noo && \
 VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make image-build image-push && \
 VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make bundle bundle-build bundle-push && \
 git tag -a "$vv" -m "$vv" && \
 git push $remote --tags

cd $path_noo
for hub in "${path_hubs[@]}"; do
  mkdir -p $hub/operators/netobserv-operator/$version && \
  cp "bundle.Dockerfile" "$hub/operators/netobserv-operator/$version" && \
  cp -r "bundle/manifests" "$hub/operators/netobserv-operator/$version" && \
  cp -r "bundle/metadata" "$hub/operators/netobserv-operator/$version"
done
for hub in "${path_hubs[@]}"; do
  cd $hub && \
  git add -A && \
  git commit -m "operators netobserv-operator ($version)" && \
  git push origin HEAD:bump-$version
done
