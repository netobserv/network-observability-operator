package loki

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/controller/constants"
	"github.com/netobserv/network-observability-operator/internal/controller/reconcilers"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
)

// LReconciler reconciles the current console plugin state with the desired configuration
type LReconciler struct {
	*reconcilers.Instance
	configMap  *corev1.ConfigMap
	deployment *appsv1.Deployment
	pvc        *corev1.PersistentVolumeClaim
	service    *corev1.Service
}

func NewReconciler(cmn *reconcilers.Instance) LReconciler {
	rec := LReconciler{
		Instance:   cmn,
		configMap:  cmn.Managed.NewConfigMap(configMapName),
		deployment: cmn.Managed.NewDeployment(constants.LokiDev),
		pvc:        cmn.Managed.NewPersistentVolumeClaim(storeVolume),
		service:    cmn.Managed.NewService(constants.LokiDev),
	}
	return rec
}

// Reconcile is the reconciler entry point to reconcile the current plugin state with the desired configuration
func (r *LReconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	l := log.FromContext(ctx).WithName("loki")
	ctx = log.IntoContext(ctx, l)

	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if desired.Spec.UseLokiDev() {
		// Create object builder
		builder := newBuilder(r.Instance, &desired.Spec, constants.LokiDev)

		cmDigest, err := r.reconcileConfigMap(ctx, &builder)
		if err != nil {
			return err
		}

		if err = r.reconcilePVC(ctx, &builder); err != nil {
			return err
		}

		if err = r.reconcileDeployment(ctx, &builder, constants.LokiDev, cmDigest); err != nil {
			return err
		}

		if err = r.reconcileServices(ctx, &builder, constants.LokiDev); err != nil {
			return err
		}
	} else {
		// delete any existing owned object
		r.Managed.TryDeleteAll(ctx)
	}

	return nil
}

func (r *LReconciler) reconcileConfigMap(ctx context.Context, builder *builder) (string, error) {
	newCM, configDigest, err := builder.configMap()
	if err != nil {
		return "", err
	}
	if !r.Managed.Exists(r.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return "", err
		}
	} else if !reflect.DeepEqual(newCM.Data, r.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.configMap, newCM); err != nil {
			return "", err
		}
	}
	return configDigest, nil
}

func (r *LReconciler) reconcilePVC(ctx context.Context, builder *builder) error {
	pvc := builder.persistentVolumeClaim()
	if !r.Managed.Exists(r.pvc) {
		if err := r.CreateOwned(ctx, pvc); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(pvc, r.pvc) {
		if err := r.UpdateIfOwned(ctx, r.pvc, pvc); err != nil {
			return err
		}
	}
	return nil
}

func (r *LReconciler) reconcileDeployment(ctx context.Context, builder *builder, name string, cm string) error {
	report := helper.NewChangeReport("Loki deployment")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDeployment(
		ctx,
		r.Instance,
		r.deployment,
		builder.deployment(name, cm),
		name,
		false,
		&report,
	)
}

func (r *LReconciler) reconcileServices(ctx context.Context, builder *builder, name string) error {
	report := helper.NewChangeReport("Loki services")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.service, builder.service(name), &report); err != nil {
		return err
	}
	return nil
}
