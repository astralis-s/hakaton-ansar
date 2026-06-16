// Package store — потокобезопасное хранилище пользователей и переписки бота с
// персистентностью в JSON-файл. Внешних зависимостей нет (только stdlib).
package store

import "time"

// SenderKind — кто отправил сообщение.
type SenderKind string

const (
	SenderClient  SenderKind = "client"  // заказчик из Telegram
	SenderManager SenderKind = "manager" // менеджер с сайта
)

// RegState — состояние конечного автомата регистрации пользователя.
type RegState string

const (
	StateNew           RegState = ""               // первый контакт, данных нет
	StateAwaitingName  RegState = "awaiting_name"  // ждём ФИО
	StateAwaitingPhone RegState = "awaiting_phone" // ждём телефон
	StateRegistered    RegState = "registered"     // данные собраны, можно общаться
)

// User — заказчик, написавший боту. ChatID — это telegram chat id (он же id ЛС).
type User struct {
	ChatID    int64     `json:"chat_id"`
	Username  string    `json:"username,omitempty"` // @username в телеграме, если есть
	FullName  string    `json:"full_name"`          // ФИО, введённое заказчиком
	Phone     string    `json:"phone"`              // контактный телефон
	State     RegState  `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Registered сообщает, завершил ли пользователь регистрацию.
func (u User) Registered() bool { return u.State == StateRegistered }

// Message — одно сообщение в переписке конкретного заказчика с менеджером.
type Message struct {
	ID        int64      `json:"id"`
	ChatID    int64      `json:"chat_id"`
	Sender    SenderKind `json:"sender"`
	Text      string     `json:"text"`
	Read      bool       `json:"read"` // для сообщений клиента: прочитано ли менеджером
	CreatedAt time.Time  `json:"created_at"`
}

// Conversation — строка списка чатов в инбоксе менеджера (производный read-model,
// не хранится отдельно). Имя — это ФИО из регистрации, а не никнейм телеграма.
type Conversation struct {
	ChatID     int64      `json:"chat_id"`
	FullName   string     `json:"full_name"`
	Phone      string     `json:"phone"`
	Username   string     `json:"username,omitempty"`
	LastText   string     `json:"last_text"`
	LastSender SenderKind `json:"last_sender"`
	LastAt     time.Time  `json:"last_at"`
	Unread     int        `json:"unread"`
}
