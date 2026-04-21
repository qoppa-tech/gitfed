package federation

import "context"

// Repository is a federation persistence port.
type Repository interface {
	ListPeers(ctx context.Context) ([]Peer, error)
	AddPeer(ctx context.Context, peer Peer) (Peer, error)
}
