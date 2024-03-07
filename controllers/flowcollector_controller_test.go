package controllers

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/onsi/ginkgo/v2"
)

const (
	timeout  = test.Timeout
	interval = test.Interval
)

var (
	updateCR = func(key types.NamespacedName, updater func(*flowslatest.FlowCollector)) {
		test.UpdateCR(ctx, k8sClient, key, updater)
	}
	cleanupCR = func(key types.NamespacedName) {
		test.CleanupCR(ctx, k8sClient, key)
	}
	expectCreation = func(namespace string, objs ...test.ResourceRef) []client.Object {
		ginkgo.GinkgoHelper()
		return test.ExpectCreation(ctx, k8sClient, namespace, objs...)
	}
	expectDeletion = func(namespace string, objs ...test.ResourceRef) {
		ginkgo.GinkgoHelper()
		test.ExpectDeletion(ctx, k8sClient, namespace, objs...)
	}
	expectOwnership = func(namespace string, objs ...test.ResourceRef) {
		ginkgo.GinkgoHelper()
		test.ExpectOwnership(ctx, k8sClient, namespace, objs...)
	}
)
