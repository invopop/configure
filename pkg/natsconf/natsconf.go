// Package natsconf helps configure NATS connections from
// a JSON/YAML configuration source.
package natsconf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	nats "github.com/nats-io/nats.go"
)

const (
	defaultMaxReconnects = -1
)

// Config defines how to connect to the NATS servers.
type Config struct {
	Name          string `json:"name"` // Optional client name for the connection
	URL           string `json:"url"`
	MaxReconnects int    `json:"max_reconnects"` // default -1
	ReconnectWait int    `json:"reconnect_wait"` // default 1000 (ms)
	JWT           string `json:"jwt"`            // JWT credential text
	NKey          string `json:"nkey"`           // NKey secret text
	Creds         string `json:"creds"`          // NKey with JWT credentials file
	TLS           struct {
		ServerName string `json:"server_name"`
		Cert       string `json:"cert"`
		Key        string `json:"key"`
		CA         string `json:"ca"`
	} `json:"tls"`
}

// Options generates an array of nats options based on the configuration.
func (conf *Config) Options() ([]nats.Option, error) {
	opts := []nats.Option{}

	if conf.Name != "" {
		opts = append(opts, nats.Name(conf.Name))
	}

	if conf.MaxReconnects == 0 {
		conf.MaxReconnects = defaultMaxReconnects
	}
	opts = append(opts, nats.MaxReconnects(conf.MaxReconnects))

	if conf.ReconnectWait == 0 {
		conf.ReconnectWait = 1000
	}
	opts = append(opts, nats.ReconnectWait(time.Duration(conf.ReconnectWait)*time.Millisecond))

	// JWT and NKey Credentials File
	if conf.JWT != "" && conf.NKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(conf.JWT, conf.NKey))
	} else if conf.Creds != "" {
		opts = append(opts, nats.UserCredentials(conf.Creds))
	}

	copt, err := conf.CertificateOption()
	if err != nil {
		return nil, err
	}
	if copt != nil {
		opts = append(opts, copt)
	}
	return opts, nil
}

// CertificateOption generates a nats.Option for the configured
// TLS certificates.
func (conf *Config) CertificateOption() (nats.Option, error) {
	if conf.TLS.Key != "" && conf.TLS.Cert != "" && conf.TLS.CA != "" {
		cert, err := tls.LoadX509KeyPair(conf.TLS.Cert, conf.TLS.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certs: %w", err)
		}
		pool := x509.NewCertPool()
		root, err := os.ReadFile(conf.TLS.CA)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA file: %w", err)
		}
		if root == nil {
			return nil, fmt.Errorf("loaded empty CA file")
		}
		ok := pool.AppendCertsFromPEM(root)
		if !ok {
			return nil, errors.New("failed to process CA certificate")
		}
		sn := conf.TLS.ServerName
		tlsconf := &tls.Config{
			ServerName:   sn,
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
			MinVersion:   tls.VersionTLS12,
		}
		return nats.Secure(tlsconf), nil
	}
	if conf.TLS.Key != "" && conf.TLS.Cert != "" {
		return nats.ClientCert(conf.TLS.Cert, conf.TLS.Key), nil
	}
	return nil, nil
}
