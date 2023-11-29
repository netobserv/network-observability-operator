package reconcilers

import (
	"context"
	"fmt"
	"reflect"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	IgnoreStatusChange = builder.WithPredicates(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Update only if spec / annotations / labels change, ie. ignore status changes
			return (e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()) ||
				!equality.Semantic.DeepEqual(e.ObjectNew.GetAnnotations(), e.ObjectOld.GetAnnotations()) ||
				!equality.Semantic.DeepEqual(e.ObjectNew.GetLabels(), e.ObjectOld.GetLabels())
		},
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		DeleteFunc:  func(e event.DeleteEvent) bool { return true },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	})
)

func ReconcileClusterRoleBinding(ctx context.Context, cl *helper.Client, desired *rbacv1.ClusterRoleBinding) error {
	actual := rbacv1.ClusterRoleBinding{}
	if err := cl.Get(ctx, types.NamespacedName{Name: desired.ObjectMeta.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return cl.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile ClusterRoleBinding %s: %w", desired.Name, err)
	}
	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		actual.RoleRef == desired.RoleRef &&
		reflect.DeepEqual(actual.Subjects, desired.Subjects) {
		if actual.RoleRef != desired.RoleRef {
			//Roleref cannot be updated deleting and creating a new rolebinding
			log := log.FromContext(ctx)
			log.Info("Deleting old ClusterRoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			err := cl.Delete(ctx, &actual)
			if err != nil {
				log.Error(err, "error deleting old ClusterRoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			}
			return cl.CreateOwned(ctx, desired)
		}
		// cluster role binding already reconciled. Exiting
		return nil
	}
	return cl.UpdateIfOwned(ctx, &actual, desired)
}

func ReconcileRoleBinding(ctx context.Context, cl *helper.Client, desired *rbacv1.RoleBinding) error {
	actual := rbacv1.RoleBinding{}
	if err := cl.Get(ctx, types.NamespacedName{Name: desired.ObjectMeta.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return cl.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile RoleBinding %s: %w", desired.Name, err)
	}
	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		actual.RoleRef == desired.RoleRef &&
		reflect.DeepEqual(actual.Subjects, desired.Subjects) {
		if actual.RoleRef != desired.RoleRef {
			//Roleref cannot be updated deleting and creating a new rolebinding
			log := log.FromContext(ctx)
			log.Info("Deleting old RoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			err := cl.Delete(ctx, &actual)
			if err != nil {
				log.Error(err, "error deleting old RoleBinding", "Namespace", actual.GetNamespace(), "Name", actual.GetName())
			}
			return cl.CreateOwned(ctx, desired)
		}
		// role binding already reconciled. Exiting
		return nil
	}
	return cl.UpdateIfOwned(ctx, &actual, desired)
}

func ReconcileClusterRole(ctx context.Context, cl *helper.Client, desired *rbacv1.ClusterRole) error {
	actual := rbacv1.ClusterRole{}
	if err := cl.Get(ctx, types.NamespacedName{Name: desired.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return cl.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile ClusterRole %s: %w", desired.Name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// cluster role already reconciled. Exiting
		return nil
	}

	return cl.UpdateIfOwned(ctx, &actual, desired)
}

func ReconcileRole(ctx context.Context, cl *helper.Client, desired *rbacv1.Role) error {
	actual := rbacv1.Role{}
	if err := cl.Get(ctx, types.NamespacedName{Name: desired.Name}, &actual); err != nil {
		if errors.IsNotFound(err) {
			return cl.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile Role %s: %w", desired.Name, err)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Rules, desired.Rules) {
		// role already reconciled. Exiting
		return nil
	}

	return cl.UpdateIfOwned(ctx, &actual, desired)
}

func ReconcileConfigMap(ctx context.Context, cl *helper.Client, desired *corev1.ConfigMap, delete bool) error {
	actual := corev1.ConfigMap{}
	if err := cl.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, &actual); err != nil {
		if errors.IsNotFound(err) {
			if delete {
				return nil
			}
			return cl.CreateOwned(ctx, desired)
		}
		return fmt.Errorf("can't reconcile Configmap %s: %w", desired.Name, err)
	}

	if delete {
		return cl.Delete(ctx, desired)
	}

	if helper.IsSubSet(actual.Labels, desired.Labels) &&
		reflect.DeepEqual(actual.Data, desired.Data) {
		// configmap already reconciled. Exiting
		return nil
	}

	return cl.UpdateIfOwned(ctx, &actual, desired)
}

func ReconcileDaemonSet(ctx context.Context, ci *Instance, old, new *appsv1.DaemonSet, containerName string, report *helper.ChangeReport) error {
	if !ci.Managed.Exists(old) {
		ci.Status.SetCreatingDaemonSet(new)
		return ci.CreateOwned(ctx, new)
	}
	ci.Status.CheckDaemonSetProgress(old)
	if helper.PodChanged(&old.Spec.Template, &new.Spec.Template, containerName, report) {
		return ci.UpdateIfOwned(ctx, old, new)
	}
	return nil
}

func ReconcileDeployment(ctx context.Context, ci *Instance, old, new *appsv1.Deployment, containerName string, replicas int32, hpa *flowslatest.FlowCollectorHPA, report *helper.ChangeReport) error {
	if !ci.Managed.Exists(old) {
		ci.Status.SetCreatingDeployment(new)
		return ci.CreateOwned(ctx, new)
	}
	ci.Status.CheckDeploymentProgress(old)
	if helper.DeploymentChanged(old, new, containerName, helper.HPADisabled(hpa), replicas, report) {
		return ci.UpdateIfOwned(ctx, old, new)
	}
	return nil
}

func ReconcileHPA(ctx context.Context, ci *Instance, old, new *ascv2.HorizontalPodAutoscaler, desired *flowslatest.FlowCollectorHPA, report *helper.ChangeReport) error {
	// Delete or Create / Update Autoscaler according to HPA option
	if helper.HPADisabled(desired) {
		ci.Managed.TryDelete(ctx, old)
	} else {
		if !ci.Managed.Exists(old) {
			return ci.CreateOwned(ctx, new)
		} else if helper.AutoScalerChanged(old, *desired, report) {
			return ci.UpdateIfOwned(ctx, old, new)
		}
	}
	return nil
}

func ReconcileService(ctx context.Context, ci *Instance, old, new *corev1.Service, report *helper.ChangeReport) error {
	if !ci.Managed.Exists(old) {
		if err := ci.CreateOwned(ctx, new); err != nil {
			return err
		}
	} else if helper.ServiceChanged(old, new, report) {
		// In case we're updating an existing service, we need to build from the old one to keep immutable fields such as clusterIP
		newSVC := old.DeepCopy()
		newSVC.Spec.Ports = new.Spec.Ports
		newSVC.ObjectMeta.Annotations = new.ObjectMeta.Annotations
		if err := ci.UpdateIfOwned(ctx, old, newSVC); err != nil {
			return err
		}
	}
	return nil
}

func GenericReconcile[K client.Object](ctx context.Context, m *NamespacedObjectManager, cl *helper.Client, old, new K, report *helper.ChangeReport, changeFunc func(old, new K, report *helper.ChangeReport) bool) error {
	if !m.Exists(old) {
		return cl.CreateOwned(ctx, new)
	}
	if changeFunc(old, new, report) {
		return cl.UpdateIfOwned(ctx, old, new)
	}
	return nil
}
