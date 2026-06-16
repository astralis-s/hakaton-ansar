// Package webui — веб-инбокс менеджера: HTTP API (список чатов, переписка,
// отправка ответа) и встроенная одностраничная статика. Доступ закрывается
// простым паролем (если он задан в конфиге).
package webui

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/app"
)

//go:embed static
var staticFS embed.FS

const sessionTTL = 12 * time.Hour

// Server обслуживает веб-инбокс менеджера.
type Server struct {
	bridge   *app.Bridge
	log      *slog.Logger
	password string

	mu       sync.Mutex
	sessions map[string]time.Time // токен -> момент истечения
}

// NewServer создаёт сервер. Пустой password отключает авторизацию.
func NewServer(bridge *app.Bridge, log *slog.Logger, password string) *Server {
	return &Server{
		bridge:   bridge,
		log:      log,
		password: password,
		sessions: map[string]time.Time{},
	}
}

// Handler собирает маршрутизатор: статика + JSON API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err) // встроенная статика отсутствует — ошибка сборки, не рантайма
	}
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("GET /api/chats", s.auth(s.handleChats))
	mux.HandleFunc("GET /api/chats/{id}/messages", s.auth(s.handleThread))
	mux.HandleFunc("POST /api/chats/{id}/messages", s.auth(s.handleSend))
	mux.HandleFunc("POST /api/chats/{id}/read", s.auth(s.handleRead))

	return mux
}

// ---- авторизация -----------------------------------------------------------

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.password == "" {
			next(w, r)
			return
		}
		if tok := bearer(r); tok != "" && s.validSession(tok) {
			next(w, r)
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
}

func (s *Server) validSession(tok string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.sessions[tok]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(s.sessions, tok)
		return false
	}
	return true
}

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"auth_required": s.password != ""})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.password == "" {
		writeJSON(w, http.StatusOK, map[string]any{"token": "", "auth_required": false})
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.password)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid password"})
		return
	}
	tok := newToken()
	s.mu.Lock()
	s.sessions[tok] = time.Now().Add(sessionTTL)
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"token": tok, "auth_required": true})
}

// ---- API инбокса -----------------------------------------------------------

func (s *Server) handleChats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.bridge.Inbox())
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid chat id"})
		return
	}
	user, ok := s.bridge.User(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "chat not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chat": map[string]any{
			"chat_id":   user.ChatID,
			"full_name": user.FullName,
			"phone":     user.Phone,
			"username":  user.Username,
		},
		"messages": s.bridge.Thread(id),
	})
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid chat id"})
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	msg, err := s.bridge.SendFromManager(r.Context(), id, req.Text)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrEmptyMessage):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty message"})
		case errors.Is(err, app.ErrUnknownChat):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "chat not found"})
		default:
			s.log.Error("send from manager", "error", err, "chat_id", id)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to deliver to telegram"})
		}
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	id, err := chatIDParam(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid chat id"})
		return
	}
	if err := s.bridge.MarkRead(id); err != nil {
		s.log.Error("mark read", "error", err, "chat_id", id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- помощники -------------------------------------------------------------

func chatIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func newToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
