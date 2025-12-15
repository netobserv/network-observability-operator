# AI Agents Best Practices for Network Observability Operator

Best practices for AI coding agents on NetObserv Operator.

> **Note**: Symlinked as [CLAUDE.md](CLAUDE.md) for Claude Code auto-loading.

## Project Context

**NetObserv Operator** - Kubernetes/OpenShift operator for network observability
(operator-sdk)

**Components:**
- **eBPF Agent**: Network flow generation from packets (DaemonSet)
- **flowlogs-pipeline**: Flow collection, enrichment, export
  (Deployment/StatefulSet)
- **Console Plugin**: OpenShift visualization (optional)
- **CRD**: `FlowCollector` v1beta2 - **single cluster-wide resource named
  `cluster`**
- **Integrations**: Loki, Prometheus, Kafka (optional)

**Key Directories:**
- `api/flowcollector/v1beta2/`: CRD definitions
- `internal/controller/`: Reconciliation logic
- `config/`: Kustomize manifests
- `docs/`: FlowCollector spec, architecture

## Critical Constraints

### 🚨 Single FlowCollector Only
Only ONE FlowCollector allowed, named `cluster`:
```go
if flowCollector.Name != constants.FlowCollectorName {
    return fmt.Errorf("only one FlowCollector allowed, named %s", constants.FlowCollectorName)
}
```

### 🚨 Backward Compatibility
FlowCollector v1beta2 is stable:
- ✅ Add optional fields with defaults, use `+optional` marker
- ❌ Never remove/rename fields or change types

### 🚨 Bundle Updates Required
After CRD/CSV changes: `make update-bundle`

### 🚨 Image References
Never hardcode. Use env vars:
- `RELATED_IMAGE_EBPF_AGENT`
- `RELATED_IMAGE_FLOWLOGS_PIPELINE`
- `RELATED_IMAGE_CONSOLE_PLUGIN`

### 🚨 Multi-Architecture
Support: amd64, arm64, ppc64le, s390x

## Effective Prompting

**Good Example:**
```
Update internal/controller/flowcollector_controller.go to add validation for
spec.agent.ebpf.logLevel (valid: trace, debug, info, warn, error).
Add webhook validation. Include unit tests and run make update-bundle.
```

**Bad Example:**
```
Add log level validation
```

**Key Principles:**
1. Specify file paths explicitly
2. Reference existing patterns
3. Mention testing requirements
4. Check dependencies in go.mod first

## Common Task Templates

### Add FlowCollector Field
```
Add spec.agent.ebpf.newFeature (bool, default: false):
1. Update api/flowcollector/v1beta2/flowcollector_types.go (+kubebuilder markers)
2. Modify internal/controller/ to use field
3. Add unit tests
4. Update docs/FlowCollector.md
5. Run make update-bundle
```

### Update Container Image
```
Update RELATED_IMAGE_FLOWLOGS_PIPELINE to vX.Y.Z.
Check main.go and internal/controller/flp/ deployment templates.
```

### Debug Controller
```
FlowCollector reconciliation failing with error "X".
Check internal/controller/flowcollector_controller.go:
- Reconcile() logic
- Error handling
- Status conditions
Suggest fixes with proper error handling patterns.
```

### Kafka Integration
```
Modify Kafka producer config in eBPF agent.
Context: spec.deploymentModel=Kafka
Update internal/controller/ for Kafka-enabled agent configuration.
```

## Code Review Checklist

```
Review for:
1. Code style consistency
2. Error handling (wrap with context)
3. Unit test coverage (Ginkgo/Gomega)
4. CRD validation markers
5. Bundle update needed?
6. Documentation updates
7. Backward compatibility
8. Security (RBAC, TLS, input validation)
```

## Testing

### Unit Tests
```
Generate tests for detectSubnets in internal/controller/flp/detect_subnets.go:
- Valid CIDR ranges
- Invalid input
- Edge cases (empty, nil)
Use Ginkgo/Gomega patterns.
```

### Integration
```
Test on Kind cluster:
1. IMAGE="quay.io/me/netobserv:test" make image-build image-push deploy
2. make deploy-sample-cr
3. Verify logs and functionality
```

## Repository-Specific Context

### Loki Modes
Three deployment modes (check `spec.loki.mode`):
- **Monolithic**: Single instance
- **LokiStack**: Loki Operator (multi-tenancy enabled)
- **Microservices**: Distributed

### Performance
- **Sampling**: Default 50 (1:50 packets). Lower = more flows/resources
- **Batching**: `cacheMaxFlows`, `cacheActiveTimeout` (agent); `writeBatchWait`,
  `writeBatchSize` (Loki)
- **Memory**: Default limits 800MB
- **Metrics**: Prefix `netobserv_*`, watch cardinality

### Namespace Handling
- **OpenShift**: `openshift-netobserv-operator`
- **Community**: `netobserv`
- Use `flowCollector.Spec.Namespace` for deployed resources

### CI/CD
Before modifying workflows:
1. Run `hack/test-workflow.sh`
2. Test on `workflow-test` branch
3. Verify images on Quay.io

## Quick Reference

**Essential Commands:**
```bash
make build test                    # Build and test
make update-bundle                 # After CRD changes
make deploy-sample-cr              # Deploy FlowCollector
make undeploy                      # Clean up
```

**Key Files:**
- CRD:
  [api/flowcollector/v1beta2/flowcollector_types.go](api/flowcollector/v1beta2/flowcollector_types.go)
- Controller:
  [internal/controller/flowcollector_controller.go](internal/controller/flowcollector_controller.go)
- FLP:
  [internal/controller/flp/flp_transfo_reconciler.go](internal/controller/flp/flp_transfo_reconciler.go)
- Docs: [docs/FlowCollector.md](docs/FlowCollector.md)
- Sample:
  [config/samples/flows_v1beta2_flowcollector.yaml](config/samples/flows_v1beta2_flowcollector.yaml)

**Version Info (v1.10.x):**
- FlowCollector: v1beta2
- Min OpenShift: 4.10+
- Min Kubernetes: 1.23+
- Operator-sdk: 1.28+

## AI Workflow Example

```
1. Research: "Explain packet drop detection in eBPF agent"
2. Plan: "Add field for drop reasons filtering - suggest changes"
3. Implement: "Implement with validation and tests"
4. Review: "Review for edge cases and errors"
5. Document: "Update FlowCollector.md"
6. Test: "Provide test scenarios"
```

## Contribution Checklist

Before commit:
1. AI code review
2. `make build test`
3. `make update-bundle` (if CRD/CSV changed)
4. Update docs
5. Conventional commit messages

## Resources

- [DEVELOPMENT.md](DEVELOPMENT.md) - Build, test, deploy
- [docs/Architecture.md](docs/Architecture.md) - Component relationships
- [docs/FlowCollector.md](docs/FlowCollector.md) - API reference
- [FAQ.md](FAQ.md) - Troubleshooting
- [Contributing](https://github.com/netobserv/documents/blob/main/CONTRIBUTING.md)



**Remember**: AI agents need clear context. Always review generated code, test
thoroughly, and follow project conventions.
