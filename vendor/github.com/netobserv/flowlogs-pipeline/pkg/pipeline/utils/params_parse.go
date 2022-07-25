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

package utils

import (
	"encoding/json"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

// ParamString returns its corresponding (json) string from config.parameters for specified params structure
func ParamString(params config.StageParam, stage string, stageType string) string {
	log.Debugf("entering paramString")
	log.Debugf("params = %v, stage = %s, stageType = %s", params, stage, stageType)

	var configMap []map[string]interface{}
	var err error
	err = json.Unmarshal([]byte(config.Opt.Parameters), &configMap)
	if err != nil {
		return ""
	}
	log.Debugf("configMap = %v", configMap)

	var returnBytes []byte
	for index := range config.Parameters {
		paramsEntry := &config.Parameters[index]
		if params.Name == paramsEntry.Name {
			log.Debugf("paramsEntry = %v", paramsEntry)
			log.Debugf("data[index][stage] = %v", configMap[index][stage])
			// convert back to string
			subField := configMap[index][stage].(map[string]interface{})
			log.Debugf("subField = %v", subField)
			returnBytes, err = json.Marshal(subField[stageType])
			if err != nil {
				return ""
			}
			break
		}
	}
	log.Debugf("returnBytes = %s", string(returnBytes))
	return string(returnBytes)
}
