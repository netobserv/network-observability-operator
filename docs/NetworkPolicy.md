# NetObserv NetworkPolicy

Depending on the Kubernetes distribution and CNI, NetObserv may come secured by default with a built-in network policy. You can force installing it or not by setting `spec.networkPolicy.enable` in `FlowCollector`. If the built-in policy does not work as intended, it is recommended to turn it off and create your own instead. NetObserv runs some highly privileged workloads, hence it is important to keep it as much isolated as possible.

If the built-in policy looks _almost_ good, but some allowed namespaces are missing, you can add allowed namespaces in `spec.networkPolicy.additionalNamespaces`.

You can find below the communication matrix that will help you create your own policy. Be aware that some pods use the host network, which not all CNI support for network policies.

## Supported environments for built-in policy

The following environments have been tested with the built-in policy, and will have it enabled by default:

- Kindnet / kind.
- OVN-Kubernetes (upstream), and API server in `kube-system` namespace.
- OpenShift with OVN-Kubernetes.

Feel free to ask & contribute to increase this list.

## Common labels

All pods deployed by NetObserv have the label `part-of: netobserv-operator`. So they can be selected via:

```yaml
  podSelector:
    matchLabels:
      part-of: netobserv-operator
```

## Namespaces

The main namespace is `netobserv` by default, and can be configured in `FlowCollector` via `spec.namespace`.
All pods managed by NetObserv are deployed there, except the `netobserv-ebpf-agent` pods, which are in the "privileged" namespace: it's the main namespace + `-privileged` suffix, so `netobserv-privileged` by default.

## Communication flows

This section describes the flows in details to help you build your network policy; However you can simplify the rules if you choose to allow in-namespace traffic, with:

```yaml
spec:
  egress:
  - to:
    - podSelector: {}
  ingress:
  - from:
    - podSelector: {}
```

and deploying Loki/Prometheus/Kafka (when relevant) in the same namespace.

### Operator

Label: `app=netobserv-operator`, default namespace: `netobserv`.

**Ingress:**

- Must allow traffic from Kube API Server to Webhooks: TCP, port 9443.
- May allow traffic from Prometheus to Metrics endpoint: TCP, port 8443 (the source depends on your Prometheus setup).

**Egress:**

- Must allow traffic to Kube API Server: TCP, port 6443.

### eBPF agents

Label: `app=netobserv-ebpf-agent`, default namespace: `netobserv-privileged`. This is host-network pods.

**Ingress:**

- May allow traffic from Prometheus to Metrics endpoint: TCP, port 9400 (the source depends on your Prometheus setup).

**Egress:**

When `spec.deploymentModel` is `Service`:
- Must allow traffic to flowlogs-pipeline (`app=flowlogs-pipeline`), TCP, default port 2055 (port configurable in `spec.processor.advanced.port`).

When `spec.deploymentModel` is `Kafka`:
- Must allow traffic to Kafka, TCP, port depends on your Kafka setup.

When `spec.deploymentModel` is `Direct`:
- Must allow traffic to flowlogs-pipeline (`app=flowlogs-pipeline`), TCP, default port 2055 (port configurable in `spec.processor.advanced.port`). `flowlogs-pipeline` are also host-network pods (same host).

### Flowlogs-pipeline

Label: `app=flowlogs-pipeline`, default namespace: `netobserv`.

**Ingress:**

- Must allow traffic from Prometheus to Metrics endpoint: TCP, port 9401 (the source depends on your Prometheus setup).

When `spec.deploymentModel` is `Service`:
- Must allow traffic from agents (`app=netobserv-ebpf-agent`), TCP, default port 2055 (port configurable in `spec.processor.advanced.port`).

When `spec.deploymentModel` is `Direct`:
- Must allow traffic from agents (`app=netobserv-ebpf-agent`), TCP, default port 2055 (port configurable in `spec.processor.advanced.port`). `flowlogs-pipeline` are also host-network pods (same host).

**Egress:**

When `spec.deploymentModel` is `Kafka`:
- Must allow traffic to Kafka, TCP, port depends on your Kafka setup.

When using Loki (`spec.loki.enabled`):
- Must allow traffic to Loki, TCP, port depends on your Loki setup (usually 3100).

When exporters are configured (`spec.exporters`):
- Must allow traffic to exporters (refer to the exporter configuration).

### Web console

Label: `app=netobserv-plugin`, default namespace: `netobserv`.

**Ingress:**

- If you set up an Ingress route/gateway to the web console, configure it accordingly to allow incoming user traffic.

- May allow traffic from Prometheus to Metrics endpoint: TCP, port 9002 (the source depends on your Prometheus setup).

**Egress:**

- Must allow traffic to Prometheus and AlertManager, TCP, as defined in `spec.prometheus.querier`.

When using Loki (`spec.loki.enabled`):
- Must allow traffic to Loki, TCP, port depends on your Loki setup (usually 3100).
