// Package tgapi is a minimal Telegram Bot API client (long polling getUpdates +
// sendMessage) built on stdlib net/http. No third-party dependencies, so the
// project keeps its canonical stack (CLAUDE.md §1).
package tgapi

// Update is one item from getUpdates. We only care about messages.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message is an inbound Telegram message.
type Message struct {
	MessageID int64    `json:"message_id"`
	From      *User    `json:"from"`
	Chat      Chat     `json:"chat"`
	Text      string   `json:"text"`
	Contact   *Contact `json:"contact"`
}

// User is the sender of a message.
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// Chat is the chat a message arrived in (private chat for a bot).
type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
}

// Contact is a contact shared via a keyboard button.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserID      int64  `json:"user_id"`
}

// Me is the getMe response (used to validate the token at startup).
type Me struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// ReplyKeyboardMarkup is a custom keyboard (e.g. a "share phone" button).
type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
}

// KeyboardButton is a keyboard button; RequestContact asks for the phone number.
type KeyboardButton struct {
	Text           string `json:"text"`
	RequestContact bool   `json:"request_contact,omitempty"`
}

// ReplyKeyboardRemove hides a previously shown custom keyboard.
type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

// ContactKeyboard is a one-button keyboard requesting the user's phone number.
func ContactKeyboard() ReplyKeyboardMarkup {
	return ReplyKeyboardMarkup{
		Keyboard:        [][]KeyboardButton{{{Text: "📱 Отправить мой номер", RequestContact: true}}},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}
