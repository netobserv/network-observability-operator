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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"time"
)

// LogRecordConfig contains mutable fields usable for constructing
// an immutable LogRecord.
type LogRecordConfig struct {
	Timestamp         *time.Time
	ObservedTimestamp time.Time
	TraceId           *trace.TraceID
	SpanId            *trace.SpanID
	TraceFlags        *trace.TraceFlags
	SeverityText      *string
	SeverityNumber    *SeverityNumber
	// Deprecated: use BodyAny instead.
	Body                 *string
	BodyAny              any
	Resource             *resource.Resource
	InstrumentationScope *instrumentation.Scope
	Attributes           *[]attribute.KeyValue
}

// NewLogRecord constructs a LogRecord using values from the provided
// LogRecordConfig.
func NewLogRecord(config LogRecordConfig) LogRecord {
	if config.BodyAny == nil && config.Body != nil {
		config.BodyAny = *config.Body
	}
	return LogRecord{
		timestamp:            config.Timestamp,
		observedTimestamp:    config.ObservedTimestamp,
		traceId:              config.TraceId,
		spanId:               config.SpanId,
		traceFlags:           config.TraceFlags,
		severityText:         config.SeverityText,
		severityNumber:       config.SeverityNumber,
		body:                 config.BodyAny,
		resource:             config.Resource,
		instrumentationScope: config.InstrumentationScope,
		attributes:           config.Attributes,
	}
}

// LogRecord is an implementation of the OpenTelemetry Log API
// representing the individual component of a log.
// see https://opentelemetry.io/docs/specs/otel/logs/data-model/#log-and-event-record-definition
type LogRecord struct {
	timestamp            *time.Time
	observedTimestamp    time.Time
	traceId              *trace.TraceID
	spanId               *trace.SpanID
	traceFlags           *trace.TraceFlags
	severityText         *string
	severityNumber       *SeverityNumber
	body                 any
	resource             *resource.Resource
	instrumentationScope *instrumentation.Scope
	attributes           *[]attribute.KeyValue
}

func (l LogRecord) Timestamp() *time.Time                        { return l.timestamp }
func (l LogRecord) ObservedTimestamp() time.Time                 { return l.observedTimestamp }
func (l LogRecord) TraceId() *trace.TraceID                      { return l.traceId }
func (l LogRecord) SpanId() *trace.SpanID                        { return l.spanId }
func (l LogRecord) TraceFlags() *trace.TraceFlags                { return l.traceFlags }
func (l LogRecord) SeverityText() *string                        { return l.severityText }
func (l LogRecord) SeverityNumber() *SeverityNumber              { return l.severityNumber }
func (l LogRecord) Body() any                                    { return l.body }
func (l LogRecord) Resource() *resource.Resource                 { return l.resource }
func (l LogRecord) InstrumentationScope() *instrumentation.Scope { return l.instrumentationScope }
func (l LogRecord) Attributes() *[]attribute.KeyValue            { return l.attributes }
func (l LogRecord) private()                                     {}

// SeverityNumber Possible values for LogRecord.SeverityNumber.
type SeverityNumber int32

const (
	// UNSPECIFIED is the default SeverityNumber, it MUST NOT be used.
	UNSPECIFIED SeverityNumber = 0
	TRACE       SeverityNumber = 1
	TRACE2      SeverityNumber = 2
	TRACE3      SeverityNumber = 3
	TRACE4      SeverityNumber = 4
	DEBUG       SeverityNumber = 5
	DEBUG2      SeverityNumber = 6
	DEBUG3      SeverityNumber = 7
	DEBUG4      SeverityNumber = 8
	INFO        SeverityNumber = 9
	INFO2       SeverityNumber = 10
	INFO3       SeverityNumber = 11
	INFO4       SeverityNumber = 12
	WARN        SeverityNumber = 13
	WARN2       SeverityNumber = 14
	WARN3       SeverityNumber = 15
	WARN4       SeverityNumber = 16
	ERROR       SeverityNumber = 17
	ERROR2      SeverityNumber = 18
	ERROR3      SeverityNumber = 19
	ERROR4      SeverityNumber = 20
	FATAL       SeverityNumber = 21
	FATAL2      SeverityNumber = 22
	FATAL3      SeverityNumber = 23
	FATAL4      SeverityNumber = 24
)

// Logger is the creator of Logs
type Logger interface {
	// Emit emits a log record
	Emit(logRecord LogRecord)
}

// LoggerProvider provides Loggers that are used by instrumentation code to
// log computational workflows.
//
// A LoggerProvider is the collection destination of logs
// provides, it represents a unique telemetry collection pipeline. How that
// pipeline is defined, meaning how those Logs are collected, processed, and
// where they are exported, depends on its implementation. Instrumentation
// authors do not need to define this implementation, rather just use the
// provided Loggers to instrument code.
type LoggerProvider interface {
	// Logger returns a unique Logger scoped to be used by instrumentation code
	// to log computational workflows. The scope and identity of that
	// instrumentation code is uniquely defined by the name and options passed.
	//
	// The passed name needs to uniquely identify instrumentation code.
	// Therefore, it is recommended that name is the Go package name of the
	// library providing instrumentation (note: not the code being
	// instrumented). Instrumentation libraries can have multiple versions,
	// therefore, the WithInstrumentationVersion option should be used to
	// distinguish these different codebases. Additionally, instrumentation
	// libraries may sometimes use traces to communicate different domains of
	// workflow data (i.e. using logs to communicate workflow events only). If
	// this is the case, the WithScopeAttributes option should be used to
	// uniquely identify Loggers that handle the different domains of workflow
	// data.
	//
	// If the same name and options are passed multiple times, the same Logger
	// will be returned (it is up to the implementation if this will be the
	// same underlying instance of that Logger or not). It is not necessary to
	// call this multiple times with the same name and options to get an
	// up-to-date Logger. All implementations will ensure any LoggerProvider
	// configuration changes are propagated to all provided Loggers.
	//
	// If name is empty, then an implementation defined default name will be
	// used instead.
	//
	// This method is safe to call concurrently.
	Logger(name string, options ...LoggerOption) Logger
}
