package tls

import (
	"crypto/tls"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SelfSignedStore {
	t.Helper()
	store, err := NewSelfSignedStore(SelfSignedConfig{
		Hosts:    []string{"localhost", "127.0.0.1"},
		Duration: 1 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestProfileServerTLSConfig(t *testing.T) {
	store := newTestStore(t)

	profile := &TLSProfile{
		ServerCert: store,
		MinVersion: tls.VersionTLS12,
	}

	cfg, err := profile.ServerTLSConfig()
	if err != nil {
		t.Fatalf("ServerTLSConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected MinVersion TLS 1.2, got %d", cfg.MinVersion)
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
	}
}

func TestProfileServerTLSConfigMTLS(t *testing.T) {
	store := newTestStore(t)

	profile := &TLSProfile{
		ServerCert: store,
		ClientCA:   store,
		MinVersion: tls.VersionTLS13,
		VerifyPeer: true,
	}

	cfg, err := profile.ServerTLSConfig()
	if err != nil {
		t.Fatalf("ServerTLSConfig: %v", err)
	}
	if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("expected RequireAndVerifyClientCert, got %v", cfg.ClientAuth)
	}
	if cfg.ClientCAs == nil {
		t.Fatal("expected non-nil ClientCAs")
	}
}

func TestProfileClientTLSConfig(t *testing.T) {
	store := newTestStore(t)

	profile := &TLSProfile{
		ServerCert: store,
		ClientCert: store,
		MinVersion: tls.VersionTLS12,
	}

	cfg, err := profile.ClientTLSConfig()
	if err != nil {
		t.Fatalf("ClientTLSConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected 1 client certificate, got %d", len(cfg.Certificates))
	}
	if cfg.RootCAs == nil {
		t.Fatal("expected non-nil RootCAs")
	}
}

func TestProfileClientTLSConfigNoClientCert(t *testing.T) {
	store := newTestStore(t)

	profile := &TLSProfile{
		ServerCert: store,
		MinVersion: tls.VersionTLS12,
	}

	cfg, err := profile.ClientTLSConfig()
	if err != nil {
		t.Fatalf("ClientTLSConfig: %v", err)
	}
	if len(cfg.Certificates) != 0 {
		t.Fatalf("expected 0 client certificates, got %d", len(cfg.Certificates))
	}
}

func TestProfileMTLSHandshake(t *testing.T) {
	serverStore := newTestStore(t)
	clientStore := newTestStore(t)

	serverProfile := &TLSProfile{
		ServerCert: serverStore,
		ClientCA:   clientStore,
		MinVersion: tls.VersionTLS13,
		VerifyPeer: true,
	}

	clientProfile := &TLSProfile{
		ServerCert: serverStore,
		ClientCert: clientStore,
		MinVersion: tls.VersionTLS13,
	}

	serverCfg, err := serverProfile.ServerTLSConfig()
	if err != nil {
		t.Fatalf("ServerTLSConfig: %v", err)
	}

	clientCfg, err := clientProfile.ClientTLSConfig()
	if err != nil {
		t.Fatalf("ClientTLSConfig: %v", err)
	}
	clientCfg.ServerName = "localhost"

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		// Force handshake
		tlsConn := conn.(*tls.Conn)
		err = tlsConn.Handshake()
		conn.Close()
		done <- err
	}()

	conn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	conn.Close()

	if err := <-done; err != nil {
		t.Fatalf("server handshake: %v", err)
	}
}
