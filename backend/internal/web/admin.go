package web

import (
	"crypto/subtle"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"mechbot/internal/store"
	"mechbot/internal/totp"
)

const (
	adminCookie = "mechbot_admin"
	// adminIdle ends an admin session after this long without any admin request; each request renews it.
	adminIdle = 30 * time.Minute
	adminUser = "sunzhenning" // fixed; not configurable
)

type adminAuth struct {
	path   string // public URL slug from ADMIN_PATH, e.g. ops-k7m2n9q4x8p3w6
	secret string

	mu       sync.Mutex
	lastCode string // prevent immediate replay of the same TOTP
}

func (s *Server) adminEnabled() bool {
	return s.admin != nil && s.admin.path != "" && s.admin.secret != ""
}

func (s *Server) adminPath() string {
	if s.admin == nil {
		return ""
	}
	return s.admin.path
}

func (s *Server) handleTrackHome(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if err := s.store.RecordHomeVisit(ip, time.Now()); err != nil {
		log.Printf("record home visit: %v", err)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAdminSetup(w http.ResponseWriter, r *http.Request) {
	if !s.adminEnabled() {
		writeError(w, http.StatusServiceUnavailable, "Admin is not configured")
		return
	}
	bound := s.store.AdminTOTPBound()
	out := map[string]any{
		"bound":    bound,
		"username": adminUser,
	}
	// Only expose the secret until the first successful TOTP login (enrollment).
	if !bound {
		url, err := totp.AuthURL("MechBot", adminUser, s.admin.secret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Invalid ADMIN_TOTP_SECRET")
			return
		}
		out["secret"] = s.admin.secret
		out["otpauthUrl"] = url
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAdminAuthLogin(w http.ResponseWriter, r *http.Request) {
	if !s.adminEnabled() {
		writeError(w, http.StatusServiceUnavailable, "Admin is not configured")
		return
	}
	var body struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	username := strings.TrimSpace(body.Username)
	code := strings.TrimSpace(body.Code)
	// Both username and TOTP must pass; any failure returns the same message.
	userOK := username != "" && checkConstUser(adminUser, username)
	codeOK := code != "" && totp.Validate(s.admin.secret, code, time.Now(), 1)

	s.admin.mu.Lock()
	replay := codeOK && code == s.admin.lastCode
	if codeOK && !replay {
		s.admin.lastCode = code
	}
	s.admin.mu.Unlock()

	if !userOK || !codeOK || replay {
		writeError(w, http.StatusUnauthorized, "Illegal login")
		return
	}

	sessionID, err := randomHex(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Sign-in failed")
		return
	}
	if err := s.store.CreateAdminSession(sessionID, adminIdle); err != nil {
		writeError(w, http.StatusInternalServerError, "Sign-in failed")
		return
	}
	if err := s.store.SetAdminTOTPBound(true); err != nil {
		log.Printf("admin totp bound: %v", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  adminCookie,
		Value: sessionID,
		Path:  "/",
		// No MaxAge: a browser session cookie. The server enforces the 30-minute idle timeout.
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(adminCookie); err == nil {
		_ = s.store.DeleteAdminSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteStrictMode})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAdminMe(w http.ResponseWriter, r *http.Request) {
	if !s.adminEnabled() {
		writeError(w, http.StatusServiceUnavailable, "Admin is not configured")
		return
	}
	if !s.adminAuthed(r) {
		writeError(w, http.StatusUnauthorized, "Please sign in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"username": adminUser,
	})
}

func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	days, err := s.store.HomeVisitStats(time.Now(), 30)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load stats")
		return
	}
	var todayVisits, todayIPs int64
	if n := len(days); n > 0 {
		todayVisits = days[n-1].Visits
		todayIPs = days[n-1].UniqueIPs
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"today": map[string]any{"visits": todayVisits, "uniqueIps": todayIPs},
		"days":  days,
	})
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	users, err := s.store.ListAdminUsers(time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load users")
		return
	}
	if users == nil {
		users = []store.AdminUser{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !s.adminEnabled() {
		writeError(w, http.StatusServiceUnavailable, "Admin is not configured")
		return false
	}
	if !s.adminAuthed(r) {
		writeError(w, http.StatusUnauthorized, "Please sign in")
		return false
	}
	return true
}

func (s *Server) adminAuthed(r *http.Request) bool {
	cookie, err := r.Cookie(adminCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	return s.store.AdminSessionValid(cookie.Value, adminIdle)
}

func checkConstUser(want, got string) bool {
	want = strings.TrimSpace(want)
	got = strings.TrimSpace(got)
	a := []byte(strings.ToLower(want))
	b := []byte(strings.ToLower(got))
	if len(a) != len(b) {
		dummy := make([]byte, len(a))
		subtle.ConstantTimeCompare(a, dummy)
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}
