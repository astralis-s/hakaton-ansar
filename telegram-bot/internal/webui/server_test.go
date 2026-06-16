package webui

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/app"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/store"
)

// noopSender реализует порт отправки в Telegram, ничего не делая.
type noopSender struct{}

func (noopSender) SendMessage(context.Context, int64, string) error                { return nil }
func (noopSender) SendMessageWithMarkup(context.Context, int64, string, any) error { return nil }

func setup(t *testing.T, password string) (http.Handler, *store.Store) {
	t.Helper()
	st, err := store.Open("")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	// Зарегистрированный клиент + одно входящее сообщение.
	if _, err := st.UpsertUser(42, func(u *store.User) {
		u.FullName = "Ибрагимов Ислам Висханович"
		u.Phone = "+79280000000"
		u.State = store.StateRegistered
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := st.AddMessage(42, store.SenderClient, "Здравствуйте"); err != nil {
		t.Fatalf("seed message: %v", err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	bridge := app.NewBridge(st, noopSender{}, log)
	return NewServer(bridge, log, password).Handler(), st
}

func TestServesStaticIndex(t *testing.T) {
	h, _ := setup(t, "")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Амана") {
		t.Fatalf("index.html не отдан (нет 'Амана' в теле)")
	}
}

func TestInboxShowsFullName(t *testing.T) {
	h, _ := setup(t, "")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/chats", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/chats = %d, want 200", rr.Code)
	}
	var chats []store.Conversation
	if err := json.Unmarshal(rr.Body.Bytes(), &chats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(chats) != 1 || chats[0].FullName != "Ибрагимов Ислам Висханович" {
		t.Fatalf("в инбоксе ожидался 1 чат с ФИО, получено %+v", chats)
	}
	if chats[0].Unread != 1 {
		t.Fatalf("unread = %d, want 1", chats[0].Unread)
	}
}

func TestAuthGate(t *testing.T) {
	h, _ := setup(t, "secret")

	// Без токена — 401.
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/chats", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("без токена GET /api/chats = %d, want 401", rr.Code)
	}

	// Неверный пароль — 401.
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, jsonReq(http.MethodPost, "/api/login", `{"password":"wrong"}`))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("неверный пароль = %d, want 401", rr.Code)
	}

	// Верный пароль — выдаётся токен.
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, jsonReq(http.MethodPost, "/api/login", `{"password":"secret"}`))
	if rr.Code != http.StatusOK {
		t.Fatalf("верный пароль = %d, want 200", rr.Code)
	}
	var login struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &login); err != nil || login.Token == "" {
		t.Fatalf("токен не выдан: body=%s err=%v", rr.Body.String(), err)
	}

	// С токеном — доступ открыт.
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/chats", nil)
	req.Header.Set("Authorization", "Bearer "+login.Token)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("с токеном GET /api/chats = %d, want 200", rr.Code)
	}
}

func jsonReq(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
