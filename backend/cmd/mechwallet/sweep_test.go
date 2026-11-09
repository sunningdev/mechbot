package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"

	"mechbot/internal/config"
	"mechbot/internal/pay"
	"mechbot/internal/store"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestCheckMnemonic(t *testing.T) {
	words := strings.Fields(testMnemonic)
	if msg := checkMnemonic(words); msg != "" {
		t.Fatalf("valid phrase rejected: %s", msg)
	}
	typo := append([]string{}, words...)
	typo[2], typo[6] = "abandn", "abnadon"
	if msg := checkMnemonic(typo); !strings.Contains(msg, "第 3、7 个词") {
		t.Fatalf("typo positions: %s", msg)
	}
	swapped := append([]string{}, words...)
	swapped[0], swapped[11] = swapped[11], swapped[0]
	if msg := checkMnemonic(swapped); !strings.Contains(msg, "校验不通过") {
		t.Fatalf("checksum: %s", msg)
	}
	if msg := checkMnemonic(words[:11]); !strings.Contains(msg, "11 个词") {
		t.Fatalf("count: %s", msg)
	}
}

func testKeys(t *testing.T) hdKeys {
	t.Helper()
	return keysFor(t, testMnemonic)
}

func keysFor(t *testing.T, mnemonic string) hdKeys {
	t.Helper()
	master, err := hdkeychain.NewMaster(bip39.NewSeed(mnemonic, ""), &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	evm, err := changeKey(master, pay.CoinEVM)
	if err != nil {
		t.Fatal(err)
	}
	tron, err := changeKey(master, pay.CoinTron)
	if err != nil {
		t.Fatal(err)
	}
	return hdKeys{evm: evm, tron: tron}
}

// Private derivation in the tool must land on the same addresses the server derives from the xpub.
func TestToolMatchesServerDerivation(t *testing.T) {
	keys := testKeys(t)
	for _, k := range []*hdkeychain.ExtendedKey{keys.evm, keys.tron} {
		pub, _ := k.Neuter()
		for i := uint32(0); i < 3; i++ {
			want, _ := pay.DeriveAddress(pub, i)
			priv, err := childPrivate(k, i)
			if err != nil {
				t.Fatal(err)
			}
			if got := pay.EncodeTronAddress(want); got != tronAddress(priv) {
				t.Fatalf("index %d: tool %s server %s", i, tronAddress(priv), got)
			}
		}
	}
}

// Dry run on mainnet: balance lookup and gas estimation run, nothing is sent.
func TestLiveSweepDryRun(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("设置 MECHBOT_LIVE=1 才访问主网")
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "w.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	// The public test phrase's addresses are blacklisted by Circle, which reverts USDC gas estimates.
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	keys := keysFor(t, mnemonic)
	evmPub, _ := keys.evm.Neuter()
	tronPub, _ := keys.tron.Neuter()
	derive := func(i uint32) (string, string, error) {
		e, _ := pay.DeriveAddress(evmPub, i)
		tr, _ := pay.DeriveAddress(tronPub, i)
		return e.Hex(), pay.EncodeTronAddress(tr), nil
	}
	if _, err := st.DepositAddress(1, derive, time.Now()); err != nil {
		t.Fatal(err)
	}
	rows, _ := st.AllDepositAddresses()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	payNet = config.PayNet{TronAPI: "https://api.trongrid.io", TronKey: os.Getenv("TRON_API_KEY")}
	for _, c := range pay.EVMChains(payNet) {
		if err := sweepEVM(ctx, c, keys, rows, common.HexToAddress(rows[0].EVM), 0, false); err != nil {
			t.Fatalf("%s: %v", c.Key, err)
		}
		if c.Key != "base" && c.Key != "optimism" {
			continue
		}
		client, err := dialEVM(ctx, c)
		if err != nil {
			t.Fatal(err)
		}
		fee, err := l1Fee(ctx, c, client, make([]byte, 68))
		client.Close()
		if err != nil || fee.Sign() <= 0 {
			t.Fatalf("%s l1 fee %v err %v", c.Key, fee, err)
		}
		t.Logf("%s l1 fee %s wei", c.Key, fee)
	}
	if payNet.TronKey == "" {
		t.Log("TRON_API_KEY not set: TronGrid allows 1 request/s without it, skipping TRON")
		return
	}
	tron, _ := chainByKey("tron")
	if err := sweepTron(ctx, tron, keys, rows, pay.TronUSDT, 0, 30_000_000, false); err != nil {
		t.Fatal(err)
	}
}
