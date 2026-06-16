package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Store — потокобезопасное хранилище. Все мутации сразу персистятся в JSON-файл
// (атомарно: запись во временный файл + rename). Объёмы переписки в рамках демо
// небольшие, поэтому полная перезапись файла на каждое изменение допустима.
type Store struct {
	mu   sync.RWMutex
	path string
	d    data
	now  func() time.Time
}

type data struct {
	Users    map[int64]*User `json:"users"`
	Messages []*Message      `json:"messages"`
	NextID   int64           `json:"next_id"`
}

// Open открывает (или создаёт) хранилище по пути path. Пустой path — режим
// in-memory без персистентности (удобно для тестов).
func Open(path string) (*Store, error) {
	s := &Store{
		path: path,
		now:  time.Now,
		d:    data{Users: map[int64]*User{}, NextID: 1},
	}
	if path != "" {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil // первый запуск — стартуем с чистого хранилища
	}
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	var d data
	if err := json.Unmarshal(b, &d); err != nil {
		return err
	}
	if d.Users == nil {
		d.Users = map[int64]*User{}
	}
	if d.NextID == 0 {
		d.NextID = 1
	}
	s.d = d
	return nil
}

// persist пишет всё хранилище на диск. Вызывается под удержанным замком.
func (s *Store) persist() error {
	if s.path == "" {
		return nil
	}
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(s.d, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// GetUser возвращает копию пользователя по chatID.
func (s *Store) GetUser(chatID int64) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.d.Users[chatID]
	if !ok {
		return User{}, false
	}
	return *u, true
}

// UpsertUser применяет mutate к пользователю (создавая его при первом обращении)
// и персистит результат. Возвращает обновлённую копию.
func (s *Store) UpsertUser(chatID int64, mutate func(*User)) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.d.Users[chatID]
	if !ok {
		u = &User{ChatID: chatID, State: StateNew, CreatedAt: s.now()}
		s.d.Users[chatID] = u
	}
	mutate(u)
	u.UpdatedAt = s.now()
	if err := s.persist(); err != nil {
		return User{}, err
	}
	return *u, nil
}

// AddMessage добавляет сообщение в переписку и персистит его. Сообщения менеджера
// сразу помечаются прочитанными (непрочитанные — это входящие от клиента).
func (s *Store) AddMessage(chatID int64, sender SenderKind, text string) (Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := &Message{
		ID:        s.d.NextID,
		ChatID:    chatID,
		Sender:    sender,
		Text:      text,
		Read:      sender == SenderManager,
		CreatedAt: s.now(),
	}
	s.d.NextID++
	s.d.Messages = append(s.d.Messages, m)
	if err := s.persist(); err != nil {
		return Message{}, err
	}
	return *m, nil
}

// Messages возвращает переписку клиента в хронологическом порядке.
func (s *Store) Messages(chatID int64) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Message, 0)
	for _, m := range s.d.Messages {
		if m.ChatID == chatID {
			out = append(out, *m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// MarkRead помечает все входящие сообщения клиента прочитанными.
func (s *Store) MarkRead(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for _, m := range s.d.Messages {
		if m.ChatID == chatID && m.Sender == SenderClient && !m.Read {
			m.Read = true
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.persist()
}

// Conversations строит список чатов для инбокса менеджера: только
// зарегистрированные пользователи (с ФИО и телефоном), отсортированные по
// последней активности. Имя в списке — ФИО из регистрации.
func (s *Store) Conversations() []Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type agg struct {
		last   *Message
		unread int
	}
	aggs := map[int64]*agg{}
	for _, m := range s.d.Messages {
		a := aggs[m.ChatID]
		if a == nil {
			a = &agg{}
			aggs[m.ChatID] = a
		}
		if a.last == nil || m.ID > a.last.ID {
			a.last = m
		}
		if m.Sender == SenderClient && !m.Read {
			a.unread++
		}
	}

	out := make([]Conversation, 0, len(s.d.Users))
	for _, u := range s.d.Users {
		if !u.Registered() {
			continue // незавершённую регистрацию в инбоксе не показываем
		}
		c := Conversation{
			ChatID:   u.ChatID,
			FullName: u.FullName,
			Phone:    u.Phone,
			Username: u.Username,
			LastAt:   u.CreatedAt,
		}
		if a := aggs[u.ChatID]; a != nil && a.last != nil {
			c.LastText = a.last.Text
			c.LastSender = a.last.Sender
			c.LastAt = a.last.CreatedAt
			c.Unread = a.unread
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastAt.After(out[j].LastAt) })
	return out
}
