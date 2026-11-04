// mechwallet manages the HD wallet behind per-user deposit addresses. Run it on your own machine, never on the server.
//
//	mechwallet init                         generate a new 24-word mnemonic
//	mechwallet xpub                         print PAY_XPUB_EVM / PAY_XPUB_TRON for .env
//	mechwallet addresses [-db mechbot.db]   list every user's deposit addresses
//	mechwallet balances [-db mechbot.db]    list deposit addresses holding USDT
//	mechwallet sweep -chain bsc -to 0x… [-min 20] [-yes]
package main

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/term"

	"mechbot/internal/config"
	"mechbot/internal/pay"
	"mechbot/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	config.LoadDotEnv(".env")
	var err error
	if payNet, err = config.LoadPayNet(); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
	if payNet.Testnet {
		fmt.Println("⚠️ 测试网模式（NETWORK=testnet）：BNB Chain Testnet / Arbitrum Sepolia / Sepolia / TRON Nile")
	}
	switch os.Args[1] {
	case "init":
		err = cmdInit()
	case "xpub":
		err = cmdXpub()
	case "addresses":
		err = cmdAddresses(os.Args[2:])
	case "balances":
		err = cmdBalances(os.Args[2:])
	case "sweep":
		err = cmdSweep(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `用法:
  mechwallet init                              生成新的 24 词助记词
  mechwallet xpub                              输入助记词，输出 .env 用的 PAY_XPUB_EVM / PAY_XPUB_TRON
  mechwallet addresses [-db mechbot.db]        列出所有用户的充值地址（不需要助记词）
  mechwallet balances [-db mechbot.db]         查看各充值地址的 USDT 余额
  mechwallet sweep -chain <tron|bsc|arbitrum|ethereum> -to <归集地址> [-min 20] [-db mechbot.db] [-yes]`)
	os.Exit(2)
}

func cmdInit() error {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return err
	}
	fmt.Println("新助记词（抄在纸上离线保存，不要截图、不要存云端、不要放到服务器）：")
	fmt.Println()
	fmt.Println("  " + mnemonic)
	fmt.Println()
	fmt.Println("下一步运行 mechwallet xpub，把输出的两行写进服务器的 .env。")
	return nil
}

func cmdXpub() error {
	keys, err := promptKeys()
	if err != nil {
		return err
	}
	evmPub, err := keys.evm.Neuter()
	if err != nil {
		return err
	}
	tronPub, err := keys.tron.Neuter()
	if err != nil {
		return err
	}
	fmt.Println("写进服务器 .env（xpub 只能看地址，不能动钱）：")
	fmt.Println()
	fmt.Println("PAY_XPUB_EVM=" + evmPub.String())
	fmt.Println("PAY_XPUB_TRON=" + tronPub.String())
	fmt.Println()
	for i := uint32(0); i < 2; i++ {
		evm, _ := pay.DeriveAddress(evmPub, i)
		tron, _ := pay.DeriveAddress(tronPub, i)
		label := "第一个用户"
		if i == 0 {
			label = "手续费钱包，归集时用来给充值地址打 gas"
		}
		fmt.Printf("index %d（%s）\n  EVM  %s\n  TRON %s\n", i, label, evm.Hex(), pay.EncodeTronAddress(tron))
	}
	fmt.Println()
	fmt.Println("核对：把助记词导入 MetaMask / TronLink，第一个账户应等于上面 index 0 的地址。")
	return nil
}

// hdKeys holds the private change-level keys m/44'/60'/0'/0 and m/44'/195'/0'/0.
type hdKeys struct {
	evm  *hdkeychain.ExtendedKey
	tron *hdkeychain.ExtendedKey
}

func promptKeys() (hdKeys, error) {
	mnemonic, err := readMnemonic()
	if err != nil {
		return hdKeys{}, err
	}
	master, err := hdkeychain.NewMaster(bip39.NewSeed(mnemonic, ""), &chaincfg.MainNetParams)
	if err != nil {
		return hdKeys{}, err
	}
	evm, err := changeKey(master, pay.CoinEVM)
	if err != nil {
		return hdKeys{}, err
	}
	tron, err := changeKey(master, pay.CoinTron)
	if err != nil {
		return hdKeys{}, err
	}
	return hdKeys{evm: evm, tron: tron}, nil
}

// readMnemonic never echoes the words, so instead it reports which position is wrong and lets the user retry.
func readMnemonic() (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		words := strings.Fields(strings.ToLower(line))
		if msg := checkMnemonic(words); msg != "" {
			return "", errors.New(msg)
		}
		return strings.Join(words, " "), nil
	}
	for try := 0; try < 3; try++ {
		fmt.Fprint(os.Stderr, "输入助记词（不显示）。整串粘贴后回车；想一个一个输入，直接按回车: ")
		raw, err := readHidden()
		if err != nil {
			return "", err
		}
		words := strings.Fields(strings.ToLower(raw))
		if len(words) == 0 {
			if words, err = readWordByWord(); err != nil {
				return "", err
			}
		}
		if n := len(words); n != 12 && n != 24 {
			fmt.Fprintf(os.Stderr, "✗ 输入了 %d 个词，应该是 24 个（或 12 个）\n", n)
			continue
		}
		ok, err := fixMnemonic(words)
		if err != nil {
			return "", err
		}
		if ok {
			fmt.Fprintf(os.Stderr, "✓ %d 个词，校验通过\n", len(words))
			return strings.Join(words, " "), nil
		}
	}
	return "", errors.New("助记词三次都不对，已退出")
}

// fixMnemonic lets the user retype single positions in place; false means they chose to start over.
func fixMnemonic(words []string) (bool, error) {
	for {
		for i, w := range words {
			if _, ok := bip39.GetWordIndex(w); ok {
				continue
			}
			fmt.Fprintf(os.Stderr, "✗ 第 %d 个词拼写不对\n", i+1)
			w, err := askWord(i, len(words))
			if err != nil {
				return false, err
			}
			words[i] = w
		}
		if bip39.IsMnemonicValid(strings.Join(words, " ")) {
			return true, nil
		}
		fmt.Fprintln(os.Stderr, "✗ 每个词都拼对了，但校验不通过：某个词抄错成了另一个合法的词，或者顺序错了。对照纸条找出来改。")
		fmt.Fprintf(os.Stderr, "  要改第几个（1-%d）？直接回车放弃，全部重输: ", len(words))
		raw, err := readHidden()
		if err != nil {
			return false, err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return false, nil
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > len(words) {
			fmt.Fprintf(os.Stderr, "  %q 不是 1-%d 的序号\n", raw, len(words))
			continue
		}
		w, err := askWord(n-1, len(words))
		if err != nil {
			return false, err
		}
		words[n-1] = w
	}
}

// askWord reads one hidden word for position i and repeats until it is on the BIP39 list.
func askWord(i, total int) (string, error) {
	for {
		fmt.Fprintf(os.Stderr, "  第 %2d/%d 个词: ", i+1, total)
		raw, err := readHidden()
		if err != nil {
			return "", err
		}
		w := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := bip39.GetWordIndex(w); ok {
			return w, nil
		}
		fmt.Fprintln(os.Stderr, "  ✗ 拼写不对，重输这一个")
	}
}

func readHidden() (string, error) {
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	return string(b), err
}

// readWordByWord checks each word against the BIP39 list as it is typed, so a typo costs one word, not all 24.
func readWordByWord() ([]string, error) {
	fmt.Fprintln(os.Stderr, "逐个输入：每个词输完回车（不显示）。拼错会让你重输这一个；输 - 撤回上一个；只有 12 个词时，第 13 个直接回车。")
	var words []string
	for len(words) < 24 {
		fmt.Fprintf(os.Stderr, "  第 %2d/24 个词: ", len(words)+1)
		raw, err := readHidden()
		if err != nil {
			return nil, err
		}
		w := strings.ToLower(strings.TrimSpace(raw))
		switch {
		case w == "" && len(words) == 12:
			return words, nil
		case w == "":
			continue
		case w == "-":
			if len(words) > 0 {
				words = words[:len(words)-1]
				fmt.Fprintf(os.Stderr, "  已撤回第 %d 个\n", len(words)+1)
			}
			continue
		}
		if _, ok := bip39.GetWordIndex(w); !ok {
			fmt.Fprintln(os.Stderr, "  ✗ 拼写不对，重输这一个")
			continue
		}
		words = append(words, w)
	}
	return words, nil
}

func checkMnemonic(words []string) string {
	if n := len(words); n != 12 && n != 24 {
		return fmt.Sprintf("输入了 %d 个词，应该是 24 个（或 12 个）", n)
	}
	var bad []string
	for i, w := range words {
		if _, ok := bip39.GetWordIndex(w); !ok {
			bad = append(bad, strconv.Itoa(i+1))
		}
	}
	if len(bad) > 0 {
		return "第 " + strings.Join(bad, "、") + " 个词拼写不对"
	}
	if !bip39.IsMnemonicValid(strings.Join(words, " ")) {
		return "每个词都拼对了，但校验不通过：可能顺序错了，或某个词写成了另一个合法的词"
	}
	return ""
}

func changeKey(master *hdkeychain.ExtendedKey, coin uint32) (*hdkeychain.ExtendedKey, error) {
	key := master
	var err error
	for _, step := range []uint32{hdkeychain.HardenedKeyStart + 44, hdkeychain.HardenedKeyStart + coin, hdkeychain.HardenedKeyStart, 0} {
		if key, err = key.Derive(step); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func childPrivate(parent *hdkeychain.ExtendedKey, index uint32) (*ecdsa.PrivateKey, error) {
	child, err := parent.Derive(index)
	if err != nil {
		return nil, err
	}
	priv, err := child.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(priv.Serialize())
}

func openDB(path string) (*store.Store, []store.DepositAddress, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil, fmt.Errorf("找不到数据库 %s（用 -db 指定，或在 backend 目录运行）", path)
	}
	st, err := store.Open(path)
	if err != nil {
		return nil, nil, err
	}
	rows, err := st.AllDepositAddresses()
	if err != nil {
		st.Close()
		return nil, nil, err
	}
	return st, rows, nil
}

// payNet follows NETWORK in .env, so sweeping reads the same chains the server watches.
var payNet config.PayNet

func dbFlag(fs *flag.FlagSet) *string {
	return fs.String("db", config.DBPath(payNet.Testnet), "数据库路径")
}

func chainByKey(key string) (pay.Chain, error) {
	if key == "tron" {
		return pay.TronChain(payNet.Testnet), nil
	}
	for _, c := range pay.EVMChains(payNet) {
		if c.Key == key {
			return c, nil
		}
	}
	return pay.Chain{}, fmt.Errorf("不支持的链 %q，可选 tron、bsc、arbitrum、ethereum", key)
}
