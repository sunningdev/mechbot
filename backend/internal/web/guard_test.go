package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiterWindow(t *testing.T) {
	l := newLimiter(2, time.Minute)
	now := time.Unix(1_000_000, 0)
	if !l.allow("a", now) || !l.allow("a", now) {
		t.Fatal("first two calls should pass")
	}
	if l.allow("a", now.Add(30*time.Second)) {
		t.Fatal("third call in the window should be blocked")
	}
	if !l.allow("b", now) {
		t.Fatal("other keys have their own budget")
	}
	if !l.allow("a", now.Add(time.Minute)) {
		t.Fatal("a new window should reset the count")
	}
	l.sweep(now.Add(3 * time.Minute))
	if len(l.hits) != 0 {
		t.Fatalf("sweep left %d stale entries", len(l.hits))
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name, remote, real, xff, want string
	}{
		{"direct client ignores headers", "203.0.113.9:5000", "1.1.1.1", "2.2.2.2", "203.0.113.9"},
		{"proxy uses X-Real-IP", "127.0.0.1:5000", "198.51.100.7", "", "198.51.100.7"},
		{"proxy falls back to last forwarded hop", "127.0.0.1:5000", "", "6.6.6.6, 198.51.100.8", "198.51.100.8"},
		{"proxy with junk headers keeps peer", "[::1]:5000", "nope", "also-nope", "::1"},
	}
	for _, c := range cases {
		r := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		r.RemoteAddr = c.remote
		if c.real != "" {
			r.Header.Set("X-Real-IP", c.real)
		}
		if c.xff != "" {
			r.Header.Set("X-Forwarded-For", c.xff)
		}
		if got := clientIP(r); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestLimitAPI(t *testing.T) {
	s := &Server{apiLimit: newLimiter(1, time.Minute)}
	h := s.limitAPI(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	call := func(path string) int {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.RemoteAddr = "203.0.113.1:1"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	if code := call("/api/me"); code != http.StatusOK {
		t.Fatalf("first call: %d", code)
	}
	if code := call("/api/me"); code != http.StatusTooManyRequests {
		t.Fatalf("second call: %d, want 429", code)
	}
	if code := call("/assets/app.js"); code != http.StatusOK {
		t.Fatalf("static files should not be limited: %d", code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	for _, secure := range []bool{false, true} {
		s := &Server{secure: secure}
		w := httptest.NewRecorder()
		s.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		h := w.Header()
		for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Content-Security-Policy"} {
			if h.Get(k) == "" {
				t.Errorf("secure=%v: missing %s", secure, k)
			}
		}
		if got := h.Get("Strict-Transport-Security") != ""; got != secure {
			t.Errorf("secure=%v: HSTS present=%v", secure, got)
		}
	}
}
