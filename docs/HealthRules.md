# Health rules in NetObserv

The NetObserv operator comes with a set of predefined health rules, based on its [metrics](./Metrics.md), that you can configure, extend or disable.
These rules are converted into a `PrometheusRule` resource, either as Alerts or as Recording rules. The alerts are then managed by Prometheus AlertManager. Both recording rules and alerts are displayed in the Network Health page of the Console.

These health rules are provided as a convenience, to take the most of NetObserv built-in metrics without requiring you to write complexe PromQL or to do fine-tuning. They give a health indication of your cluster network.

To get a detailed description of the rules, [check the runbooks](https://github.com/openshift/runbooks/tree/master/alerts/network-observability-operator).

## Default rules

By default, NetObserv creates health rules contextual to the enabled features. For example, packet drops related rules are only created if the `PacketDrop` feature is enabled. Because rules are built upon metrics, you may also see configuration warnings if some enabled rules are missing their required metrics, which can be configured in `spec.processor.metrics.includeList` (see [Metrics.md](./Metrics.md)).

These rules are installed by default:

- `PacketDropsByDevice`
- `PacketDropsByKernel`
- `IPsecErrors`
- `NetpolDenied`
- `LatencyHighTrend`
- `DNSErrors`
- `DNSNxDomain`
- `ExternalEgressHighTrend`
- `ExternalIngressHighTrend`
- `Ingress5xxErrors`
- `IngressHTTPLatencyTrend`

On top of that, there are also some operational alerts that relate to NetObserv's self health:

- `NetObservNoFlows`: triggered when no flows are being observed for a certain period.
- `NetObservLokiError`: triggered when flows are being dropped due to Loki errors.

## Other alert templates

Templates that are not enabled by default, but available for configuration: (none at this time).

## Configure predefined alerts

Alerts are configured in the `FlowCollector` custom resource, via `spec.processor.metrics.alerts`.

They are organized by templates and variants. The template names are the ones listed above, such as `PacketDropsByKernel`. For each template, you can define a list of variants, each with their thresholds and grouping configuration.

Example:

```yaml
spec:
  processor:
    metrics:
      healthRules:
      - template: PacketDropsByKernel
        mode: Alert # or Recording
        variants:
        # triggered when the whole cluster traffic (no grouping) reaches 10% of drops
        - thresholds:
            critical: "10"
        # triggered when per-node traffic reaches 5% of drops, with gradual severity
        - thresholds:
            critical: "15"
            warning: "10"
            info: "5"
          groupBy: Node
```

The `mode` setting can be either defined per variant, or for the whole template.

When you configure a template, it overrides the default configuration for that template. So, if you want to add a new rule on top of the default ones for a template, you may want to replicate the default configuration manually. All defaults are described in the [runbooks](https://github.com/openshift/runbooks/tree/master/alerts/network-observability-operator).

## Disable predefined alerts

Alert templates can be disabled in `spec.processor.metrics.disableAlerts`. This settings accepts a list of template names, as listed above.

If a template is disabled _and_ overridden in `spec.processor.metrics.healthRules`, the disable setting takes precedence: the alert rule will not be created.

## Creating your own rules that contribute to the Health dashboard

This health rule API in NetObserv `FlowCollector` is simply a mapping to the Prometheus operator API, generating a `PrometheusRule`.

You can check what is the actual generated resource by running:

```bash
kubectl get prometheusrules -n netobserv -oyaml
```

While the above sections explain how you can customize those opinionated rules, you are not limited to them: you can go further and create your own `AlertingRule` (or `PrometheusRule`) resources. You'll just need to be familiar with PromQL (or to learn).

[Click here](../config/samples/alerts) to see sample alerts, that are not built-in NetObserv.

Let's take the [incoming-traffic-surge](../config/samples/alerts/incoming-traffic-surge.yaml) as an example. What it does is raise an alert when the current ingress traffic exceeds by more than twice the traffic from the day before.

### Anatomy of the PromQL

Here's the PromQL:

```
(100 *
  (
    (sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m])) by (DstK8S_Namespace) > 1000)
    - sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m] offset 1d)) by (DstK8S_Namespace)
  )
  / sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m] offset 1d)) by (DstK8S_Namespace))
> 100
```

Let's break it down. The base query pattern is this:

`sum(rate(netobserv_workload_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m])) by (DstK8S_Namespace)`

This is the bytes rate coming from "openshift-ingress" to any of your workload's namespaces, over the last 30 minutes. This metric is provided by NetObserv (note that depending on your FlowCollector configuration, you may need to use `netobserv_namespace_ingress_bytes_total` instead of `netobserv_workload_ingress_bytes_total`).

Appending ` > 1000` to this query keeps only the rates observed greater than 1KBps, in order to eliminate the noise from low-bandwidth consumers. 1KBps still isn't a lot, you may want to increase it. Note also that the bytes rate is relative to the sampling interval defined in the `FlowCollector` agent configuration. If you have a sampling ratio of 1:100, consider that the actual traffic might be approximately 100 times higher than what is reported by the metrics. Alternatively, the metric `netobserv_agent_sampling_rate` can be use to normalize the byte rates, decoupling the promql from the sampling configuration.

In the following parts of the PromQL, you can see `offset 1d`: this is to run the same query, one day earlier. You can change that according to your needs, for instance `offset 5h` will be five hours ago.

Which gives us the formula `100 * (<query now> - <query yesterday>) / <query yesterday>`: it's the percentage of increase compared to yesterday. It can be negative, if the bytes rate today is lower than yesterday.

Finally, the last part, `> 100`, eliminates increases that are lower than 100%, so that we don't get alerted by that.

### Metadata

Some metadata is required to work with Prometheus and AlertManager (not specific to NetObserv):

```yaml
      annotations:
        message: |-
          NetObserv is detecting a surge of incoming traffic: current traffic to {{ $labels.DstK8S_Namespace }} has increased by more than 100% since yesterday.
        summary: "Surge in incoming traffic"
      labels:
        severity: warning
```

As you can see, you can leverage the output labels from the PromQL defined previously in the description. Here, since we've grouped the results per `DstK8S_Namespace`, we can use it in our text.

The severity label should be "critical", "warning" or "info".

On top of that, in order to have the rule picked up in the Health dashboard, NetObserv needs other information:

```yaml
      annotations:
        netobserv_io_network_health: '{"namespaceLabels":["DstK8S_Namespace"],"threshold":"100","unit":"%","upperBound":"500"}'
      labels:
        netobserv: "true"
```

The label `netobserv: "true"` is required.

The annotation `netobserv_io_network_health` is optional, and gives you some control on how the alert renders in the Health page. It is a JSON string that consists in:
- `namespaceLabels`: one or more labels that hold namespaces. When provided, the alert will show up under the "Namespaces" tab.
- `nodeLabels`: one or more labels that hold node names. When provided, the alert will show up under the "Nodes" tab.
- `workloadLabels`: one or more labels that hold owner/workload names. When provided alongside with `kindLabels`, the alert will show up under the "Owners" tab.
- `kindLabels`: one or more labels that hold owner/workload kinds. When provided alongside with `workloadLabels`, the alert will show up under the "Owners" tab.
- `threshold`: the alert threshold as a string, expected to match the one defined in PromQL.
- `unit`: the data unit, used only for display purpose.
- `upperBound`: an upper bound value used to compute score on a closed scale. It doesn't necessarily have to be a maximum of the metric values, but metric values will be clamped if they are above the upper bound.
- `links`: a list of links to be displayed contextually to the alert. Each link consists in:
  - `name`: display name.
  - `url`: the link URL.
- `trafficLink`: information related to the link to the Network Traffic page, for URL building. Some filters will be set automatically, such as the node or namespace filter.
  - `extraFilter`: an additional filter to inject (e.g: a DNS response code, for DNS-related alerts).
  - `backAndForth`: should the filter include return traffic? (true/false)
  - `filterDestination`: should the filter target the destination of the traffic instead of the source? (true/false)

`namespaceLabels` and `nodeLabels` are mutually exclusive. If none of them is provided, the alert will show up under the "Global" tab.
