package store

// AddLiqBorrowers records addresses seen borrowing on a liquidation-map source. Seeing an address again marks it
// for a fresh read, since a new borrow may have reopened a position that was closed.
func (s *Store) AddLiqBorrowers(source string, addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO liq_borrowers (source, address) VALUES (?, ?)
ON CONFLICT(source, address) DO UPDATE SET checked_unix = 0`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, a := range addresses {
		if _, err := stmt.Exec(source, a); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LiqBorrowersToRead lists addresses that still had debt last time or have never been read.
func (s *Store) LiqBorrowersToRead(source string) ([]string, error) {
	rows, err := s.db.Query(`SELECT address FROM liq_borrowers WHERE source = ? AND (debt_usd > 0 OR checked_unix = 0)`, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) SetLiqBorrowerDebts(source string, debts map[string]float64, nowUnix int64) error {
	if len(debts) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`UPDATE liq_borrowers SET debt_usd = ?, checked_unix = ? WHERE source = ? AND address = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for a, d := range debts {
		if _, err := stmt.Exec(d, nowUnix, source, a); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LiqBorrowerCounts returns how many addresses are known and how many had debt at the last read.
func (s *Store) LiqBorrowerCounts(source string) (known, active int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN debt_usd > 0 THEN 1 ELSE 0 END), 0) FROM liq_borrowers WHERE source = ?`, source).
		Scan(&known, &active)
	return known, active, err
}

type LiqSnapshot struct {
	AtUnix int64
	Data   []byte
}

// SaveLiqSnapshot stores a map snapshot and drops all but the newest keep rows.
func (s *Store) SaveLiqSnapshot(atUnix int64, data []byte, keep int) error {
	if _, err := s.db.Exec(`INSERT OR REPLACE INTO liq_snapshots (at_unix, data) VALUES (?, ?)`, atUnix, string(data)); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM liq_snapshots WHERE at_unix NOT IN (SELECT at_unix FROM liq_snapshots ORDER BY at_unix DESC LIMIT ?)`, keep)
	return err
}

// LiqSnapshots returns stored snapshots, oldest first.
func (s *Store) LiqSnapshots() ([]LiqSnapshot, error) {
	rows, err := s.db.Query(`SELECT at_unix, data FROM liq_snapshots ORDER BY at_unix`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LiqSnapshot
	for rows.Next() {
		var snap LiqSnapshot
		var data string
		if err := rows.Scan(&snap.AtUnix, &data); err != nil {
			return nil, err
		}
		snap.Data = []byte(data)
		out = append(out, snap)
	}
	return out, rows.Err()
}
