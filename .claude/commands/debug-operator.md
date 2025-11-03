# Debug NetObserv Operator Locally

This command prepares the environment for debugging the operator locally.

**Prerequisites:** You must have already set up the development environment using `/setup-dev-env` before running this command.

## Steps

### 1. Scale down the in-cluster operator

To debug locally, scale down the operator running in the cluster:

```bash
kubectl scale deployment netobserv-controller-manager -n netobserv --replicas=0
```

### 2. Remove the validating webhook for local development

```bash
kubectl delete validatingwebhookconfiguration netobserv-validating-webhook-configuration
```

This allows you to freely modify the FlowCollector CR while running the operator locally.

### 3. Run the operator locally

**Now ask the user which option they prefer:**

- **Option A: Run with go run** (simple execution without debugging)
- **Option B: Run with delve** (interactive debugging in terminal with dlv commands)
- **Option C: Run with delve headless + VSCode** (debugging with VSCode UI and breakpoints)

**After the user responds:**

#### If Option A (go run):
Execute this command:
```bash
go run ./main.go \
  -ebpf-agent-image=quay.io/netobserv/netobserv-ebpf-agent:main \
  -flowlogs-pipeline-image=quay.io/netobserv/flowlogs-pipeline:main \
  -console-plugin-image=quay.io/netobserv/network-observability-console-plugin:main \
  -namespace=netobserv
```

#### If Option B (delve interactive):
Execute this command:
```bash
dlv debug ./main.go -- \
  -ebpf-agent-image=quay.io/netobserv/netobserv-ebpf-agent:main \
  -flowlogs-pipeline-image=quay.io/netobserv/flowlogs-pipeline:main \
  -console-plugin-image=quay.io/netobserv/network-observability-console-plugin:main \
  -namespace=netobserv
```

This will start an interactive delve session where you can use commands like:
- `break main.main` - set breakpoint
- `continue` - continue execution
- `next` - step to next line
- `print variable` - inspect variables

#### If Option C (delve headless + VSCode):
**Step 1:** Execute this command to start Delve in headless mode:
```bash
dlv debug ./main.go --headless --listen=:2345 --api-version=2 --accept-multiclient -- \
  -ebpf-agent-image=quay.io/netobserv/netobserv-ebpf-agent:main \
  -flowlogs-pipeline-image=quay.io/netobserv/flowlogs-pipeline:main \
  -console-plugin-image=quay.io/netobserv/network-observability-console-plugin:main \
  -namespace=netobserv
```

**Step 2:** Then inform the user:
"Delve is now running in headless mode on port 2345. To connect from VSCode:
1. Open the Debug view (Ctrl+Shift+D / Cmd+Shift+D)
2. Select 'Connect to Delve (Remote - Port 2345)' from the dropdown
3. Press F5 or click 'Start Debugging'

You can now set breakpoints in VSCode and debug the operator!"

### 4. Setting your first breakpoint - FlowCollector Reconcile

**Example: Set a breakpoint in the FlowCollector reconciler**

The FlowCollector reconciliation logic is located in:
`internal/controller/flowcollector/flowcollector_controller.go`

**In VSCode:**
1. Open the file `internal/controller/flowcollector/flowcollector_controller.go`
2. Find the `Reconcile` function (around line 100-150)
3. Click on the left margin next to the line number where you want to pause (e.g., the first line inside the Reconcile function)
4. A red dot will appear indicating the breakpoint is set

**To trigger the reconciliation:**
Execute this command in a terminal to modify the FlowCollector CR:
```bash
kubectl patch flowcollector cluster --type=merge -p '{"spec":{"processor":{"logLevel":"debug"}}}'
```

This will trigger a reconciliation and the debugger will pause at your breakpoint, allowing you to:
- Inspect the FlowCollector object
- Step through the reconciliation logic
- See how the operator processes changes
- Debug any issues in the reconciliation loop

## Using Custom Images

To debug with your own component builds:

```bash
# Build and push your custom images first
cd ../netobserv-ebpf-agent
USER=myuser VERSION=dev make images

cd ../flowlogs-pipeline
USER=myuser VERSION=dev make images

# Then run the operator with your custom images
cd ../network-observability-operator

dlv debug ./main.go --headless --listen=:2345 --api-version=2 --accept-multiclient -- \
  -ebpf-agent-image=quay.io/myuser/netobserv-ebpf-agent:dev \
  -flowlogs-pipeline-image=quay.io/myuser/flowlogs-pipeline:dev \
  -console-plugin-image=quay.io/netobserv/network-observability-console-plugin:main \
  -namespace=netobserv
```

## Cleanup

When you're done debugging:

1. Stop the local operator (Ctrl+C in terminal, or stop debugging in VSCode)

2. Scale the in-cluster operator back up:
```bash
kubectl scale deployment netobserv-controller-manager -n netobserv --replicas=1
```

**Note:** The validating webhook will be automatically recreated by the operator when it starts back up, so you don't need to manually restore it.

## Troubleshooting

- **"Manager already exists" error**: The in-cluster operator is still running. Scale it to 0 replicas.
- **"Connection refused" in VSCode**: Make sure Delve is running first (step 1 of Option C).
- **Permission errors**: Ensure your kubeconfig has proper RBAC permissions.
