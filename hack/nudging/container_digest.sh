# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:2fa7ecd11f6bba41c2f048fffcd37fc4401a1cf4d9d6fc9f8504b1ffd6b38745'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:8da85229957867e0fd5d21e59cc1cc4e475a90e3867e5f8962aa1ff234a4674e'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:3ee9431f02475cde0c5f401ae1bd983b241dd39ecc0ed8776309bebca617752b'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:8f690e19a51dcaa74f2a4c29afcc5fc4273d67f4f13e17b028ddca32ae2b4162'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:cdedba76e796f23a63e1301671e137f68d70a98a4923af429f15fc879b86a326'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:394a43e50715621286f4d2795acab42291cd2d7fef788c6f78d3606f96434758'
