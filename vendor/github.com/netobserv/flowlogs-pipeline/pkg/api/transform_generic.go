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
	Policy string                 `yaml:"policy,omitempty" json:"policy,omitempty" enum:"TransformGenericOperationEnum" doc:"key replacement policy; may be one of the following:"`
	Rules  []GenericTransformRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of transform rules, each includes:"`
}

type TransformGenericOperationEnum struct {
	PreserveOriginalKeys string `yaml:"preserve_original_keys" json:"preserve_original_keys" doc:"adds new keys in addition to existing keys (default)"`
	ReplaceKeys          string `yaml:"replace_keys" json:"replace_keys" doc:"removes all old keys and uses only the new keys"`
}

func TransformGenericOperationName(operation string) string {
	return GetEnumName(TransformGenericOperationEnum{}, operation)
}

type GenericTransformRule struct {
	Input      string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Multiplier int    `yaml:"multiplier,omitempty" json:"multiplier,omitempty" doc:"scaling factor to compenstate for sampling"`
}

type GenericTransform []GenericTransformRule
