package test

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceRef struct {
	name       string
	Resource   client.Object
	kind       string
	pluralKind string
}

func (r *ResourceRef) GetKey(ns string) types.NamespacedName {
	return types.NamespacedName{Name: r.name, Namespace: ns}
}

func ExpectCreation(ctx context.Context, k8sClient client.Client, namespace string, objs ...ResourceRef) []client.Object {
	ginkgo.GinkgoHelper()
	var refs []client.Object
	for _, obj := range objs {
		refs = append(refs, obj.Resource)
	}
	for _, obj := range objs {
		ginkgo.By(fmt.Sprintf(`Expecting to create "%s" %s`, obj.name, obj.kind))
		gomega.Eventually(func() interface{} {
			return k8sClient.Get(ctx, types.NamespacedName{Name: obj.name, Namespace: namespace}, obj.Resource)
		}).WithTimeout(Timeout).WithPolling(Interval).Should(gomega.Succeed())
	}
	return refs
}

func ExpectDeletion(ctx context.Context, k8sClient client.Client, namespace string, objs ...ResourceRef) {
	ginkgo.GinkgoHelper()
	for _, obj := range objs {
		ginkgo.By(fmt.Sprintf(`Expecting to delete "%s" %s`, obj.name, obj.kind))
		gomega.Eventually(func() interface{} {
			return k8sClient.Get(ctx, types.NamespacedName{Name: obj.name, Namespace: namespace}, obj.Resource)
		}).WithTimeout(Timeout).WithPolling(Interval).Should(gomega.MatchError(fmt.Sprintf(`%s "%s" not found`, obj.pluralKind, obj.name)))
	}
}

func ExpectNoCreation(ctx context.Context, k8sClient client.Client, namespace string, objs ...ResourceRef) {
	ginkgo.GinkgoHelper()
	for _, obj := range objs {
		ginkgo.By(fmt.Sprintf(`Expecting to not create "%s" %s`, obj.name, obj.kind))
		gomega.Consistently(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: obj.name, Namespace: namespace}, obj.Resource)
		}, ConsistentlyTimeout, ConsistentlyInterval).Should(gomega.MatchError(fmt.Sprintf(`%s "%s" not found`, obj.pluralKind, obj.name)))
	}
}

func ExpectOwnership(ctx context.Context, k8sClient client.Client, namespace string, objs ...ResourceRef) {
	ginkgo.GinkgoHelper()
	// Retrieve CR to get its UID
	ginkgo.By("Getting the CR")
	flowCR := GetCR(ctx, k8sClient, types.NamespacedName{Name: "cluster"})
	for _, obj := range objs {
		gomega.Eventually(func() interface{} {
			_ = k8sClient.Get(ctx, types.NamespacedName{Name: obj.name, Namespace: namespace}, obj.Resource)
			return obj.Resource
		}).WithTimeout(Timeout).WithPolling(Interval).Should(BeGarbageCollectedBy(flowCR))
	}
}

func Namespace(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &v1.Namespace{}, kind: "Namespace", pluralKind: "namespaces"}
}

func ConfigMap(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &v1.ConfigMap{}, kind: "ConfigMap", pluralKind: "configmaps"}
}

func Service(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &v1.Service{}, kind: "Service", pluralKind: "services"}
}

func ServiceAccount(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &v1.ServiceAccount{}, kind: "ServiceAccount", pluralKind: "serviceaccounts"}
}

func Deployment(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &appsv1.Deployment{}, kind: "Deployment", pluralKind: "deployments.apps"}
}

func DaemonSet(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &appsv1.DaemonSet{}, kind: "DaemonSet", pluralKind: "daemonsets.apps"}
}

func ClusterRole(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &rbacv1.ClusterRole{}, kind: "ClusterRole", pluralKind: "clusterroles"}
}

func ClusterRoleBinding(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &rbacv1.ClusterRoleBinding{}, kind: "ClusterRoleBinding", pluralKind: "clusterrolebindings"}
}

func ServiceMonitor(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &monitoringv1.ServiceMonitor{}, kind: "ServiceMonitor", pluralKind: "servicemonitors.monitoring.coreos.com"}
}

func PrometheusRule(name string) ResourceRef {
	return ResourceRef{name: name, Resource: &monitoringv1.PrometheusRule{}, kind: "PrometheusRule", pluralKind: "prometheusrules.monitoring.coreos.com"}
}
