package pdf

import "strings"

// brand colours (Amana green).
var (
	accentR, accentG, accentB    = 10, 157, 108
	mutedR, mutedG, mutedB       = 110, 120, 116
	lineR, lineG, lineB          = 222, 226, 224
	headBgR, headBgG, headBgB    = 240, 247, 244
)

// money formats a decimal string ("100000.00", "-2250.00") for display:
// "100 000,00 ₽" → here without the ruble glyph for font safety: "100 000,00 руб.".
func money(s string) string {
	if s == "" {
		return "—"
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, dec := s, "00"
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, dec = s[:i], s[i+1:]
	}
	if len(dec) == 1 {
		dec += "0"
	}
	out := group3(intPart) + "," + dec + " руб."
	if neg {
		out = "−" + out
	}
	return out
}

// group3 inserts spaces between thousands: "100000" → "100 000".
func group3(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	first := n % 3
	if first > 0 {
		b.WriteString(s[:first])
	}
	for i := first; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
