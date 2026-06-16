package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// RegState — состояние конечного автомата регистрации Telegram-пользователя.
type RegState string

const (
	StateNew           RegState = ""               // транзиентное: записи ещё нет
	StateAwaitingName  RegState = "awaiting_name"  // ждём ФИО
	StateAwaitingPhone RegState = "awaiting_phone" // ждём телефон
	StateRegistered    RegState = "registered"     // данные собраны, привязан клиент CRM
)

const (
	minNameRunes = 3
	maxNameRunes = 120
)

// NormalizeFullName схлопывает пробелы и обрезает строку по краям.
func NormalizeFullName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// IsValidFullName проверяет, что строка похожа на ФИО: минимум два слова
// (фамилия и имя), есть буквы, разумная длина и это не команда.
func IsValidFullName(s string) bool {
	if strings.HasPrefix(s, "/") {
		return false
	}
	n := utf8.RuneCountInString(s)
	if n < minNameRunes || n > maxNameRunes {
		return false
	}
	if len(strings.Fields(s)) < 2 {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// NormalizePhone приводит номер к виду +<цифры>. Для российских номеров (11 цифр
// c 7/8 или 10 цифр) подставляет код +7. Возвращает false, если это не похоже на
// телефон.
func NormalizePhone(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	hasPlus := strings.HasPrefix(raw, "+")

	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()

	switch {
	case len(d) == 11 && (d[0] == '7' || d[0] == '8'):
		return "+7" + d[1:], true
	case len(d) == 10:
		return "+7" + d, true
	case hasPlus && len(d) >= 10 && len(d) <= 15:
		return "+" + d, true
	case !hasPlus && len(d) >= 11 && len(d) <= 15:
		return "+" + d, true
	default:
		return "", false
	}
}
