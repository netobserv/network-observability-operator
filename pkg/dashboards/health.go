package dashboards

import (
	"fmt"
)

func CreateHealthDashboard(netobsNs string) (string, error) {
	d := Dashboard{Title: "NetObserv / Health"}

	// Global stats
	// TODO after direct-FLP: if Direct mode, get flow rate from loki if enabled, else from agent
	d.Rows = append(d.Rows, NewRow("", false, "100px", []Panel{
		NewSingleStatPanel("Flows per second", PanelUnitShort, 3, NewTarget(
			`sum(rate(netobserv_ingest_flows_processed[1m]))`, "")),
		NewSingleStatPanel("Sampling", PanelUnitShort, 3, NewTarget(
			"avg(netobserv_agent_sampling_rate)", "")),
		NewSingleStatPanel("Errors last minute", PanelUnitShort, 3, NewTarget(
			`(sum(increase(netobserv_agent_errors_total[1m])) OR on() vector(0))
			+ (sum(increase(netobserv_ingest_errors[1m])) OR on() vector(0))
			+ (sum(increase(netobserv_encode_prom_errors[1m])) OR on() vector(0))
			+ (sum(increase(netobserv_loki_batch_retries_total[1m])) OR on() vector(0))
			+ (sum(increase(controller_runtime_reconcile_errors_total{job="netobserv-metrics-service"}[1m])) OR on() vector(0))
			`, "")),
		NewSingleStatPanel("Dropped flows per second", PanelUnitShort, 3, NewTarget(
			`(sum(rate(netobserv_loki_dropped_entries_total[1m])) OR on() vector(0))
			+ (sum(rate(netobserv_agent_dropped_flows_total[1m])) OR on() vector(0))
			`, "")),
	}))

	// FLP stats
	overheadQuery := fmt.Sprintf("100 * sum(rate(netobserv_namespace_flows_total{SrcK8S_Namespace='%s'}[1m]) or rate(netobserv_namespace_flows_total{SrcK8S_Namespace!='%s',DstK8S_Namespace='%s'}[1m])) / sum(rate(netobserv_namespace_flows_total[1m]))", netobsNs, netobsNs, netobsNs)
	// TODO: add FLP error
	d.Rows = append(d.Rows,
		NewRow("Flowlogs-pipeline statistics", false, "250px", []Panel{
			NewGraphPanel("Flows per second", PanelUnitShort, 4, false, []Target{
				NewTarget("sum(rate(netobserv_ingest_flows_processed[1m]))", "Flows ingested"),
				NewTarget("sum(rate(netobserv_loki_sent_entries_total[1m]))", "Flows sent to Loki"),
				NewTarget("sum(rate(netobserv_loki_dropped_entries_total[1m]))", "Flows dropped due to Loki error"),
			}),
			NewGraphPanel("Flows overhead (% generated by NetObserv own traffic)", PanelUnitShort, 4, false, []Target{
				NewTarget(overheadQuery, "% overhead"),
			}),
			NewGraphPanel("Errors per minute", PanelUnitShort, 4, true, []Target{
				NewTarget(`sum(increase(netobserv_ingest_errors[1m])) by (stage,code)`, "{{stage}} {{code}}"),
				NewTarget(`sum(increase(netobserv_encode_prom_errors[1m])) by (error)`, "metrics {{error}}"),
				NewTarget(`sum(increase(netobserv_loki_batch_retries_total[1m]))`, "loki retries"),
			}),
			NewGraphPanel("By namespace", PanelUnitShort, 6, false, []Target{
				NewTarget(`topk(10,sum(rate(netobserv_namespace_flows_total{SrcK8S_Namespace!=""}[1m])) by (SrcK8S_Namespace))`, "From {{SrcK8S_Namespace}}"),
				NewTarget(`topk(10,sum(rate(netobserv_namespace_flows_total{DstK8S_Namespace!=""}[1m])) by (DstK8S_Namespace))`, "To {{DstK8S_Namespace}}"),
			}),
			NewGraphPanel("By node", PanelUnitShort, 6, false, []Target{
				NewTarget(`topk(10,sum(rate(netobserv_node_flows_total{SrcK8S_HostName!=""}[1m])) by (SrcK8S_HostName))`, "From {{SrcK8S_HostName}}"),
				NewTarget(`topk(10,sum(rate(netobserv_node_flows_total{DstK8S_HostName!=""}[1m])) by (DstK8S_HostName))`, "To {{DstK8S_HostName}}"),
			}),
		}),
	)

	// Agent stats
	d.Rows = append(d.Rows, NewRow("eBPF agent statistics", true, "250px", []Panel{
		NewGraphPanel("Eviction rate", PanelUnitShort, 4, false, []Target{
			NewTarget("sum(rate(netobserv_agent_evictions_total[1m])) by (source, reason)", "{{source}} {{reason}}"),
		}),
		NewGraphPanel("Evicted flows rate", PanelUnitShort, 4, false, []Target{
			NewTarget("sum(rate(netobserv_agent_evicted_flows_total[1m])) by (source, reason)", "{{source}} {{reason}}"),
		}),
		NewGraphPanel("Dropped flows rate", PanelUnitShort, 4, true, []Target{
			NewTarget(`sum(rate(netobserv_agent_dropped_flows_total[1m])) by (source, reason)`, "{{source}} {{reason}}"),
		}),
		NewGraphPanel("Ringbuffer / HashMap ratio", PanelUnitShort, 4, false, []Target{
			NewTarget(`(sum(rate(netobserv_agent_evicted_flows_total{source="accounter"}[1m])) OR on() vector(0)) / sum(rate(netobserv_agent_evicted_flows_total{source="hashmap"}[1m]))`, "ratio"),
		}),
		NewGraphPanel("Buffer size", PanelUnitShort, 4, false, []Target{
			NewTarget(`sum(netobserv_agent_buffer_size) by (name)`, "{{name}}"),
		}),
		NewGraphPanel("Errors per minute", PanelUnitShort, 4, true, []Target{
			NewTarget(`sum(increase(netobserv_agent_errors_total[1m])) by (component, error)`, "{{component}} {{error}}"),
		}),
		NewGraphPanel("Filtered flows rate", PanelUnitShort, 4, false, []Target{
			NewTarget("sum(rate(netobserv_agent_filtered_flows_total[1m])) by (source, reason)", "{{source}} {{reason}}"),
		}),
	}))

	// Operator stats
	d.Rows = append(d.Rows, NewRow("Operator statistics", true, "250px", []Panel{
		NewGraphPanel("Reconcile events per minute", PanelUnitShort, 6, true, []Target{
			NewTarget(`sum(increase(controller_runtime_reconcile_total{job="netobserv-metrics-service"}[1m])) by (controller,result)`, "{{controller}}: {{result}}"),
		}),
		NewGraphPanel("Average and P99 reconcile time", PanelUnitSeconds, 6, false, []Target{
			NewTarget(`sum(rate(controller_runtime_reconcile_time_seconds_sum{job="netobserv-metrics-service"}[1m])) / sum(rate(controller_runtime_reconcile_time_seconds_count{job="netobserv-metrics-service"}[1m]))`, "average"),
			NewTarget(`histogram_quantile(0.99, sum by(le) (rate(controller_runtime_reconcile_time_seconds_bucket{job="netobserv-metrics-service"}[1m])))`, "p99"),
		}),
	}))

	// CPU and memory
	d.Rows = append(d.Rows, NewRow("Resource usage", true, "250px", []Panel{
		NewGraphPanel("Overall CPU", PanelUnitShort, 6, true, []Target{
			NewTarget(`sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{container="netobserv-ebpf-agent"})`, "eBPF agent"),
			NewTarget(`sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{container="flowlogs-pipeline"})`, "flowlogs-pipeline"),
			NewTarget(`sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{container!="",pod=~"netobserv-controller-manager.*"})`, "operator"),
		}),
		NewGraphPanel("Overall memory", PanelUnitShort, 6, true, []Target{
			NewTarget(`sum(container_memory_rss{container="netobserv-ebpf-agent"})`, "eBPF agent"),
			NewTarget(`sum(container_memory_rss{container="flowlogs-pipeline"})`, "flowlogs-pipeline"),
			NewTarget(`sum(container_memory_rss{container!="",pod=~"netobserv-controller-manager.*"})`, "operator"),
		}),
		NewGraphPanel("eBPF agent CPU - top 10 pods", PanelUnitShort, 6, true, []Target{
			NewTarget(`topk(10, node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{container="netobserv-ebpf-agent"})`, "{{pod}}"),
		}),
		NewGraphPanel("eBPF agent memory - top 10 pods", PanelUnitShort, 6, true, []Target{
			NewTarget(`topk(10, container_memory_rss{container="netobserv-ebpf-agent"})`, "{{pod}}"),
		}),
		NewGraphPanel("Flowlogs-pipeline CPU - top 10 pods", PanelUnitShort, 6, true, []Target{
			NewTarget(`topk(10, node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{container="flowlogs-pipeline"})`, "{{pod}}"),
		}),
		NewGraphPanel("Flowlogs-pipeline memory - top 10 pods", PanelUnitShort, 6, true, []Target{
			NewTarget(`topk(10, container_memory_rss{container="flowlogs-pipeline"})`, "{{pod}}"),
		}),
	}))

	return d.ToGrafanaJSON(netobsNs), nil
}
