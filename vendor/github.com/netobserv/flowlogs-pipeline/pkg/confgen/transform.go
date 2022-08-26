/*
 * Copyright (C) 2021 IBM, Inc.
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

package confgen

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

func (cg *ConfGen) parseTransport(transform *map[string]interface{}) (*api.TransformNetwork, error) {
	var jsoniterJson = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := jsoniterJson.Marshal(transform)
	if err != nil {
		log.Debugf("jsoniterJson.Marshal err: %v ", err)
		return nil, err
	}

	var jsonNetworkTransform api.TransformNetwork
	err = config.JsonUnmarshalStrict(b, &jsonNetworkTransform)
	if err != nil {
		log.Debugf("Unmarshal transform.TransformNetwork err: %v ", err)
		return nil, err
	}

	cg.transformRules = append(cg.transformRules, jsonNetworkTransform.Rules...)
	return &jsonNetworkTransform, nil
}
