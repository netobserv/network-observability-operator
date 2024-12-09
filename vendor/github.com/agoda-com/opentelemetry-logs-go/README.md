# OpenTelemetry-Logs-Go

[![Go Reference](https://pkg.go.dev/badge/github.com/agoda-com/opentelemetry-logs-go.svg)](https://pkg.go.dev/github.com/agoda-com/opentelemetry-logs-go)
[![codecov](https://codecov.io/github/agoda-com/opentelemetry-logs-go/graph/badge.svg?token=F1NW0R0W75)](https://codecov.io/github/agoda-com/opentelemetry-logs-go)

OpenTelemetry-Logs-Go is the [Go](https://golang.org) implementation of [OpenTelemetry](https://opentelemetry.io/) Logs.
It provides API to directly send logging data to observability platforms. It is an extension of official
[open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) to support Logs.

## Project Life Cycle

This project was created due log module freeze in
official [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) repository:

```
The Logs signal development is halted for this project while we stablize the Metrics SDK. 
No Logs Pull Requests are currently being accepted.
```

This project will be deprecated once official [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)
repository Logs module will have status "Stable".

## Compatibility 

Minimal supported go version `1.21`

## Project packages

| Packages                         | Description                                                                |
|----------------------------------|----------------------------------------------------------------------------|
| [autoconfigure](./autoconfigure) | Autoconfiguration SDK. Allow to configure log exporters with env variables |
| [sdk](./sdk)                     | Opentelemetry Logs SDK                                                     |
| [exporters/otlp](./exporters)    | OTLP format exporter                                                       |
| [exporters/stdout](./exporters)  | Console exporter                                                           |                                                            

## Getting Started

This is an implementation of [Logs Bridge API](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/) and not
intended to use by developers directly. It is provided for logging library authors to build log appenders, which use
this API to bridge between existing logging libraries and the OpenTelemetry log data model.

Example bellow will show how logging library could be instrumented with current API:

```go
package myInstrumentedLogger

import (
	otel "github.com/agoda-com/opentelemetry-logs-go"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	instrumentationName    = "otel/zap"
	instrumentationVersion = "0.0.1"
)

var (
	logger = otel.GetLoggerProvider().Logger(
		instrumentationName,
		logs.WithInstrumentationVersion(instrumentationVersion),
		logs.WithSchemaURL(semconv.SchemaURL),
	)
)

func (c otlpCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {

	lrc := logs.LogRecordConfig{
		Body: &ent.Message,
		...
	}
	logRecord := logs.NewLogRecord(lrc)
	logger.Emit(logRecord)
}
```

and application initialization code:

```go
package main

import (
	"os"
	"context"
	"github.com/agoda-com/opentelemetry-logs-go"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
)

func newResource() *resource.Resource {
	host, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("otlplogs-example"),
		semconv.ServiceVersion("0.0.1"),
		semconv.HostName(host),
	)
}

func main() {
	ctx := context.Background()

	exporter, _ := otlplogs.NewExporter(ctx, otlplogs.WithClient(otlplogshttp.NewClient()))
	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(newResource()),
	)
	otel.SetLoggerProvider(loggerProvider)

	myInstrumentedLogger.Info("Hello OpenTelemetry")
}
```

## References

Logger Bridge API implementations for `zap`, `slog`, `zerolog` and other
loggers can be found in https://github.com/agoda-com/opentelemetry-go

