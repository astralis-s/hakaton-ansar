// Package document holds the plain, infrastructure-free data structures for the
// business documents the application renders to PDF (contract agreements,
// finance reports). It depends only on the standard library, so any layer —
// including application services — may build these values without pulling in a
// PDF library. Rendering lives in internal/platform/pdf.
package document

// Contract is the denormalized data for a murabaha installment agreement.
type Contract struct {
	Number  string // short human contract number (e.g. first 8 of the id)
	OrgName string // the seller (organization)
	Date    string // creation date, formatted
	Status  string // human status label

	ClientName     string
	ClientPhone    string
	ClientDocument string

	ProductName string

	CostPrice      string // закупочная цена (себестоимость)
	Markup         string // наценка
	SalePrice      string // цена продажи = cost + markup
	DownPayment    string
	FinancedAmount string // к рассрочке
	Outstanding    string
	PaidAmount     string

	InstallmentsCount int
	Cadence           string // human cadence ("ежемесячно"/"еженедельно")
	StartDate         string

	Schedule []ContractLine
}

// ContractLine is one row of the installment schedule.
type ContractLine struct {
	Number  int
	DueDate string
	Amount  string
	Status  string // human status label
}

// FinanceReport is the denormalized data for the one-page finance report (P&L).
type FinanceReport struct {
	OrgName     string
	GeneratedAt string

	Revenue       string
	CostOfGoods   string
	GrossProfit   string
	OtherExpenses string
	NetProfit     string
	NetNegative   bool // true when net profit is below zero (render in red)
	SalesCount    int
	ExpensesCount int
	AvgSale       string // средний чек = revenue / salesCount

	Expenses []ReportExpense // manual expenses (rendered as a capped list)
}

// ReportExpense is one manual expense line in the finance report.
type ReportExpense struct {
	Category string
	Amount   string
	Date     string
}
