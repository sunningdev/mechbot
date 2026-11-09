package web

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// limiter is a fixed-window counter per key (client IP or user).
type limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*window
}

type window struct {
	start time.Time
	n     int
}

func newLimiter(limit int, per time.Duration) *limiter {
	return &limiter{limit: limit, window: per, hits: map[string]*window{}}
}

func (l *limiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.hits[key]
	if w == nil || now.Sub(w.start) >= l.window {
		l.hits[key] = &window{start: now, n: 1}
		return true
	}
	if w.n >= l.limit {
		return false
	}
	w.n++
	return true
}

func (l *limiter) sweep(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, w := range l.hits {
		if now.Sub(w.start) >= l.window {
			delete(l.hits, k)
		}
	}
}

const tooMany = "请求太频繁，请稍后再试"

// clientIP trusts proxy headers only when the request comes from a local reverse proxy (Nginx on the same host).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return host
	}
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(real) != nil {
		return real
	}
	// Nginx appends the peer it saw to X-Forwarded-For, so only the last entry is trustworthy.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if last := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(last) != nil {
			return last
		}
	}
	return host
}

// limitAPI applies the per-IP limit to every API call; static assets are not counted.
func (s *Server) limitAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !s.apiLimit.allow(clientIP(r), time.Now()) {
			writeError(w, http.StatusTooManyRequests, tooMany)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// perIP wraps a handler with a stricter per-IP limit for expensive or unauthenticated endpoints.
func perIP(l *limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r), time.Now()) {
			writeError(w, http.StatusTooManyRequests, tooMany)
			return
		}
		next(w, r)
	}
}

// perUser limits authenticated writes that trigger on-chain checks.
func perUser(l *limiter, next authedHandler) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, userID int64) {
		if !l.allow(userKey(userID), time.Now()) {
			writeError(w, http.StatusTooManyRequests, tooMany)
			return
		}
		next(w, r, userID)
	}
}

func userKey(id int64) string {
	return "u" + strconv.FormatInt(id, 10)
}

// The SPA loads only its own bundled scripts, styles, fonts and images; data: covers the QR code, CSS icons and the small font files Vite inlines.
// connect-src must list the Binance streams PriceTicker.tsx opens.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' wss://stream.binance.com:9443 wss://data-stream.binance.vision; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		if s.secure {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		next.ServeHTTP(w, r)
	})
}
