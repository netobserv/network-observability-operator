#!/bin/bash

# NOTE: this release script is provided for help purpose, but you should carefully read it and understand what it does before running it.
# Appropriate permissions will be required when pushing git tags / images
# Adapt the variables below to match your local files, desired version and remote target.
# Part of this workflow will be automated soon.

# Make sure also all repo have the correct HEAD & clean state

path_noo="../network-observability-operator"
version="0.1.1-rc0"

user=netobserv
remote=upstream

vv=v$version

VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make image-build
VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make bundle bundle-build
git commit -a -m "Prepare release $vv"

# Check everything is ok, then push upstream
# git push upstream HEAD:main
