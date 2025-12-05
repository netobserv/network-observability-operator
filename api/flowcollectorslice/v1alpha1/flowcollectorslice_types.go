package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FlowCollectorSliceSpec defines the desired state of FlowCollectorSlice
type FlowCollectorSliceSpec struct {
	// `subnetLabels` allows to customize subnets and IPs labelling, such as to identify cluster-external workloads or web services.
	// Beware that the subnet labels configured in FlowCollectorSlice are not limited to the flows of the related namespace: any flow
	// in the whole cluster can be labelled using this configuration. However, subnet labels defined in the cluster-scoped FlowCollector take
	// precedence in case of conflicting rules.
	//+optional
	SubnetLabels []SubnetLabel `json:"subnetLabels,omitempty"`

	// `sampling` is an optional sampling interval to apply to this slice. For example, a value of `50` means that 1 matching flow in 50 is sampled.
	//+kubebuilder:validation:Minimum=0
	// +optional
	Sampling int32 `json:"sampling,omitempty"`
}

// SubnetLabel allows to label subnets and IPs, such as to identify cluster-external workloads or web services.
type SubnetLabel struct {
	// List of CIDRs, such as `["1.2.3.4/32"]`.
	//+required
	CIDRs []string `json:"cidrs,omitempty"` // Note, starting with k8s 1.31 / ocp 4.16 there's a new way to validate CIDR such as `+kubebuilder:validation:XValidation:rule="isCIDR(self)",message="field should be in CIDR notation format"`. But older versions would reject the CRD so we cannot implement it now to maintain compatibility.
	// Label name, used to flag matching flows.
	//+required
	Name string `json:"name,omitempty"`
}

// FlowCollectorSliceStatus defines the observed state of FlowCollectorSlice
type FlowCollectorSliceStatus struct {
	// `conditions` represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`
	// Filter that is applied for flow collection
	// +optional
	FilterApplied string `json:"filterApplied"`
	// Number of subnet labels configured
	// +optional
	SubnetLabelsConfigured int `json:"subnetLabelsConfigured"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// FlowMetric is the API allowing to create custom metrics from the collected flow logs.
type FlowCollectorSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowCollectorSliceSpec   `json:"spec,omitempty"`
	Status FlowCollectorSliceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// FlowCollectorSliceList contains a list of FlowCollectorSlice
type FlowCollectorSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowCollectorSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowCollectorSlice{}, &FlowCollectorSliceList{})
}
