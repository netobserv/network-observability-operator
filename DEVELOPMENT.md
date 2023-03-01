The NetObserv Operator is meant to run in a Kubernetes cluster like OpenShift or [Kind](https://kind.sigs.k8s.io/). These are the two options most used by the development team.

[Architecture description](./docs/Architecture.md).

## Build / format / lint the code, run unit tests

```bash
make build test
```

## Build and deploy a Docker image

A way to test code changes is to build a Docker image from local sources, push it to your own Docker repository, and deploy it to an existing cluster. Do the following, but replace IMG value with your own registry and account:

```bash
IMG="quay.io/youraccount/network-observability-operator:test" make image-build image-push deploy
```

After the operator is deployed, set up Loki, which is used to store flows, install a `FlowCollector` custom resource to collect network flows, and optionally install Grafana to provide a user interface and dashboards.

This provides a quick and easy way to deploy Loki, Grafana and a `FlowCollector` with default values. Note this Loki setup is not for production use.

```bash
make deploy-loki deploy-grafana deploy-sample-cr
```

It will run Loki and Grafana locally, and set up a local port-forward to them. To avoid this, add `PORT_FWD=false` to the command above.

Creating a `FlowCollector` triggers the operator deploying the monitoring pipeline:

- Configures IPFIX exports
- Deploys the flow collector pods, `flowlogs-pipeline`
- Deploys the `netobserv-plugin` if using OpenShift Console

You should be able to see flows in OpenShift Console and Grafana. If not, wait up to 10 minutes. See the [FAQ on troubleshooting](./README.md#faq--troubleshooting) for more information.

### Test another one's pull request

To test a pull request opened by someone else, you just need to pull it locally. Using [GitHub CLI](https://cli.github.com/) is an easy way to do it. Then repeat the steps mentioned above to build, push an image, then deploy the operator and its custom resource.

## Deploy a specific image

Images are built and pushed through CI to [quay.io](https://quay.io/repository/netobserv/network-observability-operator?tab=tags).

You can refer to existing commits using their short-SHA as the image tag, or refer to existing releases. E.g:

```bash
# By commit SHA
OPERATOR_VERSION="960766c" make deploy
# By release
OPERATOR_VERSION="0.1.2" make deploy
```

It is recommended to switch to the corresponding release Git tag before deploying an old version to make sure the underlying components refer to the correct versions.

When `OPERATOR_VERSION` is not provided, it defaults to the latest released version.

To deploy all components on their `main` image tag (which correspond to their `main` branches, ie. their latest builds), you can simply run:

```bash
make deploy-latest
```

## Installing Kafka

Kafka can be used to separate flow ingestion from flow transformation. The operator does not manage kafka deployment and topic creation. We provide a quick setup for Kafka using the [strimzi operator](https://strimzi.io/).

```bash
make deploy-kafka
```

Kafka can then be enabled in the `FlowCollector` resource by setting `spec.deploymentModel` to `KAFKA`. If you use your own Kafka setup, make sure to configure `spec.kafka.address` and `spec.kafka.topic` accordingly.

## Linking with API changes in flowlogs-pipeline

To link with merged changes (but unreleased), update the FLP version by running (replacing "LONG_COMMIT_SHA"):

```bash
go get github.com/netobserv/flowlogs-pipeline@LONG_COMMIT_SHA
```

To link with unmerged changes, add this at the bottom of `go.mod`:

```
replace github.com/netobserv/flowlogs-pipeline => ../flowlogs-pipeline
```

Then run:

```bash
make vendors
```

## Deploy as bundle

For more details, refer to the [Operator Lifecycle Manager (OLM) bundle quickstart documentation](https://sdk.operatorframework.io/docs/olm-integration/quickstart-bundle/).

This task should be automatically done by the CI/CD pipeline. However, if you want to deploy as
bundle for local testing, you should execute the following commands:

```bash
export USER=<container-registry-username>
export VERSION=0.0.1
export IMG=quay.io/$USER/network-observability-operator:v$VERSION
export BUNDLE_IMG=quay.io/$USER/network-observability-operator-bundle:v$VERSION
make image-build image-push
make bundle bundle-build bundle-push
```

Optionally, you might validate the bundle:

```bash
bin/operator-sdk bundle validate $BUNDLE_IMG
```

> Note: the base64 logo can be generated with: `base64 -w 0 <image file>`, then manually pasted in the [CSV manifest file](./config/manifests/bases/netobserv-operator.clusterserviceversion.yaml) under `spec.icon`.

### Deploy as bundle from command line

This mode is recommended to quickly test the operator during its development:

```bash
bin/operator-sdk run bundle $BUNDLE_IMG
```

### Deploy as bundle from the Console's OperatorHub page

This mode is recommended when you want to test the customer experience of navigating through the
operators' catalog and installing/configuring it manually through the UI.

First, create and push a catalog image:

```bash
export CATALOG_IMG=quay.io/$USER/network-observability-operator-catalog:v$VERSION
make catalog-build catalog-push catalog-deploy
```

The NetObserv Operator is available in OperatorHub: https://operatorhub.io/operator/netobserv-operator

## Publish on central OperatorHub

See [RELEASE.md](./RELEASE.md#publishing-on-operatorhub).

## Using custom operand image

In the `manager` container of the `netobserv-controller-manager` Deployment, set any of the
following the environment variables with your custom operand image with `kubectl set env` or
`oc set env`:

* `RELATED_IMAGE_EBPF_AGENT`
* `RELATED_IMAGE_FLOWLOGS_PIPELINE`
* `RELATED_IMAGE_CONSOLE_PLUGIN`

Examples:

```bash
oc -n netobserv set env deployment/netobserv-controller-manager -c "manager" RELATED_IMAGE_EBPF_AGENT="quay.io/netobserv/netobserv-ebpf-agent:main"
oc -n netobserv set env deployment/netobserv-controller-manager -c "manager" RELATED_IMAGE_FLOWLOGS_PIPELINE="quay.io/netobserv/flowlogs-pipeline:main"
oc -n netobserv set env deployment/netobserv-controller-manager -c "manager" RELATED_IMAGE_CONSOLE_PLUGIN="quay.io/netobserv/network-observability-console-plugin:main"
```

Alternatively you can use helper make targets for the same purpose:

```bash
USER=myself VERSION=test make set-agent-image set-flp-image set-plugin-image
```

## View flowlogs-pipeline metrics in console

To view the generated flowlogs-pipeline metrics in the Openshift console, perform the following:

```
cd hack
./enable-metrics.sh
```

The metrics will be visible in the Openshift console under the tab `Observe -> Metrics.`
Look for the metrics that begin with `netobserv_.`
