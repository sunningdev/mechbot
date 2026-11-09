package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"mechbot/internal/config"
	"mechbot/internal/i18n"
)

var (
	ErrAddressExists = errors.New("地址已在监控中")
	ErrAddressLimit  = errors.New("已达到地址数量上限")
	ErrNotFound      = errors.New("没有找到")
)

const (
	FreeAddressLimit = 3
	ProAddressLimit  = 20
	// TrialDays of Pro go to every new user once; buying or earning Pro days ends the trial.
	TrialDays = 7
)

type User struct {
	TelegramID      int64
	Username        string
	FirstName       string
	RefBy           sql.NullInt64
	ExpireAt        time.Time
	IsPremium       bool
	Trial           bool
	MinHealthFactor float64
	AddressCount    int
	Lang            i18n.Lang
}

func (u User) ProActive(now time.Time) bool {
	return u.IsPremium && u.ExpireAt.After(now)
}

func (u User) AddressLimit(now time.Time) int {
	return AddressLimit(u.ProActive(now))
}

type Watch struct {
	Address    string
	TelegramID int64
	// Threshold is the address's own alert line; new addresses start at config.DefaultThreshold.
	Threshold float64
	Pro       bool
	Active    bool
	Lang      i18n.Lang
}

type AlertState struct {
	InDanger  bool
	LastAlert time.Time
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS users (
  telegram_id INTEGER PRIMARY KEY,
  ref_by INTEGER,
  expire_at TEXT,
  is_premium INTEGER NOT NULL DEFAULT 0,
  min_health_factor REAL NOT NULL DEFAULT 1.2,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS addresses (
  address TEXT NOT NULL,
  telegram_id INTEGER NOT NULL,
  min_health_factor REAL NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (address, telegram_id)
);
CREATE TABLE IF NOT EXISTS alert_state (
  address TEXT NOT NULL,
  telegram_id INTEGER NOT NULL,
  protocol_key TEXT NOT NULL,
  in_danger INTEGER NOT NULL DEFAULT 0,
  last_alert_unix INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (address, telegram_id, protocol_key)
);
CREATE TABLE IF NOT EXISTS snapshots (
  address TEXT NOT NULL,
  protocol_key TEXT NOT NULL,
  label TEXT NOT NULL,
  health_factor REAL NOT NULL DEFAULT 0,
  collateral_usd REAL NOT NULL DEFAULT 0,
  debt_usd REAL NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (address, protocol_key)
);
CREATE TABLE IF NOT EXISTS address_checks (
  address TEXT PRIMARY KEY,
  checked_unix INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS login_tokens (
  token TEXT PRIMARY KEY,
  code TEXT NOT NULL,
  ref_by INTEGER,
  telegram_id INTEGER,
  status TEXT NOT NULL DEFAULT 'pending',
  created_unix INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  telegram_id INTEGER NOT NULL,
  expires_unix INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS deposit_addresses (
  telegram_id INTEGER PRIMARY KEY,
  idx INTEGER NOT NULL UNIQUE,
  evm TEXT NOT NULL DEFAULT '',
  tron TEXT NOT NULL DEFAULT '',
  watch_until_unix INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS deposit_addresses_evm ON deposit_addresses (evm);
CREATE INDEX IF NOT EXISTS deposit_addresses_tron ON deposit_addresses (tron);
CREATE TABLE IF NOT EXISTS deposits (
  chain TEXT NOT NULL,
  tx_hash TEXT NOT NULL,
  log_index INTEGER NOT NULL,
  telegram_id INTEGER NOT NULL,
  amount_micro INTEGER NOT NULL,
  days INTEGER NOT NULL,
  status TEXT NOT NULL,
  created_unix INTEGER NOT NULL,
  PRIMARY KEY (chain, tx_hash, log_index)
);
CREATE INDEX IF NOT EXISTS deposits_user ON deposits (telegram_id, created_unix);
CREATE TABLE IF NOT EXISTS scan_cursors (
  chain TEXT PRIMARY KEY,
  block INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS liq_borrowers (
  source TEXT NOT NULL,
  address TEXT NOT NULL,
  debt_usd REAL NOT NULL DEFAULT 0,
  checked_unix INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (source, address)
);
CREATE TABLE IF NOT EXISTS liq_snapshots (
  at_unix INTEGER PRIMARY KEY,
  data TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS referral_rewards (
  referee_id INTEGER PRIMARY KEY,
  referrer_id INTEGER NOT NULL,
  days INTEGER NOT NULL,
  created_unix INTEGER NOT NULL
);`)
	if err != nil {
		return err
	}
	if err := s.ensureColumn("users", "username", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn("snapshots", "eth_price", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("snapshots", "liq_eth_price", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	for _, c := range [][2]string{
		{"coll_symbol", "TEXT NOT NULL DEFAULT ''"},
		{"coll_price", "REAL NOT NULL DEFAULT 0"},
		{"coll_lt", "REAL NOT NULL DEFAULT 0"},
		{"debt_symbol", "TEXT NOT NULL DEFAULT ''"},
		{"debt_price", "REAL NOT NULL DEFAULT 0"},
		{"app_url", "TEXT NOT NULL DEFAULT ''"},
		{"borrow_apy", "REAL NOT NULL DEFAULT 0"},
	} {
		if err := s.ensureColumn("snapshots", c[0], c[1]); err != nil {
			return err
		}
	}
	for _, c := range [][2]string{
		{"first_name", "TEXT NOT NULL DEFAULT ''"},
		{"trial", "INTEGER NOT NULL DEFAULT 0"},
		// pro_notice: 0 nothing sent for the current Pro period, 1 "ends tomorrow" sent, 2 "ended" sent.
		{"pro_notice", "INTEGER NOT NULL DEFAULT 0"},
		// lang is the language picked on the website or with /lang; tg_lang is the Telegram client's language_code.
		{"lang", "TEXT NOT NULL DEFAULT ''"},
		{"tg_lang", "TEXT NOT NULL DEFAULT ''"},
	} {
		if err := s.ensureColumn("users", c[0], c[1]); err != nil {
			return err
		}
	}
	if err := s.ensureColumn("deposits", "token", "TEXT NOT NULL DEFAULT 'USDT'"); err != nil {
		return err
	}
	for _, c := range [][2]string{
		{"stress10", "REAL NOT NULL DEFAULT 0"},
		{"stress20", "REAL NOT NULL DEFAULT 0"},
		{"stress30", "REAL NOT NULL DEFAULT 0"},
	} {
		if err := s.ensureColumn("snapshots", c[0], c[1]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureColumn(table, column, definition string) error {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
	return err
}

func (s *Store) EnsureUser(telegramID int64, refBy *int64, threshold float64) (created bool, err error) {
	var exists int
	err = s.db.QueryRow(`SELECT 1 FROM users WHERE telegram_id = ?`, telegramID).Scan(&exists)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	var ref any
	if refBy != nil && *refBy != 0 && *refBy != telegramID {
		var refExists int
		if s.db.QueryRow(`SELECT 1 FROM users WHERE telegram_id = ?`, *refBy).Scan(&refExists) == nil {
			ref = *refBy
		}
	}
	now := time.Now().UTC()
	_, err = s.db.Exec(
		`INSERT INTO users (telegram_id, ref_by, is_premium, expire_at, trial, min_health_factor, created_at) VALUES (?, ?, 1, ?, 1, ?, ?)`,
		telegramID, ref, now.Add(TrialDays*24*time.Hour).Format(time.RFC3339), threshold, now.Format(time.RFC3339),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateProfile stores the latest Telegram username. A username belongs to one account at a time, so any other
// account still recorded under it (e.g. an abandoned account whose name was taken over) loses it.
func (s *Store) UpdateProfile(telegramID int64, username, firstName string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if username != "" {
		if _, err := tx.Exec(`UPDATE users SET username = '' WHERE username = ? COLLATE NOCASE AND telegram_id != ?`, username, telegramID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE users SET username = ?, first_name = ? WHERE telegram_id = ?`, username, firstName, telegramID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetUser(telegramID int64) (User, error) {
	return s.userWhere(`u.telegram_id = ?`, telegramID)
}

// UserByUsername finds the account that has talked to the bot under this Telegram username, ignoring case and a leading @.
func (s *Store) UserByUsername(username string) (User, error) {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" {
		return User{}, ErrNotFound
	}
	return s.userWhere(`u.username = ? COLLATE NOCASE`, username)
}

func (s *Store) userWhere(where string, arg any) (User, error) {
	var u User
	var premium int
	var expire sql.NullString
	var lang, tgLang string
	err := s.db.QueryRow(`
SELECT u.telegram_id, u.username, u.first_name, u.ref_by, u.expire_at, u.is_premium, u.trial, u.min_health_factor,
       (SELECT COUNT(*) FROM addresses a WHERE a.telegram_id = u.telegram_id), u.lang, u.tg_lang
FROM users u WHERE `+where, arg).Scan(
		&u.TelegramID, &u.Username, &u.FirstName, &u.RefBy, &expire, &premium, &u.Trial, &u.MinHealthFactor, &u.AddressCount,
		&lang, &tgLang,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	u.IsPremium = premium == 1
	u.ExpireAt = parseTime(expire)
	u.Lang = i18n.Resolve(lang, tgLang)
	return u, nil
}

// SetLang records the language the user picked on the website or with /lang.
func (s *Store) SetLang(telegramID int64, l i18n.Lang) error {
	_, err := s.db.Exec(`UPDATE users SET lang = ? WHERE telegram_id = ?`, string(l), telegramID)
	return err
}

// SetTelegramLang records the Telegram client's language_code, used until the user picks a language.
func (s *Store) SetTelegramLang(telegramID int64, code string) error {
	_, err := s.db.Exec(`UPDATE users SET tg_lang = ? WHERE telegram_id = ? AND tg_lang != ?`, code, telegramID, code)
	return err
}

// UserLang is the language to answer a user in; tgFallback covers users who have not registered yet.
func (s *Store) UserLang(telegramID int64, tgFallback string) i18n.Lang {
	var lang, tgLang string
	if err := s.db.QueryRow(`SELECT lang, tg_lang FROM users WHERE telegram_id = ?`, telegramID).Scan(&lang, &tgLang); err != nil {
		return i18n.Resolve("", tgFallback)
	}
	if tgLang == "" {
		tgLang = tgFallback
	}
	return i18n.Resolve(lang, tgLang)
}

func (s *Store) SetAddressThreshold(telegramID int64, address string, threshold float64) error {
	res, err := s.db.Exec(
		`UPDATE addresses SET min_health_factor = ? WHERE telegram_id = ? AND address = ?`, threshold, telegramID, address,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AddAddress(telegramID int64, address string, limit int) error {
	address = strings.TrimSpace(address)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var known int
	err = tx.QueryRow(`SELECT 1 FROM users WHERE telegram_id = ?`, telegramID).Scan(&known)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var existing int
	err = tx.QueryRow(`SELECT 1 FROM addresses WHERE telegram_id = ? AND address = ?`, telegramID, address).Scan(&existing)
	if err == nil {
		return ErrAddressExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM addresses WHERE telegram_id = ?`, telegramID).Scan(&count); err != nil {
		return err
	}
	if count >= limit {
		return ErrAddressLimit
	}
	_, err = tx.Exec(
		`INSERT INTO addresses (address, telegram_id, min_health_factor, created_at) VALUES (?, ?, ?, ?)`,
		address, telegramID, config.DefaultThreshold, time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RemoveAddress(telegramID int64, address string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.Exec(`DELETE FROM addresses WHERE telegram_id = ? AND address = ?`, telegramID, address)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(`DELETE FROM alert_state WHERE telegram_id = ? AND address = ?`, telegramID, address); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ListAddresses(telegramID int64) ([]Watch, error) {
	all, err := s.listWatches(`WHERE a.telegram_id = ?`, telegramID)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (s *Store) ListAll() ([]Watch, error) {
	return s.listWatches(``)
}

// Active is false for addresses beyond the free limit once Pro has lapsed; they stay listed but are not polled.
func (s *Store) listWatches(where string, args ...any) ([]Watch, error) {
	rows, err := s.db.Query(`
SELECT a.address, a.telegram_id, a.min_health_factor, u.is_premium, u.expire_at, u.lang, u.tg_lang,
       ROW_NUMBER() OVER (PARTITION BY a.telegram_id ORDER BY a.created_at, a.address)
FROM addresses a JOIN users u ON u.telegram_id = a.telegram_id `+where+`
ORDER BY a.telegram_id, a.created_at, a.address`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	now := time.Now()
	var out []Watch
	for rows.Next() {
		var w Watch
		var premium, rank int
		var expire sql.NullString
		var lang, tgLang string
		if err := rows.Scan(&w.Address, &w.TelegramID, &w.Threshold, &premium, &expire, &lang, &tgLang, &rank); err != nil {
			return nil, err
		}
		w.Lang = i18n.Resolve(lang, tgLang)
		w.Pro = premium == 1 && parseTime(expire).After(now)
		w.Active = rank <= AddressLimit(w.Pro)
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) DangerKeys(address string, telegramID int64) ([]string, error) {
	rows, err := s.db.Query(
		`SELECT protocol_key FROM alert_state WHERE address = ? AND telegram_id = ? AND in_danger = 1`,
		address, telegramID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *Store) AlertState(address string, telegramID int64, protocolKey string) (AlertState, error) {
	var inDanger int
	var last int64
	err := s.db.QueryRow(
		`SELECT in_danger, last_alert_unix FROM alert_state WHERE address = ? AND telegram_id = ? AND protocol_key = ?`,
		address, telegramID, protocolKey,
	).Scan(&inDanger, &last)
	if errors.Is(err, sql.ErrNoRows) {
		return AlertState{}, nil
	}
	if err != nil {
		return AlertState{}, err
	}
	st := AlertState{InDanger: inDanger == 1}
	if last > 0 {
		st.LastAlert = time.Unix(last, 0)
	}
	return st, nil
}

func (s *Store) MarkDanger(address string, telegramID int64, protocolKey string, at time.Time) error {
	_, err := s.db.Exec(`
INSERT INTO alert_state (address, telegram_id, protocol_key, in_danger, last_alert_unix)
VALUES (?, ?, ?, 1, ?)
ON CONFLICT(address, telegram_id, protocol_key)
DO UPDATE SET in_danger = 1, last_alert_unix = excluded.last_alert_unix`,
		address, telegramID, protocolKey, at.Unix(),
	)
	return err
}

func (s *Store) MarkSafe(address string, telegramID int64, protocolKey string) error {
	_, err := s.db.Exec(
		`UPDATE alert_state SET in_danger = 0 WHERE address = ? AND telegram_id = ? AND protocol_key = ?`,
		address, telegramID, protocolKey,
	)
	return err
}

func AddressLimit(pro bool) int {
	if pro {
		return ProAddressLimit
	}
	return FreeAddressLimit
}

func FormatStatus(u User) string {
	t := i18n.T(u.Lang)
	now := time.Now()
	plan := t.StatusStarter
	expire := "—"
	if u.ProActive(now) {
		plan = t.StatusPro
		if u.Trial {
			plan = t.StatusTrial
		}
		expire = u.ExpireAt.In(time.Local).Format("2006-01-02 15:04")
	}
	ref := t.StatusNone
	if u.RefBy.Valid {
		ref = fmt.Sprintf("%d", u.RefBy.Int64)
	}
	return fmt.Sprintf(t.Status, plan, u.AddressCount, u.AddressLimit(now), expire, ref)
}

func parseTime(v sql.NullString) time.Time {
	if !v.Valid || v.String == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, v.String)
	if err != nil {
		return time.Time{}
	}
	return t
}
