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
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/encode"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	flpTracerName     = "flp_tracer"
	flpEncodeSpanName = "flp_encode"
)

type EncodeOtlpTrace struct {
	cfg api.EncodeOtlpTraces
	ctx context.Context
	res *resource.Resource
	tp  *sdktrace.TracerProvider
}

// Encode encodes a metric to be exported
func (e *EncodeOtlpTrace) Encode(entry config.GenericMap) {
	log.Tracef("entering EncodeOtlpTrace. entry = %v", entry)
	tr := e.tp.Tracer(flpTracerName)
	ll := len(e.cfg.SpanSplitter)

	// create parent span
	newCtx, span0 := tr.Start(e.ctx, flpEncodeSpanName)
	attributes := obtainAttributesFromEntry(entry)
	span0.SetAttributes(*attributes...)
	defer span0.End()
	if ll == 0 {
		return
	}
	// for each item in SpanSplitter, make a separate entry for each listed item
	// do not include fields that belong exclusively to other items
	ss := e.cfg.SpanSplitter
	records := make([]config.GenericMap, ll)
	keepItem := make([]bool, ll)
	for i := 0; i < ll; i++ {
		records[i] = make(config.GenericMap)
	}
OUTER:
	for key, value := range entry {
		for i := 0; i < ll; i++ {
			if strings.HasPrefix(key, ss[i]) {
				trimmed := strings.TrimPrefix(key, ss[i])
				records[i][trimmed] = value
				keepItem[i] = true
				continue OUTER
			}
		}
		// if we reach here, the field did not have any of the prefixes.
		// copy it into each of the records
		for i := 0; i < ll; i++ {
			records[i][key] = value
		}
	}
	// only create child spans for records that have a field directly related to their item
	for i := 0; i < ll; i++ {
		if keepItem[i] {
			_, span := tr.Start(newCtx, ss[i])
			attributes := obtainAttributesFromEntry(records[i])
			span.SetAttributes(*attributes...)
			span.End()
		}
	}
}

func (e *EncodeOtlpTrace) Update(_ config.StageParam) {
	log.Warn("EncodeOtlpTrace, update not supported")
}

func NewEncodeOtlpTraces(_ *operational.Metrics, params config.StageParam) (encode.Encoder, error) {
	log.Tracef("entering NewEncodeOtlpTraces \n")
	cfg := api.EncodeOtlpTraces{}
	if params.Encode != nil && params.Encode.OtlpTraces != nil {
		cfg = *params.Encode.OtlpTraces
	}
	log.Debugf("NewEncodeOtlpTraces cfg = %v \n", cfg)

	ctx := context.Background()
	res := newResource()

	tp, err := NewOtlpTracerProvider(ctx, params, res)
	if err != nil {
		return nil, err
	}

	w := &EncodeOtlpTrace{
		cfg: cfg,
		ctx: ctx,
		res: res,
		tp:  tp,
	}
	return w, nil
}
