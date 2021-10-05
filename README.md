# Network Observability Operator (NOO)

An OpenShift / Kubernetes operator for network observability.

This operator uses the [Go Operator Framework](https://sdk.operatorframework.io/).

...[Work in progress](https://issues.redhat.com/browse/NETOBSERV-46)...

The operator is NOT functional at this stage.

## Build / push / deploy

```bash
# With default image to quay.io/netobserv
make image-build image-push deploy

# With custom image
IMG="docker.io/myuser/netobserv:latest" make image-build image-push deploy
```

By default, the operator is deployed in namespace "network-observability-operator-system".

## Resources

- [Advanced topics](https://sdk.operatorframework.io/docs/building-operators/golang/advanced-topics/) (generic / operator framework) 
