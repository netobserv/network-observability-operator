NetObserv Operator is an OpenShift / Kubernetes operator for network observability. It deploys a monitoring pipeline to collect and enrich network flows. These flows can be produced by the NetObserv eBPF agent, or by any device or CNI able to export flows in IPFIX format, such as OVN-Kubernetes.

The operator provides dashboards, metrics, and keeps flows accessible in a queryable log store, Grafana Loki. When used in OpenShift, new views are available in the Console.

## Dependencies

### Loki

[Loki](https://grafana.com/oss/loki/), from GrafanaLabs, is the backend that is used to store all collected flows. The NetObserv Operator does not install Loki directly, however we provide some guidance to help you there.

For normal usage, we recommend two options:

- Installing the [Loki Operator](https://loki-operator.dev/docs/prologue/quickstart.md/). We have written [a guide](https://github.com/netobserv/documents/blob/main/loki_operator.md) to help you through those steps. Please note that it requires configuring an object storage. Note also that the Loki Operator can also be used for [OpenShift cluster logging](https://docs.openshift.com/container-platform/4.11/logging/cluster-logging.html). If you do so, you should not share the same `LokiStack` for Logging and NetObserv.

- Installing using [Grafana's official documentation](https://grafana.com/docs/loki/latest/). Here also we wrote a ["distributed Loki" step by step guide](https://github.com/netobserv/documents/blob/main/loki_distributed.md).

For a quick try that is not suitable for production and not scalable (it deploys a single pod, configures a 1GB storage PVC, with 24 hours of retention), you can simply run the following commands:

```
kubectl create namespace netobserv
kubectl apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/252bb624cf0425a1b66f59ce68fb939f246ef77b/examples/zero-click-loki/1-storage.yaml) -n netobserv
kubectl apply -f <(curl -L https://raw.githubusercontent.com/netobserv/documents/252bb624cf0425a1b66f59ce68fb939f246ef77b/examples/zero-click-loki/2-loki.yaml) -n netobserv
```

### Kafka

[Apache Kafka](https://kafka.apache.org/) can optionally be used for a more resilient and scalable architecture. You can use for instance [Strimzi](https://strimzi.io/), an operator-based distribution of Kafka for Kubernetes and OpenShift.

### Grafana

[Grafana](https://grafana.com/oss/grafana/) can optionally be installed for custom dashboards and query capabilities.

## Configuration

The `FlowCollector` resource is used to configure the operator and its managed components. A comprehensive documentation is [available here](https://github.com/netobserv/network-observability-operator/blob/1.0.2-rc1/docs/FlowCollector.md), and a full sample file [there](https://github.com/netobserv/network-observability-operator/blob/1.0.2-rc1/config/samples/flows_v1alpha1_flowcollector.yaml).

To edit configuration in cluster, run:

```bash
kubectl edit flowcollector cluster
```

As it operates cluster-wide, only a single `FlowCollector` is allowed, and it has to be named `cluster`.

A couple of settings deserve special attention:

- Agent (`spec.agent.type`) can be `EBPF` (default) or `IPFIX`. eBPF is recommended, as it should work in more situations and offers better performances. If you can't, or don't want to use eBPF, note that the IPFIX option is fully functional only when using [OVN-Kubernetes](https://github.com/ovn-org/ovn-kubernetes/) CNI. Other CNIs are not officially supported, but you may still be able to configure them manually if they allow IPFIX exports.

- Sampling (`spec.agent.ebpf.sampling` and `spec.agent.ipfix.sampling`): a value of `100` means: one flow every 100 is sampled. `1` means all flows are sampled. The lower it is, the more flows you get, and the more accurate are derived metrics, but the higher amount of resources are consumed. By default, sampling is set to 50 (ie. 1:50) for eBPF and 400 (1:400) for IPFIX. Note that more sampled flows also means more storage needed. We recommend to start with default values and refine empirically, to figure out which setting your cluster can manage.

- Loki (`spec.loki`): configure here how to reach Loki. The default values match the Loki quick install paths mentioned above, but you may have to configure differently if you used another installation method.

- Quick filters (`spec.consolePlugin.quickFilters`): configure preset filters to be displayed in the Console plugin. They offer a way to quickly switch from filters to others, such as showing / hiding pods network, or infrastructure network, or application network, etc. They can be tuned to reflect the different workloads running on your cluster. For a list of available filters, [check this page](https://github.com/netobserv/network-observability-operator/blob/1.0.2-rc1/docs/QuickFilters.md).

- Kafka (`spec.deploymentModel: KAFKA` and `spec.kafka`): when enabled, integrates the flow collection pipeline with Kafka, by splitting ingestion from transformation (kube enrichment, derived metrics, ...). Kafka can provide better scalability, resiliency and high availability ([view more details](https://www.redhat.com/en/topics/integration/what-is-apache-kafka)). Assumes Kafka is already deployed and a topic is created.

- Exporters (`spec.exporters`, _experimental_) an optional list of exporters to which to send enriched flows. Currently only KAFKA is supported. This allows you to define any custom storage or processing that can read from Kafka. This feature is flagged as _experimental_ as it has not been thoroughly or stress tested yet, so use at your own risk.

## Further reading

Please refer to the documentation on GitHub for more information.

This documentation includes:

- An [overview](https://github.com/netobserv/network-observability-operator#openshift-console) of the features, with screenshots
- A [performance](https://github.com/netobserv/network-observability-operator#performance-fine-tuning) section, for fine-tuning
- A [security](https://github.com/netobserv/network-observability-operator#securing-data-and-communications) section
- An [F.A.Q.](https://github.com/netobserv/network-observability-operator#faq--troubleshooting) section
