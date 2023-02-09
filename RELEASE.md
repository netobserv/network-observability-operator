## Releasing process

### Sub-components

All components deployed by this operator can be released separatly, at their own pace.

Before releasing, it's a good opportunity to check for image upgrades: [go-toolset](https://catalog.redhat.com/software/containers/ubi9/go-toolset/61e5c00b4ec9945c18787690), [node.js](https://catalog.redhat.com/software/containers/ubi9/nodejs-16/61a60604c17162a20c1c6a2e) and [ubi9-minimal](https://catalog.redhat.com/software/containers/ubi9-minimal/61832888c0d15aff4912fe0d).

To release them, a tag in the format "v0.1.2" or "v0.1.2-rc0" must be set on the desired clean HEAD state (generally, up-to-date `main` branch), then pushed. It applies to [the console plugin](https://github.com/netobserv/network-observability-console-plugin/), [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline) and [netobserv-ebpf-agent](https://github.com/netobserv/netobserv-ebpf-agent).

E.g:

```bash
version="v0.1.2-rc0"
git tag -a "$version" -m "$version"
git push upstream --tags
```

The release script should be triggered (check github actions).

When the last release candidate is accepted and the final release tag is pushed (using the same procedure), you can generate the release via the github interface:
- [console plugin](https://github.com/netobserv/network-observability-console-plugin/releases/new)
- [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline/releases/new)
- [netobserv-ebpf-agent](https://github.com/netobserv/netobserv-ebpf-agent/releases/new)

Click the "Auto-generate release note" button.

### Operator

Once all sub-components are released (or have a release candidate), we can proceed with the operator.

Edit the [Makefile](./Makefile) to update `PREVIOUS_VERSION`, `BUNDLE_VERSION`, `PLG_VERSION`, `FLP_VERSION` and `BPF_VERSION`.

```bash

make bundle

# Set desired operator version - CAREFUL, no leading "v" here
version="1.0.1"
vv=v$version
test_branch=test-$vv

git commit -a -m "Prepare release $vv"
# Push to a test branch, and tag for release
git push upstream HEAD:$test_branch
git tag -a "$version" -m "$version"
git push upstream --tags
```

The release script should be triggered ([check github actions](https://github.com/netobserv/network-observability-operator/actions)).

At this point, you can test the bundle / catalog on your cluster:

```bash
BUNDLE_VERSION="$version" make catalog-deploy
```

When everything is ok, push to main and delete the test branch

```bash
git push upstream HEAD:main
git push upstream :$test_branch
```

When the last release candidate is accepted and the final release tag is pushed (using the same procedure), you can generate the release via the github interface:
- [operator](https://github.com/netobserv/network-observability-operator/releases/new)

Click the "Auto-generate release note" button.

Add links to sub-component release notes, e.g:

```md
## Sub-component release notes:

* eBPF Agent: https://github.com/netobserv/netobserv-ebpf-agent/releases/tag/v0.1.2
* Flowlogs-pipeline: https://github.com/netobserv/flowlogs-pipeline/releases/tag/v0.1.3
* Console plugin: https://github.com/netobserv/network-observability-console-plugin/releases/tag/v0.1.4
```

Check also the "Create a discussion for this release" option, in category "Announcements".

### Testing the upgrade path

Before publishing, we should check that upgrading the operator from a previous version isn't broken. We can use `operator-sdk` for that:

```bash
previous=v0.2.2
bin/operator-sdk run bundle quay.io/netobserv/network-observability-operator-bundle:$previous --timeout 5m
bin/operator-sdk run bundle-upgrade quay.io/netobserv/network-observability-operator-bundle:$vv --timeout 5m
```

Note: currently, [seamless upgrade](https://sdk.operatorframework.io/docs/overview/operator-capabilities/#level-2---seamless-upgrades) is not fully supported because an existing custom resource needs first to be deleted before the operator is upgraded. See also: https://issues.redhat.com/browse/NETOBSERV-521.

If you need to repeat the operation several times, make sure to cleanup between attempts:

```bash
operator-sdk cleanup netobserv-operator
```


### Publishing on OperatorHub

First, do some manual cleanup. Ideally these steps should be included in the `make bundle` process (TODO).
- In [bundle.Dockerfile](./bundle.Dockerfile), remove the two "Labels for testing" and the `scorecard` reference.
- In [bundled annotations.yaml](./bundle/metadata/annotations.yaml), remove the two annotations for testing.

There's a cross-publication on two repos:
- For non-OpenShift: https://github.com/k8s-operatorhub/community-operators
- For OpenShift / community operators: https://github.com/redhat-openshift-ecosystem/community-operators-prod

After having cloned or updated these repo, copy the bundle content:

```bash
# Here, set correct paths and new version
path_k8s="../community-operators"
path_okd="../community-operators-prod"

cd $path_k8s && git fetch upstream && git rebase upstream/main
cd -
cd $path_okd && git fetch upstream && git rebase upstream/main
cd -

mkdir -p $path_k8s/operators/netobserv-operator/$version
mkdir -p $path_okd/operators/netobserv-operator/$version
cp -r "bundle/manifests" "$path_k8s/operators/netobserv-operator/$version"
cp -r "bundle/manifests" "$path_okd/operators/netobserv-operator/$version"
cp -r "bundle/metadata" "$path_k8s/operators/netobserv-operator/$version"
cp -r "bundle/metadata" "$path_okd/operators/netobserv-operator/$version"
```

You should double check eveything is correct by comparing the produced files with their equivalent in the previous release,
making sure there's nothing unexpected.

On OKD repo, edit the annotations.yaml to include the supported OpenShift versions:

```yaml
  # OpenShift annotations.
  com.redhat.openshift.versions: "v4.10-v4.12"
```

Then commit and push (commits must be signed):

```bash
cd $path_k8s
git add -A
git commit -s -m "operators netobserv-operator ($version)"
git push origin HEAD:bump-$version

cd $path_okd
git add -A
git commit -s -m "operators netobserv-operator ($version)"
git push origin HEAD:bump-$version
```

Open PRs in GitHub. A bunch of tests will be triggered, if passed the merge should be automatic.

### Writing a release summary

Create an announcement on the [dedicated page](https://github.com/netobserv/network-observability-operator/discussions/categories/announcements) with the main lines of what changed (and links to detailed release notes per component).

As a reference, we can use this template:

```md
## What's Changed

### NetObserv Operator (va.b.c -> vx.y.z)
- Feature 1
- Etc. ([view details](https://github.com/netobserv/network-observability-operator/releases/tag/x.y.z))

### Console Plugin (va.b.c -> vx.y.z)
- Feature 1
- Etc. ([view details](https://github.com/netobserv/network-observability-console-plugin/releases/tag/vx.y.z))

### FlowLogs Pipeline (va.b.c -> vx.y.z)
- Feature 1
- Etc. ([view details](https://github.com/netobserv/flowlogs-pipeline/releases/tag/vx.y.z))

### eBPF Agent (va.b.c -> vx.y.z)
- Feature 1
- Etc. ([view details](https://github.com/netobserv/netobserv-ebpf-agent/releases/tag/vx.y.z))
```

Don't forget credits involving external contributors!
