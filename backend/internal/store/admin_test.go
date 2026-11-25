package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHomeVisitStats(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	if err := st.RecordHomeVisit("1.1.1.1", now); err != nil {
		t.Fatal(err)
	}
	if err := st.RecordHomeVisit("1.1.1.1", now); err != nil {
		t.Fatal(err)
	}
	if err := st.RecordHomeVisit("2.2.2.2", now); err != nil {
		t.Fatal(err)
	}

	days, err := st.HomeVisitStats(now, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 3 {
		t.Fatalf("days %d", len(days))
	}
	today := days[2]
	if today.Day != "2026-11-20" || today.Visits != 3 || today.UniqueIPs != 2 {
		t.Fatalf("today %+v", today)
	}
}

func TestListAdminUsersPlans(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC()
	if _, err := st.EnsureUser(1, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	users, err := st.ListAdminUsers(now)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Plan != "trial" {
		t.Fatalf("want trial user, got %+v", users)
	}
}
