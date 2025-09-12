/*
 * Copyright (C) 2022 IBM, Inc.
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

import (
	"errors"
	"fmt"
	"time"

	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type WriteLoki struct {
	URL            string                       `yaml:"url,omitempty" json:"url,omitempty" doc:"the address of an existing Loki service to push the flows to"`
	TenantID       string                       `yaml:"tenantID,omitempty" json:"tenantID,omitempty" doc:"identifies the tenant for the request"`
	BatchWait      string                       `yaml:"batchWait,omitempty" json:"batchWait,omitempty" doc:"maximum amount of time to wait before sending a batch"`
	BatchSize      int                          `yaml:"batchSize,omitempty" json:"batchSize,omitempty" doc:"maximum batch size (in bytes) of logs to accumulate before sending"`
	Timeout        string                       `yaml:"timeout,omitempty" json:"timeout,omitempty" doc:"maximum time to wait for a server to respond to a request"`
	MinBackoff     string                       `yaml:"minBackoff,omitempty" json:"minBackoff,omitempty" doc:"initial backoff time for client connection between retries"`
	MaxBackoff     string                       `yaml:"maxBackoff,omitempty" json:"maxBackoff,omitempty" doc:"maximum backoff time for client connection between retries"`
	MaxRetries     int                          `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty" doc:"maximum number of retries for client connections"`
	Labels         []string                     `yaml:"labels,omitempty" json:"labels,omitempty" doc:"map of record fields to be used as labels"`
	StaticLabels   model.LabelSet               `yaml:"staticLabels,omitempty" json:"staticLabels,omitempty" doc:"map of common labels to set on each flow"`
	IgnoreList     []string                     `yaml:"ignoreList,omitempty" json:"ignoreList,omitempty" doc:"map of record fields to be removed from the record"`
	ClientConfig   *promConfig.HTTPClientConfig `yaml:"clientConfig,omitempty" json:"clientConfig,omitempty" doc:"HTTP client configuration (used only for HTTP client type)"`
	TimestampLabel model.LabelName              `yaml:"timestampLabel,omitempty" json:"timestampLabel,omitempty" doc:"label to use for time indexing"`
	// TimestampScale provides the scale in time of the units from the timestamp
	// E.g. UNIX timescale is '1s' (one second) while other clock sources might have
	// scales of '1ms' (one millisecond) or just '1' (one nanosecond)
	// Default value is '1s'
	TimestampScale string `yaml:"timestampScale,omitempty" json:"timestampScale,omitempty" doc:"timestamp units scale (e.g. for UNIX = 1s)"`
	Format         string `yaml:"format,omitempty" json:"format,omitempty" doc:"the format of each line: printf (writes using golang's default map printing), fields (writes one key and value field per line) or json (default)"`
	Reorder        bool   `yaml:"reorder,omitempty" json:"reorder,omitempty" doc:"reorder json map keys"`

	// Client type selection
	ClientType string          `yaml:"clientType,omitempty" json:"clientType,omitempty" doc:"type of client to use: 'http' or 'grpc' (default: 'http')"`
	GRPCConfig *GRPCLokiConfig `yaml:"grpcConfig,omitempty" json:"grpcConfig,omitempty" doc:"gRPC client configuration (used only for gRPC client type)"`
}

type GRPCLokiConfig struct {
	ServerAddress    string         `yaml:"serverAddress,omitempty" json:"serverAddress,omitempty" doc:"gRPC server address (host:port)"`
	MaxRecvMsgSize   int            `yaml:"maxRecvMsgSize,omitempty" json:"maxRecvMsgSize,omitempty" doc:"maximum message size the client can receive"`
	MaxSendMsgSize   int            `yaml:"maxSendMsgSize,omitempty" json:"maxSendMsgSize,omitempty" doc:"maximum message size the client can send"`
	KeepAlive        string         `yaml:"keepAlive,omitempty" json:"keepAlive,omitempty" doc:"keep alive interval"`
	KeepAliveTimeout string         `yaml:"keepAliveTimeout,omitempty" json:"keepAliveTimeout,omitempty" doc:"keep alive timeout"`
	UseStreaming     bool           `yaml:"useStreaming,omitempty" json:"useStreaming,omitempty" doc:"use streaming for real-time log pushing"`
	TLS              *GRPCTLSConfig `yaml:"tls,omitempty" json:"tls,omitempty" doc:"TLS configuration"`
}

type GRPCTLSConfig struct {
	Enabled            bool   `yaml:"enabled,omitempty" json:"enabled,omitempty" doc:"enable TLS"`
	CertFile           string `yaml:"certFile,omitempty" json:"certFile,omitempty" doc:"path to client certificate file"`
	KeyFile            string `yaml:"keyFile,omitempty" json:"keyFile,omitempty" doc:"path to client key file"`
	CAFile             string `yaml:"caFile,omitempty" json:"caFile,omitempty" doc:"path to CA certificate file"`
	ServerName         string `yaml:"serverName,omitempty" json:"serverName,omitempty" doc:"server name for certificate verification"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty" doc:"skip certificate verification (insecure)"`
}

func (w *WriteLoki) SetDefaults() {
	if w.BatchWait == "" {
		w.BatchWait = "1s"
	}
	if w.BatchSize == 0 {
		w.BatchSize = 100 * 1024
	}
	if w.Timeout == "" {
		w.Timeout = "10s"
	}
	if w.MinBackoff == "" {
		w.MinBackoff = "1s"
	}
	if w.MaxBackoff == "" {
		w.MaxBackoff = "1s"
	}
	if w.MaxRetries == 0 {
		w.MaxRetries = 10
	}
	if w.TimestampLabel == "" {
		w.TimestampLabel = "TimeReceived"
	}
	if w.TimestampScale == "" {
		w.TimestampScale = "1s"
	}
	if w.Format == "" {
		w.Format = "json"
	}
	if w.ClientType == "" {
		w.ClientType = "http"
	}

	// Set defaults for gRPC config if gRPC client type is selected
	if w.ClientType == "grpc" && w.GRPCConfig != nil {
		w.GRPCConfig.SetDefaults()
	}
}

func (g *GRPCLokiConfig) SetDefaults() {
	if g.MaxRecvMsgSize == 0 {
		g.MaxRecvMsgSize = 1024 * 1024 * 64 // 64MB
	}
	if g.MaxSendMsgSize == 0 {
		g.MaxSendMsgSize = 1024 * 1024 * 16 // 16MB
	}
	if g.KeepAlive == "" {
		g.KeepAlive = "30s"
	}
	if g.KeepAliveTimeout == "" {
		g.KeepAliveTimeout = "5s"
	}
}

func (w *WriteLoki) Validate() error {
	if w == nil {
		return errors.New("you must provide a configuration")
	}
	if w.TimestampScale == "" {
		return errors.New("timestampUnit must be a valid Duration > 0 (e.g. 1m, 1s or 1ms)")
	}
	if w.BatchSize <= 0 {
		return fmt.Errorf("invalid batchSize: %v. Required > 0", w.BatchSize)
	}

	// Validate client type
	if w.ClientType != "" && w.ClientType != "http" && w.ClientType != "grpc" {
		return fmt.Errorf("invalid clientType: %s. Must be 'http' or 'grpc'", w.ClientType)
	}

	// Validate based on client type
	switch w.ClientType {
	case "http", "":
		if w.URL == "" {
			return errors.New("url can't be empty for HTTP client")
		}
	case "grpc":
		if w.GRPCConfig == nil {
			return errors.New("grpcConfig is required when using gRPC client type")
		}
		if err := w.GRPCConfig.Validate(); err != nil {
			return fmt.Errorf("gRPC config validation failed: %w", err)
		}
	}

	return nil
}

func (g *GRPCLokiConfig) Validate() error {
	if g == nil {
		return errors.New("gRPC config cannot be nil")
	}
	if g.ServerAddress == "" {
		return errors.New("gRPC serverAddress cannot be empty")
	}

	// Validate duration fields
	if g.KeepAlive != "" {
		if _, err := time.ParseDuration(g.KeepAlive); err != nil {
			return fmt.Errorf("invalid keepAlive duration: %w", err)
		}
	}
	if g.KeepAliveTimeout != "" {
		if _, err := time.ParseDuration(g.KeepAliveTimeout); err != nil {
			return fmt.Errorf("invalid keepAliveTimeout duration: %w", err)
		}
	}

	return nil
}
