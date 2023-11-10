# Metrics in the NetObserv Operator

The NetObserv operator uses [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline/) to generate metrics out of flow logs.

They can be configured in the `FlowCollector` custom resource, via `spec.processor.metrics.includeList`. It is a list of metric names that tells which ones to generate.

The names correspond to the names in Prometheus without their prefix. For example, `namespace_egress_packets_total` will show up as `netobserv_namespace_egress_packets_total` in Prometheus.

Note that the more metrics you add, the bigger is the impact on Prometheus workload resources. Some metrics in particular have a bigger cardinality, such as all metrics starting with `workload_`, which may result in stressing Prometheus if too many of them are enabled. It is recommended to monitor the impact on Prometheus when adding more metrics.

Available names are: (names followed by `*` are enabled by default)
- `namespace_egress_bytes_total`
- `namespace_egress_packets_total`
- `namespace_ingress_bytes_total`
- `namespace_ingress_packets_total`
- `namespace_flows_total` `*`
- `node_egress_bytes_total`
- `node_egress_packets_total`
- `node_ingress_bytes_total` `*`
- `node_ingress_packets_total`
- `node_flows_total`
- `workload_egress_bytes_total`
- `workload_egress_packets_total`
- `workload_ingress_bytes_total` `*`
- `workload_ingress_packets_total`
- `workload_flows_total`

When the `PacketDrop` feature is enabled in `spec.agent.ebpf.features` (with privileged mode), additional metrics are available:
- `namespace_drop_bytes_total`
- `namespace_drop_packets_total` `*`
- `node_drop_bytes_total`
- `node_drop_packets_total`
- `workload_drop_bytes_total`
- `workload_drop_packets_total`

When the `FlowRTT` feature is enabled in `spec.agent.ebpf.features`, additional metrics are available:
- `namespace_rtt_seconds` `*`
- `node_rtt_seconds`
- `workload_rtt_seconds`

When the `DNSTracking` feature is enabled in `spec.agent.ebpf.features`, additional metrics are available:
- `namespace_dns_latency_seconds` `*`
- `node_dns_latency_seconds`
- `workload_dns_latency_seconds`
