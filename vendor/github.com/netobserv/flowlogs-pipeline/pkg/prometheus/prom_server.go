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

package prometheus

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/server"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
)

var (
	plog         = logrus.WithField("component", "prometheus")
	maybePanic   = plog.Fatalf
	SharedServer = &PromServer{}
)

type PromServer struct {
	httpServer      *http.Server
	namedRegistries sync.Map
}

func (ps *PromServer) Gather() ([]*dto.MetricFamily, error) {
	all := prom.Gatherers{}
	ps.namedRegistries.Range(func(_, value interface{}) bool {
		r := value.(prom.Gatherer)
		all = append(all, r)
		return true
	})
	return all.Gather()
}

func (ps *PromServer) Shutdown(ctx context.Context) error {
	return ps.httpServer.Shutdown(ctx)
}

func (ps *PromServer) SetRegistry(name string, registry prom.Gatherer) {
	ps.namedRegistries.Store(name, registry)
}

// InitializePrometheus starts the global Prometheus server, used for operational metrics and prom-encode stages if they don't override the server settings
func InitializePrometheus(settings *config.MetricsSettings) *PromServer {
	if settings.NoPanic {
		maybePanic = plog.Errorf
	}
	if settings.DisableGlobalServer {
		plog.Info("Disabled global Prometheus server - no operational metrics will be available")
		return nil
	}
	r := prom.DefaultGatherer
	if settings.SuppressGoMetrics {
		// set up private prometheus registry
		r = prom.NewRegistry()
	}
	SharedServer = StartServerAsync(&settings.PromConnectionInfo, "", r)
	return SharedServer
}

// StartServerAsync listens for prometheus resource usage requests
func StartServerAsync(conn *api.PromConnectionInfo, regName string, registry prom.Gatherer) *PromServer {
	// create prometheus server for operational metrics
	// if value of address is empty, then by default it will take 0.0.0.0
	port := conn.Port
	if port == 0 {
		port = 9090
	}
	addr := fmt.Sprintf("%s:%v", conn.Address, port)
	plog.Infof("StartServerAsync: addr = %s", addr)

	httpServer := http.Server{
		Addr: addr,
		// TLS clients must use TLS 1.2 or higher
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server.Default(&httpServer)

	go func() {
		var err error
		if conn.TLS != nil {
			err = httpServer.ListenAndServeTLS(conn.TLS.CertPath, conn.TLS.KeyPath)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			maybePanic("error in http.ListenAndServe: %v", err)
		}
	}()

	p := PromServer{httpServer: &httpServer}
	p.namedRegistries.Store(regName, registry)

	mux.Handle("/metrics", promhttp.HandlerFor(&p, promhttp.HandlerOpts{}))

	return &p
}
