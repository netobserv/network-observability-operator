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

type TransformGeneric struct {
	Policy TransformGenericOperationEnum `yaml:"policy,omitempty" json:"policy,omitempty" doc:"(enum) key replacement policy; may be one of the following:"`
	Rules  []GenericTransformRule        `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of transform rules, each includes:"`
}

type TransformGenericOperationEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	PreserveOriginalKeys TransformGenericOperationEnum = "preserve_original_keys" // adds new keys in addition to existing keys (default)
	ReplaceKeys          TransformGenericOperationEnum = "replace_keys"           // removes all old keys and uses only the new keys
)

type GenericTransformRule struct {
	Input      string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Multiplier int    `yaml:"multiplier,omitempty" json:"multiplier,omitempty" doc:"scaling factor to compenstate for sampling"`
}

type GenericTransform []GenericTransformRule
