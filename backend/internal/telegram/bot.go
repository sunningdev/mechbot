package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/config"
	"mechbot/internal/i18n"
	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const (
	loginPrefix = "login_"
	langData    = "lang:"
)

type Bot struct {
	client  *Client
	store   *store.Store
	chains  *protocol.Registry
	siteURL string
	// checkNow, when set, alerts right away if a newly added address is already below its line.
	checkNow func(ctx context.Context, address string, telegramID int64)
}

func NewBot(client *Client, st *store.Store, chains *protocol.Registry, siteURL string) *Bot {
	return &Bot{client: client, store: st, chains: chains, siteURL: siteURL}
}

func (b *Bot) SetCheckNow(f func(ctx context.Context, address string, telegramID int64)) {
	b.checkNow = f
}

// lang remembers the Telegram client language and returns the language to answer this user in.
func (b *Bot) lang(from User) i18n.Lang {
	if from.LanguageCode != "" {
		_ = b.store.SetTelegramLang(from.ID, from.LanguageCode)
	}
	return b.store.UserLang(from.ID, from.LanguageCode)
}

func (b *Bot) helpText(t *i18n.Text) string {
	text := t.Help
	if b.siteURL != "" {
		text += fmt.Sprintf(t.HelpConsole, b.siteURL)
	}
	return text
}

func (b *Bot) Run(ctx context.Context) {
	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := b.client.Updates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("读取 Telegram 更新失败: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			b.process(ctx, update)
		}
	}
}

// process keeps one bad message from crashing the bot for everyone.
func (b *Bot) process(ctx context.Context, update Update) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("处理 Telegram 消息崩溃已拦截 update=%d: %v\n%s", update.UpdateID, p, debug.Stack())
		}
	}()
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}
	b.handle(ctx, update)
}

func (b *Bot) handle(ctx context.Context, update Update) {
	msg := update.Message
	if msg == nil || msg.From == nil || msg.From.IsBot {
		return
	}
	cmd, args := parseCommand(msg.Text)
	if cmd == "" {
		return
	}
	t := i18n.T(b.lang(*msg.From))
	switch {
	case cmd == "/start" && strings.HasPrefix(args, loginPrefix):
		b.askLogin(ctx, msg, t, strings.TrimPrefix(args, loginPrefix))
		return
	case cmd == "/lang":
		b.askLang(ctx, msg.Chat.ID, t)
		return
	}
	reply := b.dispatch(ctx, *msg.From, t, cmd, args)
	if reply == "" {
		return
	}
	if err := b.client.Send(ctx, msg.Chat.ID, reply); err != nil {
		log.Printf("回复 Telegram 失败: %v", err)
	}
}

func (b *Bot) askLogin(ctx context.Context, msg *Message, t *i18n.Text, token string) {
	if msg.Chat.Type != "" && msg.Chat.Type != "private" {
		_ = b.client.Send(ctx, msg.Chat.ID, t.LoginPrivate)
		return
	}
	login, err := b.store.LoginToken(token)
	if err != nil || login.Status != "pending" {
		_ = b.client.Send(ctx, msg.Chat.ID, t.LoginInvalid)
		return
	}
	text := fmt.Sprintf(t.LoginPrompt, login.Code)
	if err := b.client.SendButtons(ctx, msg.Chat.ID, text, Button{Text: t.LoginButton, Data: "login:" + token}); err != nil {
		log.Printf("发送登录确认失败: %v", err)
	}
}

// askLang shows every language in its own name, two per row.
func (b *Bot) askLang(ctx context.Context, chatID int64, t *i18n.Text) {
	var rows [][]Button
	for i, l := range i18n.All {
		if i%2 == 0 {
			rows = append(rows, nil)
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], Button{Text: l.Name(), Data: langData + string(l)})
	}
	if err := b.client.SendKeyboard(ctx, chatID, t.LangPrompt, rows); err != nil {
		log.Printf("发送语言选择失败: %v", err)
	}
}

func (b *Bot) handleCallback(ctx context.Context, cb *CallbackQuery) {
	if cb.From.IsBot {
		_ = b.client.AnswerCallback(ctx, cb.ID, "")
		return
	}
	switch {
	case strings.HasPrefix(cb.Data, "login:"):
		b.confirmLogin(ctx, cb, i18n.T(b.lang(cb.From)), strings.TrimPrefix(cb.Data, "login:"))
	case strings.HasPrefix(cb.Data, langData):
		b.pickLang(ctx, cb, strings.TrimPrefix(cb.Data, langData))
	default:
		_ = b.client.AnswerCallback(ctx, cb.ID, "")
	}
}

func (b *Bot) pickLang(ctx context.Context, cb *CallbackQuery, code string) {
	l, ok := i18n.Parse(code)
	if !ok {
		_ = b.client.AnswerCallback(ctx, cb.ID, "")
		return
	}
	_, err := b.store.EnsureUser(cb.From.ID, nil, config.DefaultThreshold)
	if err == nil {
		err = b.store.SetLang(cb.From.ID, l)
	}
	if err != nil {
		log.Printf("保存语言失败: %v", err)
		_ = b.client.AnswerCallback(ctx, cb.ID, i18n.T(b.lang(cb.From)).QueryFailed)
		return
	}
	_ = b.client.AnswerCallback(ctx, cb.ID, "")
	b.editCallback(ctx, cb, i18n.T(l).LangSet)
}

func (b *Bot) confirmLogin(ctx context.Context, cb *CallbackQuery, t *i18n.Text, token string) {
	login, err := b.store.LoginToken(token)
	if err != nil {
		_ = b.client.AnswerCallback(ctx, cb.ID, t.LoginExpiredToast)
		b.editCallback(ctx, cb, t.LoginExpired)
		return
	}
	created, err := b.store.EnsureUser(cb.From.ID, login.RefBy, config.DefaultThreshold)
	if err != nil {
		log.Printf("登录注册用户失败: %v", err)
		_ = b.client.AnswerCallback(ctx, cb.ID, t.LoginFailedToast)
		return
	}
	_ = b.store.UpdateProfile(cb.From.ID, cb.From.Username, cb.From.FirstName)
	if cb.From.LanguageCode != "" {
		_ = b.store.SetTelegramLang(cb.From.ID, cb.From.LanguageCode)
	}
	if err := b.store.ConfirmLoginToken(token, cb.From.ID); err != nil {
		_ = b.client.AnswerCallback(ctx, cb.ID, t.LoginUsedToast)
		b.editCallback(ctx, cb, t.LoginUsed)
		return
	}
	_ = b.client.AnswerCallback(ctx, cb.ID, t.LoginConfirmToast)
	text := t.LoginConfirmed
	if created {
		text += trialNote(t)
	}
	b.editCallback(ctx, cb, text)
}

func trialNote(t *i18n.Text) string {
	return fmt.Sprintf(t.TrialNote, store.TrialDays)
}

func (b *Bot) editCallback(ctx context.Context, cb *CallbackQuery, text string) {
	if cb.Message == nil {
		return
	}
	if err := b.client.EditText(ctx, cb.Message.Chat.ID, cb.Message.MessageID, text); err != nil {
		log.Printf("更新按钮消息失败: %v", err)
	}
}

func (b *Bot) dispatch(ctx context.Context, from User, t *i18n.Text, cmd, args string) string {
	userID := from.ID
	switch cmd {
	case "/start":
		return b.start(from, t, args)
	case "/help":
		return b.helpText(t)
	case "/status":
		return b.status(t, userID)
	case "/add":
		return b.add(ctx, t, userID, args)
	case "/remove":
		return b.remove(t, userID, args)
	case "/set_threshold":
		return b.setThresholdMoved(t)
	case "/list":
		return b.list(ctx, t, userID)
	default:
		return t.UnknownCommand + "\n\n" + b.helpText(t)
	}
}

func (b *Bot) start(from User, t *i18n.Text, args string) string {
	var ref *int64
	if id, ok := parseRef(args); ok {
		ref = &id
	}
	created, err := b.store.EnsureUser(from.ID, ref, config.DefaultThreshold)
	if err != nil {
		log.Printf("注册用户失败: %v", err)
		return t.RegisterFailed
	}
	_ = b.store.UpdateProfile(from.ID, from.Username, from.FirstName)
	if from.LanguageCode != "" {
		_ = b.store.SetTelegramLang(from.ID, from.LanguageCode)
	}
	if created {
		return t.Registered + trialNote(t) + "\n\n" + b.helpText(t)
	}
	return t.AlreadyRegistered + "\n\n" + b.helpText(t)
}

func (b *Bot) status(t *i18n.Text, userID int64) string {
	if _, err := b.store.EnsureUser(userID, nil, config.DefaultThreshold); err != nil {
		return t.QueryFailed
	}
	user, err := b.store.GetUser(userID)
	if err != nil {
		return t.QueryFailed
	}
	return store.FormatStatus(user)
}

func (b *Bot) add(ctx context.Context, t *i18n.Text, userID int64, args string) string {
	addr, ok := NormalizeAddress(args)
	if !ok {
		return t.AddUsage
	}
	if _, err := b.store.EnsureUser(userID, nil, config.DefaultThreshold); err != nil {
		return t.AddFailed
	}
	user, err := b.store.GetUser(userID)
	if err != nil {
		return t.AddFailed
	}
	limit := user.AddressLimit(time.Now())
	err = b.store.AddAddress(userID, addr, limit)
	switch {
	case err == nil:
		callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		results := b.chains.Fetch(callCtx, common.HexToAddress(addr))
		if b.checkNow != nil {
			go b.checkNow(ctx, addr, userID)
		}
		return notify.Added(user.Lang, addr, config.DefaultThreshold, results, user.ProActive(time.Now()))
	case errors.Is(err, store.ErrAddressExists):
		return t.AddExists
	case errors.Is(err, store.ErrAddressLimit):
		return fmt.Sprintf(t.AddLimit, limit)
	default:
		log.Printf("添加地址失败: %v", err)
		return t.AddFailed
	}
}

func (b *Bot) remove(t *i18n.Text, userID int64, args string) string {
	addr, ok := NormalizeAddress(args)
	if !ok {
		return t.RemoveUsage
	}
	err := b.store.RemoveAddress(userID, addr)
	switch {
	case err == nil:
		return fmt.Sprintf(t.Removed, addr)
	case errors.Is(err, store.ErrNotFound):
		return t.RemoveMissing
	default:
		log.Printf("删除地址失败: %v", err)
		return t.RemoveFailed
	}
}

// setThresholdMoved answers the retired command: alert lines are now set per address in the web console.
func (b *Bot) setThresholdMoved(t *i18n.Text) string {
	text := t.ThresholdMoved
	if b.siteURL != "" {
		text += "\n\n" + b.siteURL
	}
	return text
}

func (b *Bot) list(ctx context.Context, t *i18n.Text, userID int64) string {
	watches, err := b.store.ListAddresses(userID)
	if err != nil {
		log.Printf("读取地址失败: %v", err)
		return t.QueryFailed
	}
	if len(watches) == 0 {
		return t.NoAddresses
	}
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	var parts []string
	for _, watch := range watches {
		results := b.chains.Fetch(callCtx, common.HexToAddress(watch.Address))
		text := notify.List(watch.Lang, watch.Address, watch.Threshold, results)
		if !watch.Active {
			text += t.Paused
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n\n")
}

func parseCommand(text string) (string, string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return "", ""
	}
	fields := strings.Fields(text)
	cmd := strings.ToLower(strings.Split(fields[0], "@")[0])
	return cmd, strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
}

func parseRef(args string) (int64, bool) {
	args = strings.TrimSpace(args)
	if !strings.HasPrefix(args, "ref_") {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(args, "ref_"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func ParseThreshold(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 1 || value > 10 {
		return 0, false
	}
	return value, true
}

func NormalizeAddress(raw string) (string, bool) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return "", false
	}
	if !common.IsHexAddress(fields[0]) {
		return "", false
	}
	addr := common.HexToAddress(fields[0])
	if addr == (common.Address{}) {
		return "", false
	}
	return addr.Hex(), true
}
