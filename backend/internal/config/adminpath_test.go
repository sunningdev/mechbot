package config

import "testing"

func TestSanitizeAdminPath(t *testing.T) {
	if got := sanitizeAdminPath("ops-k7m2n9q4x8p3w6"); got != "ops-k7m2n9q4x8p3w6" {
		t.Fatalf("got %q", got)
	}
	if sanitizeAdminPath("/ops-k7m2n9q4x8p3w6/") != "ops-k7m2n9q4x8p3w6" {
		t.Fatal("trim slashes")
	}
	for _, bad := range []string{"admin", "app", "short", "has/slash", "bad path!", ""} {
		if sanitizeAdminPath(bad) != "" {
			t.Fatalf("expected reject %q", bad)
		}
	}
}
