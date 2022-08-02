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

func (cg *ConfGen) parseEncode(encode *map[string]interface{}, followAggregate bool) (*api.PromEncode, error) {
	var jsoniterJson = jsoniter.ConfigCompatibleWithStandardLibrary
	promEncode := (*encode)["prom"]
	b, err := jsoniterJson.Marshal(promEncode)
	if err != nil {
		log.Debugf("jsoniterJson.Marshal err: %v ", err)
		return nil, err
	}

	var prom api.PromEncode
	err = config.JsonUnmarshalStrict(b, &prom)
	if err != nil {
		log.Debugf("Unmarshal aggregate.Definitions err: %v ", err)
		return nil, err
	}

	// Histograms built from Aggregate need to be flagged as they are handled in a different way in PromEncode
	if followAggregate {
		for i := range prom.Metrics {
			if prom.Metrics[i].Type == "histogram" {
				prom.Metrics[i].Type = "agg_histogram"
			}
		}
	}

	cg.promMetrics = append(cg.promMetrics, prom.Metrics...)
	return &prom, nil
}
