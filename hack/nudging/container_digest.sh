# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:546a96299190bb9f15c56164c7b68747846432813ab6814c5bb59588ad24e658'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:a99cd70d28abd406bc308da1943bdd9cbd08a768b338e154fb1d8c58e7adde0c'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:3ee9431f02475cde0c5f401ae1bd983b241dd39ecc0ed8776309bebca617752b'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:8f690e19a51dcaa74f2a4c29afcc5fc4273d67f4f13e17b028ddca32ae2b4162'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-compat-rhel9@sha256:cdedba76e796f23a63e1301671e137f68d70a98a4923af429f15fc879b86a326'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:TODO'
