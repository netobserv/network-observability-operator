package otlpconfig

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

// CreateTLSConfig creates a tls.Config from a raw certificate bytes
// to verify a server certificate.
func CreateTLSConfig(certBytes []byte) (*tls.Config, error) {
	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(certBytes); !ok {
		return nil, errors.New("failed to append certificate to the cert pool")
	}

	return &tls.Config{
		RootCAs: cp,
	}, nil
}
