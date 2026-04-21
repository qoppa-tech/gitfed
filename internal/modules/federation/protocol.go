package federation

import (
	"fmt"
	"strings"
	"time"
)

const (
	// TODO(AP-COMPAT-CRITICAL): protocol v1 is custom GitFed.
	// If AP compatibility is enabled, update canonicalization + envelope mapping
	// in lockstep with pkg/identifier/did and HTTP federation presenters.
	ProtocolVersion = "1"

	HeaderVersion   = "X-GitFed-Version"
	HeaderActorDID  = "X-GitFed-Actor"
	HeaderTimestamp = "X-GitFed-Timestamp"
	HeaderNonce     = "X-GitFed-Nonce"
	HeaderSignature = "X-GitFed-Signature"
	HeaderDigest    = "X-GitFed-Digest"
)

// CanonicalSigningInput creates the canonical request string for signatures.
func CanonicalSigningInput(method, path, actorDID string, ts time.Time, nonce, digest, version string) string {
	return strings.Join([]string{
		fmt.Sprintf("method:%s", strings.ToUpper(method)),
		fmt.Sprintf("path:%s", path),
		fmt.Sprintf("actor:%s", actorDID),
		fmt.Sprintf("timestamp:%s", ts.UTC().Format(time.RFC3339)),
		fmt.Sprintf("nonce:%s", nonce),
		fmt.Sprintf("digest:%s", digest),
		fmt.Sprintf("version:%s", version),
	}, "\n")
}
