package domain

import (
	"strings"
	"time"
)

// maxMessageLen bounds a single chat message (defensive; the column is text).
const maxMessageLen = 4000

// SenderKind is who sent a message: the client or the company's staff.
type SenderKind string

const (
	SenderClient SenderKind = "client"
	SenderStaff  SenderKind = "staff"
)

func (k SenderKind) Valid() bool { return k == SenderClient || k == SenderStaff }
func (k SenderKind) String() string { return string(k) }

// Conversation is the single chat thread between a client and the company's
// staff (one per client per organization).
type Conversation struct {
	id            string
	orgID         string
	clientID      string
	createdAt     time.Time
	lastMessageAt time.Time
}

// NewConversation creates a fresh conversation for a client.
func NewConversation(id, orgID, clientID string) (Conversation, error) {
	if id == "" {
		return Conversation{}, ErrConversationIDRequired
	}
	if orgID == "" {
		return Conversation{}, ErrOrgIDRequired
	}
	if clientID == "" {
		return Conversation{}, ErrClientIDRequired
	}
	now := time.Now().UTC()
	return Conversation{id: id, orgID: orgID, clientID: clientID, createdAt: now, lastMessageAt: now}, nil
}

// RehydrateConversation rebuilds a conversation from storage.
func RehydrateConversation(id, orgID, clientID string, createdAt, lastMessageAt time.Time) Conversation {
	return Conversation{id: id, orgID: orgID, clientID: clientID, createdAt: createdAt, lastMessageAt: lastMessageAt}
}

func (c Conversation) ID() string              { return c.id }
func (c Conversation) OrgID() string            { return c.orgID }
func (c Conversation) ClientID() string         { return c.clientID }
func (c Conversation) CreatedAt() time.Time     { return c.createdAt }
func (c Conversation) LastMessageAt() time.Time { return c.lastMessageAt }

// ConversationView is a staff-facing list item: a conversation plus a preview of
// its most recent message.
type ConversationView struct {
	ConversationID string
	ClientID       string
	LastMessage    string
	LastSenderKind SenderKind
	LastMessageAt  time.Time
}

// Message is one chat message within a conversation.
type Message struct {
	id             string
	conversationID string
	orgID          string
	senderKind     SenderKind
	senderID       string
	body           string
	createdAt      time.Time
}

// NewMessage validates invariants and creates a message.
func NewMessage(id, conversationID, orgID string, kind SenderKind, senderID, body string) (Message, error) {
	if id == "" {
		return Message{}, ErrMessageIDRequired
	}
	if conversationID == "" {
		return Message{}, ErrConversationIDRequired
	}
	if orgID == "" {
		return Message{}, ErrOrgIDRequired
	}
	if !kind.Valid() {
		return Message{}, ErrInvalidSenderKind
	}
	if senderID == "" {
		return Message{}, ErrSenderIDRequired
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return Message{}, ErrMessageBodyRequired
	}
	if len(body) > maxMessageLen {
		return Message{}, ErrMessageTooLong
	}
	return Message{
		id:             id,
		conversationID: conversationID,
		orgID:          orgID,
		senderKind:     kind,
		senderID:       senderID,
		body:           body,
		createdAt:      time.Now().UTC(),
	}, nil
}

// RehydrateMessage rebuilds a message from storage.
func RehydrateMessage(id, conversationID, orgID string, kind SenderKind, senderID, body string, createdAt time.Time) Message {
	return Message{
		id:             id,
		conversationID: conversationID,
		orgID:          orgID,
		senderKind:     kind,
		senderID:       senderID,
		body:           body,
		createdAt:      createdAt,
	}
}

func (m Message) ID() string             { return m.id }
func (m Message) ConversationID() string  { return m.conversationID }
func (m Message) OrgID() string           { return m.orgID }
func (m Message) SenderKind() SenderKind  { return m.senderKind }
func (m Message) SenderID() string        { return m.senderID }
func (m Message) Body() string            { return m.body }
func (m Message) CreatedAt() time.Time    { return m.createdAt }
