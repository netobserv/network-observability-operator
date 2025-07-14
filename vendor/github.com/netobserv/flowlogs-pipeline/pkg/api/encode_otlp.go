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

package api

type EncodeOtlpLogs struct {
	*OtlpConnectionInfo `json:",inline" doc:"OpenTelemetry connection info; includes:"`
}

type EncodeOtlpTraces struct {
	*OtlpConnectionInfo `json:",inline" doc:"OpenTelemetry connection info; includes:"`
	SpanSplitter        []string `yaml:"spanSplitter,omitempty" json:"spanSplitter,omitempty" doc:"separate span for each prefix listed"`
}

type EncodeOtlpMetrics struct {
	*OtlpConnectionInfo `json:",inline" doc:"OpenTelemetry connection info; includes:"`
	Prefix              string       `yaml:"prefix,omitempty" json:"prefix,omitempty" doc:"prefix added to each metric name"`
	Metrics             MetricsItems `yaml:"metrics,omitempty" json:"metrics,omitempty" doc:"list of metric definitions, each includes:"`
	PushTimeInterval    Duration     `yaml:"pushTimeInterval,omitempty" json:"pushTimeInterval,omitempty" doc:"how often should metrics be sent to collector (default: 20s)"`
	ExpiryTime          Duration     `yaml:"expiryTime,omitempty" json:"expiryTime,omitempty" doc:"time duration of no-flow to wait before deleting data item (default: 2m)"`
}

type OtlpConnectionInfo struct {
	Address        string            `yaml:"address" json:"address" doc:"endpoint address to expose"`
	Port           int               `yaml:"port" json:"port" doc:"endpoint port number to expose"`
	ConnectionType string            `yaml:"connectionType" json:"connectionType" doc:"interface mechanism: either http or grpc"`
	TLS            *ClientTLS        `yaml:"tls,omitempty" json:"tls,omitempty" doc:"TLS configuration for the endpoint"`
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" doc:"headers to add to messages (optional)"`
}
