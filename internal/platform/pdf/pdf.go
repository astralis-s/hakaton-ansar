// Package pdf renders business documents (contract agreements, finance reports)
// to PDF. It bundles the Cyrillic-capable Liberation Sans font (SIL OFL) so the
// Russian text renders correctly, and works with CGO disabled.
package pdf

import (
	_ "embed"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/LiberationSans-Regular.ttf
var fontRegular []byte

//go:embed fonts/LiberationSans-Bold.ttf
var fontBold []byte

// Font is the registered family name used throughout the documents.
const Font = "Liberation"

// New returns an A4 fpdf document with the Liberation Sans font registered
// (regular + bold) and selected. orientation is "P" (portrait) or "L".
func New(orientation string) *fpdf.Fpdf {
	doc := fpdf.New(orientation, "mm", "A4", "")
	doc.AddUTF8FontFromBytes(Font, "", fontRegular)
	doc.AddUTF8FontFromBytes(Font, "B", fontBold)
	doc.SetFont(Font, "", 11)
	doc.SetAutoPageBreak(true, 15)
	return doc
}
