/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package transform

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
)

var glog = logrus.WithField("component", "transform.Generic")

type Generic struct {
	policy api.TransformGenericOperationEnum
	rules  []api.GenericTransformRule
}

// Transform transforms a flow to a new set of keys
func (g *Generic) Transform(entry config.GenericMap) (config.GenericMap, bool) {
	var outputEntry config.GenericMap
	ok := true
	glog.Tracef("Transform input = %v", entry)
	if g.policy != "replace_keys" {
		outputEntry = entry.Copy()
	} else {
		outputEntry = config.GenericMap{}
	}
	for _, transformRule := range g.rules {
		if transformRule.Multiplier != 0 {
			ok = g.performMultiplier(entry, transformRule, outputEntry)
		} else {
			outputEntry[transformRule.Output] = entry[transformRule.Input]
		}
	}
	glog.Tracef("Transform output = %v", outputEntry)
	return outputEntry, ok
}

func (g *Generic) performMultiplier(entry config.GenericMap, transformRule api.GenericTransformRule, outputEntry config.GenericMap) bool {
	ok := true
	switch val := entry[transformRule.Input].(type) {
	case int:
		outputEntry[transformRule.Output] = transformRule.Multiplier * val
	case uint:
		outputEntry[transformRule.Output] = uint(transformRule.Multiplier) * val
	case int8:
		outputEntry[transformRule.Output] = int8(transformRule.Multiplier) * val
	case uint8:
		outputEntry[transformRule.Output] = uint8(transformRule.Multiplier) * val
	case int16:
		outputEntry[transformRule.Output] = int16(transformRule.Multiplier) * val
	case uint16:
		outputEntry[transformRule.Output] = uint16(transformRule.Multiplier) * val
	case int32:
		outputEntry[transformRule.Output] = int32(transformRule.Multiplier) * val
	case uint32:
		outputEntry[transformRule.Output] = uint32(transformRule.Multiplier) * val
	case int64:
		outputEntry[transformRule.Output] = int64(transformRule.Multiplier) * val
	case uint64:
		outputEntry[transformRule.Output] = uint64(transformRule.Multiplier) * val
	case float32:
		outputEntry[transformRule.Output] = float32(transformRule.Multiplier) * val
	case float64:
		outputEntry[transformRule.Output] = float64(transformRule.Multiplier) * val
	default:
		ok = false
		glog.Errorf("%s not of numerical type; cannot perform multiplication", transformRule.Output)
	}
	return ok
}

// NewTransformGeneric create a new transform
func NewTransformGeneric(params config.StageParam) (Transformer, error) {
	glog.Debugf("entering NewTransformGeneric")
	genConfig := api.TransformGeneric{}
	if params.Transform != nil && params.Transform.Generic != nil {
		genConfig = *params.Transform.Generic
	}
	glog.Debugf("params.Transform.Generic = %v", genConfig)
	rules := genConfig.Rules
	policy := genConfig.Policy
	switch policy {
	case api.ReplaceKeys, api.PreserveOriginalKeys, "":
		// valid; nothing to do
		glog.Infof("NewTransformGeneric, policy = %s", policy)
	default:
		glog.Panicf("unknown policy %s for transform.generic", policy)
	}
	transformGeneric := &Generic{
		policy: policy,
		rules:  rules,
	}
	glog.Debugf("transformGeneric = %v", transformGeneric)
	return transformGeneric, nil
}
