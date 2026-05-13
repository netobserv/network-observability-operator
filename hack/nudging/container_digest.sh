# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-rhel9-operator@sha256:546a96299190bb9f15c56164c7b68747846432813ab6814c5bb59588ad24e658'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:8da85229957867e0fd5d21e59cc1cc4e475a90e3867e5f8962aa1ff234a4674e'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:4eb005440f62b1db8d1e7299dde58866660a28e3e82cf6580cda00f39174e70a'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:14edbf32483439b404115326f698c0bd724f713ac60d0e27860f83a01ae7f8b5'
# Console plugin PF4 (default / OCP < 4.15)
export CONSOLE_PF4_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf4-rhel9@sha256:cdedba76e796f23a63e1301671e137f68d70a98a4923af429f15fc879b86a326'
# Console plugin PF5 (OCP 4.15–4.21)
export CONSOLE_PF5_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-pf5-rhel9@sha256:df10209b1966db57727d6cbc1426c2c998377362733b6dd0059d904d70cea652'
