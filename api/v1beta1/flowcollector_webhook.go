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
	"github.com/netobserv/network-observability-operator/pkg/metrics"
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
		return err
	}
	dst.Spec.Loki = restored.Spec.Loki

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
func Convert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in *v1beta2.FlowCollectorLoki, out *FlowCollectorLoki, s apiconversion.Scope) error {
	manual := helper.NewLokiConfig(in)
	out.URL = manual.IngesterURL
	out.QuerierURL = manual.QuerierURL
	out.StatusURL = manual.StatusURL
	out.TenantID = manual.TenantID
	out.AuthToken = manual.AuthToken
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(&manual.TLS, &out.TLS, nil); err != nil {
		return fmt.Errorf("copying Loki v1beta2 TLS into v1beta1 TLS: %w", err)
	}
	if err := Convert_v1beta2_ClientTLS_To_v1beta1_ClientTLS(&manual.StatusTLS, &out.StatusTLS, nil); err != nil {
		return fmt.Errorf("copying Loki v1beta2 StatusTLS into v1beta1 StatusTLS: %w", err)
	}
	return autoConvert_v1beta2_FlowCollectorLoki_To_v1beta1_FlowCollectorLoki(in, out, s)
}

// This function need to be manually created because conversion-gen not able to create it intentionally because
// we have new defined fields in v1beta2 not in v1beta1
// nolint:golint,stylecheck,revive
func Convert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in *FlowCollectorLoki, out *v1beta2.FlowCollectorLoki, s apiconversion.Scope) error {
	out.Mode = v1beta2.LokiModeManual
	out.Manual = v1beta2.LokiManualParams{
		IngesterURL: in.URL,
		QuerierURL:  in.QuerierURL,
		StatusURL:   in.StatusURL,
		TenantID:    in.TenantID,
		AuthToken:   in.AuthToken,
	}
	// fallback on ingester url if querier is not set
	if len(out.Manual.QuerierURL) == 0 {
		out.Manual.QuerierURL = out.Manual.IngesterURL
	}
	if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&in.TLS, &out.Manual.TLS, nil); err != nil {
		return fmt.Errorf("copying v1beta1.Loki.TLS into v1beta2.Loki.Manual.TLS: %w", err)
	}
	if err := Convert_v1beta1_ClientTLS_To_v1beta2_ClientTLS(&in.StatusTLS, &out.Manual.StatusTLS, nil); err != nil {
		return fmt.Errorf("copying v1beta1.Loki.StatusTLS into v1beta2.Loki.Manual.StatusTLS: %w", err)
	}
	return autoConvert_v1beta1_FlowCollectorLoki_To_v1beta2_FlowCollectorLoki(in, out, s)
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
func Convert_v1beta1_FLPMetrics_To_v1beta2_FLPMetrics(in *FLPMetrics, out *v1beta2.FLPMetrics, s apiconversion.Scope) error {
	err := autoConvert_v1beta1_FLPMetrics_To_v1beta2_FLPMetrics(in, out, s)
	if err != nil {
		return err
	}
	out.IncludeList = metrics.GetAsIncludeList(in.IgnoreTags, out.IncludeList)
	return nil
}
