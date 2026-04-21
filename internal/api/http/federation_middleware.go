package http

import (
	"context"
	"net/http"
	"time"

	"github.com/qoppa-tech/toy-gitfed/internal/modules/federation"
	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
)

const federationActorDIDKey contextKey = "federation_actor_did"

const federationTimestampSkew = 5 * time.Minute

func FederationActorDIDFromContext(ctx context.Context) (string, bool) {
	actorDID, ok := ctx.Value(federationActorDIDKey).(string)
	return actorDID, ok
}

func FederationAuth(service *federation.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if service == nil {
				writeError(r.Context(), w, apperrors.NotImplemented("federation middleware not configured"), "")
				return
			}

			version := r.Header.Get(federation.HeaderVersion)
			actor := r.Header.Get(federation.HeaderActorDID)
			timestampRaw := r.Header.Get(federation.HeaderTimestamp)
			nonce := r.Header.Get(federation.HeaderNonce)
			signature := r.Header.Get(federation.HeaderSignature)
			digest := r.Header.Get(federation.HeaderDigest)

			if version == "" || actor == "" || timestampRaw == "" || nonce == "" || signature == "" {
				writeError(r.Context(), w, apperrors.New(apperrors.CodeUnauthenticated, "missing federation signature headers"), "")
				return
			}

			timestamp, err := time.Parse(time.RFC3339, timestampRaw)
			if err != nil {
				writeError(r.Context(), w, apperrors.New(apperrors.CodeInvalidArgument, "invalid federation timestamp"), "")
				return
			}
			if time.Since(timestamp.UTC()) > federationTimestampSkew || time.Until(timestamp.UTC()) > federationTimestampSkew {
				writeError(r.Context(), w, apperrors.New(apperrors.CodeUnauthenticated, "federation timestamp outside allowed window"), "")
				return
			}

			verified, err := service.VerifyFederationRequest(r.Context(), federation.VerifyRequestInput{
				Method:    r.Method,
				Path:      r.URL.Path,
				ActorDID:  actor,
				Timestamp: timestamp,
				Nonce:     nonce,
				Digest:    digest,
				Version:   version,
				Signature: signature,
			})
			if err != nil {
				writeError(r.Context(), w, err, "")
				return
			}

			ctx := context.WithValue(r.Context(), federationActorDIDKey, verified.ActorDID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
