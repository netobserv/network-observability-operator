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

const (
	OpAddSubnet            = "add_subnet"
	OpAddLocation          = "add_location"
	OpAddService           = "add_service"
	OpAddKubernetes        = "add_kubernetes"
	OpAddKubernetesInfra   = "add_kubernetes_infra"
	OpReinterpretDirection = "reinterpret_direction"
	OpAddIPCategory        = "add_ip_category"
)

type TransformNetworkOperationEnum struct {
	AddSubnet            string `yaml:"add_subnet" json:"add_subnet" doc:"add output subnet field from input field and prefix length from parameters field"`
	AddLocation          string `yaml:"add_location" json:"add_location" doc:"add output location fields from input"`
	AddService           string `yaml:"add_service" json:"add_service" doc:"add output network service field from input port and parameters protocol field"`
	AddKubernetes        string `yaml:"add_kubernetes" json:"add_kubernetes" doc:"add output kubernetes fields from input"`
	AddKubernetesInfra   string `yaml:"add_kubernetes_infra" json:"add_kubernetes_infra" doc:"add output kubernetes isInfra field from input"`
	ReinterpretDirection string `yaml:"reinterpret_direction" json:"reinterpret_direction" doc:"reinterpret flow direction at the node level (instead of net interface), to ease the deduplication process"`
	AddIPCategory        string `yaml:"add_ip_category" json:"add_ip_category" doc:"categorize IPs based on known subnets configuration"`
}

func TransformNetworkOperationName(operation string) string {
	return GetEnumName(TransformNetworkOperationEnum{}, operation)
}

type NetworkTransformRule struct {
	Input           string        `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output          string        `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Type            string        `yaml:"type,omitempty" json:"type,omitempty" enum:"TransformNetworkOperationEnum" doc:"one of the following:"`
	Parameters      string        `yaml:"parameters,omitempty" json:"parameters,omitempty" doc:"parameters specific to type"`
	Assignee        string        `yaml:"assignee,omitempty" json:"assignee,omitempty" doc:"value needs to assign to output field"`
	KubernetesInfra *K8sInfraRule `yaml:"kubernetes_infra,omitempty" json:"kubernetes_infra,omitempty" doc:"Kubernetes infra rule specific configuration"`
	Kubernetes      *K8sRule      `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty" doc:"Kubernetes rule specific configuration"`
}

type K8sInfraRule struct {
	Inputs      []string `yaml:"inputs,omitempty" json:"inputs,omitempty" doc:"entry inputs fields"`
	Output      string   `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	InfraPrefix string   `yaml:"infra_prefixes,omitempty" json:"infra_prefixes,omitempty" doc:"Namespace prefixes that will be tagged as infra"`
}

type K8sRule struct {
	AddZone bool `yaml:"add_zone,omitempty" json:"add_zone,omitempty" doc:"If true the rule will add the zone"`
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
