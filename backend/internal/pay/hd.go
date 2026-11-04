package pay

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	CoinEVM  = 60
	CoinTron = 195
	// xpubs are exported at m/44'/coin'/0'/0, so deposit index i lives at depth 5.
	xpubDepth = 4
)

// ParseXpub accepts only a public extended key at the change level; a private key is refused so it never lands on the server.
func ParseXpub(s string) (*hdkeychain.ExtendedKey, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	key, err := hdkeychain.NewKeyFromString(s)
	if err != nil {
		return nil, fmt.Errorf("xpub 格式不对: %w", err)
	}
	if key.IsPrivate() {
		return nil, errors.New("这是私钥（xprv），服务器只能放 xpub")
	}
	if key.Depth() != xpubDepth {
		return nil, fmt.Errorf("xpub 层级是 %d，需要 m/44'/币种'/0'/0 这一层（4），用 mechwallet xpub 导出", key.Depth())
	}
	return key, nil
}

// DeriveAddress returns the account at child index i; EVM and TRON share the same key-to-address hash.
func DeriveAddress(xpub *hdkeychain.ExtendedKey, i uint32) (common.Address, error) {
	child, err := xpub.Derive(i)
	if err != nil {
		return common.Address{}, err
	}
	pub, err := child.ECPubKey()
	if err != nil {
		return common.Address{}, err
	}
	ecdsa, err := crypto.UnmarshalPubkey(pub.SerializeUncompressed())
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*ecdsa), nil
}

func EncodeTronAddress(a common.Address) string {
	payload := append([]byte{0x41}, a.Bytes()...)
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	raw := append(payload, second[:4]...)

	n := new(big.Int).SetBytes(raw)
	mod := new(big.Int)
	base := big.NewInt(58)
	var out []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for _, b := range raw {
		if b != 0 {
			break
		}
		out = append(out, base58Alphabet[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}
