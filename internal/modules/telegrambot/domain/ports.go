package domain

import "context"

// UserRepository persists Telegram users and resolves the reverse client→chat
// mapping used to deliver staff replies.
type UserRepository interface {
	// GetByChatID returns the user for a chat, or ErrUserNotFound.
	GetByChatID(ctx context.Context, chatID int64) (TelegramUser, error)
	// Save upserts the user (keyed by chat id).
	Save(ctx context.Context, u TelegramUser) error
	// FindChatByClient returns the registered Telegram chat id linked to a CRM
	// client, or found=false when the client has no Telegram chat.
	FindChatByClient(ctx context.Context, orgID, clientID string) (chatID int64, found bool, err error)
}

// OrgResolver returns the organization the single-tenant bot serves (the only
// organization in the database) when it is not pinned in configuration.
type OrgResolver interface {
	DefaultOrgID(ctx context.Context) (string, error)
}

// ClientDirectory creates a CRM client for a freshly registered Telegram user.
// Implemented in infra over the crm context's CreateClient use-case.
type ClientDirectory interface {
	CreateClient(ctx context.Context, orgID, fullName, phone string) (clientID string, err error)
}

// ChatInbox appends an inbound client message to the staff chat. Implemented in
// infra over the portal context's SendMessage use-case (so the message shows up
// in the manager inbox under the client's full name).
type ChatInbox interface {
	PostClientMessage(ctx context.Context, orgID, clientID, body string) error
}

// Messenger sends messages back to a Telegram chat. Implemented in infra over the
// Bot API client; the domain/app never touch Telegram types directly.
type Messenger interface {
	// Send delivers plain text (and clears any custom keyboard).
	Send(ctx context.Context, chatID int64, text string) error
	// AskForContact delivers text together with a "share phone" keyboard button.
	AskForContact(ctx context.Context, chatID int64, text string) error
}

// TxManager runs a function inside a single database transaction.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
