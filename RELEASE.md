## Releasing process

### Sub-components

All components deployed by this operator can be released separatly, at their own pace.

Before releasing, it's a good opportunity to check for image upgrades: [go-toolset](https://catalog.redhat.com/software/containers/ubi8/go-toolset/5ce8713aac3db925c03774d1), [node.js](https://catalog.redhat.com/software/containers/ubi8/nodejs-14/5ed7887dd70cc50e69c2fabb) and [ubi8-minimal](https://catalog.redhat.com/software/containers/ubi8-minimal/5c64772edd19c77a158ea216).

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
# Set desired operator version - CAREFUL, no leading "v" here
version="0.1.3-rc0"
# Set console plugin released version
plgv="v0.1.3-rc1"
# Set flowlogs-pipeline released version
flpv="v0.1.2-rc2"
# Set ebnpf-agent released version
bpfv="v0.1.1-rc0"

vv=v$version
test_branch=test-$vv

VERSION="$version" PLG_VERSION="$plgv" FLP_VERSION="$flpv" BPF_VERSION="$bpfv" IMAGE_TAG_BASE="quay.io/netobserv/network-observability-operator" make bundle

git commit -a -m "Prepare release $vv"
# Push to a test branch, and tag for release
git push upstream HEAD:$test_branch
git tag -a "$version" -m "$version"
git push upstream --tags
```

The release script should be triggered ([check github actions](https://github.com/netobserv/network-observability-operator/actions)).

At this point, you can test the bundle / catalog on your cluster:

```bash
VERSION="$version" make catalog-deploy
```

When everything is ok, push to main and delete the test branch

```bash
git push upstream HEAD:main
git push upstream :$test_branch
```

When the last release candidate is accepted and the final release tag is pushed (using the same procedure), you can generate the release via the github interface:
- [operator](https://github.com/netobserv/network-observability-operator/releases/new)

Click the "Auto-generate release note" button.

### Publishing on OperatorHub

First, do some manual cleanup. Ideally these steps should be included in the `make bundle` process (TODO).
- In `bundle.Dockerfile`, remove the two "Labels for testing" and the `scorecard` reference.
- In `bundle/metadata/annotations.yaml`, remove the two annotations for testing.
- In the CSV file, bump the `replaces` field so that it points to the previous version.

There's a cross-publication on two repos:
- For non-OpenShift: https://github.com/k8s-operatorhub/community-operators
- For OpenShift / community operators: https://github.com/redhat-openshift-ecosystem/community-operators-prod

After having cloned or updated these repo, copy the bundle content:

```bash
# Here, set correct paths and new version
path_k8s="../community-operators"
path_okd="../community-operators-prod"
version="0.1.2"

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
