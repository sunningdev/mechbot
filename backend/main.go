package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/config"
	"mechbot/internal/depeg"
	"mechbot/internal/liqmap"
	"mechbot/internal/monitor"
	"mechbot/internal/pay"
	"mechbot/internal/price"
	"mechbot/internal/protocol"
	"mechbot/internal/renewal"
	"mechbot/internal/store"
	"mechbot/internal/telegram"
	"mechbot/internal/watchdog"
	"mechbot/internal/web"
)

func main() {
	log.SetFlags(log.LstdFlags)
	config.LoadDotEnv(".env")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置无效: %v", err)
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer st.Close()
	seed(st, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := telegram.New(cfg.BotToken)
	meCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	me, err := client.Me(meCtx)
	cancel()
	if err != nil {
		log.Fatalf("连接 Telegram 失败，检查 TELEGRAM_BOT_TOKEN: %v", err)
	}

	chains := protocol.NewRegistry(cfg.EthRPCURL, cfg.ArbRPCURL, cfg.BaseRPCURL, cfg.OpRPCURL)
	defer chains.Close()
	payments, err := pay.New(st, client, pay.Options{
		XpubEVM:  cfg.PayXpubEVM,
		XpubTron: cfg.PayXpubTron,
		Net:      cfg.Pay,
	})
	if err != nil {
		log.Fatalf("充值配置有误: %v", err)
	}
	defer payments.Close()
	// Polls run every few minutes and deposit scans every 30s, so both alert after roughly 15 / 5 minutes of failure.
	payments.SetReporter(watchdog.New(client, cfg.AdminChatID, 10))

	bot := telegram.NewBot(client, st, chains, cfg.SiteURL)
	prices := price.NewFeed()
	go prices.Run(ctx)
	poller := monitor.New(st, chains, client, watchdog.New(client, cfg.AdminChatID, 3), prices, cfg.Interval, config.AlertCooldown)
	pegs := depeg.NewFeed()
	go pegs.Run(ctx)
	poller.SetDepeg(pegs)
	bot.SetCheckNow(poller.CheckNow)
	liqMap := liqmap.New(st, cfg.LiqmapRPCURL, cfg.EtherscanKey)
	defer liqMap.Close()

	site := web.New(web.Options{
		Store:           st,
		Monitor:         poller,
		Pay:             payments,
		LiqMap:          liqMap,
		BotUsername:     me.Username,
		SiteURL:         cfg.SiteURL,
		WebDir:          cfg.WebDir,
		Secure:          cfg.CookieSecure,
		BaseCtx:         ctx,
		WalletConnectID: cfg.WalletConnectID,
	})
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           site.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second, // above the 45s deposit check
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("网页服务启动失败: %v", err)
		}
	}()

	log.Printf("MechBot 启动，机器人 @%s，网页接口 http://%s，轮询间隔 %s", me.Username, cfg.HTTPAddr, cfg.Interval)
	if cfg.AdminChatID == 0 {
		log.Printf("未设置 ADMIN_CHAT_ID，服务异常只写日志，不发 Telegram")
	}
	if payments.Enabled() {
		names := make([]string, 0, len(payments.Chains()))
		for _, c := range payments.Chains() {
			names = append(names, c.Name)
		}
		if payments.Testnet() {
			log.Printf("⚠️ 收款为测试网模式（NETWORK=testnet），只认测试币，数据库 %s；监控与预警仍是主网", cfg.DBPath)
		}
		log.Printf("USDT 充值已开启（每个用户专属地址）: %s", strings.Join(names, "、"))
	} else {
		log.Printf("PAY_XPUB_EVM 和 PAY_XPUB_TRON 都未配置，网页充值暂不开放")
	}

	if cfg.EtherscanKey == "" {
		log.Printf("未设置 ETHERSCAN_API_KEY，清算地图只统计 Morpho，Aave 部分暂不开放")
	}

	go payments.Run(ctx)
	go poller.Run(ctx)
	go liqMap.Run(ctx)
	go renewal.Run(ctx, st, client, cfg.SiteURL)
	bot.Run(ctx)

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = httpServer.Shutdown(shutdownCtx)
	log.Printf("已退出")
}

func seed(st *store.Store, cfg config.Config) {
	if !cfg.HasSeed {
		return
	}
	if !common.IsHexAddress(cfg.SeedAddress) {
		log.Printf("WATCH_ADDRESS 不是合法地址，跳过导入")
		return
	}
	address := common.HexToAddress(cfg.SeedAddress).Hex()
	if _, err := st.EnsureUser(cfg.SeedChatID, nil, cfg.SeedThreshold); err != nil {
		log.Printf("导入用户失败: %v", err)
		return
	}
	user, err := st.GetUser(cfg.SeedChatID)
	if err != nil {
		log.Printf("读取导入用户失败: %v", err)
		return
	}
	if user.AddressCount > 0 {
		return
	}
	if err := st.AddAddress(cfg.SeedChatID, address, user.AddressLimit(time.Now())); err != nil {
		log.Printf("导入地址失败: %v", err)
		return
	}
	if err := st.SetAddressThreshold(cfg.SeedChatID, address, cfg.SeedThreshold); err != nil {
		log.Printf("导入阈值失败: %v", err)
		return
	}
	log.Printf("已从 .env 导入 %s，阈值 %.2f", address, cfg.SeedThreshold)
}
