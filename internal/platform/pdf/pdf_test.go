package pdf

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCyrillicRenders confirms the bundled font produces a valid PDF with
// Cyrillic text without error (the core risk of the whole feature).
func TestCyrillicRenders(t *testing.T) {
	doc := New("P")
	doc.AddPage()
	doc.SetFont(Font, "B", 16)
	doc.Cell(0, 10, "Договор рассрочки (мурабаха)")
	doc.Ln(12)
	doc.SetFont(Font, "", 11)
	doc.MultiCell(0, 6, "Продавец продаёт товар покупателю в рассрочку без рибы. Цена зафиксирована, долг не растёт со временем. Тест кириллицы: ЧёЖэ.", "", "L", false)

	var buf bytes.Buffer
	require.NoError(t, doc.Output(&buf))
	out := buf.Bytes()
	assert.True(t, bytes.HasPrefix(out, []byte("%PDF")), "output must be a PDF")
	assert.Greater(t, len(out), 2000, "a PDF with an embedded font subset should be non-trivial")
}
