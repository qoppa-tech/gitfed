package identity

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the port for identity persistence/resolution.
type Repository interface {
	GetUserPrincipal(ctx context.Context, userID uuid.UUID) (Principal, error)
	GetServerPrincipal(ctx context.Context) (ServerPrincipal, error)
}
