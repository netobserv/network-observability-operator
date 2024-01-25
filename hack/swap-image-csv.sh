#!/bin/bash

csv_name=$1
namespace=$2
rel_name=$3
env_name=$4
new_image=$5

if [[ ! -z $rel_name ]]; then
  idx_rel=$(oc get csv $csv_name -n $namespace -o json | jq --arg NAME "$rel_name" '.spec.relatedImages | map(.name == $NAME) | index(true)')
  patch1="{'op': 'replace', 'path': '/spec/relatedImages/$idx_rel', 'value': {'name': '$rel_name', 'image': '$new_image'}}"
fi

idx_env=$(oc get csv $csv_name -n $namespace -o json | jq --arg NAME "$env_name" '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env | map(.name == $NAME) | index(true)')
patch2="{'op': 'replace', 'path': '/spec/install/spec/deployments/0/spec/template/spec/containers/0/env/$idx_env', 'value': {'name': '$env_name', 'value': '$new_image'}}"

if [[ ! -z ${patch1+x} ]] ; then
  oc patch csv $csv_name -n $namespace --type='json' -p "[$patch1, $patch2]"
else
  oc patch csv $csv_name -n $namespace --type='json' -p "[$patch2]"
fi

if [ $? -eq 0 ]; then
  echo "Patch succeeded"
else
  echo "Patch failed!"
  exit 1
fi
