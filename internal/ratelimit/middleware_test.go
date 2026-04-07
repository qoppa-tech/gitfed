package ratelimit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubLimiter struct {
	allowed    bool
	retryAfter time.Duration
}

func (s *stubLimiter) allow(_ context.Context, _ string, _ float64, _ int) (bool, time.Duration, error) {
	return s.allowed, s.retryAfter, nil
}

func TestIPMiddleware_Allowed(t *testing.T) {
	stub := &stubLimiter{allowed: true}
	mw := IPMiddleware(stub.allow, 100, 20)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestIPMiddleware_Denied(t *testing.T) {
	stub := &stubLimiter{allowed: false, retryAfter: 3 * time.Second}
	mw := IPMiddleware(stub.allow, 100, 20)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if ra := w.Header().Get("Retry-After"); ra != "3" {
		t.Errorf("Retry-After = %q, want %q", ra, "3")
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "rate limit exceeded" {
		t.Errorf("error = %q, want %q", body["error"], "rate limit exceeded")
	}
}

func TestIPMiddleware_XForwardedFor(t *testing.T) {
	var gotKey string
	fakeFn := func(_ context.Context, key string, _ float64, _ int) (bool, time.Duration, error) {
		gotKey = key
		return true, 0, nil
	}

	mw := IPMiddleware(fakeFn, 100, 20)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotKey != "rl:ip:10.0.0.1" {
		t.Errorf("key = %q, want %q", gotKey, "rl:ip:10.0.0.1")
	}
}

func stubExtractor(id string) UserIDExtractor {
	return func(ctx context.Context) (string, bool) {
		if id == "" {
			return "", false
		}
		return id, true
	}
}

func TestUserMiddleware_Allowed(t *testing.T) {
	stub := &stubLimiter{allowed: true}
	mw := UserMiddleware(stub.allow, stubExtractor("user-abc-123"), 200, 40)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("handler should be called")
	}
}

func TestUserMiddleware_NoUser_PassesThrough(t *testing.T) {
	stub := &stubLimiter{allowed: false}
	mw := UserMiddleware(stub.allow, stubExtractor(""), 200, 40)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("handler should be called when no user in context")
	}
}
