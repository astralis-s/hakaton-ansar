package app

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// ---- fakes ------------------------------------------------------------------

type fakeRepo struct {
	users    map[int64]domain.TelegramUser
	byClient map[string]int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[int64]domain.TelegramUser{}, byClient: map[string]int64{}}
}

func (r *fakeRepo) GetByChatID(_ context.Context, chatID int64) (domain.TelegramUser, error) {
	u, ok := r.users[chatID]
	if !ok {
		return domain.TelegramUser{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeRepo) Save(_ context.Context, u domain.TelegramUser) error {
	r.users[u.ChatID()] = u
	if u.Registered() {
		r.byClient[u.OrgID()+"|"+u.ClientID()] = u.ChatID()
	}
	return nil
}

func (r *fakeRepo) FindChatByClient(_ context.Context, orgID, clientID string) (int64, bool, error) {
	id, ok := r.byClient[orgID+"|"+clientID]
	return id, ok, nil
}

type fakeClients struct {
	created             int
	lastName, lastPhone string
}

func (c *fakeClients) CreateClient(_ context.Context, _, fullName, phone string) (string, error) {
	c.created++
	c.lastName, c.lastPhone = fullName, phone
	return "client-1", nil
}

type fakeInbox struct{ posts []string }

func (i *fakeInbox) PostClientMessage(_ context.Context, _, _, body string) error {
	i.posts = append(i.posts, body)
	return nil
}

type fakeMessenger struct {
	sent    []string
	contact []string
}

func (m *fakeMessenger) Send(_ context.Context, _ int64, text string) error {
	m.sent = append(m.sent, text)
	return nil
}

func (m *fakeMessenger) AskForContact(_ context.Context, _ int64, text string) error {
	m.contact = append(m.contact, text)
	return nil
}

type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

func testLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// ---- tests ------------------------------------------------------------------

func TestRegistrationAndRelay(t *testing.T) {
	repo := newFakeRepo()
	clients := &fakeClients{}
	inbox := &fakeInbox{}
	msg := &fakeMessenger{}
	uc := NewProcessUpdate(repo, clients, inbox, msg, fakeTx{}, "org-1", testLogger())
	ctx := context.Background()
	const chatID = int64(555)

	// /start → начинает регистрацию.
	uc.HandleMessage(ctx, IncomingMessage{ChatID: chatID, Username: "ivan", Text: "/start"})
	u, _ := repo.GetByChatID(ctx, chatID)
	require.Equal(t, domain.StateAwaitingName, u.State())

	// Посторонний ввод в ожидании ФИО (одно слово) — не сохраняется, не продвигает.
	uc.HandleMessage(ctx, IncomingMessage{ChatID: chatID, Text: "привет"})
	require.Empty(t, inbox.posts)
	u, _ = repo.GetByChatID(ctx, chatID)
	require.Equal(t, domain.StateAwaitingName, u.State())

	// ФИО.
	uc.HandleMessage(ctx, IncomingMessage{ChatID: chatID, Text: "Ибрагимов Ислам Висханович"})
	u, _ = repo.GetByChatID(ctx, chatID)
	require.Equal(t, domain.StateAwaitingPhone, u.State())

	// Телефон через кнопку "поделиться контактом".
	uc.HandleMessage(ctx, IncomingMessage{ChatID: chatID, Contact: "+7 928 000-00-00"})
	u, _ = repo.GetByChatID(ctx, chatID)
	require.True(t, u.Registered())
	assert.Equal(t, 1, clients.created)
	assert.Equal(t, "Ибрагимов Ислам Висханович", clients.lastName)
	assert.Equal(t, "+79280000000", clients.lastPhone)
	assert.Equal(t, "client-1", u.ClientID())

	// Сообщение зарегистрированного клиента уходит в инбокс менеджера.
	uc.HandleMessage(ctx, IncomingMessage{ChatID: chatID, Text: "Здравствуйте, интересует рассрочка"})
	require.Equal(t, []string{"Здравствуйте, интересует рассрочка"}, inbox.posts)
}

func TestStartGreetsRegisteredUser(t *testing.T) {
	repo := newFakeRepo()
	msg := &fakeMessenger{}
	uc := NewProcessUpdate(repo, &fakeClients{}, &fakeInbox{}, msg, fakeTx{}, "org-1", testLogger())
	ctx := context.Background()

	u, _ := domain.NewTelegramUser(42, "org-1", "")
	require.NoError(t, u.RecordName("Иванов Иван"))
	require.NoError(t, u.CompleteRegistration("+79280000001", "client-9"))
	require.NoError(t, repo.Save(ctx, u))

	uc.HandleMessage(ctx, IncomingMessage{ChatID: 42, Text: "/start"})
	// Состояние не сбрасывается, отправлено приветствие.
	got, _ := repo.GetByChatID(ctx, 42)
	require.True(t, got.Registered())
	require.NotEmpty(t, msg.sent)
}

func TestDeliverStaffReply(t *testing.T) {
	repo := newFakeRepo()
	msg := &fakeMessenger{}
	ctx := context.Background()

	u, _ := domain.NewTelegramUser(777, "org-1", "")
	require.NoError(t, u.RecordName("Иванов Иван"))
	require.NoError(t, u.CompleteRegistration("+79280000001", "client-9"))
	require.NoError(t, repo.Save(ctx, u))

	d := NewDeliverStaffReply(repo, msg, testLogger())

	d.Execute(ctx, "org-1", "client-9", "Здравствуйте! Готово.")
	require.Equal(t, []string{"Здравствуйте! Готово."}, msg.sent)

	// Клиент без Telegram-чата — ничего не отправляется.
	msg.sent = nil
	d.Execute(ctx, "org-1", "client-unknown", "привет")
	require.Empty(t, msg.sent)
}
