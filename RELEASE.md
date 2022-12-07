## Releasing process

### Sub-components

All components deployed by this operator can be released separatly, at their own pace.

Before releasing, it's a good opportunity to check for image upgrades: [go-toolset](https://catalog.redhat.com/software/containers/ubi8/go-toolset/5ce8713aac3db925c03774d1), [node.js](https://catalog.redhat.com/software/containers/ubi8/nodejs-16/615aee9fc739c0a4123a87e1) and [ubi8-minimal](https://catalog.redhat.com/software/containers/ubi8-minimal/5c64772edd19c77a158ea216).

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

```bash
# Previous operator version
previous="v0.2.0"
# Set desired operator version - CAREFUL, no leading "v" here
version="0.2.1"
# Set console plugin released version
plgv="v0.1.6"
# Set flowlogs-pipeline released version
flpv="v0.1.5"
# Set ebpf-agent released version
bpfv="v0.2.2"

vv=v$version
test_branch=test-$vv

BUNDLE_VERSION="$version" PLG_VERSION="$plgv" FLP_VERSION="$flpv" BPF_VERSION="$bpfv" PREVIOUS_VERSION="$previous" make bundle

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
# NOTE: on my last try, I needed to pass an index-image that corresponds to the operator-sdk version. This is likely due to a bug and should be eventually removed (cf https://github.com/operator-framework/operator-sdk/issues/5980)
operator-sdk run bundle quay.io/netobserv/network-observability-operator-bundle:$previous --index-image quay.io/operator-framework/opm:v1.22 --timeout 5m
operator-sdk run bundle-upgrade quay.io/netobserv/network-observability-operator-bundle:$vv --timeout 5m
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
version="0.2.0"

mkdir -p $path_k8s/operators/netobserv-operator/$version
mkdir -p $path_okd/operators/netobserv-operator/$version
cp "bundle.Dockerfile" "$path_k8s/operators/netobserv-operator/$version"
# no bundle.Dockerfile for openshift's repo
cp -r "bundle/manifests" "$path_k8s/operators/netobserv-operator/$version"
cp -r "bundle/manifests" "$path_okd/operators/netobserv-operator/$version"
cp -r "bundle/metadata" "$path_k8s/operators/netobserv-operator/$version"
cp -r "bundle/metadata" "$path_okd/operators/netobserv-operator/$version"
```

You should double check eveything is correct by comparing the produced files with their equivalent in the previous release,
making sure there's nothing unexpected.

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
