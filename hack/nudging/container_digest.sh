# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:3890d8b272c9bf69fa662413ad14d800834601eec96a57a2a2596c733654f9bc'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:ad33179ffea4188c6bbd2e3b7100e64fd677202a2c5e949b003b21e65eadf78e'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:4eb005440f62b1db8d1e7299dde58866660a28e3e82cf6580cda00f39174e70a'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:844ba89f8f8df7f4b8259d3454927bb244494db686150a720124c500fb1eaecd'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:e2ae81fff5d17b39ef31ff8376e3fe619b014d85896aab6c8afb0eb3a14b2c20'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:8dc508847935b855c123cdcc0a30ac5d4205ce213dabbac3c0fcfeff01d6b3fb'
