package tls

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"
)

func TestSelfSignedStoreGetCertificate(t *testing.T) {
	store, err := NewSelfSignedStore(SelfSignedConfig{
		Hosts:    []string{"localhost", "127.0.0.1"},
		Duration: 1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSelfSignedStore: %v", err)
	}

	cert, err := store.GetCertificate()
	if err != nil {
		t.Fatalf("GetCertificate: %v", err)
	}
	if cert == nil {
		t.Fatal("expected non-nil certificate")
	}
	if len(cert.Certificate) == 0 {
		t.Fatal("expected at least one certificate in chain")
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}
	if parsed.DNSNames[0] != "localhost" {
		t.Fatalf("expected SAN localhost, got %v", parsed.DNSNames)
	}
}

func TestSelfSignedStoreGetCACertPool(t *testing.T) {
	store, err := NewSelfSignedStore(SelfSignedConfig{
		Hosts:    []string{"localhost"},
		Duration: 1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSelfSignedStore: %v", err)
	}

	pool, err := store.GetCACertPool()
	if err != nil {
		t.Fatalf("GetCACertPool: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil CA pool")
	}
}

func TestSelfSignedStoreReturnsSameCert(t *testing.T) {
	store, err := NewSelfSignedStore(SelfSignedConfig{
		Hosts:    []string{"localhost"},
		Duration: 1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSelfSignedStore: %v", err)
	}

	cert1, _ := store.GetCertificate()
	cert2, _ := store.GetCertificate()
	if &cert1.Certificate[0][0] != &cert2.Certificate[0][0] {
		t.Fatal("expected same certificate instance on repeated calls")
	}
}

func TestSelfSignedStoreTLSHandshake(t *testing.T) {
	store, err := NewSelfSignedStore(SelfSignedConfig{
		Hosts:    []string{"localhost"},
		Duration: 1 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSelfSignedStore: %v", err)
	}

	cert, _ := store.GetCertificate()
	pool, _ := store.GetCACertPool()

	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	clientCfg := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
	}

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
		// Complete TLS handshake before closing so client doesn't see EOF.
		if err := conn.(*tls.Conn).Handshake(); err != nil {
			done <- err
			return
		}
		conn.Close()
		done <- nil
	}()

	conn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	conn.Close()

	if err := <-done; err != nil {
		t.Fatalf("server: %v", err)
	}
}
