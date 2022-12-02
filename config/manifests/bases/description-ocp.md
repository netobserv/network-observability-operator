Network observability is an OpenShift operator that deploys a monitoring pipeline to collect and enrich network flows that are produced by the Network observability eBPF agent.

The operator provides dashboards, metrics, and keeps flows accessible in a queryable log store, Grafana Loki. When a `FlowCollector` instance is created, new views are available in the Console.

## Dependencies

### Loki

[Loki](https://grafana.com/oss/loki/), from GrafanaLabs, is the backend that is used to store all collected flows. The NetObserv Operator does not install Loki directly, however we provide some guidance to help you there.

For a normal usage, we recommend two options:

- Installing the [Loki Operator](https://docs.openshift.com/container-platform/4.11//logging/cluster-logging-loki.html). We have written [a guide](https://github.com/netobserv/documents/blob/main/loki_operator.md) to help you through those steps. Please note that it requires configuring an object storage. Note also that the Loki Operator can also be used for [OpenShift cluster logging](https://docs.openshift.com/container-platform/4.11/logging/cluster-logging.html). If you do so, you should not share the same `LokiStack` for Logging and NetObserv.

- Installing using [Grafana's official documentation](https://grafana.com/docs/loki/latest/). Here also we wrote a ["distributed Loki" step by step guide](https://github.com/netobserv/documents/blob/main/loki_distributed.md).

For a quick try that is not suitable for production and not scalable (it deploys a single pod, configures a 1GB storage PVC, with 24 hours of retention), you can simply run the following commands:

```
kubectl create namespace netobserv
kubectl apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/3c42f1ff9c775dd0b746a92dc08c043f9a05f47f/examples/zero-click-loki/1-storage.yaml) -n netobserv
kubectl apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/3c42f1ff9c775dd0b746a92dc08c043f9a05f47f/examples/zero-click-loki/2-loki.yaml) -n netobserv
```

### Kafka

[Apache Kafka](https://kafka.apache.org/) can optionnaly be used for a more resilient and scalable architecture. You can use for instance [Strimzi](https://strimzi.io/), an operator-based distribution of Kafka for Kubernetes and OpenShift.

### Grafana

[Grafana](https://grafana.com/oss/grafana/) can optionally be installed for custom dashboards and query capabilities.

## Configuration

The `FlowCollector` resource is used to configure the operator and its managed components. A comprehensive documentation is [available here](https://github.com/netobserv/network-observability-operator/blob/0.1.4/docs/FlowCollector.md), and a full sample file [there](https://github.com/netobserv/network-observability-operator/blob/0.1.4/config/samples/flows_v1alpha1_flowcollector.yaml).

To edit configuration in cluster, run:

```bash
kubectl edit flowcollector cluster
```

As it operates cluster-wide, only a single `FlowCollector` is allowed, and it has to be named `cluster`.

A couple of settings deserve special attention:

- Sampling (`spec.agent.ebpf.sampling`): a value of `100` means: one flow every 100 is sampled. `1` means all flows are sampled. The lower it is, the more flows you get, and the more accurate are derived metrics, but the higher amount of resources are consumed. By default, sampling is set to 50 (ie. 1:50). Note that more sampled flows also means more storage needed. We recommend to start with default values and refine empirically, to figure out which setting your cluster can manage.

- Loki (`spec.loki`): configure here how to reach Loki. The default values match the Loki quick install paths mentioned above, but you may have to configure differently if you used another installation method.

- Quick filters (`spec.consolePlugin.quickFilters`): configure preset filters to be displayed in the Console plugin. They offer a way to quickly switch from filters to others, such as showing / hiding pods network, or infrastructure network, or application network, etc. They can be tuned to reflect the different workloads running on your cluster. For a list of available filters, [check this page](https://github.com/netobserv/network-observability-operator/blob/0.1.4/docs/QuickFilters.md).

- Kafka (`spec.deploymentModel: KAFKA` and `spec.kafka`): when enabled, integrates the flow collection pipeline with Kafka, by splitting ingestion from transformation (kube enrichment, derived metrics, ...). Kafka can provide better scalability, resiliency and high availability ([view more details](https://www.redhat.com/en/topics/integration/what-is-apache-kafka)). Assumes Kafka is already deployed and a topic is created.

- Exporters (`spec.exporters`, _experimental_) an optional list of exporters to which to send enriched flows. Currently only KAFKA is supported. This allows you to define any custom storage or processing that can read from Kafka. This feature is flagged as _experimental_ as it has not been thoroughly or stress tested yet, so use at your own risk.

## Further reading

Please refer to the documentation on GitHub more more information.

This documentation includes:

- An [overview](https://github.com/netobserv/network-observability-operator#openshift-console) of the features, with screenshots
- A [performance](https://github.com/netobserv/network-observability-operator#performance-fine-tuning) section, for fine-tuning
- A [security](https://github.com/netobserv/network-observability-operator#securing-data-and-communications) section
- A [F.A.Q](https://github.com/netobserv/network-observability-operator#faq--troubleshooting) section
