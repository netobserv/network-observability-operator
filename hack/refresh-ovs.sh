#!/bin/bash

command -v yq &> /dev/null
if [ "$?" != "0" ]; then
  echo "yq not found, it can be downloaded at https://github.com/mikefarah/yq "
  echo "Exiting."
  exit 1
fi

ovnns=openshift-ovn-kubernetes

ovspods=`kubectl get pods -n $ovnns -l app=ovnkube-node --no-headers -o custom-columns=":metadata.name"`

cacheActiveTimeout=`kubectl get flowcollector cluster -o yaml | yq -e .spec.agent.ipfix.cacheActiveTimeout`
cacheMaxFlows=`kubectl get flowcollector cluster -o yaml | yq -e .spec.agent.ipfix.cacheMaxFlows`
sampling=`kubectl get flowcollector cluster -o yaml | yq -e .spec.agent.ipfix.sampling`
config="cache_active_timeout=${cacheActiveTimeout::-1} cache_max_flows=$cacheMaxFlows sampling=$sampling"

echo "Storing config: $config"

for pod in $ovspods; do
  echo "Found OVN pod: $pod"
  ip=`kubectl exec -n $ovnns $pod -c ovn-controller -- ovs-vsctl list IPFIX | grep targets | sed -r 's/.*"([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+)".*/\1/'`
  echo "Found target IP: $ip"
  echo "Clearing config"
  kubectl exec -n $ovnns $pod -c ovn-controller -- ovs-vsctl clear Bridge br-int ipfix
  echo "Sleep 1s"
  sleep 1
  echo "Reconfiguring"
  kubectl exec -n $ovnns $pod -c ovn-controller -- ovs-vsctl -- --id=@ipfix create ipfix targets=[\"$ip\"] $config -- set bridge br-int ipfix=@ipfix
done
