package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// mustUser creates a user without the sign-up trial, so payment arithmetic starts from zero.
func mustUser(t *testing.T, st *Store, id int64, ref *int64) {
	t.Helper()
	if _, err := st.EnsureUser(id, ref, 1.2); err != nil {
		t.Fatal(err)
	}
	if _, err := st.db.Exec(`UPDATE users SET is_premium = 0, expire_at = NULL, trial = 0 WHERE telegram_id = ?`, id); err != nil {
		t.Fatal(err)
	}
}

func credit(t *testing.T, st *Store, user int64, tx string, days int, now time.Time) (PaymentResult, bool) {
	t.Helper()
	res, isNew, err := st.CreditDeposit(Deposit{Chain: "bsc", TxHash: tx, TelegramID: user, AmountMicro: 5_000_000, Days: days, CreatedAt: now}, now)
	if err != nil {
		t.Fatal(err)
	}
	return res, isNew
}

func TestDepositPage(t *testing.T) {
	st := openTest(t)
	mustUser(t, st, 1, nil)
	mustUser(t, st, 2, nil)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 12; i++ {
		credit(t, st, 1, fmt.Sprintf("0x%02d", i), 30, start.Add(time.Duration(i)*time.Hour))
	}
	credit(t, st, 2, "0xother", 30, start)

	first, total, err := st.DepositPage(1, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 12 || len(first) != 10 || first[0].TxHash != "0x11" {
		t.Fatalf("first page: total %d, len %d, newest %s", total, len(first), first[0].TxHash)
	}
	second, _, err := st.DepositPage(1, 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 2 || second[1].TxHash != "0x00" {
		t.Fatalf("second page: len %d", len(second))
	}
}

func TestCreditDepositCreditsBuyerAndReferrerOnce(t *testing.T) {
	st := openTest(t)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	referrer := int64(100)
	mustUser(t, st, referrer, nil)
	mustUser(t, st, 200, &referrer)

	if _, isNew := credit(t, st, 200, "0xdust", 0, now); !isNew {
		t.Fatal("unmatched deposit should still be recorded")
	}

	day := 24 * time.Hour
	res, _ := credit(t, st, 200, "0xaaa", 30, now)
	if res.BonusDays != 7 || !res.ExpireAt.Equal(now.Add(37*day)) {
		t.Fatalf("buyer bonus %d expire %v", res.BonusDays, res.ExpireAt)
	}
	if res.ReferrerID != referrer || res.ReferrerDays != 7 || !res.ReferrerExpire.Equal(now.Add(7*day)) {
		t.Fatalf("referrer reward %+v", res)
	}

	res, _ = credit(t, st, 200, "0xbbb", 30, now)
	if res.ReferrerID != 0 || res.BonusDays != 0 {
		t.Fatal("renewal must not reward anyone again")
	}
	if !res.ExpireAt.Equal(now.Add(67 * day)) {
		t.Fatalf("renewal should stack, got %v", res.ExpireAt)
	}

	if _, isNew := credit(t, st, 200, "0xAAA", 30, now); isNew {
		t.Fatal("same transfer credited twice")
	}
	deps, err := st.Deposits(200, 10)
	if err != nil || len(deps) != 3 {
		t.Fatalf("deposits %d %v", len(deps), err)
	}

	stats, err := st.Referral(referrer)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Invited != 1 || stats.Paid != 1 || stats.DaysEarn != 7 {
		t.Fatalf("stats %+v", stats)
	}
	u, err := st.GetUser(referrer)
	if err != nil {
		t.Fatal(err)
	}
	if !u.ProActive(now) {
		t.Fatal("referrer should be Pro after reward")
	}
}

func TestRenewalStacksOnRemainingTime(t *testing.T) {
	st := openTest(t)
	mustUser(t, st, 300, nil)
	day := 24 * time.Hour
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	credit(t, st, 300, "0xm1", 30, start)
	later := start.Add(5 * day)
	res, _ := credit(t, st, 300, "0xm2", 30, later)
	if left := res.ExpireAt.Sub(later); left != 55*day {
		t.Fatalf("renewing with 25 days left should leave 55, got %v", left)
	}

	lapsed := res.ExpireAt.Add(10 * day)
	res, _ = credit(t, st, 300, "0xm3", 30, lapsed)
	if left := res.ExpireAt.Sub(lapsed); left != 30*day {
		t.Fatalf("renewing after expiry should start from now, got %v", left)
	}
}

func TestReferralTiers(t *testing.T) {
	for days, want := range map[int]int{30: 7, 90: 15, 365: 30} {
		if got := ReferralBonusDays(days); got != want {
			t.Fatalf("plan %d days: got %d want %d", days, got, want)
		}
	}
}

func TestReferralBonusByPlanAccumulates(t *testing.T) {
	st := openTest(t)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	referrer := int64(100)
	mustUser(t, st, referrer, nil)
	mustUser(t, st, 201, &referrer)
	mustUser(t, st, 202, &referrer)

	res, _ := credit(t, st, 201, "0xq", 90, now)
	if res.BonusDays != 15 || res.ReferrerDays != 15 {
		t.Fatalf("quarterly bonus %+v", res)
	}
	res, _ = credit(t, st, 202, "0xy", 365, now)
	if res.BonusDays != 30 || res.ReferrerDays != 30 {
		t.Fatalf("yearly bonus %+v", res)
	}
	if !res.ReferrerExpire.Equal(now.Add(45 * 24 * time.Hour)) {
		t.Fatalf("referrer rewards should add up, got %v", res.ReferrerExpire)
	}
	stats, err := st.Referral(referrer)
	if err != nil || stats.Paid != 2 || stats.DaysEarn != 45 {
		t.Fatalf("stats %+v %v", stats, err)
	}
}

func TestSelfAndUnknownReferrerIgnored(t *testing.T) {
	st := openTest(t)
	self := int64(1)
	mustUser(t, st, 1, &self)
	unknown := int64(999)
	mustUser(t, st, 2, &unknown)
	for _, id := range []int64{1, 2} {
		u, err := st.GetUser(id)
		if err != nil {
			t.Fatal(err)
		}
		if u.RefBy.Valid {
			t.Fatalf("user %d should have no referrer", id)
		}
	}
}

func TestLapsedProPausesExtraAddresses(t *testing.T) {
	st := openTest(t)
	mustUser(t, st, 7, nil)
	credit(t, st, 7, "0x1", 30, time.Now())
	addrs := []string{"0xA1", "0xA2", "0xA3", "0xA4", "0xA5"}
	for _, a := range addrs {
		if err := st.AddAddress(7, a, ProAddressLimit); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := st.db.Exec(`UPDATE users SET expire_at = ? WHERE telegram_id = 7`, time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	watches, err := st.ListAddresses(7)
	if err != nil {
		t.Fatal(err)
	}
	active := 0
	for i, w := range watches {
		if w.Active {
			active++
			if i >= FreeAddressLimit {
				t.Fatalf("address %d should be paused", i)
			}
		}
	}
	if active != FreeAddressLimit {
		t.Fatalf("active %d", active)
	}
}

func TestDepositAddressAllocation(t *testing.T) {
	st := openTest(t)
	derive := func(i uint32) (string, string, error) {
		return "0xABC" + string(rune('0'+i)), "T" + string(rune('0'+i)), nil
	}
	now := time.Now()
	a, err := st.DepositAddress(1, derive, now.Add(time.Hour))
	if err != nil || a.Index != 1 || a.EVM != "0xabc1" {
		t.Fatalf("first: %+v %v", a, err)
	}
	b, _ := st.DepositAddress(2, derive, now.Add(time.Hour))
	again, _ := st.DepositAddress(1, derive, now.Add(-time.Hour))
	if b.Index != 2 || again.Index != 1 {
		t.Fatalf("indexes %d %d", b.Index, again.Index)
	}
	owners, err := st.EVMDepositOwners()
	if err != nil || owners["0xabc2"] != 2 {
		t.Fatalf("owners %v %v", owners, err)
	}
	watched, _ := st.TronDepositOwners(now)
	if len(watched) != 2 {
		t.Fatalf("watch window must not shrink: %v", watched)
	}
	watched, _ = st.TronDepositOwners(now.Add(2 * time.Hour))
	if len(watched) != 0 {
		t.Fatalf("watch window expired: %v", watched)
	}
}

func TestLoginTokenSingleUse(t *testing.T) {
	st := openTest(t)
	if err := st.CreateLoginToken("tok", "1234", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ConsumeLoginToken("tok", "s1", time.Hour); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("unconfirmed token consumed: %v", err)
	}
	mustUser(t, st, 5, nil)
	if err := st.ConfirmLoginToken("tok", 5); err != nil {
		t.Fatal(err)
	}
	if err := st.ConfirmLoginToken("tok", 6); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("token confirmed twice: %v", err)
	}
	id, err := st.ConsumeLoginToken("tok", "s1", time.Hour)
	if err != nil || id != 5 {
		t.Fatalf("consume: %d %v", id, err)
	}
	if _, err := st.ConsumeLoginToken("tok", "s2", time.Hour); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("token consumed twice: %v", err)
	}
	if got, err := st.SessionUser("s1"); err != nil || got != 5 {
		t.Fatalf("session: %d %v", got, err)
	}
}

// Every address starts at the default alert line and is edited on its own.
func TestAddressThresholdPerAddress(t *testing.T) {
	st := openTest(t)
	mustUser(t, st, 400, nil)
	for _, a := range []string{"0xaaa", "0xbbb"} {
		if err := st.AddAddress(400, a, 3); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.SetAddressThreshold(400, "0xaaa", 1.5); err != nil {
		t.Fatal(err)
	}
	ws, err := st.ListAddresses(400)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, w := range ws {
		got[w.Address] = w.Threshold
	}
	if got["0xaaa"] != 1.5 || got["0xbbb"] != 1.2 {
		t.Fatalf("want 0xaaa=1.5 0xbbb=1.2, got %v", got)
	}
	if err := st.SetAddressThreshold(400, "0xccc", 1.4); err != ErrNotFound {
		t.Fatalf("unknown address should be ErrNotFound, got %v", err)
	}
}
