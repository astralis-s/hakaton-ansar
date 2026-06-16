// Package app связывает Telegram-бота и веб-инбокс менеджера: обрабатывает
// входящие апдейты (включая регистрацию ФИО + телефон через конечный автомат),
// складывает переписку в хранилище и доставляет ответы менеджера обратно в
// Telegram.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/store"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/telegram"
)

// Ошибки, возвращаемые в веб-слой при ответе менеджера.
var (
	ErrEmptyMessage = errors.New("пустое сообщение")
	ErrUnknownChat  = errors.New("чат не найден или клиент не завершил регистрацию")
)

const msgInternal = "Произошла ошибка, попробуйте, пожалуйста, ещё раз чуть позже."

// sender — порт отправки сообщений в Telegram (реализуется telegram.Client).
type sender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithMarkup(ctx context.Context, chatID int64, text string, markup any) error
}

// Bridge — мост Telegram ↔ инбокс менеджера.
type Bridge struct {
	store *store.Store
	tg    sender
	log   *slog.Logger
}

// NewBridge собирает мост из хранилища, клиента Telegram и логгера.
func NewBridge(s *store.Store, tg sender, log *slog.Logger) *Bridge {
	return &Bridge{store: s, tg: tg, log: log}
}

// HandleUpdate обрабатывает одно обновление Telegram (точка входа для поллера).
func (b *Bridge) HandleUpdate(ctx context.Context, u telegram.Update) {
	if u.Message == nil || u.Message.Chat.Type != "private" {
		return // группы/каналы и не-сообщения игнорируем
	}
	msg := u.Message
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	if text == "/start" {
		b.handleStart(ctx, chatID, msg)
		return
	}

	user, _ := b.store.GetUser(chatID)
	switch user.State {
	case store.StateNew:
		// Первый контакт без явного /start — ведём себя как при /start.
		b.handleStart(ctx, chatID, msg)
	case store.StateAwaitingName:
		b.handleName(ctx, chatID, text)
	case store.StateAwaitingPhone:
		b.handlePhone(ctx, chatID, msg, text)
	case store.StateRegistered:
		b.handleClientMessage(ctx, chatID, text)
	}
}

func (b *Bridge) handleStart(ctx context.Context, chatID int64, msg *telegram.Message) {
	if user, ok := b.store.GetUser(chatID); ok && user.Registered() {
		b.send(ctx, chatID, "Ассаламу алейкум, "+user.FullName+"! 👋\n\nВы уже зарегистрированы. Просто напишите сюда ваш вопрос — менеджер «Амана» ответит вам в этом чате.")
		return
	}
	username := ""
	if msg.From != nil {
		username = msg.From.Username
	}
	if _, err := b.store.UpsertUser(chatID, func(u *store.User) {
		u.State = store.StateAwaitingName
		u.Username = username
	}); err != nil {
		b.log.Error("upsert user (start)", "error", err)
		b.send(ctx, chatID, msgInternal)
		return
	}
	b.send(ctx, chatID, "Ассаламу алейкум! 👋\n\nЭто чат поддержки «Амана» — рассрочка без процентов (риба).\nЧтобы менеджер мог с вами работать, давайте познакомимся.\n\nНапишите, пожалуйста, ваши Фамилию, Имя и Отчество одной строкой.\nНапример: Ибрагимов Ислам Висханович")
}

func (b *Bridge) handleName(ctx context.Context, chatID int64, text string) {
	name := normalizeName(text)
	if !validName(name) {
		b.send(ctx, chatID, "Пожалуйста, укажите как минимум Фамилию и Имя одной строкой.\nНапример: Ибрагимов Ислам Висханович")
		return
	}
	if _, err := b.store.UpsertUser(chatID, func(u *store.User) {
		u.FullName = name
		u.State = store.StateAwaitingPhone
	}); err != nil {
		b.log.Error("upsert user (name)", "error", err)
		b.send(ctx, chatID, msgInternal)
		return
	}
	b.sendWithKeyboard(ctx, chatID,
		"Рады знакомству, "+name+"!\n\nТеперь отправьте, пожалуйста, ваш контактный номер телефона.\nМожно нажать кнопку ниже или ввести номер вручную (например: +7 928 000-00-00).",
		telegram.ContactKeyboard())
}

func (b *Bridge) handlePhone(ctx context.Context, chatID int64, msg *telegram.Message, text string) {
	raw := text
	if msg.Contact != nil && msg.Contact.PhoneNumber != "" {
		raw = msg.Contact.PhoneNumber
	}
	phone, ok := normalizePhone(raw)
	if !ok {
		b.sendWithKeyboard(ctx, chatID,
			"Это не похоже на номер телефона. Отправьте номер в формате +7XXXXXXXXXX или нажмите кнопку ниже.",
			telegram.ContactKeyboard())
		return
	}
	user, err := b.store.UpsertUser(chatID, func(u *store.User) {
		u.Phone = phone
		u.State = store.StateRegistered
	})
	if err != nil {
		b.log.Error("upsert user (phone)", "error", err)
		b.send(ctx, chatID, msgInternal)
		return
	}
	b.log.Info("client registered", "chat_id", chatID, "name", user.FullName)
	b.sendRemoveKeyboard(ctx, chatID,
		"Спасибо, "+user.FullName+"! Регистрация завершена. ✅\n\nТеперь просто напишите сюда ваш вопрос — менеджер «Амана» ответит вам прямо в этом чате.")
}

func (b *Bridge) handleClientMessage(ctx context.Context, chatID int64, text string) {
	if text == "" {
		b.send(ctx, chatID, "Пожалуйста, отправьте текстовое сообщение — менеджер увидит его и ответит.")
		return
	}
	if _, err := b.store.AddMessage(chatID, store.SenderClient, text); err != nil {
		b.log.Error("store client message", "error", err, "chat_id", chatID)
		b.send(ctx, chatID, msgInternal)
		return
	}
	b.log.Info("client message received", "chat_id", chatID)
	// Авто-ответа нет: на сообщение отвечает менеджер из веб-инбокса.
}

// SendFromManager доставляет ответ менеджера клиенту в Telegram и сохраняет его в
// переписке. Сообщение пишется в историю только после успешной доставки.
func (b *Bridge) SendFromManager(ctx context.Context, chatID int64, text string) (store.Message, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return store.Message{}, ErrEmptyMessage
	}
	user, ok := b.store.GetUser(chatID)
	if !ok || !user.Registered() {
		return store.Message{}, ErrUnknownChat
	}
	if err := b.tg.SendMessage(ctx, chatID, text); err != nil {
		return store.Message{}, fmt.Errorf("доставка в telegram: %w", err)
	}
	return b.store.AddMessage(chatID, store.SenderManager, text)
}

// ---- read-модели для веб-инбокса ------------------------------------------

// Inbox возвращает список чатов (с ФИО и телефоном) для менеджера.
func (b *Bridge) Inbox() []store.Conversation { return b.store.Conversations() }

// Thread возвращает переписку конкретного клиента.
func (b *Bridge) Thread(chatID int64) []store.Message { return b.store.Messages(chatID) }

// MarkRead помечает входящие клиента прочитанными.
func (b *Bridge) MarkRead(chatID int64) error { return b.store.MarkRead(chatID) }

// User возвращает данные клиента по chatID.
func (b *Bridge) User(chatID int64) (store.User, bool) { return b.store.GetUser(chatID) }

// ---- помощники отправки ----------------------------------------------------

func (b *Bridge) send(ctx context.Context, chatID int64, text string) {
	if err := b.tg.SendMessage(ctx, chatID, text); err != nil {
		b.log.Error("send telegram message", "chat_id", chatID, "error", err)
	}
}

func (b *Bridge) sendWithKeyboard(ctx context.Context, chatID int64, text string, kb telegram.ReplyKeyboardMarkup) {
	if err := b.tg.SendMessageWithMarkup(ctx, chatID, text, kb); err != nil {
		b.log.Error("send telegram message", "chat_id", chatID, "error", err)
	}
}

func (b *Bridge) sendRemoveKeyboard(ctx context.Context, chatID int64, text string) {
	if err := b.tg.SendMessageWithMarkup(ctx, chatID, text, telegram.ReplyKeyboardRemove{RemoveKeyboard: true}); err != nil {
		b.log.Error("send telegram message", "chat_id", chatID, "error", err)
	}
}
