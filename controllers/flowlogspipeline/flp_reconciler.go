package flowlogspipeline

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
)

// Type alias
type flpSpec = flowslatest.FlowCollectorFLP

// FLPReconciler reconciles the current flowlogs-pipeline state with the desired configuration
type FLPReconciler struct {
	reconcilers []singleReconciler
}

const contextReconcilerName = "FLP kind"

type singleReconciler interface {
	context(ctx context.Context) context.Context
	cleanupNamespace(ctx context.Context)
	reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error
}

type reconcilersCommonInfo struct {
	reconcilers.ClientHelper
	nobjMngr        *reconcilers.NamespacedObjectManager
	useOpenShiftSCC bool
	image           string
	availableAPIs   *discover.AvailableAPIs
}

func createCommonInfo(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS, image string, permissionsVendor *discover.Permissions, availableAPIs *discover.AvailableAPIs) *reconcilersCommonInfo {
	nobjMngr := reconcilers.NewNamespacedObjectManager(cl, ns, prevNS)
	openshift := permissionsVendor.Vendor(ctx) == discover.VendorOpenShift
	return &reconcilersCommonInfo{
		ClientHelper:    cl,
		nobjMngr:        nobjMngr,
		useOpenShiftSCC: openshift,
		image:           image,
		availableAPIs:   availableAPIs,
	}
}

func NewReconciler(ctx context.Context, cl reconcilers.ClientHelper, ns, prevNS, image string, permissionsVendor *discover.Permissions, availableAPIs *discover.AvailableAPIs) FLPReconciler {
	return FLPReconciler{
		reconcilers: []singleReconciler{
			newMonolithReconciler(createCommonInfo(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs)),
			newTransformerReconciler(createCommonInfo(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs)),
			newIngesterReconciler(createCommonInfo(ctx, cl, ns, prevNS, image, permissionsVendor, availableAPIs)),
		},
	}
}

// CleanupNamespace cleans up old namespace
func (r *FLPReconciler) CleanupNamespace(ctx context.Context) {
	for _, sr := range r.reconcilers {
		sr.cleanupNamespace(sr.context(ctx))
	}
}

func validateDesired(desired *flpSpec) error {
	if desired.Port == 4789 ||
		desired.Port == 6081 ||
		desired.Port == 500 ||
		desired.Port == 4500 {
		return fmt.Errorf("flowlogs-pipeline port value is not authorized")
	}
	return nil
}

func (r *FLPReconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	if err := validateDesired(&desired.Spec.Processor); err != nil {
		return err
	}
	for _, sr := range r.reconcilers {
		if err := sr.reconcile(sr.context(ctx), desired); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconcilersCommonInfo) reconcileDashboardConfig(ctx context.Context, dbConfigMap *corev1.ConfigMap) error {
	if dbConfigMap == nil {
		// Dashboard config not desired => delete if exists
		if err := r.Delete(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dashboardCMName,
				Namespace: dashboardCMNamespace,
			},
		}); err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("deleting %s ConfigMap: %w", dashboardCMName, err)
			}
		}
		return nil
	}
	curr := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      dashboardCMName,
		Namespace: dashboardCMNamespace,
	}, curr); err != nil {
		if errors.IsNotFound(err) {
			return r.CreateOwned(ctx, dbConfigMap)
		}
		return err
	}
	if !equality.Semantic.DeepDerivative(dbConfigMap.Data, curr.Data) {
		return r.UpdateOwned(ctx, curr, dbConfigMap)
	}
	return nil
}
