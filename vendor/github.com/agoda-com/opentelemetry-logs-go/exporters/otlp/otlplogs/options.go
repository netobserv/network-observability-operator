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

package otlplogs

import (
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/internal/otlpconfig"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
)

type ExporterConfig struct {
	client Client
}

type ExporterOption interface {
	apply(ExporterConfig) ExporterConfig
}

type exporterOptionFunc func(ExporterConfig) ExporterConfig

func (fn exporterOptionFunc) apply(config ExporterConfig) ExporterConfig {
	return fn(config)
}

// NewExporterConfig creates new configuration for exporter
func NewExporterConfig(options ...ExporterOption) ExporterConfig {

	config := ExporterConfig{}

	for _, option := range options {
		config = option.apply(config)
	}

	if config.client == nil {
		// Default is http/protobuf client
		protocol := otlpconfig.ApplyEnvProtocol(otlpconfig.ExporterProtocolHttpProtobuf)

		if protocol == otlpconfig.ExporterProtocolGrpc {
			config.client = otlplogsgrpc.NewClient()
		} else {
			config.client = otlplogshttp.NewClient()
		}
	}

	return config
}

func WithClient(client Client) ExporterOption {
	return exporterOptionFunc(func(cfg ExporterConfig) ExporterConfig {
		cfg.client = client
		return cfg
	})
}
