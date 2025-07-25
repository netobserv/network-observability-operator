# Alerts in the NetObserv Operator

The NetObserv operator comes with a set of predefined alerts, based on its [metrics](./Metrics.md), that you can configure, extend or disable.
The configured alerts generate a `PrometheusRule` resource that is used to feed Prometheus AlertManager.

These alerts are provided as a convenience, to take the most of NetObserv built-in metrics without requiring you to fine-tune anything. They provide health indication of your cluster network.

## Default alerts

By default, NetObserv creates some alerts, contextual to the enabled features. For example, packet drops related alerts are only created if the `PacketDrop` feature is enabled. Because alerts are built upon metrics, you may also see configuration warnings if some enabled alerts are missing their required metrics, which can be configured in `spec.processor.metrics.includeList` (see [Metrics.md](./Metrics.md)).

Here is the list of alerts installed by default:

- `TooManyDrops`: triggered on high percentage of packet drops; it requires the `PacketDrop` agent feature. 3 variants installed by default:
  - "Warning" severity, >=10% of drops, grouped by source nodes
  - "Warning" severity, >=10% of drops, grouped by destination nodes
  - "Info" severity, >=20% of drops, grouped by source+destination namespaces

On top of that, there are also some operational alerts that relate to NetObserv's self health:

- `NetObservNoFlows`: triggered when no flows are being observed for a certain period.
- `NetObservLokiError`: triggered when flows are being dropped due to Loki errors.

## Configure predefined alerts

Alerts are configured in the `FlowCollector` custom resource, via `spec.processor.metrics.alertGroups`.

They are organized by groups and variants. The group names are the ones listed above, such as `TooManyDrops`. For each group, you can define a list of alert rules to generate, each with their threshold, grouping configuration and severity.

Example:

```yaml
spec:
  processor:
    metrics:
      alertGroups:
      - name: TooManyDrops
        alerts:
        - severity: Critical
          threshold: "10"
          # triggered when the whole cluster traffic (no grouping) reaches 10% of drops
        - severity: Warning
          threshold: "15"
          grouping: PerNode
          groupingDirection: ByDestination
          # triggered when per-destination-node traffic reaches 15% of drops
```

When you configure an alert group, it overrides (replaces) the default configuration for that group. So, if you want to add a new alert on top of the default ones for a group, you need to replicate the default configuration manually, which is described in the section above.

## Disable predefined alerts

Alert groups can be disabled in `spec.processor.metrics.disableAlerts`. This settings accepts a list of group names, as listed above.

If a group is disabled _and_ overridden in `spec.processor.metrics.alertGroups`, the disable setting takes precedence: the alert rule will not created.

## Customizing even further

This alerting API in NetObserv `FlowCollector` is simply a mapping to the Prometheus operator API, generating a `PrometheusRule` that you can see in the `netobserv` namespace (by default) by running:

```bash
oc get prometheusrules -n netobserv -oyaml
```

The sections above explain how you can customize the alerts, but should you feel limited with this configuration API, you can go even further and create your own `AlertingRule` resources.

Here is an example to alert when the current ingress traffic exceeds by more than twice the traffic from the day before.

```yaml
apiVersion: monitoring.openshift.io/v1
kind: AlertingRule
metadata:
  name: netobserv-alerts
  namespace: openshift-monitoring
spec:
  groups:
  - name: NetObservAlerts
    rules:
    - alert: NetObservIncomingBandwidth
      annotations:
        message: |-
          NetObserv is detecting a surge of incoming traffic: current traffic to {{ $labels.DstK8S_Namespace }} has increased by more than 100% since yesterday.
        summary: "Surge in incoming traffic"
      expr: |-
        (100 * ((sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m])) by (DstK8S_Namespace) > 1000) - sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m] offset 1d)) by (DstK8S_Namespace)) / sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m] offset 1d)) by (DstK8S_Namespace)) > 100
      for: 1m
      labels:
        app: netobserv
        severity: warning
```

Let's break it down to understand the promQL expression. The base query pattern is this:

`sum(rate(netobserv_namespace_ingress_bytes_total{SrcK8S_Namespace="openshift-ingress"}[30m])) by (DstK8S_Namespace)`

This is the bytes rate coming from "openshift-ingress" to any of your workload's namespaces, over the last 30 minutes. Note that depending on your configuration, you may need to use `netobserv_workload_ingress_bytes_total` instead of `netobserv_namespace_ingress_bytes_total`.

Appending ` > 1000` to this query keeps only the rates observed greater than 1KBps, in order to eliminate the noise from low-bandwidth consumers. 1KBps still isn't a lot, you may want to increase it. Note also that the bytes rate is relative to the sampling ratio defined in the `FlowCollector` agent configuration. If you have a sampling ratio of 1:100, consider that the actual traffic might be approximately 100 times higher than what is reported by the metrics.

In the following parts of the promQL, you can see `offset 1d`: this is to run the same query, one day before. You can change that according to your needs, for instance `offset 5h` will be five hours ago.

Which gives us the formula `100 * (<query now> - <query yesterday>) / <query yesterday>`: it's the percentage of increase compared to yesterday. It can be negative, if the bytes rate today is lower than yesterday.

Finally, the last part, `> 100`, eliminates increases that are lower than 100%, so that we don't get alerted by that.
