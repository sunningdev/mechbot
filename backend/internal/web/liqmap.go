package web

import (
	"net/http"
	"time"
)

// handleLiqMap is public: visitors and Starter users get the delayed, coarse map; a Pro session gets the live one.
func (s *Server) handleLiqMap(w http.ResponseWriter, r *http.Request) {
	if s.liqmap == nil {
		writeError(w, http.StatusServiceUnavailable, "清算地图未开启")
		return
	}
	now := time.Now()
	writeJSON(w, http.StatusOK, s.liqmap.View(s.viewerPro(r, now), now))
}

func (s *Server) viewerPro(r *http.Request, now time.Time) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	userID, err := s.store.SessionUser(cookie.Value)
	if err != nil {
		return false
	}
	user, err := s.store.GetUser(userID)
	return err == nil && user.ProActive(now)
}
