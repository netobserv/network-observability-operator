## Releasing process

### Sub-components

All components deployed by this operator can be released separatly, at their own pace.

Before releasing, it's a good opportunity to check for image upgrades: Go, [node.js](https://catalog.redhat.com/software/containers/ubi9/nodejs-18/62e8e7ed22d1d3c2dfe2ca01) and [ubi9-minimal](https://catalog.redhat.com/software/containers/ubi9-minimal/61832888c0d15aff4912fe0d).

To release them, a tag in the format "v1.6.0-community" or "v1.6.0-crc0" must be set on the desired clean HEAD state (generally, up-to-date `main` branch; "crc" stands for "community release candidate"), then pushed. It applies to [the console plugin](https://github.com/netobserv/network-observability-console-plugin/), [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline) and [netobserv-ebpf-agent](https://github.com/netobserv/netobserv-ebpf-agent).

E.g:

```bash
version="v1.9.1-community"
git tag -a "$version" -m "$version"
git push upstream --tags
```

The release script should be triggered (check github actions). It will automatically draft a new release, with artifacts attached.

If the release looks good, you can publish it via the github interface:
- [console plugin](https://github.com/netobserv/network-observability-console-plugin/releases)
- [flowlogs-pipeline](https://github.com/netobserv/flowlogs-pipeline/releases)
- [netobserv-ebpf-agent](https://github.com/netobserv/netobserv-ebpf-agent/releases)

Edit the draft, set the previous tag then click the "Generate release notes" button.

If you think the "Dependencies" section is too long, you can surround it in a `<details>` block, to make it collapsed. E.g:

```yaml
<details>
<summary><b>Dependencies</b></summary>

* Bump [...] from [...] by @dependabot in...
* ...
</details>
```


### Operator

Once all sub-components are released (or have a release candidate), we can proceed with the operator.

Edit the [Makefile](./Makefile) to update `PREVIOUS_VERSION`, `BUNDLE_VERSION`, `PLG_VERSION`, `FLP_VERSION` and `BPF_VERSION`.

```bash
make update-bundle

# Set desired operator version - CAREFUL, no leading "v" here
version="1.9.1-community"
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
BUNDLE_VERSION="$version" USER=netobserv make catalog-deploy
```

When everything is ok, push to main and delete the test branch

```bash
git push upstream HEAD:main
git push upstream :$test_branch
```

If the release looks good, you can publish it via the github interface:
- [operator](https://github.com/netobserv/network-observability-operator/releases)

Edit the draft, set the previous tag then click the "Generate release notes" button.

Grab related components release notes by running:

```bash
make related-release-notes
```

The script should fetch and copy the content in the clipboard. Paste it at the end of the auto-generated release note in GitHub.

If you think the "Dependencies" section is too long, you can surround it in a `<details>` block, to make it collapsed. E.g:

```yaml
<details>
<summary><b>Dependencies</b></summary>

* Bump github.com/netobserv/flowlogs-pipeline from [...] by @dependabot in...
* ...
</details>
```

Check the "Create a discussion for this release" option, in category "Announcements".

Click on "Publish release".

### Testing HELM

Create a kind cluster, then:

```bash
helm repo add cert-manager https://charts.jetstack.io
helm install my-cert-manager cert-manager/cert-manager --set crds.enabled=true

helm install my-netobserv -n netobserv --create-namespace --set standaloneConsole.enable=true --set install.loki=true --set install.prom=true ./helm

cat <<EOF | kubectl apply -f -
apiVersion: flows.netobserv.io/v1beta2
kind: FlowCollector
metadata:
  name: cluster
spec:
  namespace: netobserv
  consolePlugin:
    advanced:
      env:
        TEST_CONSOLE: "true"
  loki:
    mode: Monolithic
    monolithic:
      url: 'http://my-netobserv-loki.netobserv.svc:3100/'
  prometheus:
    querier:
      manual:
        url: http://my-netobserv-prometheus-server.netobserv.svc/
EOF

# Check components image:
kubectl get pods -oyaml | grep image:

kubectl port-forward svc/netobserv-plugin 9001:9001 -n netobserv
```

Then open http://localhost:9001/ in your browser.

To clean up:

```bash
helm delete my-netobserv -n netobserv
```

### Publishing

```bash
helm package helm/
mv netobserv-operator-1.9.1.tgz /path/to/netobserv.github.io/static/helm/
cd /path/to/netobserv.github.io/static/
helm repo index helm/ --url https://netobserv.io/static/helm/

# Now, check there's nothing wrong in the generated files before commit
colordiff <(yq '.entries.netobserv-operator[0]' helm/index.yaml) <(yq '.entries.netobserv-operator[1]' helm/index.yaml)

git add helm/netobserv-operator-1.9.1.tgz helm/index.yaml
git commit -m "Publish helm 1.9.1-community"
git push upstream HEAD:main
```

Check ArtifactHub for update after a few minutes: https://artifacthub.io/packages/helm/netobserv/netobserv-operator

## OLD

### Testing the upgrade path

Before publishing, we should check that upgrading the operator from a previous version isn't broken. We can use `operator-sdk` for that:

```bash
previous=v1.6.1-community
bin/operator-sdk run bundle quay.io/netobserv/network-observability-operator-bundle:$previous --timeout 5m
PORT_FWD=false make deploy-loki deploy-sample-cr
bin/operator-sdk run bundle-upgrade quay.io/netobserv/network-observability-operator-bundle:$vv --timeout 5m
```

If you need to repeat the operation several times, make sure to cleanup between attempts:

```bash
bin/operator-sdk cleanup netobserv-operator
```

### Publishing on OperatorHub

There's a cross-publication on two repos:
- For non-OpenShift: https://github.com/k8s-operatorhub/community-operators
- For OpenShift / community operators: https://github.com/redhat-openshift-ecosystem/community-operators-prod

The bundle built in the steps above is set up for OpenShift, so we'll start with that one.

Run:

```bash
# OPERATORHUB_PATH defaults to "../community-operators-prod"
OPERATORHUB_PATH=/path/to/operatorhub-for-okd make prepare-operatorhub
```

You should double check eveything is correct by comparing the produced files with their equivalent in the previous release,
making sure there's nothing unexpected.

If necessary, update the "com.redhat.openshift.versions" annotation for compatible OpenShift versions.

Then commit and push (commits must be signed):

```bash
git add -A
git commit -s -m "operators netobserv-operator ($version)"
git push origin HEAD:bump-$version
```

For non-OpenShift, we need first to generate a non-OpenShift compliant bundle:

```bash
# When BUNDLE_CONFIG is "config/k8s-olm", OPERATORHUB_PATH defaults then to "../community-operators"
OPERATORHUB_PATH=/path/to/operatorhub-for-k8s BUNDLE_CONFIG=config/k8s-olm make update-bundle prepare-operatorhub
```

Then check, commit push push as above.

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
