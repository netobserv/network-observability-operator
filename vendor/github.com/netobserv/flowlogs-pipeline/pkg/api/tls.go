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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

type ClientTLS struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty" doc:"skip client verifying the server's certificate chain and host name"`
	CACertPath         string `yaml:"caCertPath,omitempty" json:"caCertPath,omitempty" doc:"path to the CA certificate"`
	UserCertPath       string `yaml:"userCertPath,omitempty" json:"userCertPath,omitempty" doc:"path to the user certificate"`
	UserKeyPath        string `yaml:"userKeyPath,omitempty" json:"userKeyPath,omitempty" doc:"path to the user private key"`
}

func (c *ClientTLS) Build() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
	}
	if c.CACertPath != "" {
		caCert, err := os.ReadFile(c.CACertPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)

		if c.UserCertPath != "" && c.UserKeyPath != "" {
			userCert, err := os.ReadFile(c.UserCertPath)
			if err != nil {
				return nil, err
			}
			userKey, err := os.ReadFile(c.UserKeyPath)
			if err != nil {
				return nil, err
			}
			pair, err := tls.X509KeyPair(userCert, userKey)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{pair}
		} else if c.UserCertPath != "" || c.UserKeyPath != "" {
			return nil, errors.New("userCertPath and userKeyPath must be both present or both absent")
		}
		return tlsConfig, nil
	}
	return tlsConfig, nil
}
