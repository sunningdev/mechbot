package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type DepositAddress struct {
	TelegramID int64
	Index      uint32
	EVM        string
	Tron       string
}

// DeriveFunc maps an HD index to its deposit addresses; an empty string means that chain family is not configured.
type DeriveFunc func(index uint32) (evm, tron string, err error)

// DepositAddress returns the user's permanent deposit address, allocating the next HD index on first use,
// and keeps the user on the watch list until watchUntil.
func (s *Store) DepositAddress(telegramID int64, derive DeriveFunc, watchUntil time.Time) (DepositAddress, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return DepositAddress{}, err
	}
	defer func() { _ = tx.Rollback() }()

	d := DepositAddress{TelegramID: telegramID}
	var storedEVM, storedTron string
	err = tx.QueryRow(`SELECT idx, evm, tron FROM deposit_addresses WHERE telegram_id = ?`, telegramID).Scan(&d.Index, &storedEVM, &storedTron)
	fresh := errors.Is(err, sql.ErrNoRows)
	if err != nil && !fresh {
		return DepositAddress{}, err
	}
	if fresh {
		if err := tx.QueryRow(`SELECT COALESCE(MAX(idx), 0) + 1 FROM deposit_addresses`).Scan(&d.Index); err != nil {
			return DepositAddress{}, err
		}
	}
	if d.EVM, d.Tron, err = derive(d.Index); err != nil {
		return DepositAddress{}, err
	}
	d.EVM = strings.ToLower(d.EVM)

	if fresh {
		_, err = tx.Exec(
			`INSERT INTO deposit_addresses (telegram_id, idx, evm, tron, watch_until_unix) VALUES (?, ?, ?, ?, ?)`,
			telegramID, d.Index, d.EVM, d.Tron, watchUntil.Unix(),
		)
	} else {
		_, err = tx.Exec(
			`UPDATE deposit_addresses SET evm = ?, tron = ?, watch_until_unix = MAX(watch_until_unix, ?) WHERE telegram_id = ?`,
			d.EVM, d.Tron, watchUntil.Unix(), telegramID,
		)
	}
	if err != nil {
		return DepositAddress{}, err
	}
	return d, tx.Commit()
}

// EVMDepositOwners maps every lowercase EVM deposit address to its user.
func (s *Store) EVMDepositOwners() (map[string]int64, error) {
	return s.depositOwners(`SELECT evm, telegram_id FROM deposit_addresses WHERE evm != ''`)
}

// TronDepositOwners maps TRON deposit addresses still inside their watch window to their user.
func (s *Store) TronDepositOwners(now time.Time) (map[string]int64, error) {
	return s.depositOwners(`SELECT tron, telegram_id FROM deposit_addresses WHERE tron != '' AND watch_until_unix >= ?`, now.Unix())
}

func (s *Store) DepositAddressOf(telegramID int64) (DepositAddress, error) {
	d := DepositAddress{TelegramID: telegramID}
	err := s.db.QueryRow(`SELECT idx, evm, tron FROM deposit_addresses WHERE telegram_id = ?`, telegramID).Scan(&d.Index, &d.EVM, &d.Tron)
	if errors.Is(err, sql.ErrNoRows) {
		return DepositAddress{}, ErrNotFound
	}
	return d, err
}

func (s *Store) AllDepositAddresses() ([]DepositAddress, error) {
	rows, err := s.db.Query(`SELECT telegram_id, idx, evm, tron FROM deposit_addresses ORDER BY idx`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DepositAddress
	for rows.Next() {
		var d DepositAddress
		if err := rows.Scan(&d.TelegramID, &d.Index, &d.EVM, &d.Tron); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) depositOwners(query string, args ...any) (map[string]int64, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var addr string
		var id int64
		if err := rows.Scan(&addr, &id); err != nil {
			return nil, err
		}
		out[addr] = id
	}
	return out, rows.Err()
}

func (s *Store) ScanCursor(chain string) (uint64, bool, error) {
	var block uint64
	err := s.db.QueryRow(`SELECT block FROM scan_cursors WHERE chain = ?`, chain).Scan(&block)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return block, err == nil, err
}

func (s *Store) SetScanCursor(chain string, block uint64) error {
	_, err := s.db.Exec(
		`INSERT INTO scan_cursors (chain, block) VALUES (?, ?) ON CONFLICT(chain) DO UPDATE SET block = excluded.block`,
		chain, block,
	)
	return err
}

const (
	DepositCredited  = "credited"
	DepositUnmatched = "unmatched"
)

type Deposit struct {
	Chain       string
	TxHash      string
	LogIndex    int
	TelegramID  int64
	AmountMicro int64
	Days        int // 0 means the amount matched no plan
	Status      string
	CreatedAt   time.Time
}

type PaymentResult struct {
	ExpireAt       time.Time
	BonusDays      int // referral bonus added to the buyer on top of the plan
	ReferrerID     int64
	ReferrerDays   int
	ReferrerExpire time.Time
}

// CreditDeposit records an on-chain transfer once. New reports whether this call recorded it;
// a credited deposit extends Pro and, on the user's first credited deposit, rewards both the buyer and the referrer.
func (s *Store) CreditDeposit(d Deposit, now time.Time) (result PaymentResult, isNew bool, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return PaymentResult{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	status := DepositUnmatched
	if d.Days > 0 {
		status = DepositCredited
	}
	var paidBefore int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM deposits WHERE telegram_id = ? AND status = ?`, d.TelegramID, DepositCredited).Scan(&paidBefore); err != nil {
		return PaymentResult{}, false, err
	}
	res, err := tx.Exec(
		`INSERT OR IGNORE INTO deposits (chain, tx_hash, log_index, telegram_id, amount_micro, days, status, created_unix) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.Chain, strings.ToLower(d.TxHash), d.LogIndex, d.TelegramID, d.AmountMicro, d.Days, status, d.CreatedAt.Unix(),
	)
	if err != nil {
		return PaymentResult{}, false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return PaymentResult{}, false, nil
	}
	if status == DepositUnmatched {
		return PaymentResult{}, true, tx.Commit()
	}

	var referrer int64
	bonus := 0
	if paidBefore == 0 {
		if referrer, bonus, err = claimReferral(tx, d.TelegramID, d.Days, now); err != nil {
			return PaymentResult{}, false, err
		}
	}
	expire, err := extendPro(tx, d.TelegramID, d.Days+bonus, now)
	if err != nil {
		return PaymentResult{}, false, err
	}
	result = PaymentResult{ExpireAt: expire, BonusDays: bonus}
	if referrer != 0 {
		refExpire, err := extendPro(tx, referrer, bonus, now)
		if err != nil {
			return PaymentResult{}, false, err
		}
		result.ReferrerID, result.ReferrerDays, result.ReferrerExpire = referrer, bonus, refExpire
	}
	return result, true, tx.Commit()
}

// claimReferral records the one-time reward for a referred user's first purchase; bonus is 0 when nothing is due.
func claimReferral(tx *sql.Tx, telegramID int64, planDays int, now time.Time) (referrer int64, bonus int, err error) {
	var ref sql.NullInt64
	if err := tx.QueryRow(`SELECT ref_by FROM users WHERE telegram_id = ?`, telegramID).Scan(&ref); err != nil {
		return 0, 0, err
	}
	if !ref.Valid || ref.Int64 == telegramID {
		return 0, 0, nil
	}
	bonus = ReferralBonusDays(planDays)
	res, err := tx.Exec(
		`INSERT OR IGNORE INTO referral_rewards (referee_id, referrer_id, days, created_unix) VALUES (?, ?, ?, ?)`,
		telegramID, ref.Int64, bonus, now.Unix(),
	)
	if err != nil {
		return 0, 0, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return 0, 0, nil
	}
	return ref.Int64, bonus, nil
}

func (s *Store) Deposits(telegramID int64, limit int) ([]Deposit, error) {
	return s.depositRows(telegramID, limit, 0)
}

// DepositPage returns one page of a user's deposits, newest first, plus the total count.
func (s *Store) DepositPage(telegramID int64, limit, offset int) ([]Deposit, int, error) {
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM deposits WHERE telegram_id = ?`, telegramID).Scan(&total); err != nil {
		return nil, 0, err
	}
	out, err := s.depositRows(telegramID, limit, offset)
	return out, total, err
}

func (s *Store) depositRows(telegramID int64, limit, offset int) ([]Deposit, error) {
	rows, err := s.db.Query(
		`SELECT chain, tx_hash, log_index, telegram_id, amount_micro, days, status, created_unix FROM deposits WHERE telegram_id = ?
ORDER BY created_unix DESC, tx_hash, log_index LIMIT ? OFFSET ?`,
		telegramID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Deposit
	for rows.Next() {
		var d Deposit
		var created int64
		if err := rows.Scan(&d.Chain, &d.TxHash, &d.LogIndex, &d.TelegramID, &d.AmountMicro, &d.Days, &d.Status, &created); err != nil {
			return nil, err
		}
		d.CreatedAt = time.Unix(created, 0)
		out = append(out, d)
	}
	return out, rows.Err()
}
