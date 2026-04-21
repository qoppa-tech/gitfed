package federation

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
)

func TestVerifyFederationRequest_UnsupportedVersion(t *testing.T) {
	svc := NewService(nil)

	_, err := svc.VerifyFederationRequest(context.Background(), VerifyRequestInput{
		Method:    "GET",
		Path:      "/federation/peers",
		ActorDID:  "did:gitfed:server:remote.example",
		Timestamp: time.Now().UTC(),
		Nonce:     "nonce-1",
		Version:   "999",
		Signature: "ZmFrZQ==",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !apperrors.IsCode(err, apperrors.CodeInvalidArgument) {
		t.Fatalf("expected invalid_argument code, got %v", err)
	}
}

func TestVerifyFederationRequest_NoVerifierConfigured(t *testing.T) {
	svc := NewService(nil)

	_, err := svc.VerifyFederationRequest(context.Background(), VerifyRequestInput{
		Method:    "GET",
		Path:      "/federation/peers",
		ActorDID:  "did:gitfed:server:remote.example",
		Timestamp: time.Now().UTC(),
		Nonce:     "nonce-1",
		Version:   ProtocolVersion,
		Signature: "ZmFrZQ==",
	})
	if err == nil {
		t.Fatalf("expected not implemented error")
	}
	if !apperrors.IsCode(err, apperrors.CodeNotImplemented) {
		t.Fatalf("expected not_implemented code, got %v", err)
	}
}
