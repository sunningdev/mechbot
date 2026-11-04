package store

import (
	"testing"
	"time"
)

func TestNewUserGetsTrialOnce(t *testing.T) {
	st := openTest(t)
	before := time.Now()
	if created, err := st.EnsureUser(1, nil, 1.2); err != nil || !created {
		t.Fatalf("created %v %v", created, err)
	}
	u, err := st.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Trial || !u.ProActive(before) || u.AddressLimit(before) != ProAddressLimit {
		t.Fatalf("new user should be on a Pro trial: %+v", u)
	}
	if left := u.ExpireAt.Sub(before); left < TrialDays*24*time.Hour-time.Minute || left > TrialDays*24*time.Hour+time.Minute {
		t.Fatalf("trial length %v", left)
	}

	// Seeing the user again (e.g. /start twice) must not restart the trial.
	if _, err := st.db.Exec(`UPDATE users SET expire_at = ? WHERE telegram_id = 1`, before.Add(-time.Hour).UTC().Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if created, err := st.EnsureUser(1, nil, 1.2); err != nil || created {
		t.Fatalf("existing user re-created: %v %v", created, err)
	}
	u, _ = st.GetUser(1)
	if u.ProActive(before) {
		t.Fatal("trial must not be granted twice")
	}
	if u.AddressLimit(before) != FreeAddressLimit {
		t.Fatalf("lapsed trial should fall back to Starter, limit %d", u.AddressLimit(before))
	}
}

func TestBuyingDuringTrialKeepsTrialDaysAndEndsTrial(t *testing.T) {
	st := openTest(t)
	if _, err := st.EnsureUser(2, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	u, _ := st.GetUser(2)
	now := time.Now()
	res, _ := credit(t, st, 2, "0xbuy", 30, now)
	if want := u.ExpireAt.Add(30 * 24 * time.Hour); !res.ExpireAt.Equal(want.UTC().Truncate(time.Second)) {
		t.Fatalf("purchase should stack on the trial: got %v want %v", res.ExpireAt, want)
	}
	u, _ = st.GetUser(2)
	if u.Trial {
		t.Fatal("paying ends the trial")
	}
	// The trial is not a purchase, so the first real payment still counts as first for referral purposes.
	if res.BonusDays != 0 {
		t.Fatalf("no referrer, no bonus: %+v", res)
	}
}

func TestTrialFirstPurchaseStillRewardsReferrer(t *testing.T) {
	st := openTest(t)
	ref := int64(10)
	if _, err := st.EnsureUser(ref, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	if _, err := st.EnsureUser(11, &ref, 1.2); err != nil {
		t.Fatal(err)
	}
	res, _ := credit(t, st, 11, "0xfirst", 30, time.Now())
	if res.ReferrerID != ref || res.BonusDays != 7 {
		t.Fatalf("trial must not count as the first purchase: %+v", res)
	}
}

func TestProNoticesStepForwardOnce(t *testing.T) {
	st := openTest(t)
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	set := func(id int64, expire time.Time, trial int) {
		t.Helper()
		if _, err := st.EnsureUser(id, nil, 1.2); err != nil {
			t.Fatal(err)
		}
		if _, err := st.db.Exec(`UPDATE users SET expire_at = ?, trial = ?, pro_notice = 0 WHERE telegram_id = ?`,
			expire.Format(time.RFC3339), trial, id); err != nil {
			t.Fatal(err)
		}
	}
	set(1, now.Add(48*time.Hour), 1)     // not due yet
	set(2, now.Add(5*time.Hour), 1)      // ending soon
	set(3, now.Add(-time.Hour), 0)       // just ended
	set(4, now.Add(-10*24*time.Hour), 0) // ended long ago

	due := func() map[int64]ProNotice {
		t.Helper()
		list, err := st.ProNoticesDue(now, 24*time.Hour, 72*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		out := map[int64]ProNotice{}
		for _, n := range list {
			out[n.TelegramID] = n
		}
		return out
	}
	got := due()
	if _, ok := got[1]; ok {
		t.Fatal("user 1 is not due")
	}
	if n := got[2]; n.Stage != NoticeEndingSoon || !n.Trial {
		t.Fatalf("user 2: %+v", n)
	}
	if n := got[3]; n.Stage != NoticeEnded || n.Stale {
		t.Fatalf("user 3: %+v", n)
	}
	if n := got[4]; n.Stage != NoticeEnded || !n.Stale {
		t.Fatalf("user 4: %+v", n)
	}
	for _, n := range got {
		if err := st.SetProNotice(n); err != nil {
			t.Fatal(err)
		}
	}
	got = due()
	if len(got) != 0 {
		t.Fatalf("notices must not repeat: %+v", got)
	}

	// When the period actually ends, user 2 gets the second notice.
	later := now.Add(6 * time.Hour)
	list, _ := st.ProNoticesDue(later, 24*time.Hour, 72*time.Hour)
	if len(list) != 1 || list[0].TelegramID != 2 || list[0].Stage != NoticeEnded {
		t.Fatalf("ended notice: %+v", list)
	}

	// Renewing starts a new period with fresh notices.
	credit(t, st, 3, "0xrenew", 30, now)
	list, _ = st.ProNoticesDue(now.Add(30*24*time.Hour-12*time.Hour), 24*time.Hour, 72*time.Hour)
	found := false
	for _, n := range list {
		if n.TelegramID == 3 && n.Stage == NoticeEndingSoon && !n.Trial {
			found = true
		}
	}
	if !found {
		t.Fatalf("renewed user should be reminded again: %+v", list)
	}
}
