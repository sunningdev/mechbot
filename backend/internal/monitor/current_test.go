package monitor

import (
	"path/filepath"
	"testing"

	"mechbot/internal/store"
)

// A poll round lists watches before it reaches each address; alerts must use what the user set since then.
func TestCurrentRereadsWatches(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	const addr = "0xc936e848688c9f035fa0e7a0e4dbcf26a01245f3"
	for _, id := range []int64{1, 2} {
		if _, err := st.EnsureUser(id, nil, 1.2); err != nil {
			t.Fatal(err)
		}
		if err := st.AddAddress(id, addr, 20); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := st.ListAll()
	if err != nil || len(snapshot) != 2 {
		t.Fatalf("snapshot %v %v", snapshot, err)
	}
	m := &Monitor{store: st}

	if err := st.SetAddressThreshold(1, addr, 1.35); err != nil {
		t.Fatal(err)
	}
	if err := st.RemoveAddress(2, addr); err != nil {
		t.Fatal(err)
	}
	got := m.current(addr, snapshot)
	if len(got) != 1 || got[0].TelegramID != 1 || got[0].Threshold != 1.35 {
		t.Fatalf("got %+v, want user 1 at 1.35 only", got)
	}
	if got := m.current(addr, nil); len(got) != 0 {
		t.Fatalf("no watches given, got %+v", got)
	}
}
