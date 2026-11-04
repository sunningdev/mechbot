package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultThreshold = 1.2
	DefaultInterval  = 5 * time.Minute
	AlertCooldown    = time.Hour
)

type Config struct {
	BotToken      string
	DBPath        string
	EthRPCURL     string
	ArbRPCURL     string
	OpRPCURL      string
	Interval      time.Duration
	SeedAddress   string
	SeedChatID    int64
	SeedThreshold float64
	HasSeed       bool
	HTTPAddr      string
	SiteURL       string
	WebDir        string
	CookieSecure  bool
	PayXpubEVM    string
	PayXpubTron   string
	Pay           PayNet
	AdminChatID   int64 // receives service-health alerts; falls back to TELEGRAM_CHAT_ID
	// LiqmapRPCURL feeds the liquidation-map scanner; kept separate so its bulk reads never slow alerts.
	LiqmapRPCURL string
	EtherscanKey string
}

// PayNet is where deposits are watched. NETWORK=testnet moves only payments to testnets; monitoring always reads mainnet.
type PayNet struct {
	Testnet bool
	EthRPC  string // comma-separated; evmrpc appends public backups for the chosen network
	BscRPC  string
	ArbRPC  string
	TronAPI string
	TronKey string
}

func LoadPayNet() (PayNet, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NETWORK"))) {
	case "", "mainnet":
		return PayNet{
			EthRPC:  firstNonEmpty(os.Getenv("ETH_RPC_URL"), "https://ethereum.publicnode.com"),
			BscRPC:  firstNonEmpty(os.Getenv("BSC_RPC_URL"), "https://bsc-rpc.publicnode.com"),
			ArbRPC:  firstNonEmpty(os.Getenv("ARB_RPC_URL"), "https://arb1.arbitrum.io/rpc"),
			TronAPI: firstNonEmpty(os.Getenv("TRON_API_URL"), "https://api.trongrid.io"),
			TronKey: strings.TrimSpace(os.Getenv("TRON_API_KEY")),
		}, nil
	case "testnet":
		// Separate variables so a mainnet node configured for monitoring is never used to watch test deposits.
		return PayNet{
			Testnet: true,
			EthRPC:  strings.TrimSpace(os.Getenv("TESTNET_ETH_RPC_URL")),
			BscRPC:  strings.TrimSpace(os.Getenv("TESTNET_BSC_RPC_URL")),
			ArbRPC:  strings.TrimSpace(os.Getenv("TESTNET_ARB_RPC_URL")),
			TronAPI: firstNonEmpty(os.Getenv("TESTNET_TRON_API_URL"), "https://nile.trongrid.io"),
		}, nil
	default:
		return PayNet{}, errString("NETWORK 只能是 mainnet 或 testnet")
	}
}

// DBPath keeps testnet payments in their own database so test Pro never mixes with real data.
func DBPath(testnet bool) string {
	if testnet {
		return firstNonEmpty(os.Getenv("DB_PATH"), "mechbot-testnet.db")
	}
	return firstNonEmpty(os.Getenv("DB_PATH"), "mechbot.db")
}

func Load() (Config, error) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return Config{}, errString("TELEGRAM_BOT_TOKEN 为空，在 .env 里填 @MechFiBot 的 token")
	}

	interval := DefaultInterval
	if raw := strings.TrimSpace(os.Getenv("POLL_INTERVAL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, errString("POLL_INTERVAL 必须是正的时间间隔，例如 5m")
		}
		interval = parsed
	}

	threshold := DefaultThreshold
	if raw := strings.TrimSpace(os.Getenv("ALERT_THRESHOLD")); raw != "" {
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil || parsed <= 1 {
			return Config{}, errString("ALERT_THRESHOLD 必须大于 1")
		}
		threshold = parsed
	}

	payNet, err := LoadPayNet()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		BotToken:      token,
		DBPath:        DBPath(payNet.Testnet),
		Pay:           payNet,
		EthRPCURL:     firstNonEmpty(os.Getenv("ETH_RPC_URL"), "https://ethereum.publicnode.com"),
		ArbRPCURL:     firstNonEmpty(os.Getenv("ARB_RPC_URL"), "https://arb1.arbitrum.io/rpc"),
		OpRPCURL:      firstNonEmpty(os.Getenv("OP_RPC_URL"), "https://optimism.publicnode.com"),
		Interval:      interval,
		SeedAddress:   strings.TrimSpace(os.Getenv("WATCH_ADDRESS")),
		SeedThreshold: threshold,
		HTTPAddr:      firstNonEmpty(os.Getenv("HTTP_ADDR"), "127.0.0.1:8080"),
		SiteURL:       strings.TrimRight(firstNonEmpty(os.Getenv("SITE_URL"), "http://localhost:5173"), "/"),
		WebDir:        firstNonEmpty(os.Getenv("WEB_DIR"), "../frontend/dist"),
		CookieSecure:  strings.EqualFold(strings.TrimSpace(os.Getenv("COOKIE_SECURE")), "true"),
		PayXpubEVM:    strings.TrimSpace(os.Getenv("PAY_XPUB_EVM")),
		PayXpubTron:   strings.TrimSpace(os.Getenv("PAY_XPUB_TRON")),
		LiqmapRPCURL:  strings.TrimSpace(os.Getenv("LIQMAP_ETH_RPC_URL")),
		EtherscanKey:  strings.TrimSpace(os.Getenv("ETHERSCAN_API_KEY")),
	}
	if raw := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id == 0 {
			return Config{}, errString("TELEGRAM_CHAT_ID 必须是数字")
		}
		cfg.SeedChatID = id
	}
	cfg.HasSeed = cfg.SeedAddress != "" && cfg.SeedChatID != 0
	cfg.AdminChatID = cfg.SeedChatID
	if raw := strings.TrimSpace(os.Getenv("ADMIN_CHAT_ID")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id == 0 {
			return Config{}, errString("ADMIN_CHAT_ID 必须是数字")
		}
		cfg.AdminChatID = id
	}
	return cfg, nil
}

func LoadDotEnv(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type errString string

func (e errString) Error() string { return string(e) }
