package identity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
	"github.com/qoppa-tech/toy-gitfed/pkg/identifier/did"
)

func TestResolveUserDID_NoRepositoryConfigured(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.ResolveUserDID(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected not implemented error")
	}
	if !apperrors.IsCode(err, apperrors.CodeNotImplemented) {
		t.Fatalf("expected not_implemented code, got %v", err)
	}
}

func TestVerifyPrincipalOwnership_RequiresUserDID(t *testing.T) {
	svc := NewService(nil)
	err := svc.VerifyPrincipalOwnership(context.Background(), did.DID{
		Method:        did.MethodGitFed,
		PrincipalType: did.PrincipalTypeServer,
		Host:          "node.example",
	}, uuid.New())
	if err == nil {
		t.Fatalf("expected invalid argument error")
	}
	if !apperrors.IsCode(err, apperrors.CodeInvalidArgument) {
		t.Fatalf("expected invalid_argument code, got %v", err)
	}
}
