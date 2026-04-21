package identity

import (
	"github.com/google/uuid"
	"github.com/qoppa-tech/toy-gitfed/pkg/identifier/did"
)

// Principal binds an internal subject to a DID and key material.
type Principal struct {
	SubjectID uuid.UUID
	DID       did.DID
	PublicKey []byte
}

// ServerPrincipal holds server-level DID and key material.
type ServerPrincipal struct {
	DID       did.DID
	PublicKey []byte
}
