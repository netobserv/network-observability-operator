Network Observability is an OpenShift operator that deploys a monitoring pipeline consisting in:
- an eBPF agent, that generates network flows from captured packets
- flowlogs-pipeline, a component that collects, enriches and exports these flows
- a Console plugin for flows visualization with powerful filtering options, a topology representation and more

Flow data is then available in multiple ways, each optional:

- As Cluster Monitoring metrics
- As raw flow logs stored in Grafana Loki
- As raw flow logs exported to a collector

## Dependencies

### Loki

[Loki](https://grafana.com/oss/loki/), from GrafanaLabs, can optionally be used as the backend to store all collected flows. The Network Observability operator does not install Loki directly, however we provide some guidance to help you there.

- For a production or production-like environment usage, refer to [the operator documentation](https://docs.openshift.com/container-platform/latest/network_observability/installing-operators.html).

- For a quick try that is not suitable for production and not scalable (it deploys a single pod, configures a 10GB storage PVC, with 24 hours of retention), you can simply run the following commands:

```
oc create namespace netobserv
oc apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/5410e65b8e05aaabd1244a9524cfedd8ac8c56b5/examples/zero-click-loki/1-storage.yaml) -n netobserv
oc apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/5410e65b8e05aaabd1244a9524cfedd8ac8c56b5/examples/zero-click-loki/2-loki.yaml) -n netobserv
```

If you prefer to not use Loki, you must set `spec.loki.enable` to `false` in `FlowCollector`.
In that case, you can still get the Prometheus metrics or export raw flows to a custom collector. But be aware that some of the Console plugin features will be disabled. For instance, you will not be able to view raw flows there, and the metrics / topology will have a more limited level of details, missing information such as pods or IPs.

### Kafka

[Apache Kafka](https://kafka.apache.org/) can optionally be used for a more resilient and scalable architecture. You can use for example [Strimzi](https://strimzi.io/), an operator-based distribution of Kafka for Kubernetes and OpenShift.

### Grafana

[Grafana](https://grafana.com/oss/grafana/) can optionally be installed for custom dashboards and query capabilities.

## Configuration

The `FlowCollector` resource is used to configure the operator and its managed components. A comprehensive documentation is [available here](https://github.com/netobserv/network-observability-operator/blob/1.9.1-community/docs/FlowCollector.md), and a full sample file [there](https://github.com/netobserv/network-observability-operator/blob/1.9.1-community/config/samples/flows_v1beta2_flowcollector.yaml).

To edit configuration in cluster, run:

```bash
oc edit flowcollector cluster
```

As it operates cluster-wide on every node, only a single `FlowCollector` is allowed, and it has to be named `cluster`.

A couple of settings deserve special attention:

- Sampling (`spec.agent.ebpf.sampling`): a value of `100` means: one flow every 100 is sampled. `1` means all flows are sampled. The lower it is, the more flows you get, and the more accurate are derived metrics, but the higher amount of resources are consumed. By default, sampling is set to 50 (ie. 1:50). Note that more sampled flows also means more storage needed. We recommend to start with default values and refine empirically, to figure out which setting your cluster can manage.

- Loki (`spec.loki`): configure here how to reach Loki. The default values match the Loki quick install paths mentioned above, but you might have to configure differently if you used another installation method. Make sure to disable it (`spec.loki.enable`) if you don't want to use Loki.

- Quick filters (`spec.consolePlugin.quickFilters`): configure preset filters to be displayed in the Console plugin. They offer a way to quickly switch from filters to others, such as showing / hiding pods network, or infrastructure network, or application network, etc. They can be tuned to reflect the different workloads running on your cluster. For a list of available filters, [check this page](https://github.com/netobserv/network-observability-operator/blob/1.9.1-community/docs/QuickFilters.md).

- Kafka (`spec.deploymentModel: KAFKA` and `spec.kafka`): when enabled, integrates the flow collection pipeline with Kafka, by splitting ingestion from transformation (kube enrichment, derived metrics, ...). Kafka can provide better scalability, resiliency and high availability ([view more details](https://www.redhat.com/en/topics/integration/what-is-apache-kafka)). Assumes Kafka is already deployed and a topic is created.

- Exporters (`spec.exporters`) an optional list of exporters to which to send enriched flows. KAFKA and IPFIX exporters are supported. This allows you to define any custom storage or processing that can read from Kafka or use the IPFIX standard.

- To enable availability zones awareness, set `spec.processor.addZone` to `true`.

## Resource considerations

The following table outlines examples of resource considerations for clusters with certain workload sizes.
The examples outlined in the table demonstrate scenarios that are tailored to specific workloads. Consider each example only as a baseline from which adjustments can be made to accommodate your workload needs.


| Resource recommendations                        | Extra small (10 nodes) | Small (25 nodes)       | Medium (65 nodes) **    | Large (120 nodes) **          |
| ----------------------------------------------- | ---------------------- | ---------------------- | ----------------------- | ----------------------------- |
| *Worker Node vCPU and memory*                   | 4 vCPUs\| 16GiB mem *  | 16 vCPUs\| 64GiB mem * | 16 vCPUs\| 64GiB mem  * |16 vCPUs\| 64GiB Mem *         |
| *LokiStack size*                                | `1x.extra-small`       | `1x.small`             | `1x.small`              | `1x.medium`                   |
| *Network Observability controller memory limit* | 400Mi (default)        | 400Mi (default)        | 400Mi (default)         | 800Mi                         |
| *eBPF sampling rate*                            | 50 (default)           | 50 (default)           | 50 (default)            | 50 (default)                  |
| *eBPF memory limit*                             | 800Mi (default)        | 800Mi (default)        | 2000Mi                  | 800Mi (default)               |
| *FLP memory limit*                              | 800Mi (default)        | 800Mi (default)        | 800Mi (default)         | 800Mi (default)               |
| *FLP Kafka partitions*                          | N/A                    | 48                     | 48                      | 48                            |
| *Kafka consumer replicas*                       | N/A                    | 24                     | 24                      | 24                            |
| *Kafka brokers*                                 | N/A                    | 3 (default)            | 3 (default)             | 3 (default)                   |

*. Tested with AWS M6i instances.
**. In addition to this worker and its controller, 3 infra nodes (size `M6i.12xlarge`) and 1 workload node (size `M6i.8xlarge`) were tested.

## Further reading

Please refer to the documentation on GitHub for more information.

This documentation includes:

- An [overview](https://github.com/netobserv/network-observability-operator#openshift-console) of the features, with screenshots
- More information on [configuring metrics](https://github.com/netobserv/network-observability-operator/blob/1.9.1-community/docs/Metrics.md).
- A [performance](https://github.com/netobserv/network-observability-operator#performance-fine-tuning) section, for fine-tuning
- A [security](https://github.com/netobserv/network-observability-operator#securing-data-and-communications) section
- An [F.A.Q.](https://github.com/netobserv/network-observability-operator#faq--troubleshooting) section
