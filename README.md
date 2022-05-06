# Network Observability Operator (NOO)

[![Docker Repository on Quay](https://quay.io/repository/netobserv/network-observability-operator/status "Docker Repository on Quay")](https://quay.io/repository/netobserv/network-observability-operator)

An OpenShift / Kubernetes operator for network observability. It deploys a flow collector (IPFIX standard), an OpenShift console plugin (if working with OpenShift) and it configures the OpenShift Cluster Network Operator to enable flow exports. The `OVN-Kubernetes` CNI is required.

A Grafana dashboard is also provided.

It is also possible to use without OpenShift:
- Using the upstream [ovn-kubernetes](https://github.com/ovn-org/ovn-kubernetes/) with any supported Kubernetes flavour.
- If you don't use ovn-kubernetes but still can manage having IPFIX exports by a different mean, you're more on your own, but still should be able to use this operator. You will need to configure the IPFIX export to push flows to the `flowlogs-pipeline` component deployed by this operator. You could also consider using [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline) directly.

Managed components are deployed in a namespace configured via a Custom Resource (see [FlowCollector custom resource](#flowcollector-custom-resource) section below).

## Deploy an existing image

Images are built and pushed through CI to [quay.io](https://quay.io/repository/netobserv/network-observability-operator?tab=tags).

The operator isn't yet bundled for OLM or Openshift Operator Hub, so you need to clone this repo and deploy from there at the moment.

To refer to the latest version of the `main` branch, use `IMG=quay.io/netobserv/network-observability-operator:main` or simply `VERSION=main`. To refer to older versions, use the commit short-SHA as the image tag. By default, `main` will be used.

E.g. to deploy the latest build:

```bash
make deploy
```

You must then install the custom resource, e.g.:

```bash
kubectl apply -f ./config/samples/flows_v1alpha1_flowcollector.yaml
```

## Build / push / deploy

The repository `quay.io/netobserv/network-observability-operator` is only writable by the CI, so you need to use another repository (such as your own one) if you want to use your own build.

For instance, to build from a pull-request, checkout that PR (e.g. using github CLI or `git fetch upstream pull/99/head:pr-99 && git checkout pr-99` (replace `99` with the PR ID)), then run:

```bash
IMG="quay.io/youraccount/network-observability-operator:v0.0.1" make image-build image-push deploy
```

Note, the default image pull policy is `IfNotPresent`, so if you previously deployed the operator on a cluster and then create another build with the same image name/tag, it won't be pulled in the cluster registry. So you need either to provide a different image name/tag for every build, or modify [manager.yaml](./config/manager/manager.yaml) to set `imagePullPolicy: Always`, then re-deploy.

Then, you can deploy a custom resource, e.g.:

```bash
kubectl apply -f ./config/samples/flows_v1alpha1_flowcollector.yaml

# or using make
make deploy-sample-cr
```

## Deploy as bundle

For more details, refer to the [Operator Lifecycle Manager (OLM) bundle quickstart documentation](https://sdk.operatorframework.io/docs/olm-integration/quickstart-bundle/).

This task should be automatically done by the CI/CD pipeline. However, if you want to deploy as
bundle for local testing, you should execute the following commands:

```
export USER=<container-registry-username>
export VERSION=0.0.1
export IMG=quay.io/$USER/network-observability-operator:v$VERSION
export BUNDLE_IMG=quay.io/$USER/network-observability-operator-bundle:v$VERSION
make image-build image-push
make bundle bundle-build bundle-push
```

Optionally, you might validate the bundle:
```
operator-sdk bundle validate $BUNDLE_IMG
```

Note: the base64 logo can be generated with: `base64 -w 0 <image file>`

### Deploy as bundle from command line

This mode is recommended to quickly test the operator during its development:

```
operator-sdk run bundle $BUNDLE_IMG
```

### Deploy as bundle from the Console's OperatorHub page

This mode is recommended when you want to test the customer experience of navigating through the
operators' catalog and installing/configuring it manually through the UI.

First, create and push a catalog image:

```
export CATALOG_IMG=quay.io/$USER/network-observability-operator-catalog:v$VERSION
make catalog-build catalog-push catalog-deploy
```

The NetObserv Operator is available in OperatorHub: https://operatorhub.io/operator/netobserv-operator

## Publish on central OperatorHub

See [RELEASE.md](./RELEASE.md#publishing-on-operatorhub).

## FlowCollector custom resource

The `FlowCollector` custom resource is used to configure the operator and its managed components. You can read its [full documentation](https://github.com/netobserv/network-observability-operator/blob/main/docs/FlowCollector.md) and check this [sample file](./config/samples/flows_v1alpha1_flowcollector.yaml) that you can copy, edit and install.

Note that the `FlowCollector` resource must be unique and must be named `cluster`. It applies to the whole cluster.

## Enabling OVS IPFIX export

If you use OpenShift 4.10 or the upstream ovn-kubernetes without OpenShift, you don't have anything to do: the operator will configure OVS *via* OpenShift Cluster Network Operator, or the ovn-kubernetes layer directly.

Else if you use OpenShift 4.8 or 4.9, some manual steps are still required

```bash
FLP_IP=`oc get svc flowlogs-pipeline -n network-observability -ojsonpath='{.spec.clusterIP}'` && echo $FLP_IP
oc patch networks.operator.openshift.io cluster --type='json' -p "[{'op': 'add', 'path': '/spec', 'value': {'exportNetworkFlows': {'ipfix': { 'collectors': ['$FLP_IP:2055']}}}}]"
```

OpenShift versions older than 4.8 don't support IPFIX exports.

## Installing Loki

Loki is used to store the flows, however its installation is not managed directly by the operator. There are several options to install Loki, like using the `loki-operator` or the helm charts. Get some help about it on [this page](https://github.com/netobserv/documents/blob/main/hack_loki.md).

Once Loki is setup, you may have to update the `flowcollector` CR to update the Loki URL (use an URL that is accessible in-cluster by the `flowlogs-pipeline` pods; default is `http://loki:3100/`).

## OpenShift Console plugin

The operator deploys a console dynamic plugin when used in OpenShift, and should register it automatically if `spec.consolePlugin.register` is set to `true` (default).

If it's set to `false`, or if for any reason the registration failed, you can still do it manually by editing
`console.operator.openshift.io/cluster` to add the plugin reference:

```yaml
spec:
  plugins:
  - network-observability-plugin
```

Or simply execute:

```bash
oc patch console.operator.openshift.io cluster --type='json' -p '[{"op": "add", "path": "/spec/plugins/-", "value": "network-observability-plugin"}]'
```

The plugin provides new views in the OpenShift Console: a new submenu _Network Traffic_ in _Observe_, and new tabs in several details views (Pods, Deployments, Services...).

![Main view](./docs/assets/network-traffic-main.png)
_Main view_

![Pod traffic](./docs/assets/network-traffic-pod.png)
_Pod traffic_

### Grafana dashboard

Grafana can be used to retrieve and show the collected flows from Loki. You can [find here](https://github.com/netobserv/documents/blob/main/hack_loki.md#grafana) some help to install Grafana if needed.

Then import [this dashboard](./config/samples/dashboards/Network%20Observability.json) in Grafana. It includes a table of the flows and some graphs showing the volumetry per source or destination namespaces or workload:

![Grafana dashboard](./docs/assets/netobserv-grafana-dashboard.png)
