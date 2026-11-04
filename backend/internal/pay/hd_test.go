package pay

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func testXpub(t *testing.T, coin uint32) *hdkeychain.ExtendedKey {
	t.Helper()
	key, err := hdkeychain.NewMaster(bip39.NewSeed(testMnemonic, ""), &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []uint32{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart + coin, hdkeychain.HardenedKeyStart, 0} {
		if key, err = key.Derive(step); err != nil {
			t.Fatal(err)
		}
	}
	pub, err := key.Neuter()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseXpub(pub.String())
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestDeriveMatchesWallets(t *testing.T) {
	evm, err := DeriveAddress(testXpub(t, CoinEVM), 0)
	if err != nil {
		t.Fatal(err)
	}
	if evm.Hex() != "0x9858EfFD232B4033E47d90003D41EC34EcaEda94" {
		t.Fatalf("MetaMask index 0: %s", evm.Hex())
	}
	tron, err := DeriveAddress(testXpub(t, CoinTron), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := EncodeTronAddress(tron); got != "TUEZSdKsoDHQMeZwihtdoBiN46zxhGWYdH" {
		t.Fatalf("TronLink index 0: %s", got)
	}
}

func TestTronAddressRoundTrip(t *testing.T) {
	body, ok := DecodeTronAddress(TronUSDT)
	if !ok || EncodeTronAddress(body) != TronUSDT {
		t.Fatalf("round trip: %s", EncodeTronAddress(body))
	}
}

func TestParseXpubRejects(t *testing.T) {
	master, err := hdkeychain.NewMaster(bip39.NewSeed(testMnemonic, ""), &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseXpub(master.String()); err == nil {
		t.Fatal("accepted xprv")
	}
	pub, _ := master.Neuter()
	if _, err := ParseXpub(pub.String()); err == nil {
		t.Fatal("accepted master xpub at depth 0")
	}
	if k, err := ParseXpub(""); k != nil || err != nil {
		t.Fatal("empty should be disabled, not an error")
	}
}
