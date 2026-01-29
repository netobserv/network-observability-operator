package alerts

import (
	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
)

// TemplateInfo contains display information for a health rule template
type TemplateInfo struct {
	Summary            string
	DescriptionPattern string // Pattern with placeholders for threshold, legend, etc.
}

// TemplateMetadata provides centralized metadata for all health rule templates
// This is used both for generating alert annotations and for console plugin configuration
var TemplateMetadata = map[flowslatest.HealthRuleTemplate]TemplateInfo{
	flowslatest.HealthRulePacketDropsByKernel: {
		Summary:            "Too many packets dropped by the kernel",
		DescriptionPattern: "NetObserv is detecting more than %s%% of packets dropped by the kernel%s",
	},
	flowslatest.HealthRulePacketDropsByDevice: {
		Summary:            "Too many drops from device",
		DescriptionPattern: "node-exporter is reporting more than %s%% of dropped packets%s",
	},
	flowslatest.HealthRuleIPsecErrors: {
		Summary:            "Too many IPsec errors",
		DescriptionPattern: "NetObserv is detecting more than %s%% of IPsec errors%s",
	},
	flowslatest.HealthRuleDNSErrors: {
		Summary:            "Too many DNS errors",
		DescriptionPattern: "NetObserv is detecting more than %s%% of DNS errors%s (other than NX_DOMAIN)",
	},
	flowslatest.HealthRuleDNSNxDomain: {
		Summary:            "Too many DNS NX_DOMAIN errors",
		DescriptionPattern: "NetObserv is detecting more than %s%% of DNS NX_DOMAIN errors%s. In Kubernetes, this is a common error due to the resolution using several search suffixes. It can be optimized by using trailing dots in domain names",
	},
	flowslatest.HealthRuleNetpolDenied: {
		Summary:            "Traffic denied by Network Policies",
		DescriptionPattern: "NetObserv is detecting more than %s%% of denied traffic due to Network Policies%s",
	},
	flowslatest.HealthRuleLatencyHighTrend: {
		Summary:            "TCP latency increase",
		DescriptionPattern: "NetObserv is detecting TCP latency increased by more than %s%%%s, compared to baseline (offset: %s)",
	},
	flowslatest.HealthRuleExternalEgressHighTrend: {
		Summary:            "External egress traffic increase",
		DescriptionPattern: "NetObserv is detecting external egress traffic increased by more than %s%%%s, compared to baseline (offset: %s)",
	},
	flowslatest.HealthRuleExternalIngressHighTrend: {
		Summary:            "External ingress traffic increase",
		DescriptionPattern: "NetObserv is detecting external ingress traffic increased by more than %s%%%s, compared to baseline (offset: %s)",
	},
	flowslatest.HealthRuleIngress5xxErrors: {
		Summary:            "Too many ingress 5xx errors",
		DescriptionPattern: "HAProxy is reporting more than %s%% of 5xx HTTP response codes from ingress traffic%s",
	},
	flowslatest.HealthRuleIngressHTTPLatencyTrend: {
		Summary:            "Ingress HTTP latency increase",
		DescriptionPattern: "HAProxy ingress average HTTP response latency increased by more than %s%%%s, compared to baseline (offset: %s)",
	},
}
