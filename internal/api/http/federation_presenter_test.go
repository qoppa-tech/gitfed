package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qoppa-tech/toy-gitfed/internal/modules/federation"
	"github.com/qoppa-tech/toy-gitfed/internal/modules/git"
)

func newFederationServerForTest(t *testing.T, federationSvc *federation.Service) *Server {
	t.Helper()
	reposDir := t.TempDir()
	if federationSvc == nil {
		federationSvc = federation.NewService(nil)
	}
	return NewServer(Config{
		ReposDir:           reposDir,
		Address:            "127.0.0.1:0",
		GitService:         git.NewService(reposDir),
		FederationService:  federationSvc,
		FederationInstance: FederationInstanceConfig{InstanceURL: "https://node.example", InstanceName: "Node A", AdminEmail: "admin@node.example"},
	})
}

func TestFederationMetadataEndpoint(t *testing.T) {
	srv := newFederationServerForTest(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/gitfed", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode metadata response: %v", err)
	}
	if body["instance_url"] != "https://node.example" {
		t.Fatalf("instance_url = %v", body["instance_url"])
	}
}

func TestFederationPeersRoute_RequiresSignatureHeaders(t *testing.T) {
	srv := newFederationServerForTest(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/federation/peers", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestFederationPeersRoute_NotImplementedIsMappedTo501(t *testing.T) {
	srv := newFederationServerForTest(t, federation.NewService(nil))

	req := httptest.NewRequest(http.MethodGet, "/federation/peers", nil)
	req.Header.Set(federation.HeaderVersion, federation.ProtocolVersion)
	req.Header.Set(federation.HeaderActorDID, "did:gitfed:server:remote.example")
	req.Header.Set(federation.HeaderTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(federation.HeaderNonce, "nonce-1")
	req.Header.Set(federation.HeaderSignature, "ZmFrZS1zaWduYXR1cmU=")

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}
}
