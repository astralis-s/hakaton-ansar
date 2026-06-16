// Package app holds the telegrambot use-cases: ProcessUpdate drives the
// registration FSM and relays inbound client messages into the staff chat;
// DeliverStaffReply pushes manager replies back to Telegram.
package app

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// Bot copy shown to users.
const (
	msgInternal = "Произошла ошибка, попробуйте, пожалуйста, ещё раз чуть позже."
	msgNotReady = "Сервис пока не готов принимать обращения. Попробуйте позже, иншаАллах."

	msgAskName = "Ассаламу алейкум! 👋\n\nЭто чат поддержки «Амана» — рассрочка без процентов (риба).\n" +
		"Чтобы менеджер мог с вами работать, давайте познакомимся.\n\n" +
		"Напишите, пожалуйста, ваши Фамилию, Имя и Отчество одной строкой.\n" +
		"Например: Ибрагимов Ислам Висханович"
	msgBadName = "Пожалуйста, укажите как минимум Фамилию и Имя одной строкой.\n" +
		"Например: Ибрагимов Ислам Висханович"
	msgBadPhone = "Это не похоже на номер телефона. Отправьте номер в формате +7XXXXXXXXXX " +
		"или нажмите кнопку ниже."
	msgEmptyBody = "Пожалуйста, отправьте текстовое сообщение — менеджер увидит его и ответит."
)

// IncomingMessage is the transport-agnostic view of one Telegram message. The
// infra poller maps a Telegram update onto it, so the use-case never imports
// Telegram types.
type IncomingMessage struct {
	ChatID   int64
	Username string
	Text     string
	Contact  string // phone number from a shared contact, if any
}

// ProcessUpdate handles one inbound Telegram message through the registration
// FSM, creating the CRM client on completion and relaying chat messages.
type ProcessUpdate struct {
	repo      domain.UserRepository
	clients   domain.ClientDirectory
	inbox     domain.ChatInbox
	messenger domain.Messenger
	tx        domain.TxManager
	orgID     string
	log       *slog.Logger
}

func NewProcessUpdate(repo domain.UserRepository, clients domain.ClientDirectory, inbox domain.ChatInbox, messenger domain.Messenger, tx domain.TxManager, orgID string, log *slog.Logger) *ProcessUpdate {
	return &ProcessUpdate{repo: repo, clients: clients, inbox: inbox, messenger: messenger, tx: tx, orgID: orgID, log: log}
}

// HandleMessage processes a single inbound message. Errors are logged and a
// friendly notice is sent to the user; the poller loop never fails on one update.
func (uc *ProcessUpdate) HandleMessage(ctx context.Context, in IncomingMessage) {
	text := strings.TrimSpace(in.Text)

	user, err := uc.repo.GetByChatID(ctx, in.ChatID)
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		uc.begin(ctx, in)
		return
	case err != nil:
		uc.log.Error("load telegram user", "error", err, "chat_id", in.ChatID)
		return
	}

	if text == "/start" {
		uc.handleStart(ctx, &user, in)
		return
	}

	switch user.State() {
	case domain.StateAwaitingName:
		uc.handleName(ctx, &user, text)
	case domain.StateAwaitingPhone:
		uc.handlePhone(ctx, &user, text, in.Contact)
	case domain.StateRegistered:
		uc.handleClientMessage(ctx, &user, text)
	default:
		uc.askName(ctx, in.ChatID)
	}
}

func (uc *ProcessUpdate) begin(ctx context.Context, in IncomingMessage) {
	if uc.orgID == "" {
		uc.log.Error("telegram bot has no organization configured")
		uc.send(ctx, in.ChatID, msgNotReady)
		return
	}
	user, err := domain.NewTelegramUser(in.ChatID, uc.orgID, in.Username)
	if err != nil {
		uc.log.Error("new telegram user", "error", err, "chat_id", in.ChatID)
		return
	}
	if err := uc.repo.Save(ctx, user); err != nil {
		uc.log.Error("save telegram user", "error", err, "chat_id", in.ChatID)
		uc.send(ctx, in.ChatID, msgInternal)
		return
	}
	uc.send(ctx, in.ChatID, msgAskName)
}

func (uc *ProcessUpdate) handleStart(ctx context.Context, user *domain.TelegramUser, in IncomingMessage) {
	user.SetUsername(in.Username)
	if user.Registered() {
		_ = uc.repo.Save(ctx, *user)
		uc.send(ctx, user.ChatID(), "Ассаламу алейкум, "+user.FullName()+"! 👋\n\nВы уже зарегистрированы. Просто напишите сюда ваш вопрос — менеджер «Амана» ответит вам в этом чате.")
		return
	}
	user.RestartRegistration()
	if err := uc.repo.Save(ctx, *user); err != nil {
		uc.log.Error("save telegram user", "error", err, "chat_id", user.ChatID())
		uc.send(ctx, user.ChatID(), msgInternal)
		return
	}
	uc.send(ctx, user.ChatID(), msgAskName)
}

func (uc *ProcessUpdate) handleName(ctx context.Context, user *domain.TelegramUser, text string) {
	if err := user.RecordName(text); err != nil {
		uc.send(ctx, user.ChatID(), msgBadName)
		return
	}
	if err := uc.repo.Save(ctx, *user); err != nil {
		uc.log.Error("save telegram user", "error", err, "chat_id", user.ChatID())
		uc.send(ctx, user.ChatID(), msgInternal)
		return
	}
	uc.askPhone(ctx, user.ChatID(), user.FullName())
}

func (uc *ProcessUpdate) handlePhone(ctx context.Context, user *domain.TelegramUser, text, contact string) {
	raw := text
	if contact != "" {
		raw = contact
	}
	phone, ok := domain.NormalizePhone(raw)
	if !ok {
		_ = uc.messenger.AskForContact(ctx, user.ChatID(), msgBadPhone)
		return
	}

	// Create the CRM client and link it to the user atomically.
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		clientID, err := uc.clients.CreateClient(ctx, user.OrgID(), user.FullName(), phone)
		if err != nil {
			return err
		}
		if err := user.CompleteRegistration(phone, clientID); err != nil {
			return err
		}
		return uc.repo.Save(ctx, *user)
	})
	if err != nil {
		uc.log.Error("complete telegram registration", "error", err, "chat_id", user.ChatID())
		uc.send(ctx, user.ChatID(), msgInternal)
		return
	}
	uc.log.Info("telegram user registered", "chat_id", user.ChatID(), "client_id", user.ClientID())
	uc.send(ctx, user.ChatID(), "Спасибо, "+user.FullName()+"! Регистрация завершена. ✅\n\nТеперь просто напишите сюда ваш вопрос — менеджер «Амана» ответит вам прямо в этом чате.")
}

func (uc *ProcessUpdate) handleClientMessage(ctx context.Context, user *domain.TelegramUser, text string) {
	if text == "" {
		uc.send(ctx, user.ChatID(), msgEmptyBody)
		return
	}
	if err := uc.inbox.PostClientMessage(ctx, user.OrgID(), user.ClientID(), text); err != nil {
		uc.log.Error("relay client message", "error", err, "chat_id", user.ChatID())
		uc.send(ctx, user.ChatID(), msgInternal)
		return
	}
	uc.log.Info("client message relayed to manager inbox", "chat_id", user.ChatID(), "client_id", user.ClientID())
}

func (uc *ProcessUpdate) askName(ctx context.Context, chatID int64) {
	uc.send(ctx, chatID, msgAskName)
}

func (uc *ProcessUpdate) askPhone(ctx context.Context, chatID int64, name string) {
	if err := uc.messenger.AskForContact(ctx, chatID,
		"Рады знакомству, "+name+"!\n\nТеперь отправьте, пожалуйста, ваш контактный номер телефона.\n"+
			"Можно нажать кнопку ниже или ввести номер вручную (например: +7 928 000-00-00)."); err != nil {
		uc.log.Error("send telegram message", "error", err, "chat_id", chatID)
	}
}

func (uc *ProcessUpdate) send(ctx context.Context, chatID int64, text string) {
	if err := uc.messenger.Send(ctx, chatID, text); err != nil {
		uc.log.Error("send telegram message", "error", err, "chat_id", chatID)
	}
}
