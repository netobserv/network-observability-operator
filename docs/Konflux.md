# NetObserv Operator Konflux builds

> [!WARNING]
> This documentation is about the downstream CI/CD used to build and release Network Observability product in Openshift.
> Some of the links here only work with Red Hat vpn.

## Links

Useful links

- [NetObserv konflux console](https://console.redhat.com/application-pipeline/workspaces/ocp-network-observab/applications)
- [Konflux documentation](https://konflux.pages.redhat.com/docs/users/)
- [Konflux build task reference](https://github.com/konflux-ci/build-definitions/tree/main/task) contains the definitions of tasks that can be used in konflux pipeline
- [Konflux release configuration](https://gitlab.cee.redhat.com/releng/konflux-release-data) contains release configuration
- [Konflux sample project](https://github.com/konflux-ci/olm-operator-konflux-sample) is a sample project built as an example on how to use konflux
- [Lightspeed operator](https://github.com/openshift/lightspeed-operator/blob/main/.tekton/fbc-v4-15-pull-request.yaml) probably one of the first project to release with konflux

## Builds

Builds are defined in the .tekton/ directory of each project, to ensure pre-merge build and post merge buids are identical and make maintenance easier, pipeline references has been centralized in a dediceted pipeline-ref file.

### Nudging

Nudging is the konflux mecanism to create dependencies and update between components. When a component A is nudging a component B, Konflux will automatically create PR to update component A reference in component B.

FLP, the ebpf-agent, the console plugin and the operator component are nudging the bundle build once finished and the bundle is nudging the FBC build once finished.

### File Based Catalog

When using Konflux to release an operator, it is required to use a File Based Catalog image. This imply some changes:
- FBC will not be additive to another one, the released FBC must contain all previous versions
- one FBC version must be built per Openshift version, the base image used will define the targeted Openshift version
- FBC build must be the only component in an application

### Konflux pull requests

Konflux will regulary create new pull requests, there are three categories :

- [Nudging pull request](https://github.com/netobserv/network-observability-operator/pull/969) To upgrade the reference to another component
- [Konflux tasks update](https://github.com/netobserv/network-observability-operator/pull/787) Up to date tasks are required to pass security check during release. Also the migration note sometimes contains instruction to some required actions
- [Dependencies update](https://github.com/netobserv/network-observability-operator/pull/962) Kondlux internally use [https://github.com/renovatebot/renovate](renovate) to automatically create this PR.

## Deploying

An `ImageDigestMirrorSet` is required:

```yaml
apiVersion: config.openshift.io/v1
kind: ImageDigestMirrorSet
metadata:
  name: netobserv-image-digest-mirror-set
spec:
  imageDigestMirrors:
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-operator-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-operator-zstream
      source: registry.redhat.io/network-observability/network-observability-rhel9-operator
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/flowlogs-pipeline-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/flowlogs-pipeline-ztream
      source: registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/netobserv-ebpf-agent-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/netobserv-ebpf-agent-zstream
      source: registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-console-plugin-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-console-plugin-ztream
      source: registry.redhat.io/network-observability/network-observability-console-plugin-rhel9
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-cli-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-cli-zstream
      source: registry.redhat.io/network-observability/network-observability-cli-rhel9
    - mirrors:
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-operator-bundle-ystream
      - quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-operator-bundle-zstream
      source: registry.redhat.io/network-observability/network-observability-operator-bundle
```

The testing FBC image can be added as a CatalogSource:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: netobserv-konflux
  namespace: openshift-marketplace
spec:
  displayName: netobserv-konflux
  image: 'quay.io/redhat-user-workloads/ocp-network-observab-tenant/catalog-ystream:latest'
  # for z-stream, use instead:
  # image: 'quay.io/redhat-user-workloads/ocp-network-observab-tenant/catalog-zstream:latest'
  publisher: Netobserv team
  sourceType: grpc
```

## Release pipeline setup

[Konflux release configuration](https://gitlab.cee.redhat.com/releng/konflux-release-data) define the release process for konflux built projects.

For new releases two differents directory are important :
- The [ReleasePlanAdmission](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/config/stone-prd-rh01.pg1f.p1/product/ReleasePlanAdmission/ocp-network-observab) directory. The `ReleasePlanAdmission` defines the images that are going to be released and the release target.

- The [ReleasePlan](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config/cluster/stone-prd-rh01/tenants/ocp-network-observab-tenant?ref_type=heads) directory. The `ReleasePlan` object define which `ReleasePlanAdmission` are going to be part of the release.

To be able to see release pipeline, a read access to the `rhtap-releng` namespace is required, this access must be requested in the konflux-user slack channel.

### Branching

After creating a new release branch, the following steps need to be done:
- create a new konflux component
- create new component inside the new application, one for each image except the FBC
- edit the new component build pipeline to point to the pipeline-ref file
- creating the new `ReleasePlanAdmission` objects, one for staging one for production
- creating the new `ReleasePlan` objects, one for staging, one for production, note that the `auto-release` label in the production file must be false

## Release

Once it is ready to be released, a new `Release` object needs to be created to trigger the production release pipeline:

```yaml
apiVersion: appstudio.redhat.com/v1alpha1
kind: Release
metadata:
  name: release-netobserv-1-8-0-0                      # name+version - last digit is the attempt number, in case you need to retry
  namespace: ocp-network-observab-tenant
  labels:
    release.appstudio.openshift.io/author: 'jtakvori'  # your konflux / redhat user
spec:
  releasePlan: netobserv-1-8
  snapshot: netobserv-operator-1-8-9ms9w               # the validated snapshot
```

It must be created on the OCP instance that runs Konflux (ask for the address if you don't have it).

For the record, store the created Release in the `releases` directory of this repo.

## After release

After a release, the following steps should be done:
- moving release candidate file in the catalog to the already released directory
- bumping version label inside the different conteiner images
