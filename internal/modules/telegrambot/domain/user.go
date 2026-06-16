// Package domain is the telegrambot bounded context core: the Telegram user
// aggregate with its registration FSM, the registration value objects, the
// typed errors and the ports the module depends on. Imports only stdlib.
package domain

import (
	"strings"
	"time"
)

// TelegramUser is a customer who contacts the company through the Telegram bot.
// ChatID is the Telegram private-chat id (stable per user). The aggregate owns
// the registration flow: a user must provide a valid full name and phone (which
// links them to a CRM client) before they can chat with a manager.
type TelegramUser struct {
	chatID    int64
	orgID     string
	clientID  string // linked CRM client; empty until registration completes
	username  string
	fullName  string
	phone     string
	state     RegState
	createdAt time.Time
	updatedAt time.Time
}

// NewTelegramUser starts a fresh registration (state = awaiting name).
func NewTelegramUser(chatID int64, orgID, username string) (TelegramUser, error) {
	if chatID == 0 {
		return TelegramUser{}, ErrChatIDRequired
	}
	if orgID == "" {
		return TelegramUser{}, ErrOrgIDRequired
	}
	now := time.Now().UTC()
	return TelegramUser{
		chatID:    chatID,
		orgID:     orgID,
		username:  strings.TrimSpace(username),
		state:     StateAwaitingName,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RehydrateTelegramUser rebuilds a user from trusted storage.
func RehydrateTelegramUser(chatID int64, orgID, clientID, username, fullName, phone string, state RegState, createdAt, updatedAt time.Time) TelegramUser {
	return TelegramUser{
		chatID:    chatID,
		orgID:     orgID,
		clientID:  clientID,
		username:  username,
		fullName:  fullName,
		phone:     phone,
		state:     state,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// RecordName validates and stores the full name, advancing to the phone step.
func (u *TelegramUser) RecordName(raw string) error {
	if u.state != StateAwaitingName {
		return ErrInvalidState
	}
	name := NormalizeFullName(raw)
	if !IsValidFullName(name) {
		return ErrInvalidFullName
	}
	u.fullName = name
	u.state = StateAwaitingPhone
	u.touch()
	return nil
}

// CompleteRegistration stores the phone, links the CRM client and marks the user
// registered. The phone must already be normalized/validated by the caller.
func (u *TelegramUser) CompleteRegistration(phone, clientID string) error {
	if u.state != StateAwaitingPhone {
		return ErrInvalidState
	}
	if strings.TrimSpace(phone) == "" {
		return ErrInvalidPhone
	}
	if clientID == "" {
		return ErrClientIDRequired
	}
	u.phone = phone
	u.clientID = clientID
	u.state = StateRegistered
	u.touch()
	return nil
}

// RestartRegistration resets an unfinished registration back to the name step
// (used when an unregistered user sends /start again).
func (u *TelegramUser) RestartRegistration() {
	u.clientID = ""
	u.fullName = ""
	u.phone = ""
	u.state = StateAwaitingName
	u.touch()
}

// SetUsername updates the cached Telegram @username (metadata only).
func (u *TelegramUser) SetUsername(username string) {
	if username = strings.TrimSpace(username); username != "" {
		u.username = username
	}
}

func (u TelegramUser) Registered() bool { return u.state == StateRegistered }

func (u TelegramUser) ChatID() int64        { return u.chatID }
func (u TelegramUser) OrgID() string        { return u.orgID }
func (u TelegramUser) ClientID() string     { return u.clientID }
func (u TelegramUser) Username() string     { return u.username }
func (u TelegramUser) FullName() string     { return u.fullName }
func (u TelegramUser) Phone() string        { return u.phone }
func (u TelegramUser) State() RegState      { return u.state }
func (u TelegramUser) CreatedAt() time.Time { return u.createdAt }
func (u TelegramUser) UpdatedAt() time.Time { return u.updatedAt }

func (u *TelegramUser) touch() { u.updatedAt = time.Now().UTC() }
