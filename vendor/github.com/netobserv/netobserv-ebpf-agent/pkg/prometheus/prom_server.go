package prometheus

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	plog       = logrus.WithField("component", "prometheus")
	maybePanic = plog.Fatalf
)

// InitializePrometheus starts the global Prometheus server, used for operational metrics and prom-encode stages if they don't override the server settings
func InitializePrometheus(settings *metrics.Settings) *http.Server {
	return StartServerAsync(settings, nil)
}

// StartServerAsync listens for prometheus resource usage requests
func StartServerAsync(conn *metrics.Settings, registry *prom.Registry) *http.Server {
	// create prometheus server for operational metrics
	// if value of address is empty, then by default it will take 0.0.0.0
	port := conn.Port
	if port == 0 {
		port = 9090
	}
	addr := fmt.Sprintf("%s:%v", conn.Address, port)
	plog.Infof("StartServerAsync: addr = %s", addr)

	httpServer := &http.Server{
		Addr: addr,
		// TLS clients must use TLS 1.2 or higher
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	mux := http.NewServeMux()
	if registry == nil {
		mux.Handle("/metrics", promhttp.Handler())
	} else {
		mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	}
	httpServer.Handler = mux
	httpServer = defaultServer(httpServer)

	go func() {
		var err error
		if conn.TLS != nil {
			err = httpServer.ListenAndServeTLS(conn.TLS.CertPath, conn.TLS.KeyPath)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			maybePanic("error in http.ListenAndServe: %v", err)
		}
	}()

	return httpServer
}

func defaultServer(srv *http.Server) *http.Server {
	// defaults taken from https://bruinsslot.jp/post/go-secure-webserver/ can be overriden by caller
	if srv.Handler != nil {
		// No more than 2MB body
		srv.Handler = http.MaxBytesHandler(srv.Handler, 2<<20)
	} else {
		plog.Warnf("Handler not yet set on server while securing defaults. Make sure a MaxByte middleware is used.")
	}
	if srv.ReadTimeout == 0 {
		srv.ReadTimeout = 10 * time.Second
	}
	if srv.ReadHeaderTimeout == 0 {
		srv.ReadHeaderTimeout = 5 * time.Second
	}
	if srv.WriteTimeout == 0 {
		srv.WriteTimeout = 10 * time.Second
	}
	if srv.IdleTimeout == 0 {
		srv.IdleTimeout = 120 * time.Second
	}
	if srv.MaxHeaderBytes == 0 {
		srv.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}
	if srv.TLSConfig.MinVersion == 0 {
		srv.TLSConfig.MinVersion = tls.VersionTLS13
	}
	// Disable http/2
	srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

	return srv
}
