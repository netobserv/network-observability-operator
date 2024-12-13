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

package semconv

import "go.opentelemetry.io/otel/attribute"

// Describes Log Record attributes.
// see also https://opentelemetry.io/docs/specs/otel/logs/semantic_conventions/exceptions/#attributes
const (
	// ExceptionMessageKey is the attribute Key conforming to the "exception.message"
	// semantic conventions. It represents the exception message.
	//
	// Type: string
	// RequirementLevel: Required
	// Stability: stable
	ExceptionMessageKey = attribute.Key("exception.message")

	// ExceptionStacktraceKey is the attribute Key conforming to the "exception.stacktrace"
	// semantic conventions. It represents the stacktrace message of exception.
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: stable
	ExceptionStacktraceKey = attribute.Key("exception.stacktrace")

	// ExceptionTypeKey is the attribute Key conforming to the "exception.type"
	// semantic conventions. It represents the type of exception
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: stable
	ExceptionTypeKey = attribute.Key("exception.type")
)

// ExceptionMessage returns an attribute KeyValue conforming to the
// "exception.message" semantic conventions. It represents the exception
// message
// Examples: Division by zero; Can't convert 'int' object to str implicitly
func ExceptionMessage(val string) attribute.KeyValue {
	return ExceptionMessageKey.String(val)
}

// ExceptionStacktrace returns an attribute KeyValue conforming to the
// "exception.stacktrace" semantic conventions. It represents the exception
// stacktrace
// Examples: Exception in thread "main" java.lang.RuntimeException: ...
func ExceptionStacktrace(val string) attribute.KeyValue {
	return ExceptionStacktraceKey.String(val)
}

// ExceptionType returns an attribute KeyValue conforming to the
// "exception.type" semantic conventions. It represents the exception type
// Examples: java.net.ConnectException; OSError
func ExceptionType(val string) attribute.KeyValue {
	return ExceptionTypeKey.String(val)
}
