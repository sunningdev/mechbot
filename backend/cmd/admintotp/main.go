// Generate an admin TOTP secret for Google Authenticator.
// Usage: go run ./cmd/admintotp
package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"os"

	"mechbot/internal/totp"
)

const account = "sunzhenning"

func main() {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		fmt.Fprintf(os.Stderr, "rand: %v\n", err)
		os.Exit(1)
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
	url, err := totp.AuthURL("MechBot", account, secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "url: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Add to Google Authenticator (scan QR from any otpauth QR tool, or enter secret manually):")
	fmt.Println()
	fmt.Println("  username: " + account + " (fixed)")
	fmt.Println("  ADMIN_TOTP_SECRET=" + secret)
	fmt.Println()
	fmt.Println(url)
}
