## Releasing process

### Sub-components

All components deployed by this operator can be released separatly, at their own pace.
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
version="0.1.2-rc0"
# Set console plugin released version
plgv="v0.1.2-rc0"
# Set flowlogs-pipeline released version
flpv="v0.1.2-rc0"

vv=v$version
test_branch=test-$vv

VERSION="$version" PLG_VERSION="$plgv" FLP_VERSION="$flpv" IMAGE_TAG_BASE="quay.io/netobserv/network-observability-operator" make bundle

git commit -a -m "Prepare release $vv"
# Push to a test branch, and tag for release
git push upstream HEAD:$test_branch
git tag -a "$version" -m "$version"
git push upstream --tags
```

The release script should be triggered ([check github actions](https://github.com/netobserv/network-observability-operator/actions)).

At this point, you can test the bundle / catalog on your cluster:

```bash
# Set user to point to your quay account.
user=your_name
VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make bundle-build bundle-push catalog-build catalog-push catalog-deploy
```

Other users (e.g. QE) can use the same catalog:
```bash
VERSION="$version" IMAGE_TAG_BASE="quay.io/$user/network-observability-operator" make catalog-deploy
```

When everything is ok, push to main and delete the test branch

```bash
git push upstream HEAD:main
git push upstream :$test_branch
```


When the last release candidate is accepted and the final release tag is pushed (using the same procedure), you can generate the release via the github interface:
- [operator](https://github.com/netobserv/network-observability-operator/releases/new)

Click the "Auto-generate release note" button.
