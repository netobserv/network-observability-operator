# Network Observability Operator (NOO)

An OpenShift / Kubernetes operator for network observability.

This operator uses the [Go Operator Framework](https://sdk.operatorframework.io/).

...[Work in progress](https://issues.redhat.com/browse/NETOBSERV-46)...

The operator is NOT functional at this stage.

By default, the operator is deployed in namespace "network-observability".

## Deploy an existing image

Images are built and pushed through CI to [quay.io](https://quay.io/repository/netobserv/network-observability-operator?tab=tags). To refer to the latest version of the `main` branch, use `IMG=quay.io/netobserv/network-observability-operator:main` or simply `VERSION=main`. To refer to older versions, use the commit short-SHA as the image tag. By default, `main` will be used.

E.g. to deploy the latest build:

```bash
make deploy
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
```

### Enabling OVS IPFIX export

This part will eventually be done automatically by the operator, but for the time being it requires manual intervention.

#### On KIND

```bash
GF_IP=`kubectl get svc goflow-kube -n network-observability -ojsonpath='{.spec.clusterIP}'` && echo $GF_IP
kubectl set env daemonset/ovnkube-node -c ovnkube-node -n ovn-kubernetes OVN_IPFIX_TARGETS="$GF_IP:2055"
```

#### On OpenShift

In OpenShift, a difference with the upstream `ovn-kubernetes` is that the flows export config is managed by the `ClusterNetworkOperator`.

```bash
GF_IP=`oc get svc goflow-kube -n network-observability -ojsonpath='{.spec.clusterIP}'` && echo $GF_IP
oc patch networks.operator.openshift.io cluster --type='json' -p "$(sed -e "s/GF_IP/$GF_IP/" ./config/samples/net-cluster-patch.json)"
```

### Enabling the console plugin

The plugin automatically deploy an OpenShift console dynamic plugin.

The plugin then needs to be enabled through the console configuration:

```
$ oc edit console.operator.openshift.io cluster
```

```
spec:
  plugins:
  - network-observability-plugin
```

## Resources

- [Advanced topics](https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/) (generic / operator framework)
