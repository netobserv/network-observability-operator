/*
 * Copyright (C) 2023 IBM, Inc.
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

package opentelemetry

import (
	"context"

	sdklog "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/encode"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/resource"
)

type EncodeOtlpLogs struct {
	cfg api.EncodeOtlpLogs
	ctx context.Context
	res *resource.Resource
	lp  *sdklog.LoggerProvider
}

// Encode encodes a log entry to be exported
func (e *EncodeOtlpLogs) Encode(entry config.GenericMap) {
	log.Tracef("entering EncodeOtlpLogs. entry = %v", entry)
	e.LogWrite(entry)
}

func (e *EncodeOtlpLogs) Update(_ config.StageParam) {
	log.Warn("EncodeOtlpLogs, update not supported")
}

func NewEncodeOtlpLogs(_ *operational.Metrics, params config.StageParam) (encode.Encoder, error) {
	log.Tracef("entering NewEncodeOtlpLogs \n")
	cfg := api.EncodeOtlpLogs{}
	if params.Encode != nil && params.Encode.OtlpLogs != nil {
		cfg = *params.Encode.OtlpLogs
	}
	log.Debugf("NewEncodeOtlpLogs cfg = %v \n", cfg)

	ctx := context.Background()
	res := newResource()

	lp, err := NewOtlpLoggerProvider(ctx, params, res)
	if err != nil {
		return nil, err
	}

	w := &EncodeOtlpLogs{
		cfg: cfg,
		ctx: ctx,
		res: res,
		lp:  lp,
	}
	return w, nil
}
