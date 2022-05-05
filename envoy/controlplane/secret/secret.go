package secret

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func NewSecret(name, certPEM, keyPEM string, password string) *tls.Secret {
	return &tls.Secret{
		Name: name,
		Type: &tls.Secret_TlsCertificate{
			TlsCertificate: NewTlsCertificate(certPEM, keyPEM, password),
		},
	}
}

func NewTlsCertificate(certPEM, keyPEM string, password string) *tls.TlsCertificate {
	return &tls.TlsCertificate{
		CertificateChain: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: certPEM,
			},
		},
		PrivateKey: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: keyPEM,
			},
		},
		Password: func() *core.DataSource {
			if password == "" {
				return nil
			}

			return &core.DataSource{
				Specifier: &core.DataSource_InlineString{
					InlineString: password,
				},
			}
		}(),
	}
}

func NewCommonTlsContext_ValidationContext(name, chainPEM, peerSAN string, allowExpired bool) *tls.CommonTlsContext_ValidationContext {
	return &tls.CommonTlsContext_ValidationContext{
		ValidationContext: newValidationContext(name, chainPEM, peerSAN, allowExpired),
	}
}

func newValidationContext(name, chainPEM, peerSAN string, allowExpired bool) *tls.CertificateValidationContext {
	return &tls.CertificateValidationContext{
		TrustedCa: &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: chainPEM,
			},
		},
		MatchSubjectAltNames: []*matcher.StringMatcher{
			{
				MatchPattern: &matcher.StringMatcher_Exact{
					Exact: peerSAN,
				},
			},
		},
		AllowExpiredCertificate: allowExpired,
	}
}

func NewTlsParameters() *tls.TlsParameters {
	return &tls.TlsParameters{
		TlsMinimumProtocolVersion: tls.TlsParameters_TLSv1_2,
		TlsMaximumProtocolVersion: tls.TlsParameters_TLSv1_3,
		CipherSuites: []string{
			"ECDHE-ECDSA-AES128-GCM-SHA256",
			"ECDHE-ECDSA-CHACHA20-POLY1305",
			"ECDHE-RSA-AES128-GCM-SHA256",
			"ECDHE-RSA-CHACHA20-POLY1305",
		},
		EcdhCurves: []string{
			"X25519",
			"P-256",
		},
	}
}
