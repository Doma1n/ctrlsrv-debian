package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// QUICServer handles QUIC/HTTP3 requests
type QUICServer struct {
	config    *Config
	server    *http3.Server
	handler   http.Handler
	tlsConfig *tls.Config
}

// NewQUICServer creates a new QUIC server
func NewQUICServer(cfg *Config) *QUICServer {
	// Generate self-signed certificate for development
	tlsConfig := generateTLSConfig()

	// Use the same handler as HTTP API
	apiServer := NewAPIServer(cfg)

	return &QUICServer{
		config:    cfg,
		handler:   apiServer.mux,
		tlsConfig: tlsConfig,
	}
}

// Start starts the QUIC server
func (s *QUICServer) Start() error {
	log.Printf("Starting QUIC/HTTP3 server on %s", s.config.Server.QUICAddr)

	s.server = &http3.Server{
		Addr:      s.config.Server.QUICAddr,
		Handler:   s.handler,
		TLSConfig: s.tlsConfig,
		QUICConfig: &quic.Config{
			MaxIdleTimeout:  30 * time.Second,
			KeepAlivePeriod: 10 * time.Second,
		},
	}

	return s.server.ListenAndServe()
}

// Stop stops the QUIC server
func (s *QUICServer) Stop() error {
	if s.server == nil {
		return nil
	}

	return s.server.Close()
}

// generateTLSConfig creates a self-signed certificate for development
func generateTLSConfig() *tls.Config {
	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ctrlsrv"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Create TLS certificate
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("Failed to create TLS certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3"},
	}
}
