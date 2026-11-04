package store

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTokenInvalid = errors.New("登录链接无效或已过期")
)

const LoginTokenTTL = 10 * time.Minute

type Snapshot struct {
	ProtocolKey   string
	Label         string
	HealthFactor  float64
	CollateralUSD float64
	DebtUSD       float64
	Error         string
	ETHPrice      float64
	LiqETHPrice   float64
	CollSymbol    string
	CollPrice     float64
	CollLT        float64
	DebtSymbol    string
	DebtPrice     float64
	AppURL        string
	BorrowAPY     float64
	// Stress holds the health factor after ETH falls 10%, 20% and 30%; zeros when unknown.
	Stress [3]float64
}

func (s *Store) SaveSnapshots(address string, rows []Snapshot, at time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM snapshots WHERE address = ?`, address); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO snapshots (address, protocol_key, label, health_factor, collateral_usd, debt_usd, error, eth_price, liq_eth_price,
				coll_symbol, coll_price, coll_lt, debt_symbol, debt_price, app_url, borrow_apy, stress10, stress20, stress30)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			address, r.ProtocolKey, r.Label, r.HealthFactor, r.CollateralUSD, r.DebtUSD, r.Error, r.ETHPrice, r.LiqETHPrice,
			r.CollSymbol, r.CollPrice, r.CollLT, r.DebtSymbol, r.DebtPrice, r.AppURL, r.BorrowAPY, r.Stress[0], r.Stress[1], r.Stress[2],
		); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(
		`INSERT INTO address_checks (address, checked_unix) VALUES (?, ?) ON CONFLICT(address) DO UPDATE SET checked_unix = excluded.checked_unix`,
		address, at.Unix(),
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Snapshots(address string) ([]Snapshot, time.Time, error) {
	var checked int64
	err := s.db.QueryRow(`SELECT checked_unix FROM address_checks WHERE address = ?`, address).Scan(&checked)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, time.Time{}, err
	}
	rows, err := s.db.Query(
		`SELECT protocol_key, label, health_factor, collateral_usd, debt_usd, error, eth_price, liq_eth_price,
			coll_symbol, coll_price, coll_lt, debt_symbol, debt_price, app_url, borrow_apy, stress10, stress20, stress30
		FROM snapshots WHERE address = ? ORDER BY error = '', health_factor`,
		address,
	)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer rows.Close()
	var out []Snapshot
	for rows.Next() {
		var r Snapshot
		if err := rows.Scan(&r.ProtocolKey, &r.Label, &r.HealthFactor, &r.CollateralUSD, &r.DebtUSD, &r.Error, &r.ETHPrice, &r.LiqETHPrice,
			&r.CollSymbol, &r.CollPrice, &r.CollLT, &r.DebtSymbol, &r.DebtPrice, &r.AppURL, &r.BorrowAPY,
			&r.Stress[0], &r.Stress[1], &r.Stress[2]); err != nil {
			return nil, time.Time{}, err
		}
		out = append(out, r)
	}
	var at time.Time
	if checked > 0 {
		at = time.Unix(checked, 0)
	}
	return out, at, rows.Err()
}

func (s *Store) CreateLoginToken(token, code string, refBy *int64) error {
	var ref any
	if refBy != nil && *refBy != 0 {
		ref = *refBy
	}
	_, err := s.db.Exec(
		`INSERT INTO login_tokens (token, code, ref_by, status, created_unix) VALUES (?, ?, ?, 'pending', ?)`,
		token, code, ref, time.Now().Unix(),
	)
	return err
}

type LoginToken struct {
	Code       string
	RefBy      *int64
	TelegramID int64
	Status     string
}

func (s *Store) LoginToken(token string) (LoginToken, error) {
	var t LoginToken
	var ref, tg sql.NullInt64
	var created int64
	err := s.db.QueryRow(
		`SELECT code, ref_by, telegram_id, status, created_unix FROM login_tokens WHERE token = ?`, token,
	).Scan(&t.Code, &ref, &tg, &t.Status, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return LoginToken{}, ErrTokenInvalid
	}
	if err != nil {
		return LoginToken{}, err
	}
	if time.Since(time.Unix(created, 0)) > LoginTokenTTL {
		return LoginToken{}, ErrTokenInvalid
	}
	if ref.Valid {
		v := ref.Int64
		t.RefBy = &v
	}
	if tg.Valid {
		t.TelegramID = tg.Int64
	}
	return t, nil
}

func (s *Store) ConfirmLoginToken(token string, telegramID int64) error {
	res, err := s.db.Exec(
		`UPDATE login_tokens SET status = 'confirmed', telegram_id = ? WHERE token = ? AND status = 'pending' AND created_unix >= ?`,
		telegramID, token, time.Now().Add(-LoginTokenTTL).Unix(),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrTokenInvalid
	}
	return nil
}

// ConsumeLoginToken turns a confirmed token into a session exactly once.
func (s *Store) ConsumeLoginToken(token, sessionID string, ttl time.Duration) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var tg int64
	err = tx.QueryRow(
		`SELECT telegram_id FROM login_tokens WHERE token = ? AND status = 'confirmed' AND created_unix >= ?`,
		token, time.Now().Add(-LoginTokenTTL).Unix(),
	).Scan(&tg)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenInvalid
	}
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`UPDATE login_tokens SET status = 'used' WHERE token = ?`, token); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO sessions (id, telegram_id, expires_unix) VALUES (?, ?, ?)`,
		sessionID, tg, time.Now().Add(ttl).Unix(),
	); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`DELETE FROM login_tokens WHERE created_unix < ?`, time.Now().Add(-24*time.Hour).Unix()); err != nil {
		return 0, err
	}
	return tg, tx.Commit()
}

func (s *Store) SessionUser(id string) (int64, error) {
	var tg, expires int64
	err := s.db.QueryRow(`SELECT telegram_id, expires_unix FROM sessions WHERE id = ?`, id).Scan(&tg, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if time.Now().Unix() > expires {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
		return 0, ErrNotFound
	}
	return tg, nil
}

func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func extendPro(tx *sql.Tx, telegramID int64, days int, now time.Time) (time.Time, error) {
	var expire sql.NullString
	if err := tx.QueryRow(`SELECT expire_at FROM users WHERE telegram_id = ?`, telegramID).Scan(&expire); err != nil {
		return time.Time{}, err
	}
	base := parseTime(expire)
	if base.Before(now) {
		base = now
	}
	next := base.Add(time.Duration(days) * 24 * time.Hour).UTC()
	_, err := tx.Exec(
		`UPDATE users SET is_premium = 1, expire_at = ?, trial = 0, pro_notice = 0 WHERE telegram_id = ?`, next.Format(time.RFC3339), telegramID,
	)
	return next, err
}

const (
	NoticeEndingSoon = 1
	NoticeEnded      = 2
)

type ProNotice struct {
	TelegramID   int64
	ExpireAt     time.Time
	Trial        bool
	Stage        int  // the notice due now
	Stale        bool // ended too long ago to be worth a message; only mark it
	AddressCount int
	expireRaw    string
}

// ProNoticesDue lists users who should hear that their Pro ends within notifyBefore, or has ended. Periods that ended
// more than staleAfter ago are only marked, so a deploy doesn't message every user whose Pro lapsed long ago.
func (s *Store) ProNoticesDue(now time.Time, notifyBefore, staleAfter time.Duration) ([]ProNotice, error) {
	rows, err := s.db.Query(`
SELECT u.telegram_id, u.expire_at, u.trial, u.pro_notice,
       (SELECT COUNT(*) FROM addresses a WHERE a.telegram_id = u.telegram_id)
FROM users u WHERE u.is_premium = 1 AND u.expire_at IS NOT NULL AND u.pro_notice < ?`, NoticeEnded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProNotice
	for rows.Next() {
		var n ProNotice
		var expire sql.NullString
		var sent int
		if err := rows.Scan(&n.TelegramID, &expire, &n.Trial, &sent, &n.AddressCount); err != nil {
			return nil, err
		}
		n.ExpireAt, n.expireRaw = parseTime(expire), expire.String
		if n.ExpireAt.IsZero() {
			continue
		}
		switch {
		case !n.ExpireAt.After(now):
			n.Stage = NoticeEnded
			n.Stale = now.Sub(n.ExpireAt) > staleAfter
		case n.ExpireAt.Sub(now) <= notifyBefore && sent < NoticeEndingSoon:
			n.Stage = NoticeEndingSoon
		default:
			continue
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// SetProNotice records the notice sent for the current Pro period; it only moves forward, and only for the same
// expiry, so a renewal that lands meanwhile isn't overwritten.
func (s *Store) SetProNotice(n ProNotice) error {
	_, err := s.db.Exec(`UPDATE users SET pro_notice = ? WHERE telegram_id = ? AND expire_at = ? AND pro_notice < ?`,
		n.Stage, n.TelegramID, n.expireRaw, n.Stage)
	return err
}

// ReferralBonusDays is what both the referrer and the referred user get for the referred user's first plan.
func ReferralBonusDays(planDays int) int {
	switch {
	case planDays >= 365:
		return 30
	case planDays >= 90:
		return 15
	default:
		return 7
	}
}

type ReferralStats struct {
	Invited  int
	Paid     int
	DaysEarn int
}

func (s *Store) Referral(telegramID int64) (ReferralStats, error) {
	var st ReferralStats
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE ref_by = ?`, telegramID).Scan(&st.Invited); err != nil {
		return st, err
	}
	err := s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(days), 0) FROM referral_rewards WHERE referrer_id = ?`, telegramID,
	).Scan(&st.Paid, &st.DaysEarn)
	return st, err
}
