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

import (
	"errors"
	"fmt"

	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type WriteLoki struct {
	URL            string                       `yaml:"url,omitempty" json:"url,omitempty" doc:"the address of an existing Loki service to push the flows to"`
	TenantID       string                       `yaml:"tenantID,omitempty" json:"tenantID,omitempty" doc:"identifies the tenant for the request"`
	BatchWait      string                       `yaml:"batchWait,omitempty" json:"batchWait,omitempty" doc:"maximum amount of time to wait before sending a batch"`
	BatchSize      int                          `yaml:"batchSize,omitempty" json:"batchSize,omitempty" doc:"maximum batch size (in bytes) of logs to accumulate before sending"`
	Timeout        string                       `yaml:"timeout,omitempty" json:"timeout,omitempty" doc:"maximum time to wait for a server to respond to a request"`
	MinBackoff     string                       `yaml:"minBackoff,omitempty" json:"minBackoff,omitempty" doc:"initial backoff time for client connection between retries"`
	MaxBackoff     string                       `yaml:"maxBackoff,omitempty" json:"maxBackoff,omitempty" doc:"maximum backoff time for client connection between retries"`
	MaxRetries     int                          `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty" doc:"maximum number of retries for client connections"`
	Labels         []string                     `yaml:"labels,omitempty" json:"labels,omitempty" doc:"map of record fields to be used as labels"`
	StaticLabels   model.LabelSet               `yaml:"staticLabels,omitempty" json:"staticLabels,omitempty" doc:"map of common labels to set on each flow"`
	IgnoreList     []string                     `yaml:"ignoreList,omitempty" json:"ignoreList,omitempty" doc:"map of record fields to be removed from the record"`
	ClientConfig   *promConfig.HTTPClientConfig `yaml:"clientConfig,omitempty" json:"clientConfig,omitempty" doc:"clientConfig"`
	TimestampLabel model.LabelName              `yaml:"timestampLabel,omitempty" json:"timestampLabel,omitempty" doc:"label to use for time indexing"`
	// TimestampScale provides the scale in time of the units from the timestamp
	// E.g. UNIX timescale is '1s' (one second) while other clock sources might have
	// scales of '1ms' (one millisecond) or just '1' (one nanosecond)
	// Default value is '1s'
	TimestampScale string `yaml:"timestampScale,omitempty" json:"timestampScale,omitempty" doc:"timestamp units scale (e.g. for UNIX = 1s)"`
}

func GetWriteLokiDefaults() WriteLoki {
	return WriteLoki{
		URL:            "http://loki:3100/",
		BatchWait:      "1s",
		BatchSize:      100 * 1024,
		Timeout:        "10s",
		MinBackoff:     "1s",
		MaxBackoff:     "5m",
		MaxRetries:     10,
		StaticLabels:   model.LabelSet{},
		TimestampLabel: "TimeReceived",
		TimestampScale: "1s",
	}
}

func (wl *WriteLoki) Validate() error {
	if wl == nil {
		return errors.New("you must provide a configuration")
	}
	if wl.TimestampScale == "" {
		return errors.New("timestampUnit must be a valid Duration > 0 (e.g. 1m, 1s or 1ms)")
	}
	if wl.URL == "" {
		return errors.New("url can't be empty")
	}
	if wl.BatchSize <= 0 {
		return fmt.Errorf("invalid batchSize: %v. Required > 0", wl.BatchSize)
	}
	return nil
}
