# Do not remove comment lines, they are there to reduce conflicts
# Operator
export OPERATOR_IMAGE_PULLSPEC='quay.io/redhat-user-workloads/ocp-network-observab-tenant/network-observability-operator-ystream:on-pr-c48e27be5c75d094e10a9efa16561fc1eba24105'
# eBPF agent
export EBPF_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-ebpf-agent-rhel9@sha256:2475ac0482f27f598324ccbdbf3099f93820bed25a05148b9d9fcd4c1369577c'
# Flowlogs-pipeline
export FLP_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-flowlogs-pipeline-rhel9@sha256:e3027a51cf804b4a0e54f38f07b2e061081fcb47c1f5f8061df9ebf822170873'
# Console plugin
export CONSOLE_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-rhel9@sha256:e01f1da6b901abdbdb964ef01792cb36e7f14350436c3079d27ee6b86b96a68e'
# Compatibility Console plugin
export CONSOLE_COMPAT_IMAGE_PULLSPEC='registry.redhat.io/network-observability/network-observability-console-plugin-compat-rhel9@sha256:b7e5d3ed5fdeb117e949934e964b9396ec9af134b8d6b21b6058a7865b6dfc48'
