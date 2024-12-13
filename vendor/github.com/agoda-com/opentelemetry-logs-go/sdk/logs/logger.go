/*
Copyright Agoda Services Co.,Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	"github.com/agoda-com/opentelemetry-logs-go/semconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

type logger struct {
	provider             *LoggerProvider
	instrumentationScope instrumentation.Scope
}

var _ logs.Logger = &logger{}

func (l logger) Emit(logRecord logs.LogRecord) {
	lps := l.provider.getLogRecordProcessorStates()
	if len(lps) == 0 {
		return
	}

	pr, err := resource.Merge(l.provider.resource, logRecord.Resource())
	if err != nil {
		return
	}

	elr := &exportableLogRecord{
		timestamp:            logRecord.Timestamp(),
		observedTimestamp:    logRecord.ObservedTimestamp(),
		traceId:              logRecord.TraceId(),
		spanId:               logRecord.SpanId(),
		traceFlags:           logRecord.TraceFlags(),
		severityText:         logRecord.SeverityText(),
		severityNumber:       logRecord.SeverityNumber(),
		body:                 logRecord.Body(),
		resource:             pr,
		instrumentationScope: logRecord.InstrumentationScope(),
		attributes:           logRecord.Attributes(),
	}

	for _, lp := range lps {
		lp.lp.OnEmit(elr)
	}
}

// ReadableLogRecord Log structure
// see https://opentelemetry.io/docs/specs/otel/logs/data-model/#log-and-event-record-definition
// see https://opentelemetry.io/docs/specs/otel/logs/sdk/#readablelogrecord
type ReadableLogRecord interface {
	// Timestamp Time when the event occurred.
	Timestamp() *time.Time
	// ObservedTimestamp	Time when the event was observed.
	ObservedTimestamp() time.Time
	// TraceId Request trace id.
	TraceId() *trace.TraceID
	// SpanId Request span id.
	SpanId() *trace.SpanID
	// TraceFlags W3C trace flag.
	TraceFlags() *trace.TraceFlags
	// SeverityText This is the original string representation of the severityNumber as it is known at the source
	SeverityText() *string
	// SeverityNumber	Numerical value of the severityNumber.
	SeverityNumber() *logs.SeverityNumber
	// Body The body of the log record.
	Body() *string
	// Resource 	Describes the source of the log.
	Resource() *resource.Resource
	// InstrumentationScope returns information about the instrumentation
	// scope that created the log.
	InstrumentationScope() *instrumentation.Scope
	// Attributes describe the aspects of the event.
	Attributes() *[]attribute.KeyValue

	// A private method to prevent users implementing the
	// interface and so future additions to it will not
	// violate compatibility.
	private()
}

type ReadWriteLogRecord interface {
	SetResource(resource *resource.Resource)
	// RecordException message, stacktrace, type
	RecordException(*string, *string, *string)
	ReadableLogRecord
}

// exportableLogRecord is an implementation of the OpenTelemetry Log API
// representing the individual component of a log.
type exportableLogRecord struct {
	// mu protects the contents of this log.
	mu                   sync.Mutex
	timestamp            *time.Time
	observedTimestamp    time.Time
	traceId              *trace.TraceID
	spanId               *trace.SpanID
	traceFlags           *trace.TraceFlags
	severityText         *string
	severityNumber       *logs.SeverityNumber
	body                 *string
	resource             *resource.Resource
	instrumentationScope *instrumentation.Scope
	attributes           *[]attribute.KeyValue
}

// newReadWriteLogRecord create
// This method may change in the future
func newReadWriteLogRecord(
	ctx *trace.SpanContext,
	body *string,
	severityText *string,
	severityNumber *logs.SeverityNumber,
	resource *resource.Resource,
	instrumentationScope *instrumentation.Scope,
	attributes *[]attribute.KeyValue,
	timestamp *time.Time) ReadWriteLogRecord {

	traceId := ctx.TraceID()
	spanId := ctx.SpanID()
	traceFlags := ctx.TraceFlags()

	return &exportableLogRecord{
		timestamp:            timestamp,
		observedTimestamp:    time.Now(),
		traceId:              &traceId,
		spanId:               &spanId,
		traceFlags:           &traceFlags,
		severityText:         severityText,
		severityNumber:       severityNumber,
		body:                 body,
		resource:             resource,
		instrumentationScope: instrumentationScope,
		attributes:           attributes,
	}
}

func (r *exportableLogRecord) SetResource(resource *resource.Resource) { r.resource = resource }

// RecordException helper to add Exception related information as attributes of Log Record
// see https://opentelemetry.io/docs/specs/otel/logs/semantic_conventions/exceptions/#recording-an-exception
func (r *exportableLogRecord) RecordException(message *string, stacktrace *string, exceptionType *string) {
	if message == nil && exceptionType == nil {
		// one of the fields must present
		return
	}
	if message != nil {
		*r.attributes = append(*r.attributes, semconv.ExceptionMessage(*message))
	}
	if stacktrace != nil {
		*r.attributes = append(*r.attributes, semconv.ExceptionStacktrace(*stacktrace))
	}
	if exceptionType != nil {
		*r.attributes = append(*r.attributes, semconv.ExceptionType(*exceptionType))
	}
}

func (r *exportableLogRecord) Timestamp() *time.Time        { return r.timestamp }
func (r *exportableLogRecord) ObservedTimestamp() time.Time { return r.observedTimestamp }
func (r *exportableLogRecord) TraceId() *trace.TraceID      { return r.traceId }

func (r *exportableLogRecord) SpanId() *trace.SpanID         { return r.spanId }
func (r *exportableLogRecord) TraceFlags() *trace.TraceFlags { return r.traceFlags }
func (r *exportableLogRecord) InstrumentationScope() *instrumentation.Scope {
	return r.instrumentationScope
}
func (r *exportableLogRecord) SeverityText() *string                { return r.severityText }
func (r *exportableLogRecord) SeverityNumber() *logs.SeverityNumber { return r.severityNumber }
func (r *exportableLogRecord) Body() *string                        { return r.body }
func (r *exportableLogRecord) Resource() *resource.Resource         { return r.resource }
func (r *exportableLogRecord) Attributes() *[]attribute.KeyValue    { return r.attributes }
func (r *exportableLogRecord) private()                             {}
