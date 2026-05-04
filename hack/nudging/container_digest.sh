# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:e7791e072d5b3b7fd9f5df5fae9c6f00a0b0ff971ea4d7b14984359e19965fce'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:da1b88f64a8dec15dbd4fee877e62cdc134ff23b1f94a3ec384a5a154bbb938d'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:14f5d89958ae540251490b1becb5ceda5ed2b5e421945105dd86ddaba42aae27'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:9a310cb12c887c7102c2393efc7ba77adcf08991dde04ff9d82bb83ea03fc5d6'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-compat-rhel9@sha256:aa95016b777baadce46d482d0af167697ce8857c6db134e50bfac8b1bda2b9c8'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:TODO'
