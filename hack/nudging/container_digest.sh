# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:e7791e072d5b3b7fd9f5df5fae9c6f00a0b0ff971ea4d7b14984359e19965fce'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:d97a44d9bb346516be5abb5c300a4947ff4cb0543665745f6440f81b8003680c'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:93268130749b55f472dd5283c60d17e34a615b4d01ab9704f9b4a2798b290162'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:8f690e19a51dcaa74f2a4c29afcc5fc4273d67f4f13e17b028ddca32ae2b4162'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-compat-rhel9@sha256:aa95016b777baadce46d482d0af167697ce8857c6db134e50bfac8b1bda2b9c8'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:TODO'
