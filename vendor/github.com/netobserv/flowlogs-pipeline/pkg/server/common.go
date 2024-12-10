package server

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var slog = logrus.WithField("module", "server")

func Default(srv *http.Server) *http.Server {
	// defaults taken from https://bruinsslot.jp/post/go-secure-webserver/ can be overriden by caller
	if srv.Handler != nil {
		// No more than 2MB body
		srv.Handler = http.MaxBytesHandler(srv.Handler, 2<<20)
	} else {
		slog.Warnf("Handler not yet set on server while securing defaults. Make sure a MaxByte middleware is used.")
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
