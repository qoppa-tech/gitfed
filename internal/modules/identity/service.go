package identity

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
	"github.com/qoppa-tech/toy-gitfed/pkg/identifier/did"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ResolveUserDID(ctx context.Context, userID uuid.UUID) (did.DID, error) {
	if s.repo == nil {
		return did.DID{}, apperrors.NotImplemented("identity user did resolution not implemented")
	}

	principal, err := s.repo.GetUserPrincipal(ctx, userID)
	if err != nil {
		return did.DID{}, err
	}
	return principal.DID, nil
}

func (s *Service) ResolveServerDID(ctx context.Context) (did.DID, error) {
	if s.repo == nil {
		return did.DID{}, apperrors.NotImplemented("identity server did resolution not implemented")
	}

	principal, err := s.repo.GetServerPrincipal(ctx)
	if err != nil {
		return did.DID{}, err
	}
	return principal.DID, nil
}

func (s *Service) VerifyPrincipalOwnership(ctx context.Context, principalDID did.DID, subjectID uuid.UUID) error {
	if principalDID.PrincipalType != did.PrincipalTypeUser {
		return apperrors.New(apperrors.CodeInvalidArgument, "ownership verification requires a user did")
	}

	resolved, err := s.ResolveUserDID(ctx, subjectID)
	if err != nil {
		return err
	}
	if resolved.String() != principalDID.String() {
		return apperrors.New(apperrors.CodeForbidden, "did does not belong to subject")
	}
	return nil
}

func (s *Service) VerifySignature(_ context.Context, _ did.DID, _ []byte, _ []byte) error {
	return apperrors.NotImplemented("signature verification not implemented")
}
