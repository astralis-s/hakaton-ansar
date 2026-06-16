package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiBase = "https://api.telegram.org"

// Client — тонкий клиент Telegram Bot API поверх net/http.
type Client struct {
	token string
	http  *http.Client
}

// New создаёт клиент. HTTP-таймаут заведомо больше таймаута long polling, чтобы
// долгий getUpdates не обрывался преждевременно.
func New(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 70 * time.Second},
	}
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
	ErrorCode   int             `json:"error_code"`
}

// call выполняет POST-запрос к методу Bot API и при out != nil декодирует result.
func (c *Client) call(ctx context.Context, method string, payload, out any) error {
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return err
		}
	}
	url := fmt.Sprintf("%s/bot%s/%s", apiBase, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ar apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return fmt.Errorf("декодирование ответа telegram (%s): %w", method, err)
	}
	if !ar.OK {
		return fmt.Errorf("telegram api %s: %s (код %d)", method, ar.Description, ar.ErrorCode)
	}
	if out != nil && len(ar.Result) > 0 {
		if err := json.Unmarshal(ar.Result, out); err != nil {
			return err
		}
	}
	return nil
}

// GetMe проверяет токен и возвращает данные бота.
func (c *Client) GetMe(ctx context.Context) (Me, error) {
	var me Me
	err := c.call(ctx, "getMe", nil, &me)
	return me, err
}

type getUpdatesRequest struct {
	Offset         int64    `json:"offset,omitempty"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

// GetUpdates делает long polling: ждёт до timeoutSec секунд новые обновления,
// начиная с offset.
func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	var updates []Update
	err := c.call(ctx, "getUpdates", getUpdatesRequest{
		Offset:         offset,
		Timeout:        timeoutSec,
		AllowedUpdates: []string{"message"},
	}, &updates)
	return updates, err
}

type sendMessageRequest struct {
	ChatID      int64  `json:"chat_id"`
	Text        string `json:"text"`
	ReplyMarkup any    `json:"reply_markup,omitempty"`
}

// SendMessage отправляет простое текстовое сообщение.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.call(ctx, "sendMessage", sendMessageRequest{ChatID: chatID, Text: text}, nil)
}

// SendMessageWithMarkup отправляет сообщение вместе с reply-разметкой
// (кастомной клавиатурой или её удалением).
func (c *Client) SendMessageWithMarkup(ctx context.Context, chatID int64, text string, markup any) error {
	return c.call(ctx, "sendMessage", sendMessageRequest{ChatID: chatID, Text: text, ReplyMarkup: markup}, nil)
}
