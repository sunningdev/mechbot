package store

import (
	"database/sql"
	"math"
	"strings"
	"time"
)

const homePath = "/"

// RecordHomeVisit counts one homepage hit for ip on the given UTC day.
// Unique IPs are one row per (day, path, ip); hits accumulate for repeat visits.
func (s *Store) RecordHomeVisit(ip string, at time.Time) error {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	day := at.UTC().Format("2006-01-02")
	_, err := s.db.Exec(
		`INSERT INTO page_visits (day, path, ip, hits) VALUES (?, ?, ?, 1)
		 ON CONFLICT(day, path, ip) DO UPDATE SET hits = hits + 1`,
		day, homePath, ip,
	)
	return err
}

type DayVisitStats struct {
	Day       string `json:"day"`
	Visits    int64  `json:"visits"`
	UniqueIPs int64  `json:"uniqueIps"`
}

// HomeVisitStats returns daily homepage stats for the last n days (UTC), oldest first.
func (s *Store) HomeVisitStats(now time.Time, days int) ([]DayVisitStats, error) {
	if days < 1 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	start := now.UTC().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	rows, err := s.db.Query(
		`SELECT day, COALESCE(SUM(hits), 0), COUNT(*) FROM page_visits
		 WHERE path = ? AND day >= ? GROUP BY day ORDER BY day`,
		homePath, start,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byDay := map[string]DayVisitStats{}
	for rows.Next() {
		var d DayVisitStats
		if err := rows.Scan(&d.Day, &d.Visits, &d.UniqueIPs); err != nil {
			return nil, err
		}
		byDay[d.Day] = d
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]DayVisitStats, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := now.UTC().AddDate(0, 0, -i).Format("2006-01-02")
		if st, ok := byDay[day]; ok {
			out = append(out, st)
		} else {
			out = append(out, DayVisitStats{Day: day})
		}
	}
	return out, nil
}

func (s *Store) CreateAdminSession(id string, ttl time.Duration) error {
	_, err := s.db.Exec(
		`INSERT INTO admin_sessions (id, expires_unix) VALUES (?, ?)`,
		id, time.Now().Add(ttl).Unix(),
	)
	return err
}

func (s *Store) AdminSessionValid(id string) bool {
	var expires int64
	err := s.db.QueryRow(`SELECT expires_unix FROM admin_sessions WHERE id = ?`, id).Scan(&expires)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expires {
		_, _ = s.db.Exec(`DELETE FROM admin_sessions WHERE id = ?`, id)
		return false
	}
	return true
}

func (s *Store) DeleteAdminSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM admin_sessions WHERE id = ?`, id)
	return err
}

func (s *Store) PurgeAdminSessions(now time.Time) error {
	_, err := s.db.Exec(`DELETE FROM admin_sessions WHERE expires_unix < ?`, now.Unix())
	return err
}

const adminMetaTOTPBound = "totp_bound"

func (s *Store) AdminTOTPBound() bool {
	var v string
	err := s.db.QueryRow(`SELECT value FROM admin_meta WHERE key = ?`, adminMetaTOTPBound).Scan(&v)
	return err == nil && v == "1"
}

func (s *Store) SetAdminTOTPBound(bound bool) error {
	val := "0"
	if bound {
		val = "1"
	}
	_, err := s.db.Exec(
		`INSERT INTO admin_meta (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		adminMetaTOTPBound, val,
	)
	return err
}

type AdminUser struct {
	TelegramID int64     `json:"telegramId"`
	Username   string    `json:"username"`
	FirstName  string    `json:"firstName"`
	Plan       string    `json:"plan"` // starter | trial | pro
	ExpireAt   time.Time `json:"expireAt,omitempty"`
	DaysLeft   int       `json:"daysLeft"`
	PaidUSD    float64   `json:"paidUsd"`
	Chains     []string  `json:"chains"`
	LastPlan   string    `json:"lastPlan,omitempty"` // month | quarter | year from last credited deposit days
	CreatedAt  time.Time `json:"createdAt,omitempty"`
}

// ListAdminUsers returns every user with deposit totals for the admin panel.
func (s *Store) ListAdminUsers(now time.Time) ([]AdminUser, error) {
	rows, err := s.db.Query(`
		SELECT u.telegram_id, u.username, u.first_name, u.expire_at, u.is_premium, u.trial, u.created_at,
			COALESCE((SELECT SUM(amount_micro) FROM deposits d WHERE d.telegram_id = u.telegram_id AND d.status = ?), 0),
			COALESCE((SELECT GROUP_CONCAT(DISTINCT chain) FROM deposits d WHERE d.telegram_id = u.telegram_id AND d.status = ?), ''),
			COALESCE((SELECT days FROM deposits d WHERE d.telegram_id = u.telegram_id AND d.status = ? ORDER BY created_unix DESC LIMIT 1), 0)
		FROM users u
		ORDER BY u.created_at DESC`, DepositCredited, DepositCredited, DepositCredited)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AdminUser
	for rows.Next() {
		var u AdminUser
		var expire, created sql.NullString
		var premium, trial int
		var paidMicro int64
		var chains string
		var lastDays int
		if err := rows.Scan(&u.TelegramID, &u.Username, &u.FirstName, &expire, &premium, &trial, &created,
			&paidMicro, &chains, &lastDays); err != nil {
			return nil, err
		}
		u.ExpireAt = parseTime(expire)
		u.CreatedAt = parseTime(created)
		u.PaidUSD = float64(paidMicro) / 1_000_000
		if chains != "" {
			u.Chains = strings.Split(chains, ",")
		} else {
			u.Chains = []string{}
		}
		u.LastPlan = planKeyFromDays(lastDays)
		active := premium == 1 && u.ExpireAt.After(now)
		switch {
		case active && trial == 1:
			u.Plan = "trial"
		case active:
			u.Plan = "pro"
		default:
			u.Plan = "starter"
		}
		if active {
			// Round up, same as the user-facing proDaysLeft: 30 days minus a few minutes is still 30.
			left := int(math.Ceil(u.ExpireAt.Sub(now).Hours() / 24))
			if left < 0 {
				left = 0
			}
			u.DaysLeft = left
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func planKeyFromDays(days int) string {
	switch days {
	case 30:
		return "month"
	case 90:
		return "quarter"
	case 365:
		return "year"
	default:
		if days > 0 {
			return "custom"
		}
		return ""
	}
}