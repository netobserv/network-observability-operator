# Setup NetObserv Development Environment

This command sets up a complete development environment for NetObserv.

**First, check if the user specified custom images:**
- Did the user provide a **USER** (quay.io username/repo)?
- Did the user provide a **VERSION** (image tag)?

If the user specified custom images (e.g., "using the image from my repo, leandroberetta with tag v1.2.3"), extract those values. Otherwise, use the defaults: `USER=netobserv` and `VERSION=main`.

**Next, determine your platform:**
- Are you using **OpenShift** or **vanilla Kubernetes**?

This will determine whether to run `make set-release-kind-downstream` in step 2.

## Steps

### 1. Deploy the operator to the cluster

Deploy the operator with the specified or default images:

**If custom USER and VERSION were provided:**
```bash
USER=<user_value> VERSION=<version_value> make deploy
```

**If using defaults (no custom images specified):**
```bash
USER=netobserv make deploy
```

**Note:**
- Using `USER=netobserv` (default) ensures the operator uses public images from `quay.io/netobserv` without authentication.
- Custom USER values like `USER=leandroberetta` will use images from `quay.io/leandroberetta/`.
- The VERSION parameter controls the image tag (e.g., `VERSION=v1.2.3` or `VERSION=main`).

### 2. Configure for OpenShift (if applicable)

**Only if you're using OpenShift**, run:

```bash
make set-release-kind-downstream
```

**If you're using vanilla Kubernetes**, skip this step entirely.

### 3. Deploy Loki

```bash
make deploy-loki
```

This will deploy Loki and make it available at http://localhost:3100

### 4. Deploy sample FlowCollector CR

```bash
make deploy-sample-cr
```

### 5. Verify deployment

Check that all components are running:

```bash
kubectl get pods -n netobserv
kubectl get flowcollector cluster -o yaml
```

## Cleanup

To clean up the entire environment:

```bash
make undeploy
```