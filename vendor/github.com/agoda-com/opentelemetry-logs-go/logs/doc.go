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

/*
Package logs provides an implementation of the logging part of the
OpenTelemetry API.

This package defines a log backend API. The API is not intended to be called by application developers directly.
It is provided for logging library authors to build log appenders, which use this API to bridge between existing
logging libraries and the OpenTelemetry log data model.

To participate in logging a LogRecord needs to be created for the
operation being performed as part of a logging workflow. In its simplest form:

		var logger logger.Logger

		func init() {
			logger = otel.Logger()
		}

		func operation(ctx context.Context) {
	        logRecord := logger.NewLogRecord(..)
	        logger.Emit(logRecord)
		}

A Logger is unique to the instrumentation and is used to create Logs.
Instrumentation should be designed to accept a LoggerProvider from which it
can create its own unique Logger. Alternatively, the registered global
LoggerProvider from the github.com/agoda-com/opentelemetry-logs-go package can be used as
a default.

	const (
		name    = "instrumentation/package/name"
		version = "0.1.0"
	)

	type Instrumentation struct {
		logger logging.Logger
	}

	func NewInstrumentation(tp logging.LoggerProvider) *Instrumentation {
		if lp == nil {
			lp = otel.LoggerProvider()
		}
		return &Instrumentation{
			logger: lp.Logger(name, logs.WithInstrumentationVersion(version)),
		}
	}

	func operation(ctx context.Context, inst *Instrumentation) {

		// ...
	}
*/
package logs // import "github.com/agoda-com/opentelemetry-logs-go/logs"
