package http

import (
	"net/http"
	"testing"

	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
)

func TestStatusFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid argument", err: apperrors.New(apperrors.CodeInvalidArgument, "bad"), want: http.StatusBadRequest},
		{name: "unauthenticated", err: apperrors.New(apperrors.CodeUnauthenticated, "bad"), want: http.StatusUnauthorized},
		{name: "forbidden", err: apperrors.New(apperrors.CodeForbidden, "bad"), want: http.StatusForbidden},
		{name: "not found", err: apperrors.New(apperrors.CodeNotFound, "bad"), want: http.StatusNotFound},
		{name: "conflict", err: apperrors.New(apperrors.CodeConflict, "bad"), want: http.StatusConflict},
		{name: "not implemented", err: apperrors.New(apperrors.CodeNotImplemented, "bad"), want: http.StatusNotImplemented},
		{name: "internal", err: apperrors.New(apperrors.CodeInternal, "bad"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFromError(tt.err); got != tt.want {
				t.Fatalf("statusFromError() = %d, want %d", got, tt.want)
			}
		})
	}
}
