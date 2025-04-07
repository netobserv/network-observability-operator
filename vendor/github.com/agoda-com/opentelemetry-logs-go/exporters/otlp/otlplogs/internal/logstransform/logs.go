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

package logstransform

import (
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Logs transforms OpenTelemetry LogRecord's into a OTLP ResourceLogs
func Logs(sdl []sdk.ReadableLogRecord) []*logspb.ResourceLogs {

	var resourceLogs []*logspb.ResourceLogs

	for _, sd := range sdl {

		lr := logRecord(sd)

		var is *commonpb.InstrumentationScope
		var schemaURL = ""
		if sd.InstrumentationScope() != nil {
			is = &commonpb.InstrumentationScope{
				Name:    sd.InstrumentationScope().Name,
				Version: sd.InstrumentationScope().Version,
			}
			schemaURL = sd.InstrumentationScope().SchemaURL
		}

		// Create a log resource
		resourceLog := &logspb.ResourceLogs{
			Resource: &resourcepb.Resource{
				Attributes: KeyValues(sd.Resource().Attributes()),
			},
			// provide a resource description if available
			ScopeLogs: []*logspb.ScopeLogs{
				{
					Scope:      is,
					SchemaUrl:  schemaURL,
					LogRecords: []*logspb.LogRecord{lr},
				},
			},
		}

		resourceLogs = append(resourceLogs, resourceLog)
	}

	return resourceLogs
}

func logRecord(record sdk.ReadableLogRecord) *logspb.LogRecord {
	var traceIDBytes []byte
	if record.TraceId() != nil {
		tid := *record.TraceId()
		traceIDBytes = tid[:]
	}
	var spanIDBytes []byte
	if record.SpanId() != nil {
		sid := *record.SpanId()
		spanIDBytes = sid[:]
	}
	var traceFlags byte = 0
	if record.TraceFlags() != nil {
		tf := *record.TraceFlags()
		traceFlags = byte(tf)
	}
	var ts time.Time
	if record.Timestamp() != nil {
		ts = *record.Timestamp()
	} else {
		ts = record.ObservedTimestamp()
	}

	var kv []*commonpb.KeyValue
	if record.Attributes() != nil {
		kv = KeyValues(*record.Attributes())
	}

	var st = ""
	if record.SeverityText() != nil {
		st = *record.SeverityText()
	}

	var sn = logspb.SeverityNumber_SEVERITY_NUMBER_UNSPECIFIED
	if record.SeverityNumber() != nil {
		sn = logspb.SeverityNumber(*record.SeverityNumber())
	}

	logRecord := &logspb.LogRecord{
		TimeUnixNano:         uint64(ts.UnixNano()),
		ObservedTimeUnixNano: uint64(record.ObservedTimestamp().UnixNano()),
		TraceId:              traceIDBytes,                   // provide the associated trace ID if available
		SpanId:               spanIDBytes,                    // provide the associated span ID if available
		Flags:                uint32(traceFlags),             // provide the associated trace flags
		Body:                 valueToAnyValue(record.Body()), // provide the associated log body if available
		Attributes:           kv,                             // provide additional log attributes if available
		SeverityText:         st,
		SeverityNumber:       sn,
	}
	return logRecord
}

func valueToAnyValue(value any) *commonpb.AnyValue {
	if value == nil {
		return nil
	}
	typ := reflect.TypeOf(value)
	val := reflect.ValueOf(value)
	if valueIsNil(typ, val) {
		return nil
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	switch {
	case val.CanFloat():
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_DoubleValue{
				DoubleValue: val.Float(),
			},
		}
	case val.CanInt():
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: val.Int(),
			},
		}
	case val.Kind() == reflect.Bool:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BoolValue{
				BoolValue: val.Bool(),
			},
		}
	case val.CanUint():
		valUInt := val.Uint()
		if valUInt > math.MaxInt64 {
			valUInt = math.MaxInt64
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: int64(valUInt),
			},
		}
	case val.Kind() == reflect.String:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: val.String(),
			},
		}
	case typ.ConvertibleTo(reflect.TypeOf(time.Time{})):
		valTime := val.Convert(reflect.TypeOf(time.Time{})).Interface().(time.Time)
		if valTime.IsZero() {
			return nil
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: valTime.Format(time.RFC3339Nano),
			},
		}
	case (typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) && typ.Elem() == reflect.TypeOf(byte(0)):
		return byteSliceToAnyValue(val)
	case typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array:
		return sliceToAnyValue(val)
	case typ.Kind() == reflect.Map:
		return mapToAnyValue(val)
	case typ.Kind() == reflect.Struct:
		return structToAnyValue(typ, val)
	default:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: val.String(),
			},
		}
	}
}

func byteSliceToAnyValue(val reflect.Value) *commonpb.AnyValue {
	sliceLen := val.Len()
	if sliceLen == 0 {
		return nil
	}
	out := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(byte(1))), sliceLen, sliceLen)
	reflect.Copy(out, val)
	return &commonpb.AnyValue{
		Value: &commonpb.AnyValue_BytesValue{
			BytesValue: out.Interface().([]byte),
		},
	}
}

func sliceToAnyValue(val reflect.Value) *commonpb.AnyValue {
	sliceLen := val.Len()
	if sliceLen == 0 {
		return nil
	}
	elems := make([]*commonpb.AnyValue, 0, sliceLen)
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if !elem.CanInterface() {
			continue
		}
		elems = append(elems, valueToAnyValue(elem.Interface()))
	}
	return &commonpb.AnyValue{
		Value: &commonpb.AnyValue_ArrayValue{
			ArrayValue: &commonpb.ArrayValue{
				Values: elems,
			},
		},
	}
}

func structToAnyValue(typ reflect.Type, val reflect.Value) *commonpb.AnyValue {
	nFields := typ.NumField()
	attrs := make([]*commonpb.KeyValue, 0, nFields)
	for i := 0; i < nFields; i++ {
		fieldType := typ.Field(i)
		if !fieldType.IsExported() {
			continue
		}
		fieldValue := val.Field(i)
		if !fieldValue.CanInterface() {
			continue
		}
		fieldName := fieldType.Name
		if jsonTag, exists := fieldType.Tag.Lookup("json"); exists {
			fieldName = strings.Split(jsonTag, ",")[0]
		}
		fieldAnyValue := valueToAnyValue(fieldValue.Interface())
		if fieldAnyValue == nil {
			continue
		}
		attr := &commonpb.KeyValue{
			Key:   fieldName,
			Value: fieldAnyValue,
		}
		attrs = append(attrs, attr)
	}
	return &commonpb.AnyValue{
		Value: &commonpb.AnyValue_KvlistValue{
			KvlistValue: &commonpb.KeyValueList{
				Values: attrs,
			},
		},
	}
}

func mapToAnyValue(val reflect.Value) *commonpb.AnyValue {
	nFields := val.Len()
	elems := make([]*commonpb.KeyValue, 0, nFields)
	mapKeys := val.MapKeys()
	sort.SliceStable(mapKeys, func(i, j int) bool {
		return mapKeys[i].String() < mapKeys[j].String()
	})
	for _, mapKey := range mapKeys {
		mapValue := val.MapIndex(mapKey)
		if !mapValue.CanInterface() {
			continue
		}
		mapValueAnyValue := valueToAnyValue(mapValue.Interface())
		if mapValueAnyValue == nil {
			continue
		}
		elem := &commonpb.KeyValue{
			Key:   mapKey.String(),
			Value: mapValueAnyValue,
		}
		elems = append(elems, elem)
	}
	return &commonpb.AnyValue{
		Value: &commonpb.AnyValue_KvlistValue{
			KvlistValue: &commonpb.KeyValueList{
				Values: elems,
			},
		},
	}
}

func valueIsNil(typ reflect.Type, val reflect.Value) bool {
	if typ == nil {
		return true
	}
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
