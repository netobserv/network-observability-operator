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
	"encoding/json"
	"fmt"
	"time"

	otel2 "github.com/agoda-com/opentelemetry-logs-go"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	sdklog2 "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc/credentials"
)

// Note:
// As of the writing of this module, go.opentelemetry.io does not provide interfaces for logs.
// We therefore temporarily use agoda-com/opentelemetry-logs-go for logs.
// When go.opentelemetry.io provides interfaces for logs, the code here should be updated to use those interfaces.

const (
	flpOtlpLoggerName      = "flp-otlp-logger"
	defaultTimeInterval    = time.Duration(20 * time.Second)
	flpOtlpResourceVersion = "v0.1.0"
	flpOtlpResourceName    = "netobserv-otlp"
	grpcType               = "grpc"
	httpType               = "http"
)

func NewOtlpTracerProvider(ctx context.Context, params config.StageParam, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	cfg := api.EncodeOtlpTraces{}
	if params.Encode != nil && params.Encode.OtlpTraces != nil {
		cfg = *params.Encode.OtlpTraces
	}
	if cfg.OtlpConnectionInfo == nil {
		return nil, fmt.Errorf("otlptraces missing connection info")
	}
	addr := fmt.Sprintf("%s:%v", cfg.OtlpConnectionInfo.Address, cfg.OtlpConnectionInfo.Port)
	var err error
	var traceProvider *sdktrace.TracerProvider
	var traceExporter *otlptrace.Exporter
	if cfg.ConnectionType == grpcType {
		var expOption otlptracegrpc.Option
		var tlsOption otlptracegrpc.Option
		tlsOption = otlptracegrpc.WithInsecure()
		if cfg.TLS != nil {
			tlsConfig, err := cfg.OtlpConnectionInfo.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig))
		}
		expOption = otlptracegrpc.WithEndpoint(addr)
		traceExporter, err = otlptracegrpc.New(ctx,
			expOption,
			tlsOption,
			otlptracegrpc.WithHeaders(cfg.Headers))
		if err != nil {
			return nil, err
		}
	} else if cfg.ConnectionType == httpType {
		var expOption otlptracehttp.Option
		var tlsOption otlptracehttp.Option
		tlsOption = otlptracehttp.WithInsecure()
		if cfg.TLS != nil {
			tlsConfig, err := cfg.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlptracehttp.WithTLSClientConfig(tlsConfig)
		}
		expOption = otlptracehttp.WithEndpoint(addr)
		traceExporter, err = otlptracehttp.New(ctx,
			expOption,
			tlsOption,
			otlptracehttp.WithHeaders(cfg.Headers))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("must specify grpcaddress or httpaddress")
	}
	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExporter)),
	)

	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func NewOtlpMetricsProvider(ctx context.Context, params config.StageParam, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	cfg := api.EncodeOtlpMetrics{}
	if params.Encode != nil && params.Encode.OtlpMetrics != nil {
		cfg = *params.Encode.OtlpMetrics
	}
	timeInterval := cfg.PushTimeInterval
	if timeInterval.Duration == 0 {
		timeInterval.Duration = defaultTimeInterval
	}
	if cfg.OtlpConnectionInfo == nil {
		return nil, fmt.Errorf("otlpmetrics missing connection info")
	}
	addr := fmt.Sprintf("%s:%v", cfg.OtlpConnectionInfo.Address, cfg.OtlpConnectionInfo.Port)
	var err error
	var meterProvider *sdkmetric.MeterProvider
	if cfg.ConnectionType == grpcType {
		var metricExporter *otlpmetricgrpc.Exporter
		var expOption otlpmetricgrpc.Option
		var tlsOption otlpmetricgrpc.Option
		tlsOption = otlpmetricgrpc.WithInsecure()
		if cfg.TLS != nil {
			tlsConfig, err := cfg.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig))
		}
		expOption = otlpmetricgrpc.WithEndpoint(addr)
		metricExporter, err = otlpmetricgrpc.New(ctx, expOption, tlsOption)
		if err != nil {
			return nil, err
		}
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(timeInterval.Duration))),
		)
	} else if cfg.ConnectionType == httpType {
		var metricExporter *otlpmetrichttp.Exporter
		var expOption otlpmetrichttp.Option
		var tlsOption otlpmetrichttp.Option
		tlsOption = otlpmetrichttp.WithInsecure()
		if cfg.TLS != nil {
			tlsConfig, err := cfg.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlpmetrichttp.WithTLSClientConfig(tlsConfig)
		}
		expOption = otlpmetrichttp.WithEndpoint(addr)
		metricExporter, err = otlpmetrichttp.New(ctx, expOption, tlsOption)
		if err != nil {
			return nil, err
		}
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(timeInterval.Duration))),
		)
	} else {
		return nil, fmt.Errorf("must specify grpcaddress or httpaddress")
	}

	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func NewOtlpLoggerProvider(ctx context.Context, params config.StageParam, res *resource.Resource) (*sdklog2.LoggerProvider, error) {
	cfg := api.EncodeOtlpLogs{}
	if params.Encode != nil && params.Encode.OtlpLogs != nil {
		cfg = *params.Encode.OtlpLogs
	}
	if cfg.OtlpConnectionInfo == nil {
		return nil, fmt.Errorf("otlplogs missing connection info")
	}
	addr := fmt.Sprintf("%s:%v", cfg.OtlpConnectionInfo.Address, cfg.OtlpConnectionInfo.Port)
	var expOption otlplogs.ExporterOption
	if cfg.ConnectionType == grpcType {
		var tlsOption otlplogsgrpc.Option
		tlsOption = otlplogsgrpc.WithInsecure()
		if params.Encode.OtlpLogs.TLS != nil {
			tlsConfig, err := cfg.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlplogsgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig))
		}
		expOption = otlplogs.WithClient(otlplogsgrpc.NewClient(
			otlplogsgrpc.WithEndpoint(addr),
			tlsOption,
			otlplogsgrpc.WithHeaders(cfg.Headers),
		))
	} else if cfg.ConnectionType == httpType {
		var tlsOption otlplogshttp.Option
		tlsOption = otlplogshttp.WithInsecure()
		if params.Encode.OtlpLogs.TLS != nil {
			tlsConfig, err := cfg.TLS.Build()
			if err != nil {
				return nil, err
			}
			tlsOption = otlplogshttp.WithTLSClientConfig(tlsConfig)
		}
		expOption = otlplogs.WithClient(otlplogshttp.NewClient(
			otlplogshttp.WithEndpoint(addr),
			tlsOption,
			otlplogshttp.WithHeaders(cfg.Headers),
		))
	} else {
		return nil, fmt.Errorf("must specify grpcaddress or httpaddress")
	}
	logExporter, err := otlplogs.NewExporter(ctx, expOption)
	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog2.NewLoggerProvider(
		sdklog2.WithBatcher(logExporter),
		sdklog2.WithResource(res),
	)
	otel2.SetLoggerProvider(loggerProvider)
	return loggerProvider, nil
}

func (e *EncodeOtlpLogs) LogWrite(entry config.GenericMap) {
	now := time.Now()
	sn := logs.INFO
	st := "INFO"
	msgByteArray, _ := json.Marshal(entry)
	msg := string(msgByteArray)
	// TODO: Decide whether the content should be delivered as Body or as Attributes
	lrc := logs.LogRecordConfig{
		// Timestamp:         &now, // take timestamp from entry, if present?
		ObservedTimestamp: now,
		SeverityNumber:    &sn,
		SeverityText:      &st,
		Resource:          e.res,
		Body:              &msg,
		Attributes:        obtainAttributesFromEntry(entry),
	}
	logRecord := logs.NewLogRecord(lrc)

	logger := otel2.GetLoggerProvider().Logger(
		flpOtlpLoggerName,
		logs.WithSchemaURL(semconv.SchemaURL),
	)
	logger.Emit(logRecord)
}

func obtainAttributesFromEntry(entry config.GenericMap) *[]attribute.KeyValue {
	// convert the entry fields to Attributes of the message
	var att = make([]attribute.KeyValue, len(entry))
	index := 0
	for k, v := range entry {
		switch v := v.(type) {
		case []string:
			att[index] = attribute.StringSlice(k, v)
		case string:
			att[index] = attribute.String(k, v)
		case []int:
			att[index] = attribute.IntSlice(k, v)
		case []int32:
			valInt64Slice := []int64{}
			for _, valInt32 := range v {
				valInt64, _ := utils.ConvertToInt64(valInt32)
				valInt64Slice = append(valInt64Slice, valInt64)
			}
			att[index] = attribute.Int64Slice(k, valInt64Slice)
		case []int64:
			att[index] = attribute.Int64Slice(k, v)
		case int:
			att[index] = attribute.Int(k, v)
		case int32, int64, int16, uint, uint8, uint16, uint32, uint64:
			valInt, _ := utils.ConvertToInt64(v)
			att[index] = attribute.Int64(k, valInt)
		case []float32:
			valFloat64Slice := []float64{}
			for _, valFloat32 := range v {
				valFloat64, _ := utils.ConvertToFloat64(valFloat32)
				valFloat64Slice = append(valFloat64Slice, valFloat64)
			}
			att[index] = attribute.Float64Slice(k, valFloat64Slice)
		case []float64:
			att[index] = attribute.Float64Slice(k, v)
		case float32:
			valFloat, _ := utils.ConvertToFloat64(v)
			att[index] = attribute.Float64(k, valFloat)
		case float64:
			att[index] = attribute.Float64(k, v)
		case []bool:
			att[index] = attribute.BoolSlice(k, v)
		case bool:
			att[index] = attribute.Bool(k, v)
		case nil:
			// skip this field
			continue
		}
		index++
	}
	addjustedAtt := att[0:index]
	return &addjustedAtt
}

func obtainAttributesFromLabels(labels map[string]string) []attribute.KeyValue {
	// convert the entry fields to Attributes of the message
	var att = make([]attribute.KeyValue, len(labels))
	index := 0
	for k, v := range labels {
		att[index] = attribute.String(k, v)
		index++
	}
	return att
}

func (e *EncodeOtlpMetrics) MetricWrite(_ config.GenericMap) {
	// nothing more to do at present
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(flpOtlpResourceName),
			semconv.ServiceVersion(flpOtlpResourceVersion),
		),
	)
	return r
}
