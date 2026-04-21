package federation

import (
	"context"
	"encoding/base64"
	"time"

	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
	"github.com/qoppa-tech/toy-gitfed/pkg/identifier/did"
)

type SignatureVerifier interface {
	VerifySignature(ctx context.Context, principalDID did.DID, payload []byte, signature []byte) error
}

type Service struct {
	repo     Repository
	verifier SignatureVerifier
}

func NewService(verifier SignatureVerifier) *Service {
	return &Service{verifier: verifier}
}

func NewServiceWithRepo(repo Repository, verifier SignatureVerifier) *Service {
	return &Service{repo: repo, verifier: verifier}
}

func (s *Service) VerifyFederationRequest(ctx context.Context, in VerifyRequestInput) (VerifiedRequest, error) {
	if in.Version != ProtocolVersion {
		return VerifiedRequest{}, apperrors.New(apperrors.CodeInvalidArgument, "unsupported federation protocol version")
	}

	actor, err := did.Parse(in.ActorDID)
	if err != nil {
		return VerifiedRequest{}, err
	}

	sig, err := base64.StdEncoding.DecodeString(in.Signature)
	if err != nil {
		return VerifiedRequest{}, apperrors.New(apperrors.CodeInvalidArgument, "invalid federation signature encoding")
	}

	if s.verifier == nil {
		return VerifiedRequest{}, apperrors.NotImplemented("federation signature verification not implemented")
	}

	signingInput := CanonicalSigningInput(in.Method, in.Path, actor.String(), in.Timestamp, in.Nonce, in.Digest, in.Version)
	if err := s.verifier.VerifySignature(ctx, actor, []byte(signingInput), sig); err != nil {
		return VerifiedRequest{}, err
	}

	return VerifiedRequest{ActorDID: actor.String()}, nil
}

func (s *Service) ListPeers(ctx context.Context) ([]Peer, error) {
	if s.repo == nil {
		return nil, apperrors.NotImplemented("federation peers listing not implemented")
	}
	return s.repo.ListPeers(ctx)
}

func (s *Service) AddPeer(ctx context.Context, peer Peer) (Peer, error) {
	if s.repo == nil {
		return Peer{}, apperrors.NotImplemented("federation peer registration not implemented")
	}
	if peer.InstanceURL == "" {
		return Peer{}, apperrors.New(apperrors.CodeInvalidArgument, "instance url is required")
	}
	if peer.AddedAt.IsZero() {
		peer.AddedAt = time.Now().UTC()
	}
	return s.repo.AddPeer(ctx, peer)
}

func (s *Service) ResolveDID(_ context.Context, raw string) (DIDResolution, error) {
	if _, err := did.Parse(raw); err != nil {
		return DIDResolution{}, err
	}
	return DIDResolution{}, apperrors.NotImplemented("did resolution endpoint not implemented")
}
