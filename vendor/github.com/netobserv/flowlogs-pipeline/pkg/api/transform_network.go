/*
 * Copyright (C) 2022 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package api

type TransformNetwork struct {
	Rules          NetworkTransformRules         `yaml:"rules" json:"rules" doc:"list of transform rules, each includes:"`
	KubeConfigPath string                        `yaml:"kubeConfigPath,omitempty" json:"kubeConfigPath,omitempty" doc:"path to kubeconfig file (optional)"`
	ServicesFile   string                        `yaml:"servicesFile,omitempty" json:"servicesFile,omitempty" doc:"path to services file (optional, default: /etc/services)"`
	ProtocolsFile  string                        `yaml:"protocolsFile,omitempty" json:"protocolsFile,omitempty" doc:"path to protocols file (optional, default: /etc/protocols)"`
	IPCategories   []NetworkTransformIPCategory  `yaml:"ipCategories,omitempty" json:"ipCategories,omitempty" doc:"configure IP categories"`
	DirectionInfo  NetworkTransformDirectionInfo `yaml:"directionInfo,omitempty" json:"directionInfo,omitempty" doc:"information to reinterpret flow direction (optional, to use with reinterpret_direction rule)"`
}

func (tn *TransformNetwork) GetServiceFiles() (string, string) {
	p := tn.ProtocolsFile
	if p == "" {
		p = "/etc/protocols"
	}
	s := tn.ServicesFile
	if s == "" {
		s = "/etc/services"
	}
	return p, s
}

type TransformNetworkOperationEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	NetworkAddSubnet            TransformNetworkOperationEnum = "add_subnet"            // add output subnet field from input field and prefix length from parameters field
	NetworkAddLocation          TransformNetworkOperationEnum = "add_location"          // add output location fields from input
	NetworkAddService           TransformNetworkOperationEnum = "add_service"           // add output network service field from input port and parameters protocol field
	NetworkAddKubernetes        TransformNetworkOperationEnum = "add_kubernetes"        // add output kubernetes fields from input
	NetworkAddKubernetesInfra   TransformNetworkOperationEnum = "add_kubernetes_infra"  // add output kubernetes isInfra field from input
	NetworkReinterpretDirection TransformNetworkOperationEnum = "reinterpret_direction" // reinterpret flow direction at the node level (instead of net interface), to ease the deduplication process
	NetworkAddIPCategory        TransformNetworkOperationEnum = "add_ip_category"       // categorize IPs based on known subnets configuration
)

type NetworkTransformRule struct {
	Type            TransformNetworkOperationEnum `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	KubernetesInfra *K8sInfraRule                 `yaml:"kubernetes_infra,omitempty" json:"kubernetes_infra,omitempty" doc:"Kubernetes infra rule configuration"`
	Kubernetes      *K8sRule                      `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty" doc:"Kubernetes rule configuration"`
	AddSubnet       *NetworkAddSubnetRule         `yaml:"add_subnet,omitempty" json:"add_subnet,omitempty" doc:"Add subnet rule configuration"`
	AddLocation     *NetworkGenericRule           `yaml:"add_location,omitempty" json:"add_location,omitempty" doc:"Add location rule configuration"`
	AddIPCategory   *NetworkGenericRule           `yaml:"add_ip_category,omitempty" json:"add_ip_category,omitempty" doc:"Add ip category rule configuration"`
	AddService      *NetworkAddServiceRule        `yaml:"add_service,omitempty" json:"add_service,omitempty" doc:"Add service rule configuration"`
}

type K8sInfraRule struct {
	Inputs        []string       `yaml:"inputs,omitempty" json:"inputs,omitempty" doc:"entry inputs fields"`
	Output        string         `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	InfraPrefixes []string       `yaml:"infra_prefixes,omitempty" json:"infra_prefixes,omitempty" doc:"Namespace prefixes that will be tagged as infra"`
	InfraRefs     []K8sReference `yaml:"infra_refs,omitempty" json:"infra_refs,omitempty" doc:"Additional object references to be tagged as infra"`
}

type K8sReference struct {
	Name      string `yaml:"name,omitempty" json:"name,omitempty" doc:"name of the object"`
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty" doc:"namespace of the object"`
}

type K8sRule struct {
	Input        string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output       string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Assignee     string `yaml:"assignee,omitempty" json:"assignee,omitempty" doc:"value needs to assign to output field"`
	LabelsPrefix string `yaml:"labels_prefix,omitempty" json:"labels_prefix,omitempty" doc:"labels prefix to use to copy input lables, if empty labels will not be copied"`
	AddZone      bool   `yaml:"add_zone,omitempty" json:"add_zone,omitempty" doc:"If true the rule will add the zone"`
}

type NetworkGenericRule struct {
	Input  string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
}

type NetworkAddSubnetRule struct {
	Input      string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	SubnetMask string `yaml:"subnet_mask,omitempty" json:"subnet_mask,omitempty" doc:"subnet mask field"`
}

type NetworkAddServiceRule struct {
	Input    string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output   string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty" doc:"entry protocol field"`
}

type NetworkTransformDirectionInfo struct {
	ReporterIPField    string `yaml:"reporterIPField,omitempty" json:"reporterIPField,omitempty" doc:"field providing the reporter (agent) host IP"`
	SrcHostField       string `yaml:"srcHostField,omitempty" json:"srcHostField,omitempty" doc:"source host field"`
	DstHostField       string `yaml:"dstHostField,omitempty" json:"dstHostField,omitempty" doc:"destination host field"`
	FlowDirectionField string `yaml:"flowDirectionField,omitempty" json:"flowDirectionField,omitempty" doc:"field providing the flow direction in the input entries; it will be rewritten"`
	IfDirectionField   string `yaml:"ifDirectionField,omitempty" json:"ifDirectionField,omitempty" doc:"interface-level field for flow direction, to create in output"`
}

type NetworkTransformRules []NetworkTransformRule

type NetworkTransformIPCategory struct {
	CIDRs []string `yaml:"cidrs,omitempty" json:"cidrs,omitempty" doc:"list of CIDRs to match a category"`
	Name  string   `yaml:"name,omitempty" json:"name,omitempty" doc:"name of the category"`
}
