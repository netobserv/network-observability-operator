package v1alpha1

import (
	"context"
	"fmt"

	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/helper/cardinality"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var flowmetriclog = logf.Log.WithName("flowmetric-resource")

type FlowMetricWebhook struct {
	FlowMetric
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-flows-netobserv-io-v1alpha1-flowmetric,mutating=false,failurePolicy=fail,sideEffects=None,groups=flows.netobserv.io,resources=flowmetrics,versions=v1alpha1,name=flowmetricvalidationwebhook.netobserv.io,admissionReviewVersions=v1
var (
	_ webhook.CustomValidator = &FlowMetricWebhook{FlowMetric{}}
)

func (r *FlowMetricWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&FlowMetric{}).
		WithValidator(&FlowMetricWebhook{}).
		Complete()
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowMetricWebhook) ValidateCreate(ctx context.Context, newObj runtime.Object) (warnings admission.Warnings, err error) {
	flowmetriclog.Info("validate create", "name", r.Name)
	newFlowMetric, ok := newObj.(*FlowMetric)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an FlowMetric but got a %T", newObj))
	}
	return nil, validateFlowMetric(ctx, newFlowMetric)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowMetricWebhook) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (warnings admission.Warnings, err error) {
	flowmetriclog.Info("validate update", "name", r.Name)
	newFlowMetric, ok := newObj.(*FlowMetric)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an FlowMetric but got a %T", newObj))
	}
	return nil, validateFlowMetric(ctx, newFlowMetric)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FlowMetricWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (warnings admission.Warnings, err error) {
	flowmetriclog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func checkFlowMetricCartinality(fMetric *FlowMetric) {
	r, err := cardinality.CheckCardinality(fMetric.Spec.Labels...)
	if err != nil {
		flowmetriclog.WithValues("FlowMetric name", fMetric.Name).Error(err, "Could not check metrics cardinality")
	}
	overallCardinality := r.GetOverall()
	if overallCardinality == cardinality.WarnAvoid || overallCardinality == cardinality.WarnUnknown {
		flowmetriclog.WithValues("FlowMetric name", fMetric.Name).Info("Warning: unsafe metric detected with potentially very high cardinality, please check its definition.", "Details", r.GetDetails())
	}
}

func validateFlowMetric(_ context.Context, fMetric *FlowMetric) error {
	var str []string
	var allErrs field.ErrorList

	for _, f := range fMetric.Spec.Filters {
		str = append(str, f.Field)
	}

	if len(str) != 0 {
		if !helper.FindFields(str, false) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "filters"), str,
				fmt.Sprintf("invalid filter field: %s", str)))
		}
	}

	if len(fMetric.Spec.Labels) != 0 {
		if !helper.FindFields(fMetric.Spec.Labels, false) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "labels"), fMetric.Spec.Labels,
				fmt.Sprintf("invalid label name: %s", fMetric.Spec.Labels)))
		}

		// Only fields defined as Labels are valid for remapping
		if len(fMetric.Spec.Remap) != 0 {
			labelsMap := make(map[string]any, len(fMetric.Spec.Labels))
			for _, label := range fMetric.Spec.Labels {
				labelsMap[label] = nil
			}
			var invalidMapping []string
			for toRemap := range fMetric.Spec.Remap {
				if _, ok := labelsMap[toRemap]; !ok {
					invalidMapping = append(invalidMapping, toRemap)
				}
			}
			if len(invalidMapping) > 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "remap"), fMetric.Spec.Remap,
					fmt.Sprintf("some fields defined for remapping are not defined as labels: %v", invalidMapping)))
			}
		}
	}

	if fMetric.Spec.ValueField != "" {
		if !helper.FindFields([]string{fMetric.Spec.ValueField}, true) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "valueField"), fMetric.Spec.ValueField,
				fmt.Sprintf("invalid value field: %s", fMetric.Spec.ValueField)))
		}
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: FlowMetric{}.Kind},
			fMetric.Name, allErrs)
	}
	checkFlowMetricCartinality(fMetric)
	return nil
}
