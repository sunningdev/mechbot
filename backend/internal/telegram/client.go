package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	token string
	http  *http.Client
}

func New(token string) *Client {
	return &Client{token: token, http: &http.Client{Timeout: 40 * time.Second}}
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	From      *User  `json:"from"`
	Chat      Chat   `json:"chat"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	// LanguageCode is the IETF tag of the user's Telegram client language, when Telegram shares it.
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type Button struct {
	Text string `json:"text"`
	Data string `json:"callback_data"`
}

func (c *Client) Me(ctx context.Context) (User, error) {
	var wrapped struct {
		OK     bool `json:"ok"`
		Result User `json:"result"`
	}
	if err := c.call(ctx, "getMe", nil, &wrapped); err != nil {
		return User{}, err
	}
	if !wrapped.OK {
		return User{}, fmt.Errorf("getMe 返回失败")
	}
	return wrapped.Result, nil
}

func (c *Client) Updates(ctx context.Context, offset int64) ([]Update, error) {
	var wrapped struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	payload := map[string]any{
		"timeout":         25,
		"offset":          offset,
		"allowed_updates": []string{"message", "callback_query"},
	}
	if err := c.call(ctx, "getUpdates", payload, &wrapped); err != nil {
		return nil, err
	}
	if !wrapped.OK {
		return nil, fmt.Errorf("getUpdates 返回失败")
	}
	return wrapped.Result, nil
}

func (c *Client) Send(ctx context.Context, chatID int64, text string) error {
	for _, part := range chunks(text, 3500) {
		if err := c.sendOne(ctx, chatID, part, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) SendButtons(ctx context.Context, chatID int64, text string, buttons ...Button) error {
	return c.SendKeyboard(ctx, chatID, text, [][]Button{buttons})
}

// SendKeyboard sends text with an inline keyboard laid out row by row.
func (c *Client) SendKeyboard(ctx context.Context, chatID int64, text string, rows [][]Button) error {
	return c.sendOne(ctx, chatID, text, rows)
}

func (c *Client) sendOne(ctx context.Context, chatID int64, text string, rows [][]Button) error {
	payload := map[string]any{"chat_id": chatID, "text": text, "disable_web_page_preview": true}
	if len(rows) > 0 {
		payload["reply_markup"] = map[string]any{"inline_keyboard": rows}
	}
	return c.expectOK(ctx, "sendMessage", payload)
}

func (c *Client) EditText(ctx context.Context, chatID, messageID int64, text string) error {
	return c.expectOK(ctx, "editMessageText", map[string]any{"chat_id": chatID, "message_id": messageID, "text": text})
}

func (c *Client) AnswerCallback(ctx context.Context, id, text string) error {
	return c.expectOK(ctx, "answerCallbackQuery", map[string]any{"callback_query_id": id, "text": text})
}

func (c *Client) expectOK(ctx context.Context, method string, payload any) error {
	var wrapped struct {
		OK bool `json:"ok"`
	}
	if err := c.call(ctx, method, payload, &wrapped); err != nil {
		return err
	}
	if !wrapped.OK {
		return fmt.Errorf("%s 返回失败", method)
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, payload any, dest any) error {
	if payload == nil {
		payload = map[string]any{}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint := "https://api.telegram.org/bot" + c.token + "/" + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		// url.Error embeds the request URL, which contains the bot token; keep it out of logs.
		var uerr *url.Error
		if errors.As(err, &uerr) {
			return fmt.Errorf("telegram %s: %w", method, uerr.Err)
		}
		return fmt.Errorf("telegram %s: %w", method, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram %s status %d: %s", method, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return json.Unmarshal(raw, dest)
}

func chunks(text string, size int) []string {
	runes := []rune(text)
	if len(runes) <= size {
		return []string{text}
	}
	var out []string
	for len(runes) > 0 {
		n := size
		if n > len(runes) {
			n = len(runes)
		}
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}
