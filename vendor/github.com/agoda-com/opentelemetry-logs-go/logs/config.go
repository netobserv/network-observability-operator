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

package logs // Package logs import "github.com/agoda-com/opentelemetry-logs-go/logs"

import "go.opentelemetry.io/otel/attribute"

// LoggerConfig is a group of options for a Logger.
type LoggerConfig struct {
	instrumentationVersion string
	// Schema URL of the telemetry emitted by the Logger.
	schemaURL string
	attrs     attribute.Set
}

// InstrumentationVersion returns the version of the library providing instrumentation.
func (t *LoggerConfig) InstrumentationVersion() string {
	return t.instrumentationVersion
}

// InstrumentationAttributes returns the attributes associated with the library
// providing instrumentation.
func (t *LoggerConfig) InstrumentationAttributes() attribute.Set {
	return t.attrs
}

// SchemaURL returns the Schema URL of the telemetry emitted by the Logger.
func (t *LoggerConfig) SchemaURL() string {
	return t.schemaURL
}

// NewLoggerConfig applies all the options to a returned LoggerConfig.
func NewLoggerConfig(options ...LoggerOption) LoggerConfig {
	var config LoggerConfig
	for _, option := range options {
		config = option.apply(config)
	}
	return config
}

// LoggerOption applies an option to a LoggerConfig.
type LoggerOption interface {
	apply(LoggerConfig) LoggerConfig
}

type loggerOptionFunc func(LoggerConfig) LoggerConfig

func (fn loggerOptionFunc) apply(cfg LoggerConfig) LoggerConfig {
	return fn(cfg)
}

// WithInstrumentationVersion sets the instrumentation version.
func WithInstrumentationVersion(version string) LoggerOption {
	return loggerOptionFunc(func(cfg LoggerConfig) LoggerConfig {
		cfg.instrumentationVersion = version
		return cfg
	})
}

// WithInstrumentationAttributes sets the instrumentation attributes.
//
// The passed attributes will be de-duplicated.
func WithInstrumentationAttributes(attr ...attribute.KeyValue) LoggerOption {
	return loggerOptionFunc(func(config LoggerConfig) LoggerConfig {
		config.attrs = attribute.NewSet(attr...)
		return config
	})
}

// WithSchemaURL sets the schema URL for the Logger.
func WithSchemaURL(schemaURL string) LoggerOption {
	return loggerOptionFunc(func(cfg LoggerConfig) LoggerConfig {
		cfg.schemaURL = schemaURL
		return cfg
	})
}
