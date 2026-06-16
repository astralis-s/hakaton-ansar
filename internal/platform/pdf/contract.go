package pdf

import (
	"bytes"

	"github.com/go-pdf/fpdf"

	"github.com/astralis-s/hakaton-ansar/internal/shared/document"
)

// disclosures are the murabaha / riba-prohibition clauses printed on every
// contract — a prototype disclosure block adapted to the Amana model.
var disclosures = []string{
	"Настоящий договор заключён по исламской модели купли-продажи в рассрочку «Мурабаха». Продавец раскрывает покупателю себестоимость товара и размер своей наценки (прибыли); итоговая цена продажи согласована сторонами до подписания.",
	"Цена продажи зафиксирована в момент заключения договора и НЕ изменяется со временем. Проценты (риба) не начисляются ни при каких условиях.",
	"При просрочке платежа сумма долга не увеличивается. Дополнительные проценты или штрафы, увеличивающие задолженность, не применяются — это прямой запрет (риба).",
	"Покупатель вправе досрочно погасить остаток задолженности в любой момент без каких-либо доплат и комиссий.",
	"Право собственности на товар переходит к покупателю в момент продажи. Договор является реальной сделкой купли-продажи актива, а не процентным займом.",
	"График платежей, приведённый ниже, является неотъемлемой частью договора. Стороны подтверждают согласие с суммами и датами платежей.",
}

// RenderContract produces the PDF bytes of a murabaha installment agreement.
func RenderContract(c document.Contract) ([]byte, error) {
	doc := New("P")
	doc.SetTitle("Договор рассрочки "+c.Number, true)
	doc.SetAuthor("Амана CRM", true)
	doc.SetMargins(18, 16, 18)
	doc.SetFooterFunc(func() { footer(doc) })
	doc.AddPage()

	// --- Header ---------------------------------------------------------------
	doc.SetFont(Font, "B", 16)
	doc.SetTextColor(20, 24, 22)
	doc.CellFormat(0, 8, "Договор рассрочки № "+c.Number, "", 1, "L", false, 0, "")
	doc.SetFont(Font, "", 10.5)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(0, 6, "по исламской модели «Мурабаха» — честная рассрочка без рибы", "", 1, "L", false, 0, "")
	doc.SetFont(Font, "", 10)
	doc.CellFormat(0, 6, "Продавец: "+orDash(c.OrgName)+"     •     Дата: "+orDash(c.Date)+"     •     Статус: "+orDash(c.Status), "", 1, "L", false, 0, "")
	doc.Ln(3)

	// --- Parties --------------------------------------------------------------
	section(doc, "Стороны договора")
	kv(doc, "Продавец", orDash(c.OrgName))
	kv(doc, "Покупатель", orDash(c.ClientName))
	kv(doc, "Телефон покупателя", orDash(c.ClientPhone))
	kv(doc, "Документ покупателя", orDash(c.ClientDocument))
	doc.Ln(2)

	// --- Subject (cost + markup disclosure — the heart of murabaha) -----------
	section(doc, "Предмет договора и раскрытие цены")
	kv(doc, "Товар", orDash(c.ProductName))
	kv(doc, "Себестоимость (закупка)", money(c.CostPrice))
	kv(doc, "Наценка продавца (прибыль)", money(c.Markup))
	kvStrong(doc, "Цена продажи", money(c.SalePrice))
	doc.Ln(2)

	// --- Terms ----------------------------------------------------------------
	section(doc, "Условия рассрочки")
	twoCol(doc,
		[][2]string{
			{"Первоначальный взнос", money(c.DownPayment)},
			{"Сумма рассрочки", money(c.FinancedAmount)},
			{"Оплачено", money(c.PaidAmount)},
		},
		[][2]string{
			{"Количество платежей", itoa(c.InstallmentsCount) + " (" + orDash(c.Cadence) + ")"},
			{"Дата первого платежа", orDash(c.StartDate)},
			{"Остаток задолженности", money(c.Outstanding)},
		},
	)
	doc.Ln(2)

	// --- Schedule -------------------------------------------------------------
	section(doc, "График платежей")
	scheduleTable(doc, c.Schedule)
	doc.Ln(2)

	// --- Disclosures ----------------------------------------------------------
	section(doc, "Существенные условия и раскрытия")
	doc.SetFont(Font, "", 9.5)
	doc.SetTextColor(40, 44, 42)
	for i, d := range disclosures {
		doc.MultiCell(0, 5, itoa(i+1)+". "+d, "", "L", false)
		doc.Ln(0.5)
	}
	doc.Ln(4)

	// --- Signatures -----------------------------------------------------------
	signatures(doc)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- layout helpers ---------------------------------------------------------

func section(doc *fpdf.Fpdf, title string) {
	doc.SetFillColor(headBgR, headBgG, headBgB)
	doc.SetTextColor(accentR, accentG, accentB)
	doc.SetFont(Font, "B", 11.5)
	doc.CellFormat(0, 8, title, "", 1, "L", true, 0, "")
	doc.Ln(1.5)
}

func kv(doc *fpdf.Fpdf, label, value string) {
	doc.SetFont(Font, "", 10)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(58, 6, label, "", 0, "L", false, 0, "")
	doc.SetTextColor(30, 34, 32)
	doc.CellFormat(0, 6, value, "", 1, "L", false, 0, "")
}

func kvStrong(doc *fpdf.Fpdf, label, value string) {
	doc.SetFont(Font, "", 10)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(58, 7, label, "", 0, "L", false, 0, "")
	doc.SetFont(Font, "B", 12)
	doc.SetTextColor(accentR, accentG, accentB)
	doc.CellFormat(0, 7, value, "", 1, "L", false, 0, "")
}

// twoCol prints two columns of key/value pairs side by side.
func twoCol(doc *fpdf.Fpdf, left, right [][2]string) {
	n := len(left)
	if len(right) > n {
		n = len(right)
	}
	colW := 87.0
	for i := 0; i < n; i++ {
		y := doc.GetY()
		if i < len(left) {
			cellKV(doc, 18, y, colW, left[i][0], left[i][1])
		}
		if i < len(right) {
			cellKV(doc, 18+colW, y, colW, right[i][0], right[i][1])
		}
		doc.SetY(y + 6)
	}
}

func cellKV(doc *fpdf.Fpdf, x, y, w float64, label, value string) {
	doc.SetXY(x, y)
	doc.SetFont(Font, "", 10)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(40, 6, label, "", 0, "L", false, 0, "")
	doc.SetTextColor(30, 34, 32)
	doc.CellFormat(w-40, 6, value, "", 0, "L", false, 0, "")
}

func scheduleTable(doc *fpdf.Fpdf, lines []document.ContractLine) {
	widths := []float64{16, 60, 50, 0} // № / дата / сумма / статус (0 = rest)
	headers := []string{"№", "Дата платежа", "Сумма", "Статус"}
	doc.SetFont(Font, "B", 9.5)
	doc.SetFillColor(headBgR, headBgG, headBgB)
	doc.SetTextColor(accentR, accentG, accentB)
	for i, h := range headers {
		doc.CellFormat(widths[i], 7, h, "", 0, "L", true, 0, "")
	}
	doc.Ln(-1)
	doc.SetFont(Font, "", 9.5)
	doc.SetTextColor(40, 44, 42)
	doc.SetDrawColor(lineR, lineG, lineB)
	for _, l := range lines {
		doc.CellFormat(widths[0], 6.5, itoa(l.Number), "B", 0, "L", false, 0, "")
		doc.CellFormat(widths[1], 6.5, orDash(l.DueDate), "B", 0, "L", false, 0, "")
		doc.CellFormat(widths[2], 6.5, money(l.Amount), "B", 0, "L", false, 0, "")
		doc.CellFormat(widths[3], 6.5, orDash(l.Status), "B", 1, "L", false, 0, "")
	}
}

func signatures(doc *fpdf.Fpdf) {
	doc.Ln(8)
	y := doc.GetY()
	doc.SetDrawColor(60, 64, 62)
	doc.Line(18, y, 88, y)
	doc.Line(122, y, 192, y)
	doc.SetXY(18, y+1)
	doc.SetFont(Font, "", 9)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(70, 5, "Продавец (подпись)", "", 0, "C", false, 0, "")
	doc.SetXY(122, y+1)
	doc.CellFormat(70, 5, "Покупатель (подпись)", "", 0, "C", false, 0, "")
}

func footer(doc *fpdf.Fpdf) {
	doc.SetY(-12)
	doc.SetFont(Font, "", 8)
	doc.SetTextColor(mutedR, mutedG, mutedB)
	doc.CellFormat(0, 6, "Сформировано в CRM «Амана» — рассрочка без рибы. Документ-прототип.", "", 0, "L", false, 0, "")
	doc.CellFormat(0, 6, "Стр. "+itoa(doc.PageNo()), "", 0, "R", false, 0, "")
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
