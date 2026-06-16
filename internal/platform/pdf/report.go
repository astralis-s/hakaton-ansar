package pdf

import (
	"bytes"

	"github.com/go-pdf/fpdf"

	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
)

// maxExpenseRows caps the manual-expense list so the report stays on one page.
const maxExpenseRows = 8

// RenderFinanceReport produces a one-page A4 P&L report.
func RenderFinanceReport(r document.FinanceReport) ([]byte, error) {
	doc := New("P")
	doc.SetTitle("Финансовый отчёт", true)
	doc.SetAuthor("Амана CRM", true)
	doc.SetMargins(16, 14, 16)
	doc.SetAutoPageBreak(false, 0) // strictly one page
	doc.SetFooterFunc(func() { footer(doc) })
	doc.AddPage()

	// --- Header ---------------------------------------------------------------
	doc.SetFont(Font, "B", 17)
	doc.SetTextColor(20, 24, 22)
	doc.CellFormat(0, 9, "Финансовый отчёт", "", 1, "L", false, 0, "")
	doc.SetFont(Font, "", 10.5)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(0, 6, orDash(r.OrgName)+"     •     сформирован "+orDash(r.GeneratedAt), "", 1, "L", false, 0, "")
	doc.Ln(2)
	doc.SetDrawColor(accentR, accentG, accentB)
	y := doc.GetY()
	doc.SetLineWidth(0.6)
	doc.Line(16, y, 194, y)
	doc.SetLineWidth(0.2)
	doc.Ln(4)

	// --- Headline net profit --------------------------------------------------
	netR, netG, netB := accentR, accentG, accentB
	if r.NetNegative {
		netR, netG, netB = 168, 64, 79 // red
	}
	hy := doc.GetY()
	doc.SetFillColor(headBgR, headBgG, headBgB)
	doc.Rect(16, hy, 178, 20, "F")
	doc.SetXY(20, hy+3)
	doc.SetFont(Font, "", 10.5)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(0, 6, "Чистая прибыль за период (доход − все расходы)", "", 2, "L", false, 0, "")
	doc.SetX(20)
	doc.SetFont(Font, "B", 20)
	doc.SetTextColor(netR, netG, netB)
	doc.CellFormat(0, 9, money(r.NetProfit), "", 0, "L", false, 0, "")
	doc.SetY(hy + 24)

	// --- KPI cards ------------------------------------------------------------
	cy := doc.GetY()
	cardW, gap := 56.67, 4.0
	statCard(doc, 16, cy, cardW, "Выручка", money(r.Revenue), false)
	statCard(doc, 16+cardW+gap, cy, cardW, "Валовая прибыль", money(r.GrossProfit), true)
	statCard(doc, 16+2*(cardW+gap), cy, cardW, "Прочие расходы", money(r.OtherExpenses), false)
	doc.SetY(cy + 22)

	// --- Доходы ---------------------------------------------------------------
	section(doc, "Доходы от продаж")
	kv(doc, "Выручка (сумма продаж)", money(r.Revenue))
	kv(doc, "Количество продаж", itoa(r.SalesCount))
	kv(doc, "Средний чек", money(r.AvgSale))
	kv(doc, "Себестоимость проданного", money(r.CostOfGoods))
	kvStrong(doc, "Валовая прибыль", money(r.GrossProfit))
	doc.Ln(2)

	// --- Расходы --------------------------------------------------------------
	section(doc, "Расходы")
	kv(doc, "Себестоимость проданных товаров", money(r.CostOfGoods))
	shown := r.Expenses
	hidden := 0
	if len(shown) > maxExpenseRows {
		hidden = len(shown) - maxExpenseRows
		shown = shown[:maxExpenseRows]
	}
	for _, e := range shown {
		label := orDash(e.Category)
		if e.Date != "" {
			label += " (" + e.Date + ")"
		}
		kv(doc, label, money(e.Amount))
	}
	if hidden > 0 {
		doc.SetFont(Font, "", 9.5)
		doc.SetTextColor(mutedR, mutedG, mutedB)
		doc.CellFormat(0, 6, "… и ещё "+itoa(hidden)+" расходов", "", 1, "L", false, 0, "")
	}
	kv(doc, "Прочие расходы, итого", money(r.OtherExpenses))
	kvStrong(doc, "Чистая прибыль", money(r.NetProfit))

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// statCard draws a filled KPI card with a label and value.
func statCard(doc *fpdf.Fpdf, x, y, w float64, label, value string, accent bool) {
	doc.SetFillColor(247, 249, 248)
	doc.SetDrawColor(lineR, lineG, lineB)
	doc.Rect(x, y, w, 18, "FD")
	doc.SetXY(x+3, y+3)
	doc.SetFont(Font, "", 8.5)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(w-6, 4, label, "", 2, "L", false, 0, "")
	doc.SetX(x + 3)
	doc.SetFont(Font, "B", 11.5)
	if accent {
		doc.SetTextColor(accentR, accentG, accentB)
	} else {
		doc.SetTextColor(30, 34, 32)
	}
	doc.CellFormat(w-6, 7, value, "", 0, "L", false, 0, "")
}
