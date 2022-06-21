The NetObserv operator is meant to run in a Kubernetes cluster like OpenShift or KIND (these are the two options most used by the dev team).

## Build / format / lint the code, run unit tests

```bash
make build test
```

## Build and deploy a Docker image

A way to test code changes is to build a Docker image from local sources, push it to your own Docker repository and deploy it to an existing cluster:

(replace `quay.io/youraccount` with your own registry and account)

```bash
IMG="quay.io/youraccount/network-observability-operator:test" make image-build image-push deploy
```

After the operator is deployed, you need to set up Loki (the flows store), install a `FlowCollector` custom resource (which stands for the operator configuration), and optionally install Grafana.

We provide a quick & easy way to deploy Loki (not for production use), Grafana and a `FlowCollector` with default values:

```bash
make deploy-loki deploy-grafana deploy-sample-cr
```

It will set up a local port-forward to Grafana and Loki. To avoid it, pass `PORT_FWD=false` with the command above.

Creating a `FlowCollector` triggers the operator deploying the monitoring pipeline:

- Configures IPFIX exports
- Deploys the flow collector pods, `flowlogs-pipeline`
- Deploys the `network-observability-plugin` for OpenShift console (when used in OpenShift)

You should shortly see flows coming in Grafana or the OpenShift Console ([if not, wait at least 10 minutes](./README.md#faq--troubleshooting)).

### Test another one's pull request

To test a pull request opened by someone else, you just need to pull it locally. Using [GitHub CLI](https://cli.github.com/) is an easy way to do it. Then repeat the steps mentioned above to build, push an image, then deploy the operator and its custom resource.

## Deploy a specific image

Images are built and pushed through CI to [quay.io](https://quay.io/repository/netobserv/network-observability-operator?tab=tags).

You can refer to existing commits using their short-SHA as the image tag, or refer to existing releases. E.g:

```bash
# By commit SHA
VERSION="960766c" make deploy
# By release
VERSION="0.1.2" make deploy
```

Beware that, by referring to an old image, you increase chances to hit breaking changes with the other underlying components, such as [Flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline). It is recommended to switch to the corresponding release GIT tag before deploying an old version, to make sure underlying components refer to correct versions.

## Installing Kafka

Kafka can be used to separate flow ingestion from flow transformation. The operator does not manage kafka deployment and topic creation. We provide a quick setup for Kafka using the [strimzi operator](https://strimzi.io/).

```bash
make deploy-kafka
```

Kafka can then be enabled in the `FlowCollector` CR. If Kafka was deployed using the Makefile, switching the `kafka.enable` flag to `true` in the sample file should be enough. Otherwise, the Kafka address and topic name should be configured.

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
operator-sdk bundle validate $BUNDLE_IMG
```

> Note: the base64 logo can be generated with: `base64 -w 0 <image file>`

### Deploy as bundle from command line

This mode is recommended to quickly test the operator during its development:

```bash
operator-sdk run bundle $BUNDLE_IMG
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
