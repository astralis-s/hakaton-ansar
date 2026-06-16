package app

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/store"
	"github.com/astralis-s/hakaton-ansar/telegram-bot/internal/telegram"
)

// fakeSender записывает отправленные в Telegram сообщения вместо реального вызова API.
type fakeSender struct {
	texts []string
}

func (f *fakeSender) SendMessage(_ context.Context, _ int64, text string) error {
	f.texts = append(f.texts, text)
	return nil
}

func (f *fakeSender) SendMessageWithMarkup(_ context.Context, _ int64, text string, _ any) error {
	f.texts = append(f.texts, text)
	return nil
}

func newTestBridge(t *testing.T) (*Bridge, *fakeSender) {
	t.Helper()
	st, err := store.Open("") // in-memory
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	fake := &fakeSender{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewBridge(st, fake, log), fake
}

func textUpdate(chatID int64, text string) telegram.Update {
	return telegram.Update{Message: &telegram.Message{
		Chat: telegram.Chat{ID: chatID, Type: "private"},
		From: &telegram.User{ID: chatID, Username: "tg_user"},
		Text: text,
	}}
}

// TestRegistrationFlow проверяет весь путь: /start → ФИО → телефон → переписка,
// и что в инбоксе менеджера отображается ФИО, а не никнейм.
func TestRegistrationFlow(t *testing.T) {
	b, _ := newTestBridge(t)
	ctx := context.Background()
	const chatID = int64(555)

	b.HandleUpdate(ctx, textUpdate(chatID, "/start"))
	if u, _ := b.User(chatID); u.State != store.StateAwaitingName {
		t.Fatalf("после /start состояние = %q, want awaiting_name", u.State)
	}

	// Пока ждём ФИО, посторонний ввод (не похожий на ФИО) не попадает в переписку
	// и не продвигает регистрацию.
	b.HandleUpdate(ctx, textUpdate(chatID, "привет"))
	if got := len(b.Thread(chatID)); got != 0 {
		t.Fatalf("до регистрации сообщений в треде = %d, want 0", got)
	}
	if u, _ := b.User(chatID); u.State != store.StateAwaitingName {
		t.Fatalf("после невалидного ФИО состояние = %q, want awaiting_name", u.State)
	}

	b.HandleUpdate(ctx, textUpdate(chatID, "Ибрагимов Ислам Висханович"))
	if u, _ := b.User(chatID); u.State != store.StateAwaitingPhone {
		t.Fatalf("после ФИО состояние = %q, want awaiting_phone", u.State)
	}

	b.HandleUpdate(ctx, textUpdate(chatID, "+7 928 000-00-00"))
	u, _ := b.User(chatID)
	if !u.Registered() {
		t.Fatalf("после телефона пользователь не зарегистрирован: %+v", u)
	}
	if u.FullName != "Ибрагимов Ислам Висханович" {
		t.Fatalf("ФИО = %q", u.FullName)
	}
	if u.Phone != "+79280000000" {
		t.Fatalf("телефон = %q", u.Phone)
	}

	// Теперь сообщение клиента попадает в переписку.
	b.HandleUpdate(ctx, textUpdate(chatID, "Здравствуйте, интересует рассрочка"))
	thread := b.Thread(chatID)
	if len(thread) != 1 || thread[0].Sender != store.SenderClient {
		t.Fatalf("ожидалось 1 клиентское сообщение, получено %+v", thread)
	}

	// Инбокс менеджера показывает ФИО + телефон и 1 непрочитанное.
	inbox := b.Inbox()
	if len(inbox) != 1 {
		t.Fatalf("в инбоксе %d чатов, want 1", len(inbox))
	}
	if inbox[0].FullName != "Ибрагимов Ислам Висханович" {
		t.Fatalf("в инбоксе имя = %q, want ФИО", inbox[0].FullName)
	}
	if inbox[0].Phone != "+79280000000" {
		t.Fatalf("в инбоксе телефон = %q", inbox[0].Phone)
	}
	if inbox[0].Unread != 1 {
		t.Fatalf("непрочитанных = %d, want 1", inbox[0].Unread)
	}
}

// TestManagerReply проверяет доставку ответа менеджера в Telegram и его попадание
// в историю, а также сброс счётчика непрочитанных после MarkRead.
func TestManagerReply(t *testing.T) {
	b, fake := newTestBridge(t)
	ctx := context.Background()
	const chatID = int64(777)

	// Быстрая регистрация.
	b.HandleUpdate(ctx, textUpdate(chatID, "/start"))
	b.HandleUpdate(ctx, textUpdate(chatID, "Иванов Иван"))
	b.HandleUpdate(ctx, textUpdate(chatID, "89280000001"))
	b.HandleUpdate(ctx, textUpdate(chatID, "вопрос по договору"))

	before := len(fake.texts)
	msg, err := b.SendFromManager(ctx, chatID, "Здравствуйте! Сейчас всё расскажу.")
	if err != nil {
		t.Fatalf("SendFromManager: %v", err)
	}
	if msg.Sender != store.SenderManager {
		t.Fatalf("сообщение менеджера имеет sender = %q", msg.Sender)
	}
	if len(fake.texts) != before+1 || fake.texts[len(fake.texts)-1] != "Здравствуйте! Сейчас всё расскажу." {
		t.Fatalf("ответ менеджера не доставлен в telegram: %+v", fake.texts)
	}

	thread := b.Thread(chatID)
	if len(thread) != 2 {
		t.Fatalf("в треде %d сообщений, want 2", len(thread))
	}

	if err := b.MarkRead(chatID); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if got := b.Inbox()[0].Unread; got != 0 {
		t.Fatalf("после MarkRead непрочитанных = %d, want 0", got)
	}
}

// TestManagerReplyUnknownChat: нельзя писать в незарегистрированный чат.
func TestManagerReplyUnknownChat(t *testing.T) {
	b, _ := newTestBridge(t)
	if _, err := b.SendFromManager(context.Background(), 999, "привет"); err != ErrUnknownChat {
		t.Fatalf("err = %v, want ErrUnknownChat", err)
	}
}
