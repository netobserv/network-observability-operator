# NetObserv Operator

NetObserv Operator is a Kubernetes / OpenShift operator for network observability. It deploys a monitoring pipeline that consists in:
- an eBPF agent, that generates network flows from captured packets
- flowlogs-pipeline, a component that collects, enriches and exports these flows
- when used in OpenShift, a Console plugin for flows visualization with powerful filtering options, a topology representation and more

Flow data is then available in multiple ways, each optional:

- As Prometheus metrics
- As raw flow logs stored in Loki
- As raw flow logs exported to a collector

## Getting Started

You can install the NetObserv Operator using [Helm](https://helm.sh/), or directly from sources.

In OpenShift, NetObserv is named Network Observability operator and can be found in OperatorHub as an OLM operator. This section does not apply to it: please refer to the [OpenShift documentation](docs.redhat.com/en/documentation/openshift_container_platform/latest/html/network_observability/installing-network-observability-operators) in that case.

### Pre-requisite

The following architectures are supported: amd64, arm64, ppc64le and s390x.

NetObserv has a couple of dependencies that must be installed on your cluster:

- Cert-manager
- Prometheus
- Loki

Cert-manager has to be installed separately. For example, using helm:

```bash
helm repo add cert-manager https://charts.jetstack.io
helm install my-cert-manager cert-manager/cert-manager --set crds.enabled=true
```

Prometheus and Loki can be installed separately, or as dependencies of NetObserv (see below).

Loki is not mandatory but improves the overall experience with NetObserv.

### Install with Helm

> [!TIP]
> See it also on [ArtifactHub](https://artifacthub.io/packages/helm/netobserv/netobserv-operator).

```bash
helm repo add netobserv https://netobserv.io/static/helm/ --force-update

# Standalone install, including dependencies:
helm install my-netobserv -n netobserv --create-namespace --set standaloneConsole.enable=true --set install.loki=true --set install.prom=true netobserv/netobserv-operator

# OR minimal install (Prometheus/Loki are installed separately)
helm install my-netobserv -n netobserv --create-namespace --set standaloneConsole.enable=true netobserv/netobserv-operator

# If you're in OpenShift, you can omit "--set standaloneConsole.enable=true" to use the Console plugin instead.
```

You can now create a `FlowCollector` resource ([full API reference](https://github.com/netobserv/network-observability-operator/blob/main/docs/FlowCollector.md#flowsnetobserviov1beta2)). A short `FlowCollector` should work, using most default values, plus with the standalone console enabled:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: flows.netobserv.io/v1beta2
kind: FlowCollector
metadata:
  name: cluster
spec:
  namespace: netobserv
  networkPolicy:
    enable: false
  consolePlugin:
    advanced:
      env:
        TEST_CONSOLE: "true"
  loki:
    mode: Monolithic
    monolithic:
      url: 'http://my-netobserv-loki.netobserv.svc.cluster.local.:3100/'
  prometheus:
    querier:
      manual:
        url: http://my-netobserv-prometheus-server.netobserv.svc.cluster.local./
EOF
```

A few remarks:
- While the [web console](https://github.com/netobserv/network-observability-console-plugin) is primarily designed as a plugin for the OpenShift Console, it is still possible to deploy it as a standalone, which the dev team sometimes use for testing. This is why it is mentioned as "TEST_CONSOLE" here.
- If you're in OpenShift, you should omit "TEST_CONSOLE: true" to use the Console plugin instead, which offers a better / more integrated experience.
- You can change the Prometheus and Loki URLs depending on your installation. This example works if you use the "standalone" installation described above, with `install.loki=true` and `install.prom=true`. Check more configuration options for [Prometheus](https://github.com/netobserv/network-observability-operator/blob/main/docs/FlowCollector.md#flowcollectorspecprometheus-1) and [Loki](https://github.com/netobserv/network-observability-operator/blob/main/docs/FlowCollector.md#flowcollectorspecloki-1).
- You can enable networkPolicy, which makes the operator lock down the namespaces that it manages; however, this is highly dependent on your cluster topology, and may cause malfunctions, such as preventing NetObserv pods from communicating with the Kube API server.

To view the test console, you can port-forward 9001:

```bash
kubectl port-forward svc/netobserv-plugin 9001:9001 -n netobserv
```

Then open http://localhost:9001/ in your browser.

### More information

Refer to the GitHub repository README for more information: https://github.com/netobserv/network-observability-operator/blob/main/README.md
