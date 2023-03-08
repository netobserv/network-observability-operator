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

package utils

import (
	"net/http"
	"os"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// StartPromServer listens for prometheus resource usage requests
func StartPromServer(tlsConfig *api.PromTLSConf, server *http.Server, panicOnError bool) {
	logrus.Debugf("entering StartPromServer")

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())

	var err error
	if tlsConfig != nil {
		err = server.ListenAndServeTLS(tlsConfig.CertPath, tlsConfig.KeyPath)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("error in http.ListenAndServe: %v", err)
		if panicOnError {
			os.Exit(1)
		}
	}
}
