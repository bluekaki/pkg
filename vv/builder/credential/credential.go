package credential

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/bluekaki/pkg/errors"

	"google.golang.org/grpc/credentials"
)

// ClientOption option for client credential
type ClientOption struct {
	// ChainPEMBlock chain of server issuer's
	ChainPEMBlock []byte
	// CertPEMBlock used by if server require and verify client cert; optional
	CertPEMBlock []byte
	// KeyPEMBlock used by if server require and verify client cert; optional
	KeyPEMBlock []byte
}

// ServerOption option for server credential
type ServerOption struct {
	// ChainPEMBlock chain of client issuer's
	ChainPEMBlock []byte
	// CertPEMBlock PEM encoded data
	CertPEMBlock []byte
	// KeyPEMBlock PEM encoded data
	KeyPEMBlock []byte
	// RequireAndVerifyClientCert enable verify client's cert
	RequireAndVerifyClientCert bool
}

// NewClient create client's credential
func NewClient(option ClientOption) (credentials.TransportCredentials, error) {
	var pool *x509.CertPool
	if len(option.ChainPEMBlock) > 0 {
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(option.ChainPEMBlock) {
			return nil, errors.New("ChainPEMBlocks illegal")
		}
	}

	if len(option.CertPEMBlock) == 0 {
		return nil, errors.New("CertPEMBlock required")
	}
	if len(option.KeyPEMBlock) == 0 {
		return nil, errors.New("KeyPEMBlock required")
	}

	cert, err := tls.X509KeyPair(option.CertPEMBlock, option.KeyPEMBlock)
	if err != nil {
		return nil, errors.New("CertPEMBlock or KeyPEMBlock illegal")
	}
	certs := []tls.Certificate{cert}

	return credentials.NewTLS(&tls.Config{
		Certificates: certs,
		RootCAs:      pool,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.X25519},
	}), nil
}

// NewServer create server's credential
func NewServer(option ServerOption) (credentials.TransportCredentials, error) {
	var pool *x509.CertPool
	if len(option.ChainPEMBlock) > 0 {
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(option.ChainPEMBlock) {
			return nil, errors.New("ChainPEMBlocks illegal")
		}
	}

	if len(option.CertPEMBlock) == 0 {
		return nil, errors.New("CertPEMBlock required")
	}
	if len(option.KeyPEMBlock) == 0 {
		return nil, errors.New("KeyPEMBlock required")
	}

	cert, err := tls.X509KeyPair(option.CertPEMBlock, option.KeyPEMBlock)
	if err != nil {
		return nil, errors.New("CertPEMBlock or KeyPEMBlock illegal")
	}
	certs := []tls.Certificate{cert}

	clientAuth := tls.NoClientCert
	if option.RequireAndVerifyClientCert {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: certs,
		ClientAuth:   clientAuth,
		ClientCAs:    pool,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.X25519},
	}), nil
}
