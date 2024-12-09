package agent

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

func buildTLSConfig(cfg *Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.KafkaTLSInsecureSkipVerify,
	}
	if cfg.KafkaTLSCACertPath != "" {
		caCert, err := os.ReadFile(cfg.KafkaTLSCACertPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)

		if cfg.KafkaTLSUserCertPath != "" && cfg.KafkaTLSUserKeyPath != "" {
			userCert, err := os.ReadFile(cfg.KafkaTLSUserCertPath)
			if err != nil {
				return nil, err
			}
			userKey, err := os.ReadFile(cfg.KafkaTLSUserKeyPath)
			if err != nil {
				return nil, err
			}
			pair, err := tls.X509KeyPair([]byte(userCert), []byte(userKey))
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{pair}
		}
	}
	return tlsConfig, nil
}
