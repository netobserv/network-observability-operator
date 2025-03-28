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

In OpenShift, NetObserv is named Network Observability operator and can be found in OperatorHub as an OLM operator. This section does not apply to it: please refer to the [OpenShift documentation](https://docs.openshift.com/container-platform/latest/observability/network_observability/installing-operators.html) in that case.

### Pre-requisite

The following architectures are supported: amd64, arm64, ppc64le and s390x.

NetObserv has a couple of dependencies that must be installed on your cluster:

- Cert-manager
- Prometheus
- Loki

Loki is not mandatory but improves the overall experience with NetObserv.
If you don't have these dependencies already, some convenience scripts are available from the repository, provided for demo purpose:

```bash
git clone https://github.com/netobserv/network-observability-operator.git && cd network-observability-operator
PORT_FWD=false make deploy-prometheus deploy-loki install-cert-manager
# (it is expected to see errors while running this script, since it runs several attempts creating a certificate for testing, before eventually succeeding)
```

### Install with Helm

```bash
helm repo add netobserv https://netobserv.io/static/helm/ --force-update
helm install my-netobserv --set standaloneConsole.enable=true netobserv/netobserv-operator
# If you're in OpenShift, you can omit "--set standaloneConsole.enable=true" to use the Console plugin instead.
```

You can now create a `FlowCollector` resource. A very short `FlowCollector` should work, using default values, plus with the standalone console enabled:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: flows.netobserv.io/v1beta2
kind: FlowCollector
metadata:
  name: cluster
spec:
  namespace: netobserv
  consolePlugin:
    advanced:
      env:
        TEST_CONSOLE: "true"
  prometheus:
    querier:
      manual:
        url: http://prometheus:9090
EOF
```

A few remarks:
- While the [web console](https://github.com/netobserv/network-observability-console-plugin) is primarily designed as a plugin for the OpenShift Console, it is still possible to deploy it as a standalone, which the dev team sometimes use for testing. This is why it is mentioned as "TEST_CONSOLE" here.
- If you're in OpenShift, you should omit "TEST_CONSOLE: true" to use the Console plugin instead, which offers a better / more integrated experience.
- You can change the Prometheus URL depending on your installation. This example URL works if you use the `make deploy-prometheus` script from the repository. Prometheus configuration options are documented [here](https://github.com/netobserv/network-observability-operator/blob/main/docs/FlowCollector.md#flowcollectorspecprometheus-1).

To view the test console, you can port-forward 9001:

```bash
kubectl port-forward svc/netobserv-plugin 9001:9001 -n netobserv
```

Then open http://localhost:9001/ in your browser.

### More information

Refer to the GitHub repository README for more information: https://github.com/netobserv/network-observability-operator/blob/main/README.md
