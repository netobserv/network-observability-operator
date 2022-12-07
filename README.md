# NetObserv Operator

NetObserv Operator is a Kubernetes / OpenShift operator for network observability. It deploys a monitoring pipeline to collect and enrich network flows. These flows can be produced by the NetObserv eBPF agent, or by any device or CNI able to export flows in IPFIX format, such as OVN-Kubernetes.

The operator provides dashboards, metrics, and keeps flows accessible in a queryable log store, Grafana Loki. When used in OpenShift, new views are available in the Console.

## Getting Started

You can install NetObserv Operator using [OLM](https://olm.operatorframework.io/) if it is available in your cluster, or directly from its repository.

### Install with OLM

NetObserv Operator is available in [OperatorHub](https://operatorhub.io/operator/netobserv-operator) with guided steps on how to install this. It is also available in the OperatorHub catalog directly in the OpenShift Console.

![OpenShift OperatorHub search](./docs/assets/operatorhub-search.png)

Please read the [operator description in OLM](./config/manifests/bases/description-upstream.md). You will need to install Loki, some instructions are provided there.

After the operator is installed, create a `FlowCollector` resource:

![OpenShift OperatorHub FlowCollector](./docs/assets/operatorhub-flowcollector.png)

Refer to the [Configuration section](#configuration) of this document.

### Install from repository

A couple of `make` targets are provided in this repository to allow installing without OLM:

```bash
git clone https://github.com/netobserv/network-observability-operator.git && cd network-observability-operator
make deploy deploy-loki deploy-grafana
```

It will deploy the operator in its latest version, with port-forwarded Loki and Grafana.

> Note: the `loki-deploy` script is provided as a quick install path and is not suitable for production. It deploys a single pod, configures a 1GB storage PVC, with 24 hours of retention. For a scalable deployment, please refer to [our distributed Loki guide](https://github.com/netobserv/documents/blob/main/loki_distributed.md) or [Grafana's official documentation](https://grafana.com/docs/loki/latest/).

To deploy the monitoring pipeline, this `make` target installs a `FlowCollector` with default values:

```bash
make deploy-sample-cr
```

Alternatively, you can [grab and edit](./config/samples/flows_v1alpha1_flowcollector.yaml) this config before installing it.

You can still edit the `FlowCollector` after it's installed: the operator will take care about reconciling everything with the updated configuration:

```bash
kubectl edit flowcollector cluster
```

Refer to the [Configuration section](#configuration) of this document.

#### Install older versions

To deploy a specific version of the operator, you need to switch to the related git branch, then add a `VERSION` env to the above make command, e.g:

```bash
git checkout 0.1.2
VERSION=0.1.2 make deploy deploy-loki deploy-grafana
kubectl apply -f ./config/samples/flows_v1alpha1_flowcollector_versioned.yaml
```

Beware that the version of the underlying components, such as flowlogs-pipeline, may be tied to the version of the operator (this is why we recommend switching the git branch). Breaking this correlation may result in crashes. The versions of the underlying components are defined in the `FlowCollector` resource as image tags.

### OpenShift Console

_Pre-requisite: OpenShift 4.10 or above_

If the OpenShift Console is detected in the cluster, a console plugin is deployed when a `FlowCollector` is installed. It adds new pages and tabs to the console:

#### Overview metrics

Charts on this page show overall, aggregated metrics on the cluster network. The stats can be refined with comprehensive filtering and display options. Different levels of aggregations are available: per node, per namespace, per owner or per pod/service). For instance, it allows to identify biggest talkers in different contexts: top X inter-namespace flows, or top X pod-to-pod flows within a namespace, etc.

The watched time interval can be adjusted, as well as the refresh frequency, hence you can get an almost live view on the cluster traffic. This also applies to the other pages described below.

![Overview](./docs/assets/overview-dashboard.png)

#### Topology

The topology view represents traffic between elements as a graph. The same filtering and aggregation options as described above are available, plus extra display options e.g. to group element by node, namespaces, etc. A side panel provides contextual information and metrics related to the selected element.

![Topology](./docs/assets/topology-main.png)
_This screenshot shows the NetObserv architecture itself: Nodes (via eBPF agents) sending traffic (flows) to the collector flowlogs-pipeline, which in turn sends data to Loki. The NetObserv console plugin fetches these flows from Loki._

#### Flow table

The table view shows raw flows, ie. non aggregated, still with the same filtering options, and configurable columns.

![Flow table](./docs/assets/network-traffic-main.png)

#### Integration with existing console views

These views are accessible directly from the main menu, and also as contextual tabs for any Pod, Deployment, Service (etc.) in their details page, with filters set to focus on that particular resource.

![Contextual topology](./docs/assets/topology-pod.png)


### Grafana

Grafana can be used to retrieve and show the collected flows from Loki. If you used the `make` commands provided above to install NetObserv from the repository, you should already have Grafana installed and configured with Loki data source. Otherwise, you can install Grafana by following the instructions [here](https://github.com/netobserv/documents/blob/main/hack_loki.md#grafana), and add a new Loki data source that matches your setup. If you used the provided quick install path for Loki, its access URL is `http://loki:3100`.

To get dashboards, import [this file](./config/samples/dashboards/Network%20Observability.json) into Grafana. It includes a table of the flows and some graphs showing the volumetry per source or destination namespaces or workload:

![Grafana dashboard](./docs/assets/netobserv-grafana-dashboard.png)

## Configuration

The `FlowCollector` resource is used to configure the operator and its managed components. A comprehensive documentation is [available here](./docs/FlowCollector.md), and a full sample file [there](./config/samples/flows_v1alpha1_flowcollector.yaml).

To edit configuration in cluster, run:

```bash
kubectl edit flowcollector cluster
```

As it operates cluster-wide, only a single `FlowCollector` is allowed, and it has to be named `cluster`.

A couple of settings deserve special attention:

- Agent (`spec.agent.type`) can be `EBPF` (default) or `IPFIX`. eBPF is recommended, as it should work in more situations and offers better performances. If you can't, or don't want to use eBPF, note that the IPFIX option is fully functional only when using [OVN-Kubernetes](https://github.com/ovn-org/ovn-kubernetes/) CNI. Other CNIs are not officially supported, but you may still be able to configure them manually if they allow IPFIX exports.

- Sampling (`spec.agent.ebpf.sampling` and `spec.agent.ipfix.sampling`): a value of `100` means: one flow every 100 is sampled. `1` means all flows are sampled. The lower it is, the more flows you get, and the more accurate are derived metrics, but the higher amount of resources are consumed. By default, sampling is set to 50 (ie. 1:50) for eBPF and 400 (1:400) for IPFIX. Note that more sampled flows also means more storage needed. We recommend to start with default values and refine empirically, to figure out which setting your cluster can manage.

- Loki (`spec.loki`): configure here how to reach Loki. The default URL values match the Loki quick install paths mentioned in the _Getting Started_ section, but you may have to configure differently if you used another installation method. You will find more information in our guides for deploying Loki: [with Loki Operator](https://github.com/netobserv/documents/blob/main/loki_operator.md), or our alternative ["distributed Loki" guide](https://github.com/netobserv/documents/blob/main/loki_distributed.md).

- Quick filters (`spec.consolePlugin.quickFilters`): configure preset filters to be displayed in the Console plugin. They offer a way to quickly switch from filters to others, such as showing / hiding pods network, or infrastructure network, or application network, etc. They can be tuned to reflect the different workloads running on your cluster. For a list of available filters, [check this page](./docs/QuickFilters.md).

- Kafka (`spec.deploymentModel: KAFKA` and `spec.kafka`): when enabled, integrates the flow collection pipeline with Kafka, by splitting ingestion from transformation (kube enrichment, derived metrics, ...). Kafka can provide better scalability, resiliency and high availability. It's also an option to consider when you have a bursty traffic. [This page](https://www.redhat.com/en/topics/integration/what-is-apache-kafka) provides some guidance on why to use Kafka. When configured to use Kafka, NetObserv operator assumes it is already deployed and a topic is created. For convenience, we provide a quick deployment using [Strimzi](https://strimzi.io/): run `make deploy-kafka` from the repository.

- Exporters (`spec.exporters`, _experimental_) an optional list of exporters to which to send enriched flows. Currently only KAFKA is supported. This allows you to define any custom storage or processing that can read from Kafka. This feature is flagged as _experimental_ as it has not been thoroughly or stress tested yet, so use at your own risk.

### Performance fine-tuning

In addition to sampling and using Kafka or not, other settings can help you get an optimal setup without compromising on the observability.

Here is what you should pay attention to:

- Resource requirements and limits (`spec.agent.ebpf.resources`, `spec.agent.processor.resources`): adapt the resource requirements and limits to the load and memory usage you expect on your cluster. The default limits (800MB) should be sufficient for most medium sized clusters. You can read more about reqs and limits [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

- eBPF agent's cache max flows (`spec.agent.ebpf.cacheMaxFlows`) and timeout (`spec.agent.ebpf.cacheActiveTimeout`) control how often flows are reported by the agents. The higher are `cacheMaxFlows` and `cacheActiveTimeout`, the less traffic will be generated by the agents themselves, which also ties with less CPU load. But on the flip side, it leads to a slightly higher memory consumption, and might generate more latency in the flow collection.

- It is possible to reduce the overall observed traffic by restricting or excluding interfaces via `spec.agent.ebpf.interfaces` and `spec.agent.ebpf.excludeInterfaces`. Note that the interface names may vary according to the CNI used.

- The eBPF agent offers more advanced settings via [environment variables](https://github.com/netobserv/netobserv-ebpf-agent/blob/main/docs/config.md) that you can set through `spec.agent.ebpf.env`.

#### Loki

The `FlowCollector` resource includes configuration of the Loki client, which is used by the processor (`flowlogs-pipeline`) to connect and send data to Loki for storage. They impact two things: batches and retries.

- `spec.loki.batchWait` and `spec.loki.batchSize` control the batching mechanism, ie. how often data is flushed out to Loki. Like in the eBPF agent batching, higher values will generate fewer traffic and consume less CPU, however it will increase a bit the memory consumption of `flowlogs-pipeline`, and may increase a bit collection latency.

- `spec.loki.minBackoff`, `spec.loki.maxBackoff` and `spec.loki.maxRetries` control the retry mechanism. Retries may happen when Loki is unreachable or when it returns errors. Often, it is due to the rate limits configured on Loki server. When such situation occurs, it might not always be the best solution to increase rate limits (on server configuration side) or to increase retries. Increasing rate limits will put more pressure on Loki, so expect more memory and CPU usage, and also more traffic. Increasing retries will put more pressure on `flowlogs-pipeline`, as it will retain data for longer and accumulate more flows to send. When all the retry attempts fail, flows are simply dropped. Flow drops are counted in the metric `netobserv_loki_dropped_entries_total`.

On the Loki server side, configuration differs depending on how Loki was installed, e.g. via Helm chart, Loki Operator, etc. Nevertheless, here are a couple of settings that may impact the flow processing pipeline:

- Rate limits ([cf Loki documentation](https://grafana.com/docs/loki/latest/configuration/#limits_config)), especially ingestion rate limit, ingestion burst size, per-stream rate limit and burst size. When these rate limits are reached, Loki returns an error when `flowlogs-pipeline` tries to send batches, visible in logs. A good practice is to define an alert, to get notified when these limits are reached: [cf this example](https://github.com/netobserv/documents/blob/main/examples/distributed-loki/alerting/loki-ratelimit-alert.yaml). It uses a metrics provided by the Loki operator: `loki_request_duration_seconds_count`. In case you don't use the Loki operator, you can replace it by the same metric provided by NetObserv Loki client, named `netobserv_loki_request_duration_seconds_count`.

- Max active streams / max streams per user: this limit is reached when too many streams are created. In Loki terminology, a stream is a given set of labels (keys and values) used for indexing. NetObserv defines labels for source and destination namespaces and pod owners (ie. aggregated workloads, such as Deployments). So the more workloads are running and generating traffic on the cluster, the more chances there are to hit this limit, when it's set. We recommend setting a high limit or turning it off (0 stands for unlimited).

#### With Kafka

More performance fine-tuning is possible when using Kafka, ie. with `spec.deploymentModel` set to `KAFKA`:

- You can set the size of the batches (in bytes) sent by the eBPF agent to Kafka, with `spec.agent.ebpf.kafkaBatchSize`. It has a similar impact than `cacheMaxFlows` mentioned above, with higher values generating less traffic and less CPU usage, but more memory consumption and more latency. It is recommended to keep these two settings somewhat aligned (ie. do not set a super low `cacheMaxFlows` with high `kafkaBatchSize`, or the other way around). We expect the default values to be a good fit for most environments.

- If you find that the Kafka consumer might be a bottleneck, you can increase the number of replicas with `spec.processor.kafkaConsumerReplicas`, or set up an horizontal autoscaler with `spec.processor.kafkaConsumerAutoscaler`.

- Other advanced settings for Kafka include `spec.processor.kafkaConsumerQueueCapacity`, that defines the capacity of the internal message queue used in the Kafka consumer client, and `spec.processor.kafkaConsumerBatchSize`, which indicates to the broker the maximum batch size, in bytes, that the consumer will read.


### Securing data and communications

NetObserv is currently meant to be used by cluster / network admins. There is not yet multi-tenancy implemented, which means a user having access to the data will be able to see *all* the data. It is up to the cluster admins to restrict the access to the NetObserv and Loki namespaces and services according to their security policy.

By default, communications between internal components are not secured. It is possible to set up TLS for several communications:

- Between processor (`flowlogs-pipeline`) and Loki, by setting `spec.loki.tls`.
- With Kafka (both on producer and consumer sides), by setting `spec.kafka.tls`. Mutual TLS is supported here.
- The metrics server running in the processor (`flowlogs-pipeline`) can listen using TLS, via `spec.processor.metrics.server.tls`.
- The Console plugin server always uses TLS.

To make authenticated queries to Loki, when using the Loki Operator, there are two options:

- per-user auth: set `spec.loki.authToken` to `FORWARD` in the `FlowCollector` resource. The console plugin will forward the logged-in Console user token to Loki. To get access to the flow logs, users must have the following permission:

```yaml
- apiGroups:
  - 'loki.grafana.com'
  resources:
  - network
  resourceNames:
  - logs
  verbs:
  - 'get'
```

- per-app auth, ie. with NetObserv components having access to the logs using their own service account, regardless of the user permissions. For this mode, set `spec.loki.authToken` to `HOST` and setup `ClusterRole` for `flowlogs-pipeline` and `netobserv-plugin`, e.g: [role.yaml](https://github.com/netobserv/documents/blob/main/examples/loki-stack/role.yaml).


## Development & building from sources

Please refer to [this documentation](./DEVELOPMENT.md) for everything related to building, deploying or bundling from sources.

## F.A.Q / Troubleshooting

If you can't find help here, don't hesitate to open [an issue](https://github.com/netobserv/network-observability-operator/issues) or a [Q&A](https://github.com/netobserv/network-observability-operator/discussions/categories/q-a). There are several repositories under _netobserv_ github org, but it is fine to centralize these in _network-observability-operator_.

### Is it for OpenShift only?

No! While some features are developed primarily for OpenShift, we want to keep it on track with other / "vanilla" Kubes. For instance, there has been some work to make the console plugin [run as a standalone](https://github.com/netobserv/network-observability-console-plugin/pull/163), or the operator to manage upstream (non-OpenShift) [ovn-kubernetes](https://github.com/netobserv/network-observability-operator/pull/97).

And if something is not working as hoped with your setup, you are welcome to contribute to the project ;-)

### Which version of Kubernetes / OpenShift is supported?

It depends on which `agent` you want to use: `ebpf` or `ipfix`, and whether you want to get the OpenShift Console plugin.

#### To run the eBPF agent

What matters is the version of the Linux kernel: 4.18 or more is supported. Earlier versions are not tested.

Other than that, there are no known restrictions on the Kubernetes version.

#### To use IPFIX exports

OpenShift 4.10 or above, or upstream OVN-Kubernetes, are recommended, as the operator will configure OVS for you.

For other CNIs, you need to find out if they can export IPFIX, and configure them accordingly.

#### To get the OpenShift Console plugin

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

## Contributions

This project is licensed under [Apache 2.0](./LICENSE) and accepts contributions via GitHub pull requests. Other related `netobserv` projects follow the same rules:
- [Flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline)
- [eBPF agent](https://github.com/netobserv/netobserv-ebpf-agent)
- [OpenShift Console plugin](https://github.com/netobserv/network-observability-console-plugin)

External contributions are welcome and can take various forms:

- Providing feedback, by starting [discussions](https://github.com/netobserv/network-observability-operator/discussions) or opening [issues](https://github.com/netobserv/network-observability-operator/issues).
- Code / doc contributions. You will [find here](./DEVELOPMENT.md) some help on how to build, run and test your code changes. Don't hesitate to [ask for help](https://github.com/netobserv/network-observability-operator/discussions/categories/q-a).
