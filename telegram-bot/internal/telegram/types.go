// Package telegram — минимальный клиент Telegram Bot API на stdlib (long polling
// getUpdates + sendMessage) и используемые типы. Сознательно без сторонних
// библиотек, чтобы модуль оставался полностью автономным.
package telegram

// Update — одно обновление из getUpdates. Нас интересуют только сообщения.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message — входящее сообщение в Telegram.
type Message struct {
	MessageID int64    `json:"message_id"`
	From      *User    `json:"from"`
	Chat      Chat     `json:"chat"`
	Text      string   `json:"text"`
	Contact   *Contact `json:"contact"` // приходит при нажатии кнопки "отправить контакт"
}

// User — отправитель сообщения в Telegram.
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// Chat — чат, в котором пришло сообщение. Для бота это личка ("private").
type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
}

// Contact — контакт, которым пользователь поделился через кнопку клавиатуры.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserID      int64  `json:"user_id"`
}

// Me — ответ getMe (используем для проверки токена на старте).
type Me struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// ReplyKeyboardMarkup — кастомная клавиатура (например, кнопка "отправить номер").
type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
}

// KeyboardButton — кнопка клавиатуры. RequestContact=true просит поделиться телефоном.
type KeyboardButton struct {
	Text           string `json:"text"`
	RequestContact bool   `json:"request_contact,omitempty"`
}

// ReplyKeyboardRemove убирает ранее показанную кастомную клавиатуру.
type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

// ContactKeyboard — клавиатура с единственной кнопкой "поделиться телефоном".
func ContactKeyboard() ReplyKeyboardMarkup {
	return ReplyKeyboardMarkup{
		Keyboard: [][]KeyboardButton{{
			{Text: "📱 Отправить мой номер", RequestContact: true},
		}},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}
