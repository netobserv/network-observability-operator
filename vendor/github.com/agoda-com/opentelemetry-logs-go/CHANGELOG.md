# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.6.0] 2025-02-11

### Changed

- opentelemetry updated to 1.34.0

## [v0.5.1] 2024-06-17

## Fixed 

- fix: make ReadableLogRecord accept the "any" type as Body (#38)

## [v0.5.0] 2024-04-21

### Changed

- opentelemetry updated to 1.25.0

### Removed 

- Drop support for [Go 1.20](https://go.dev/doc/go1.20)

## [v0.4.3] 2023-11-02

### Fixed

- fix: race on batch processing (#30)

## [v0.4.2] 2023-10-30

### Fixed

- fix: accept any 2xx status code in otlplogshttp client (#26)
- fix: show the error body when status code is unknown (#27)
- fix: grpc rapid-reset vulnerability (#28)

## [v0.4.1] 2023-10-13

### Fixed

- autoconfiguration always emit error message on initialization (#23)
- fix variables and private methods names (#22)
- merge the logRecord resources with those provided by the logProvider (#21)

## [v0.4.0] 2023-10-02

### Changed

- opentelemetry updated to 1.19.0
- drop compatibility guarantee of Go [1.19](https://go.dev/doc/go1.19)

## [v0.3.0] 2023-09-13

### Changed

- opentelemetry updated to 1.18.0

### Fixed

- stdoutlogs writer parameter was ignored

## [v0.2.0] 2023-08-30

### Changed

- opentelemetry updated to 1.17.0
- `github.com/golang/protobuf/proto` replaced with `google.golang.org/protobuf`
- `otlp/internal` package moved to `otlp/otlplogs/internal`
- more unit tests added

## [v0.1.2] 2023-08-05

### Fixed

- reverted to all-in-one package
- inconsistent v0.1.0 go package

## [v0.1.0] 2023-08-05

### Added

- otlplogsgrpc exporter with `grpc` protocol
- `http/json` protocol supported in otlplogshttp exporter
- `stdout` logs logger
- Package split into separate `otel`, `sdk`, `exporters/otlp/otlplogs` and `exporters/stdout/stdoutlogs` packages
- `OTEL_EXPORTER_OTLP_PROTOCOL` env variable to configure `grpc`, `http/protobuf` and `http/json` otlp formats with OTEL
  logs exporter
- `autoconfigure` sdk package with `OTEL_LOGS_EXPORTER` env variable support with `none`,`otlp` and `logging` options to
  autoconfigure logger provider

## [v0.0.1] 2023-07-25

### Added

- implementation of [Logs Bridge API](https://opentelemetry.io/docs/specs/otel/logs/bridge-api) with Stable API and SDK
  API interfaces.
- Package all-in-one for logs `github.com/agoda-com/opentelemetry-logs-go`
- Package module `semconv`
  with [Logs Exceptions Semantic Conventions](https://opentelemetry.io/docs/specs/otel/logs/semantic_conventions/exceptions/#attributes)
- Package module `logs`
  with [Stable Log Model](https://opentelemetry.io/docs/specs/otel/logs/data-model), [Logger](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#logger)
  and [LoggerProvider](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/#loggerprovider) interfaces
- Package module `sdk` with [Logger SDK](https://opentelemetry.io/docs/specs/otel/logs/sdk/) implementation
- Package module `exporters`
  with [Built-in processors](https://opentelemetry.io/docs/specs/otel/logs/sdk/#built-in-processors), `otlp` interface
  and `noop` and `http/protobuf` exporters
