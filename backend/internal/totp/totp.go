package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	Digits = 6
	Period = 30
)

// NormalizeSecret accepts spaced or unpadded base32 secrets from authenticator apps.
func NormalizeSecret(secret string) (string, error) {
	s := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	if s == "" {
		return "", fmt.Errorf("empty totp secret")
	}
	if _, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s); err != nil {
		if _, err2 := base32.StdEncoding.DecodeString(s); err2 != nil {
			return "", fmt.Errorf("invalid base32 totp secret")
		}
	}
	return s, nil
}

func decodeSecret(secret string) ([]byte, error) {
	s, err := NormalizeSecret(secret)
	if err != nil {
		return nil, err
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		return base32.StdEncoding.DecodeString(s)
	}
	return key, nil
}

func hotp(key []byte, counter uint64, digits int) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	mod := 1
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", digits, code%mod)
}

// Code returns the TOTP for secret at time t.
func Code(secret string, t time.Time) (string, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix() / Period)
	return hotp(key, counter, Digits), nil
}

// Validate checks code against secret at time t, allowing skew steps (±window).
func Validate(secret, code string, t time.Time, window int) bool {
	code = strings.TrimSpace(code)
	if len(code) != Digits {
		return false
	}
	key, err := decodeSecret(secret)
	if err != nil {
		return false
	}
	if window < 0 {
		window = 0
	}
	counter := t.Unix() / Period
	for i := -window; i <= window; i++ {
		want := hotp(key, uint64(counter+int64(i)), Digits)
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// AuthURL builds an otpauth:// URI for Google Authenticator.
func AuthURL(issuer, account, secret string) (string, error) {
	s, err := NormalizeSecret(secret)
	if err != nil {
		return "", err
	}
	issuer = strings.TrimSpace(issuer)
	account = strings.TrimSpace(account)
	if issuer == "" {
		issuer = "MechBot"
	}
	if account == "" {
		account = "admin"
	}
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuer, account, s, issuer, Digits, Period,
	), nil
}
