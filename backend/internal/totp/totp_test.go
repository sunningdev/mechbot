package totp

import (
	"testing"
	"time"
)

func TestValidateKnownVector(t *testing.T) {
	// RFC 6238 appendix B uses a different key length; use a fixed secret and round-trip.
	secret := "JBSWY3DPEHPK3PXP" // "Hello!"
	now := time.Unix(1_111_111_111, 0)
	code, err := Code(secret, now)
	if err != nil {
		t.Fatal(err)
	}
	if !Validate(secret, code, now, 0) {
		t.Fatalf("expected valid code %s", code)
	}
	if Validate(secret, "000000", now, 0) {
		t.Fatal("expected invalid")
	}
	if !Validate(secret, code, now.Add(20*time.Second), 1) {
		t.Fatal("expected skew window to accept")
	}
}
