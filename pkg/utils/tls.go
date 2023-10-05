package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/numaproj-contrib/kafka-source-go/pkg/config"
)

/**
 * This entire file is a copy of https://github.com/numaproj/numaflow/blob/main/pkg/shared/util/tls_config.go with small modifications
 */

// GetTLSConfig is a utility function to get tls.Config
func GetTLSConfig(config *config.TLS, reader VolumeReader) (*tls.Config, error) {
	if config == nil {
		return nil, nil
	}

	var caCertPath, certPath, keyPath string
	var err error
	if config.CACertSecret != nil {
		caCertPath, err = reader.GetSecretVolumePath(config.CACertSecret)
		if err != nil {
			return nil, err
		}
	}

	if config.CertSecret != nil {
		certPath, err = reader.GetSecretVolumePath(config.CertSecret)
		if err != nil {
			return nil, err
		}
	}

	if config.KeySecret != nil {
		keyPath, err = reader.GetSecretVolumePath(config.KeySecret)
		if err != nil {
			return nil, err
		}
	}

	if len(certPath)+len(keyPath) > 0 && len(certPath)*len(keyPath) == 0 {
		// Only one of certSecret and keySecret is configured
		return nil, fmt.Errorf("invalid tls config, both certSecret and keySecret need to be configured")
	}

	c := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}
	if len(caCertPath) > 0 {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read ca cert file %s, %w", caCertPath, err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
		c.RootCAs = pool
	}

	if len(certPath) > 0 && len(keyPath) > 0 {
		clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert key pair (%s, %s), %w", certPath, keyPath, err)
		}
		c.Certificates = []tls.Certificate{clientCert}
	}
	return c, nil
}
