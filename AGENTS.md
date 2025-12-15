# AI Agents Best Practices for Network Observability Operator

This guide provides best practices for working with AI coding agents (such as
Claude Code, GitHub Copilot, etc.) on the NetObserv Operator codebase. Following
these practices will help you get better results and maintain code quality.

> **Note**: This file is also symlinked as [CLAUDE.md](CLAUDE.md) for Claude Code to auto-load project-specific instructions.

## Table of Contents

- [Understanding the Codebase](#understanding-the-codebase)
- [Effective Prompting](#effective-prompting)
- [Common Tasks](#common-tasks)
- [Code Review with AI](#code-review-with-ai)
- [Testing and Validation](#testing-and-validation)
- [Pitfalls to Avoid](#pitfalls-to-avoid)
- [Repository-Specific Context](#repository-specific-context)

## Understanding the Codebase

### Project Overview

NetObserv Operator is a Kubernetes/OpenShift operator for network observability
built using the operator-sdk. Key components:

- **eBPF Agent**: Generates network flows from captured packets
- **flowlogs-pipeline**: Collects, enriches, and exports flows
- **Console Plugin**: Provides visualization in OpenShift Console
- **CRD**: `FlowCollector` (v1beta2) - single cluster-wide resource named
  `cluster`

### Architecture Context

Before asking an AI agent to make changes, provide context:

```
I'm working on the NetObserv Operator, which manages:
- eBPF agents (DaemonSet)
- flowlogs-pipeline (Deployment/StatefulSet)
- Console plugin (optional)
- Integration with Loki, Prometheus, and optionally Kafka

The main controller reconciles FlowCollector resources.
```

### Key Directories

When working with AI agents, reference these directories:

- `api/`: CRD definitions and API types
- `internal/controller/`: Main reconciliation logic
- `config/`: Kustomize manifests for deployment
- `helm/`: Helm chart definitions
- `hack/`: Development and build scripts
- `docs/`: Documentation including FlowCollector spec

## Effective Prompting

### DO: Provide Specific Context

**Good:**
```
Update the FlowCollector controller in internal/controller/flowcollector_controller.go
to add validation for the new spec.agent.ebpf.logLevel field. The valid values should
be: trace, debug, info, warn, error. Add webhook validation.
```

**Bad:**
```
Add log level validation
```

### DO: Reference Existing Patterns

**Good:**
```
Add a new metric for tracking flow drops, similar to how
netobserv_loki_dropped_entries_total is implemented in the FLP pipeline builder.
Follow the same pattern in internal/controller/flp/flp_pipeline_builder.go
```

### DO: Specify Testing Requirements

**Good:**
```
Implement the new feature and include:
1. Unit tests following the pattern in flp_controller_test.go
2. Update the bundle with: make update-bundle
3. Verify the changes work with: make build test
```

### DON'T: Make Assumptions About Dependencies

**Bad:**
```
Add support for the latest Loki features
```

**Good:**
```
Check the current Loki version in go.mod and add support for retention
policies compatible with that version. If upgrade needed, document it.
```

## Common Tasks

### Adding a New Field to FlowCollector

Prompt template:
```
I want to add a new field spec.agent.ebpf.newFeature (type: bool, default: false).
Please:
1. Add the field to api/flowcollector/v1beta2/flowcollector_types.go
2. Update the CRD markers with appropriate validation
3. Update internal/controller/ to use this field
4. Add unit tests
5. Update docs/FlowCollector.md
6. Run make update-bundle
```

### Updating Container Images

```
Update the RELATED_IMAGE_FLOWLOGS_PIPELINE to use version X.Y.Z.
Check main.go for where this is referenced and update the
deployment templates in internal/controller/flp/
```

### Debugging Controller Issues

```
The FlowCollector reconciliation is failing with error "X".
Please examine internal/controller/flowcollector_controller.go
and check:
1. The reconciliation logic
2. Error handling in the Reconcile() method
3. Status conditions updates
Suggest fixes with proper error handling patterns.
```

### Working with Kafka Integration

```
I need to modify the Kafka producer configuration in the eBPF agent.
Context: deploymentModel is set to Kafka in FlowCollector spec.
Please update the relevant parts in internal/controller/ that
generate the agent configuration when Kafka is enabled.
```

## Code Review with AI

### Pre-commit Checklist Prompt

```
Review my changes for:
1. Consistency with existing code style
2. Proper error handling
3. Unit test coverage
4. CRD validation markers
5. Whether bundle update is needed
6. Documentation updates needed
7. Backward compatibility concerns
```

### Security Review

```
Review this code for security issues:
- Sensitive data handling
- RBAC implications
- Certificate/TLS configuration
- Input validation
Focus on areas handling user input from FlowCollector spec.
```

## Testing and Validation

### Test Generation

**Good prompt:**
```
Generate unit tests for the new detectSubnets function in
internal/controller/flp/detect_subnets.go. Include test cases for:
- Valid CIDR ranges
- Invalid input
- Edge cases (empty, nil)
Follow the existing test patterns using Ginkgo/Gomega.
```

### Integration Testing

```
I need to test this change on a Kind cluster. Provide:
1. Steps to build and deploy: IMAGE="quay.io/me/netobserv:test" make image-build image-push deploy
2. How to create a test FlowCollector
3. What to check in logs
4. How to verify the feature works
```

### Bundle Validation

```
I've made CRD changes. Help me:
1. Understand if bundle update is needed
2. Run: make update-bundle
3. Verify the bundle with operator-sdk
4. Check what files changed and if they're expected
```

## Pitfalls to Avoid

### ❌ Don't: Ignore the Single FlowCollector Constraint

The operator only supports ONE FlowCollector named "cluster". Always validate
this in code:

```go
if flowCollector.Name != constants.FlowCollectorName {
    // Handle error
}
```

### ❌ Don't: Break Backward Compatibility

FlowCollector v1beta2 must maintain compatibility. When adding fields:
- Use optional fields with proper defaults
- Don't remove or rename existing fields
- Use +optional marker in CRD

### ❌ Don't: Forget Bundle Updates

After changing CRD or CSV-related files, always run:
```bash
make update-bundle
```

Ask the AI: "Do I need to run make update-bundle after these changes?"

### ❌ Don't: Hardcode Image References

Use environment variables:
- `RELATED_IMAGE_EBPF_AGENT`
- `RELATED_IMAGE_FLOWLOGS_PIPELINE`
- `RELATED_IMAGE_CONSOLE_PLUGIN`

### ❌ Don't: Ignore Namespace Handling

The operator can run in different namespaces:
- OpenShift: typically `openshift-netobserv-operator`
- Community: typically `netobserv`

Always use `flowCollector.Spec.Namespace` for deployed resources.

## Repository-Specific Context

### Workflow Testing

When modifying CI/CD:
```
I'm updating .github/workflows/push_image.yml. Before committing:
1. Run hack/test-workflow.sh
2. Test on workflow-test branch
3. Verify images on Quay.io after merge
See DEVELOPMENT.md "Testing the github workflow" section.
```

### Loki Integration

```
Context: NetObserv supports multiple Loki deployment modes:
- Monolithic: single Loki instance
- LokiStack: via Loki Operator (enables multi-tenancy)
- Microservices: distributed deployment

When working on Loki client code in internal/controller/,
consider all three modes. Check spec.loki.mode field.
```

### Performance Considerations

```
When modifying agent or FLP configuration:
- Consider impact on sampling rate (spec.agent.ebpf.sampling)
- Check cacheMaxFlows and cacheActiveTimeout
- Memory limits are typically 800MB
- Test with realistic cluster sizes
Review docs on performance tuning before making changes.
```

### Metrics and Monitoring

```
NetObserv exposes metrics with prefix "netobserv_".
When adding new metrics:
- Follow Prometheus naming conventions
- Add to ServiceMonitor in config/openshift-olm/default/
- Document in docs/Metrics.md
- Consider cardinality impact
```

### Multi-Architecture Support

The operator supports: amd64, arm64, ppc64le, s390x

When working on container builds:
```
Ensure changes work across all architectures.
CI builds multi-arch manifests. Test locally:
IMAGE="quay.io/me/netobserv:test" make image-build
```

## Advanced AI Agent Usage

### Refactoring Large Functions

```
The function reconcileFlowlogsPipeline in internal/controller/flp/flp_transfo_reconciler.go
is complex. Help me refactor it:
1. Identify logical sections
2. Propose helper functions
3. Maintain existing behavior
4. Add tests for new functions
5. Ensure no regressions
```

### Documentation Generation

```
Generate user-facing documentation for the new field
spec.processor.logTypes.flows.enable in the style of docs/FlowCollector.md.
Include:
- Description
- Default value
- Example usage
- Impact on deployment
```

### Troubleshooting Guide

```
Create a troubleshooting guide for when the Console Plugin fails to start.
Base it on common issues in internal/controller/consoleplugin/ and
follow the FAQ.md structure. Include:
- Symptoms
- How to check logs
- Common causes
- Solutions
```

## Quick Reference

### Essential Commands for AI Context

```bash
# Build and test
make build test

# Update bundle after CRD changes
make update-bundle

# Deploy to cluster
IMAGE="quay.io/me/netobserv:test" make image-build image-push deploy

# Deploy sample FlowCollector
make deploy-sample-cr

# Clean up
make undeploy
```

### Key Files to Reference

- **CRD Schema**:
  [api/flowcollector/v1beta2/flowcollector_types.go](api/flowcollector/v1beta2/flowcollector_types.go)
- **Main Controller**:
  [internal/controller/flowcollector_controller.go](internal/controller/flowcollector_controller.go)
- **FLP Reconciler**:
  [internal/controller/flp/flp_transfo_reconciler.go](internal/controller/flp/flp_transfo_reconciler.go)
- **Documentation**: [docs/FlowCollector.md](docs/FlowCollector.md)
- **Sample CR**:
  [config/samples/flows_v1beta2_flowcollector.yaml](config/samples/flows_v1beta2_flowcollector.yaml)

### Getting Help from AI

When stuck, ask the AI to:
```
1. Explain the architecture around [component]
2. Find where [feature] is implemented
3. Suggest how to implement [new feature] following existing patterns
4. Review my code for [specific concerns]
5. Generate tests for [function/feature]
```

## Version-Specific Notes

### Current Version (v1.10.x)

- FlowCollector API: v1beta2
- Minimum OpenShift: 4.10+
- Minimum Kubernetes: 1.23+
- Controller-runtime: check go.mod
- Operator-sdk: 1.28+

When asking AI for help, mention the version context if working with older
releases.

## Contributing Changes

Before finalizing AI-assisted changes:

1. **Code Review**: Have the AI review its own code
2. **Testing**: Run `make build test`
3. **Bundle**: Run `make update-bundle` if needed
4. **Documentation**: Update relevant docs
5. **Commit Messages**: Follow conventional commits
6. **PR Description**: Explain the change clearly

### Example Workflow with AI

```
1. Research Phase:
   "Explain how packet drop detection works in the eBPF agent"

2. Planning Phase:
   "I want to add a field to configure drop reasons filtering.
    Suggest where to add it and what changes are needed"

3. Implementation Phase:
   "Implement the field with proper validation and tests"

4. Review Phase:
   "Review this implementation for edge cases and errors"

5. Documentation Phase:
   "Update FlowCollector.md with this new field"

6. Testing Phase:
   "Provide test scenarios for this feature"
```

## Additional Resources

- [Development Guide](DEVELOPMENT.md) - Build, test, deploy procedures
- [Architecture](docs/Architecture.md) - Component relationships
- [FlowCollector Spec](docs/FlowCollector.md) - Complete API reference
- [FAQ](FAQ.md) - Common issues and solutions
- [Contributing](https://github.com/netobserv/documents/blob/main/CONTRIBUTING.md) - Contribution guidelines

## Feedback and Improvements

This guide evolves with the project. If you discover better prompting strategies
or AI agent workflows, please contribute updates to this file.



**Remember**: AI agents are powerful tools but require clear context and
validation. Always review generated code, test thoroughly, and ensure it follows
project conventions.
