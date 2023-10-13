/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"fmt"

	"github.com/netobserv/network-observability-operator/api/v1beta2"
	utilconversion "github.com/netobserv/network-observability-operator/pkg/conversion"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 FlowCollector to its v1beta2 equivalent (the conversion Hub)
// https://book.kubebuilder.io/multiversion-tutorial/conversion.html
func (r *FlowCollector) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1beta1_FlowCollector_To_v1beta2_FlowCollector(r, dst, nil); err != nil {
		return fmt.Errorf("copying v1beta1.FlowCollector into v1beta2.FlowCollector: %w", err)
	}
	dst.Status.Conditions = make([]v1.Condition, len(r.Status.Conditions))
	copy(dst.Status.Conditions, r.Status.Conditions)

	// Manually restore data.
	restored := &v1beta2.FlowCollector{}
	if ok, err := utilconversion.UnmarshalData(r, restored); err != nil || !ok {
		// fallback on current loki config as Manual mode if metadata are not available
		dst.Spec.Loki.Mode = v1beta2.LokiModeManual
		dst.Spec.Loki.Manual.IngesterURL = r.Spec.Loki.URL
		dst.Spec.Loki.Manual.QuerierURL = r.Spec.Loki.QuerierURL
		dst.Spec.Loki.Manual.StatusURL = r.Spec.Loki.StatusURL
		dst.Spec.Loki.Manual.TenantID = r.Spec.Loki.TenantID
		dst.Spec.Loki.Manual.AuthToken = r.Spec.Loki.AuthToken
		if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&r.Spec.Loki.TLS, &dst.Spec.Loki.Manual.TLS, nil); err != nil {
			return fmt.Errorf("copying v1beta1.Loki.TLS into v1beta2.Loki.Manual.TLS: %w", err)
		}
		if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&r.Spec.Loki.StatusTLS, &dst.Spec.Loki.Manual.StatusTLS, nil); err != nil {
			return fmt.Errorf("copying v1beta1.Loki.StatusTLS into v1beta2.Loki.Manual.StatusTLS: %w", err)
		}
		return err
	}

	// Loki
	dst.Spec.Loki.Mode = restored.Spec.Loki.Mode
	dst.Spec.Loki.Manual = restored.Spec.Loki.Manual
	if restored.Spec.Loki.LokiStack != nil {
		dst.Spec.Loki.LokiStack = restored.Spec.Loki.LokiStack
	}

	return nil
}

// ConvertFrom converts the hub version v1beta2 FlowCollector object to v1beta1
func (r *FlowCollector) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollector)

	if err := Convert_v1beta2_FlowCollector_To_v1beta1_FlowCollector(src, r, nil); err != nil {
		return fmt.Errorf("copying v1beta2.FlowCollector into v1beta1.FlowCollector: %w", err)
	}
	r.Status.Conditions = make([]v1.Condition, len(src.Status.Conditions))
	copy(r.Status.Conditions, src.Status.Conditions)

	r.Spec.Loki.URL = helper.LokiIngesterURL(&src.Spec.Loki)
	r.Spec.Loki.QuerierURL = helper.LokiQuerierURL(&src.Spec.Loki)
	r.Spec.Loki.StatusURL = helper.LokiStatusURL(&src.Spec.Loki)
	r.Spec.Loki.TenantID = helper.LokiTenantID(&src.Spec.Loki)
	switch src.Spec.Loki.Mode {
	case v1beta2.LokiModeManual:
		r.Spec.Loki.AuthToken = src.Spec.Loki.Manual.AuthToken
	case v1beta2.LokiModeLokiStack:
		r.Spec.Loki.AuthToken = v1beta2.LokiAuthForwardUserToken
	}
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(helper.LokiTLS(&src.Spec.Loki), &r.Spec.Loki.TLS, nil); err != nil {
		return fmt.Errorf("copying v1beta2.LokiTLS into v1beta1.LokiTLS: %w", err)
	}
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(helper.LokiStatusTLS(&src.Spec.Loki), &r.Spec.Loki.StatusTLS, nil); err != nil {
		return fmt.Errorf("copying v1beta2.LokiStatusTLS into v1beta1.LokiStatusTLS: %w", err)
	}

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, r)
}

func (r *FlowCollectorList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1beta1_FlowCollectorList_To_v1beta2_FlowCollectorList(r, dst, nil)
}

func (r *FlowCollectorList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.FlowCollectorList)
	return Convert_v1beta2_FlowCollectorList_To_v1beta1_FlowCollectorList(src, r, nil)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorFLP_To_v1beta1_FlowCollectorFLP(in *v1beta2.FlowCollectorFLP, out *FlowCollectorFLP, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorFLP_To_v1beta1_FlowCollectorFLP(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FLPMetrics_To_v1beta1_FLPMetrics(in *v1beta2.FLPMetrics, out *FLPMetrics, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FLPMetrics_To_v1beta1_FLPMetrics(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in *v1beta2.FlowCollectorLoki, out *FlowCollectorLoki, s apiconversion.Scope) error {
	return autoConvert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in *FlowCollectorLoki, out *v1beta2.FlowCollectorLoki, s apiconversion.Scope) error {
	return autoConvert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in, out, s)
}
