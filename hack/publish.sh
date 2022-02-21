#!/bin/bash

# NOTE: this publish script is provided for help purpose, but you should carefully read it and understand what it does before running it.
# Appropriate permissions will be required when pushing git tags / images
# Adapt the variables below to match your local files, desired version and remote target.
# Part of this workflow may be automated soon.

path_noo="../network-observability-operator"
path_hubs=("../community-operators" "../community-operators-okd")
version="0.1.0"

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
  git commit -s -m "operators netobserv-operator ($version)" && \
  git push origin HEAD:bump-$version
done
