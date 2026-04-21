package http

import (
	"context"
	"net/http"

	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
)

func statusFromError(err error) int {
	switch {
	case apperrors.IsCode(err, apperrors.CodeInvalidArgument):
		return http.StatusBadRequest
	case apperrors.IsCode(err, apperrors.CodeUnauthenticated):
		return http.StatusUnauthorized
	case apperrors.IsCode(err, apperrors.CodeForbidden):
		return http.StatusForbidden
	case apperrors.IsCode(err, apperrors.CodeNotFound):
		return http.StatusNotFound
	case apperrors.IsCode(err, apperrors.CodeConflict):
		return http.StatusConflict
	case apperrors.IsCode(err, apperrors.CodeNotImplemented):
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

func writeError(ctx context.Context, w http.ResponseWriter, err error, fallback string) {
	status := statusFromError(err)
	message := fallback
	if appErr, ok := apperrors.As(err); ok && appErr.Message() != "" {
		message = appErr.Message()
	}
	if message == "" {
		message = http.StatusText(status)
	}
	writeJSON(ctx, w, status, map[string]string{"error": message})
}
