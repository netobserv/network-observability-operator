# F.A.Q / Troubleshooting

If you can't find help here, don't hesitate to open [an issue](https://github.com/netobserv/network-observability-operator/issues) or a [Q&A](https://github.com/netobserv/network-observability-operator/discussions/categories/q-a). There are several repositories under _netobserv_ github org, but it is fine to centralize these in _network-observability-operator_.

## Table of Contents

* Q&A
  * [Is it for OpenShift only?](#is-it-for-openshift-only)
  * [Which version of Kubernetes / OpenShift is supported?](#which-version-of-kubernetes--openshift-is-supported)
* How-to
  * [To run the eBPF agent](#to-run-the-ebpf-agent)
  * [To use IPFIX exports](#to-use-ipfix-exports)
  * [To get the OpenShift Console plugin](#to-get-the-openshift-console-plugin)
  * [How can I make sure everything is correctly deployed?](#how-can-i-make-sure-everything-is-correctly-deployed)
* Troubleshooting
  * [Everything seems correctly deployed but there isn't any flow showing up](#everything-seems-correctly-deployed-but-there-isnt-any-flow-showing-up)
  * [There is no Network Traffic menu entry in OpenShift Console](#there-is-no-network-traffic-menu-entry-in-openshift-console)
  * [I first deployed flowcollector, and then kafka. Flowlogs-pipeline is not consuming any flow from Kafka](#i-first-deployed-flowcollector-and-then-kafka-flowlogs-pipeline-is-not-consuming-any-flow-from-kafka)
  * [I don't see flows from either the `br-int` or `br-ex` interfaces](#i-dont-see-flows-from-either-the-br-int-or-br-ex-interfaces)

## Q&A

### Is it for OpenShift only?

No! While some features are developed primarily for OpenShift, we want to keep it on track with other / "vanilla" Kubes. For instance, there has been some work to make the console plugin [run as a standalone](https://github.com/netobserv/network-observability-console-plugin/pull/163), or the operator to manage upstream (non-OpenShift) [ovn-kubernetes](https://github.com/netobserv/network-observability-operator/pull/97).

And if something is not working as hoped with your setup, you are welcome to contribute to the project ;-)

### Which version of Kubernetes / OpenShift is supported?

It depends on which `agent` you want to use: `ebpf` or `ipfix`, and whether you want to get the OpenShift Console plugin.

## How to

### To run the eBPF agent

What matters is the version of the Linux kernel: 4.18 or more is supported. Earlier versions are not tested.

Other than that, there are no known restrictions on the Kubernetes version.

### To use IPFIX exports

OpenShift 4.10 or above, or upstream OVN-Kubernetes, are recommended, as the operator will configure OVS for you.

For other CNIs, you need to find out if they can export IPFIX, and configure them accordingly.

### To get the OpenShift Console plugin

OpenShift 4.10 or above is required.

### How can I make sure everything is correctly deployed?

Make sure all pods are up and running:

```bash
# Assuming configured namespace is netobserv (default)
kubectl get pods -n netobserv
```

Should provide results similar to this:

```
NAME                                            READY   STATUS    RESTARTS   AGE
flowlogs-pipeline-5rrg2                         1/1     Running   0          43m
flowlogs-pipeline-cp2lb                         1/1     Running   0          43m
flowlogs-pipeline-hmwxd                         1/1     Running   0          43m
flowlogs-pipeline-wmx4z                         1/1     Running   0          43m
grafana-6dbddc9869-sxn62                        1/1     Running   0          31m
loki                                            1/1     Running   0          43m
netobserv-controller-manager-7487d87dc-2ltq2    2/2     Running   0          43m
netobserv-plugin-7fb8c5477b-drg2z               1/1     Running   0          43m
```

Results may slightly differ depending on the installation method and the `FlowCollector` configuration. At least you should see `flowlogs-pipeline` pods in a `Running` state.

If you use the eBPF agent, check also for pods in privileged namespace:

```bash
# Assuming configured namespace is netobserv (default)
kubectl get pods -n netobserv-privileged
```

```
NAME                         READY   STATUS    RESTARTS   AGE
netobserv-ebpf-agent-7rwtk   1/1     Running   0          7s
netobserv-ebpf-agent-c7nkv   1/1     Running   0          7s
netobserv-ebpf-agent-hbjz8   1/1     Running   0          7s
netobserv-ebpf-agent-ldj66   1/1     Running   0          7s

```

Finally, make sure Loki is correctly deployed, and reachable from pods via the URL defined in `spec.loki.url`. You can for instance check using this command:

```bash
kubectl exec $(kubectl get pod -l "app=flowlogs-pipeline" -o name)  -- curl  -G -s "`kubectl get flowcollector cluster -o=jsonpath={.spec.loki.url}`loki/api/v1/query" --data-urlencode 'query={app="netobserv-flowcollector"}' --data-urlencode 'limit=1'
```

It should return some json in this form:

```
{"status":"success","data":{"resultType":"streams","result":[...],"stats":{...}}}
```

## Troubleshooting

### Everything seems correctly deployed but there isn't any flow showing up

If using IPFIX (ie. `spec.agent.type` is `IPFIX` in FlowCollector), wait 10 minutes and check again. There is sometimes a delay, up to 10 minutes, before the flows appear. This is due to the IPFIX protocol requiring exporter and collector to exchange record template definitions as a preliminary step. The eBPF agent doesn't have such a delay.

Else, check for any suspicious error in logs, especially in the `flowlogs-pipeline` pods and the eBPF agent pods. You may also take a look at prometheus metrics prefixed with `netobserv_`: they can give you clues if flows are processed, if errors are reported, etc.

Finally, don't hesitate to [open an issue](https://github.com/netobserv/network-observability-operator/issues).

### There is no Network Traffic menu entry in OpenShift Console

Make sure your cluster version is at least OpenShift 4.10: prior versions have no (or incompatible) console plugin SDK.

Make sure that `spec.consolePlugin.register` is set to `true` (default).

If not, or if for any reason the registration seems to have failed, you can still do it manually by editing the Console Operator config:

```bash
kubectl edit console.operator.openshift.io cluster
```

If it's not already there, add the plugin reference:

```yaml
spec:
  plugins:
  - netobserv-plugin
```

If the new views still don't show up, try clearing your browser cache and refreshing. Check also the `netobserv-console-plugin-...` pod status and logs.

```bash
kubectl get pods -n netobserv -l app=netobserv-plugin
kubectl logs -n netobserv -l app=netobserv-plugin
```

It is also possible that the OpenShift Console is failing to restart. There is for instance a [known issue](https://issues.redhat.com/browse/OCPBUGS-810) with very small clusters, causing this problem.
```bash
kubectl get pods -n openshift-console
```

```
NAME                         READY   STATUS    RESTARTS   AGE
console-65486f987c-bk6zn     1/1     Running   0          44m
console-67fd7b6f49-l5nmq     0/1     Pending   0          10m
console-67fd7b6f49-wmhp4     0/1     Pending   0          10m
downloads-5fc6cc467f-2fttx   1/1     Running   0          45m
downloads-5fc6cc467f-2gfpw   1/1     Running   0          45m
```

To force a restart kill the running console pod, note that this will make the console anavailable for some time.

### I first deployed flowcollector, and then kafka. Flowlogs-pipeline is not consuming any flow from Kafka

This is a [known bug](https://github.com/segmentio/kafka-go/issues/1044) in one of flowlogs-pipeline dependencies.

Please recreate the flowlogs-pipeline pods by either killing them maunally or deleting and recreating the flow collector object.

### I don't see flows from either the `br-int` or `br-ex` interfaces

[`br-ex` and `br-int` are virtual bridge devices](https://access.redhat.com/documentation/en-us/red_hat_openstack_platform/16.0/html/networking_guide/bridge-mappings),
so they operate at OSI Layer 2 (e.g. Ethernet level). The eBPF agent works at Layers 3 and 4
(IP and TCP level), so it is expected that traffic passing through `br-int` and `br-ex` is captured
by the agent when it is processed by other interfaces (e.g. physical host or virtual pod interfaces).

This means that, if you restrict the agent interfaces (using the `interfaces` or `excludeInterfaces`
properties) to attach only to `br-int` and/or `br-ex`, you won't be able to see any flow.
