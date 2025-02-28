package ebpf

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go:generate bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS -target amd64,arm64,ppc64le,s390x -type flow_metrics_t -type flow_id_t -type flow_record_t -type pkt_drops_t -type dns_record_t -type global_counters_key_t -type direction_t -type filter_action_t -type tcp_flags_t -type translated_flow_t Bpf ../../bpf/flows.c -- -I../../bpf/headers
