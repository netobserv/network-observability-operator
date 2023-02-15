package flowlogspipeline

import (
	"context"
	"fmt"

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
	initStaticResources(ctx context.Context) error
	prepareNamespaceChange(ctx context.Context) error
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

// InitStaticResources inits some "static" / one-shot resources, usually not subject to reconciliation
func (r *FLPReconciler) InitStaticResources(ctx context.Context) error {
	for _, sr := range r.reconcilers {
		if err := sr.initStaticResources(sr.context(ctx)); err != nil {
			return err
		}
	}
	return nil
}

// PrepareNamespaceChange cleans up old namespace and restore the relevant "static" resources
func (r *FLPReconciler) PrepareNamespaceChange(ctx context.Context) error {
	for _, sr := range r.reconcilers {
		if err := sr.prepareNamespaceChange(sr.context(ctx)); err != nil {
			return err
		}
	}
	return nil
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
