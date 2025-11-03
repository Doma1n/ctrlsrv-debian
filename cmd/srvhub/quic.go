package main

import (
	"fmt"
	"log"
)

// QUICServer handles HTTP/3 (QUIC) requests
type QUICServer struct {
	config *Config
}

// NewQUICServer creates a new QUIC server
func NewQUICServer(cfg *Config) *QUICServer {
	return &QUICServer{
		config: cfg,
	}
}

// Start starts the QUIC server
func (s *QUICServer) Start() error {
	// TODO: Implement HTTP/3 QUIC server
	// This requires:
	// - github.com/quic-go/quic-go
	// - TLS certificates
	// - HTTP/3 handler similar to API server

	log.Printf("QUIC server would listen on %s", s.config.Server.QUICAddr)
	log.Println("QUIC/HTTP3 implementation coming soon...")

	// For now, just block to prevent immediate exit
	select {}

	return fmt.Errorf("QUIC server not yet implemented")
}
