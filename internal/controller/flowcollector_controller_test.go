package controllers

import (
	"k8s.io/apimachinery/pkg/types"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/pkg/test"
)

const (
	timeout  = test.Timeout
	interval = test.Interval
)

var (
	updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
		test.UpdateCR(ctx, k8sClient, key, updater)
	}
	getCR = func(key types.NamespacedName) *flowslatest.FlowCollector {
		return test.GetCR(ctx, k8sClient, key)
	}
	cleanupCR = func(key types.NamespacedName) {
		test.CleanupCR(ctx, k8sClient, key)
	}
)
