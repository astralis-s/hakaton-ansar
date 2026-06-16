// Package seed inserts demo data (Grozny flavour) so the app is presentable
// immediately after `docker compose up`. It is idempotent: it does nothing if an
// organization already exists. It builds the data through the same domain
// constructors and application use-cases as the running app, so the demo is
// internally consistent (bcrypt passwords, real murabaha schedules, etc.).
package seed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	catalogapp "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/app"
	catalogdomain "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	cataloginfra "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/infra"
	crmdomain "github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	crminfra "github.com/astralis-s/hakaton-ansar/internal/modules/crm/infra"
	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	financinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra"
	iamdomain "github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	iaminfra "github.com/astralis-s/hakaton-ansar/internal/modules/iam/infra"
	ledgerapp "github.com/astralis-s/hakaton-ansar/internal/modules/ledger/app"
	ledgerinfra "github.com/astralis-s/hakaton-ansar/internal/modules/ledger/infra"
	portalapp "github.com/astralis-s/hakaton-ansar/internal/modules/portal/app"
	portaldomain "github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
	portalinfra "github.com/astralis-s/hakaton-ansar/internal/modules/portal/infra"
	schedulingapp "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/app"
	schedulingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	schedulinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/infra"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/shared/money"
)

// Demo credentials (printed at boot; documented in README).
const (
	OwnerEmail    = "owner@amana.ru"
	OwnerPassword = "owner12345"
	ManagerEmail  = "manager@amana.ru"
	ManagerPass   = "manager12345"
	DemoAPIKey    = "amana_demo_marketplace_key_2026"
)

// Config carries the prayer settings the seeder needs for the reminder slots.
type Config struct {
	Lat      float64
	Lon      float64
	Madhab   string
	Method   string
	Timezone *time.Location
}

// Run seeds demo data if the database is empty.
func Run(ctx context.Context, pool *pgxpool.Pool, cfg Config, log *slog.Logger) error {
	orgRepo := iaminfra.NewOrganizationRepository(pool)
	count, err := orgRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("seed: count organizations: %w", err)
	}
	if count > 0 {
		log.Info("seed: skipped (database already initialized)")
		return nil
	}

	userRepo := iaminfra.NewUserRepository(pool)
	keyRepo := iaminfra.NewApiKeyRepository(pool)
	hasher := iaminfra.NewBcryptHasher()
	productRepo := cataloginfra.NewProductRepository(pool)
	stockRepo := cataloginfra.NewStockRepository(pool)
	clientRepo := crminfra.NewClientRepository(pool)
	contractRepo := financinginfra.NewContractRepository(pool)
	reminderRepo := schedulinginfra.NewReminderRepository(pool)
	tx := database.NewTxManager(pool)

	// --- Organization + users -------------------------------------------------
	org, err := iamdomain.NewOrganization(uuid.NewString(), "Грозный Мебель", "RUB")
	if err != nil {
		return seedErr("organization", err)
	}
	if _, err := orgRepo.Create(ctx, org); err != nil {
		return seedErr("create organization", err)
	}

	owner, err := newUser(org.ID(), "Адам Магомадов", OwnerEmail, OwnerPassword, iamdomain.RoleOwner, hasher)
	if err != nil {
		return seedErr("owner", err)
	}
	if _, err := userRepo.Create(ctx, owner); err != nil {
		return seedErr("create owner", err)
	}
	manager, err := newUser(org.ID(), "Иса Дудаев", ManagerEmail, ManagerPass, iamdomain.RoleManager, hasher)
	if err != nil {
		return seedErr("manager", err)
	}
	if _, err := userRepo.Create(ctx, manager); err != nil {
		return seedErr("create manager", err)
	}

	// Known demo API key for the live Swagger demo.
	apiKey, err := iamdomain.NewApiKey(uuid.NewString(), org.ID(), "Демо-маркетплейс", DemoAPIKey[:14], iamdomain.HashAPIKey(DemoAPIKey))
	if err != nil {
		return seedErr("api key", err)
	}
	if _, err := keyRepo.Create(ctx, apiKey); err != nil {
		return seedErr("create api key", err)
	}

	// --- Clients --------------------------------------------------------------
	clients := []struct{ name, phone, doc string }{
		{"Магомед Алиев", "+7 928 000-11-22", "96 00 123456"},
		{"Хеда Юсупова", "+7 928 000-33-44", "96 01 654321"},
		{"Рустам Бараев", "+7 928 000-55-66", "96 02 112233"},
		{"Зарема Бисултанова", "+7 928 000-77-88", "96 03 778899"},
		{"Тамерлан Мусаев", "+7 928 000-99-00", "96 04 445566"},
		{"Мадина Абдулкеримова", "+7 928 000-12-34", "96 05 223344"},
	}
	clientIDs := make([]string, 0, len(clients))
	for _, c := range clients {
		cl, err := crmdomain.NewClient(uuid.NewString(), org.ID(), c.name, c.phone, c.doc)
		if err != nil {
			return seedErr("client", err)
		}
		created, err := clientRepo.Create(ctx, cl)
		if err != nil {
			return seedErr("create client", err)
		}
		clientIDs = append(clientIDs, created.ID())
	}

	// --- Products (mixed halal statuses) + initial stock receipts ------------
	// Each product enters the warehouse with zero stock, then a logged receipt
	// brings it on hand — so every unit is traceable in товарооборот.
	adjustStockUC := catalogapp.NewAdjustStock(stockRepo, tx)
	products := []struct {
		name, category, cost string
		status               catalogdomain.HalalStatus
		stock                int
	}{
		{"Диван угловой «Кавказ»", "Мебель", "85000.00", catalogdomain.HalalStatusHalal, 12},
		{"Кухонный гарнитур «Беркат»", "Мебель", "120000.00", catalogdomain.HalalStatusHalal, 8},
		{"Холодильник Bosch", "Техника", "65000.00", catalogdomain.HalalStatusHalal, 15},
		{"Подарочный набор (сомнительный поставщик)", "Прочее", "5000.00", catalogdomain.HalalStatusDoubtful, 6},
		{"Вино столовое", "Напитки", "3000.00", catalogdomain.HalalStatusHaram, 4},
	}
	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		cost, _ := money.FromString(p.cost, "RUB")
		prod, err := catalogdomain.NewProduct(uuid.NewString(), org.ID(), p.name, p.category, cost, p.status, 0)
		if err != nil {
			return seedErr("product", err)
		}
		created, err := productRepo.Create(ctx, prod)
		if err != nil {
			return seedErr("create product", err)
		}
		if _, _, err := adjustStockUC.Execute(ctx, catalogapp.AdjustStockInput{
			OrgID:     org.ID(),
			ProductID: created.ID(),
			Delta:     p.stock,
			Reason:    string(catalogdomain.StockReceipt),
			Note:      "Поступление на склад (стартовый запас)",
		}); err != nil {
			return seedErr("stock receipt", err)
		}
		productIDs = append(productIDs, created.ID())
	}

	// --- Contracts (different states) ----------------------------------------
	productReader := financinginfra.NewProductReader(productRepo)
	clientReader := financinginfra.NewClientReader(clientRepo)
	stockReserver := financinginfra.NewStockReserver(stockRepo)
	createUC := financingapp.NewCreateContract(contractRepo, productReader, clientReader, stockReserver, tx)

	now := time.Now()

	type contractSeed struct {
		label            string
		clientIdx        int
		productIdx       int
		cost             string
		markup           string
		down             string
		startDate        time.Time
		paidInstallments int
	}
	seeds := []contractSeed{
		{
			label:            "contract adam 1",
			clientIdx:        0,
			productIdx:       0,
			cost:             "85000.00",
			markup:           "15000.00",
			down:             "10000.00",
			startDate:        weekDate(now, time.Monday).AddDate(0, 0, -9*7),
			paidInstallments: 6,
		},
		{
			label:            "contract adam 2",
			clientIdx:        1,
			productIdx:       1,
			cost:             "120000.00",
			markup:           "18000.00",
			down:             "18000.00",
			startDate:        weekDate(now, time.Wednesday).AddDate(0, 0, -8*7),
			paidInstallments: 7,
		},
		{
			label:            "contract adam 3",
			clientIdx:        2,
			productIdx:       2,
			cost:             "65000.00",
			markup:           "9750.00",
			down:             "15000.00",
			startDate:        weekDate(now, time.Thursday).AddDate(0, 0, -8*7),
			paidInstallments: 8,
		},
		{
			label:            "contract adam 4",
			clientIdx:        3,
			productIdx:       0,
			cost:             "91000.00",
			markup:           "14000.00",
			down:             "15000.00",
			startDate:        weekDate(now, time.Friday).AddDate(0, 0, -8*7),
			paidInstallments: 8,
		},
		{
			label:            "contract adam 5",
			clientIdx:        4,
			productIdx:       1,
			cost:             "134000.00",
			markup:           "19000.00",
			down:             "24000.00",
			startDate:        weekDate(now, time.Saturday).AddDate(0, 0, -8*7),
			paidInstallments: 8,
		},
		{
			label:            "contract adam 6",
			clientIdx:        5,
			productIdx:       2,
			cost:             "78000.00",
			markup:           "12000.00",
			down:             "10000.00",
			startDate:        weekDate(now, time.Sunday).AddDate(0, 0, -8*7),
			paidInstallments: 8,
		},
	}

	seededContracts := make([]*financingdomain.Contract, 0, len(seeds))
	for _, s := range seeds {
		contract, err := createUC.Execute(ctx, financingapp.CreateContractInput{
			OrgID:        org.ID(),
			ClientID:     clientIDs[s.clientIdx],
			ProductID:    productIDs[s.productIdx],
			CostPrice:    rub(s.cost),
			Markup:       markup(s.markup),
			DownPayment:  rub(s.down),
			Installments: 10,
			Cadence:      financingdomain.CadenceWeekly,
			StartDate:    s.startDate,
		})
		if err != nil {
			return seedErr(s.label, err)
		}

		if err := seedHistoricInstallmentPayments(ctx, tx, contractRepo, contract, s.paidInstallments); err != nil {
			return seedErr(s.label+" payments", err)
		}
		seededContracts = append(seededContracts, contract)
	}

	// --- Manual expenses (расходы) -------------------------------------------
	// Income (продажа − покупка) is derived from the contracts above; these are
	// the other business costs that the owner records by hand.
	expenseRepo := ledgerinfra.NewExpenseRepository(pool)
	createExpenseUC := ledgerapp.NewCreateExpense(expenseRepo)
	expenses := []struct {
		category, amount, note string
		daysAgo                int
	}{
		{"Аренда", "45000.00", "Аренда торгового зала и склада", 20},
		{"Ремонт", "12000.00", "Ремонт витрины после доставки", 12},
		{"Логистика", "8000.00", "Доставка мебели клиентам", 6},
		{"Реклама", "15000.00", "Продвижение в соцсетях", 3},
	}
	for _, e := range expenses {
		if _, err := createExpenseUC.Execute(ctx, ledgerapp.CreateExpenseInput{
			OrgID:    org.ID(),
			Category: e.category,
			Amount:   rub(e.amount),
			Note:     e.note,
			SpentAt:  now.AddDate(0, 0, -e.daysAgo),
		}); err != nil {
			return seedErr("expense", err)
		}
	}

	chatRepo := portalinfra.NewChatRepository(pool)
	sendMsgUC := portalapp.NewSendMessage(chatRepo, tx)
	demoChats := []struct {
		clientIdx int
		kind      portaldomain.SenderKind
		senderID  string
		body      string
	}{
		{0, portaldomain.SenderClient, clientIDs[0], "Ассаламу алейкум. Хотел уточнить, когда лучше внести следующий платеж?"},
		{0, portaldomain.SenderStaff, owner.ID(), "Ва алейкум ассалам. На этой неделе до пятницы, чтобы не ушло в просрочку."},
		{1, portaldomain.SenderClient, clientIDs[1], "Можно ли перенести встречу по кухонному гарнитуру на вечер после намаза?"},
		{1, portaldomain.SenderStaff, owner.ID(), "Да, поставил на 19:30. Если что, напишите сюда."},
		{2, portaldomain.SenderStaff, owner.ID(), "Напоминаю: завтра ожидается платеж по договору. Если удобно, подтвердите время."},
	}
	for _, msg := range demoChats {
		if _, err := sendMsgUC.Execute(ctx, portalapp.SendMessageInput{
			OrgID:      org.ID(),
			ClientID:   clientIDs[msg.clientIdx],
			SenderKind: msg.kind,
			SenderID:   msg.senderID,
			Body:       msg.body,
		}); err != nil {
			return seedErr("chat message", err)
		}
	}

	// --- Reminders (namaz-aware) ---------------------------------------------
	loc := schedulingdomain.Location{Lat: cfg.Lat, Lon: cfg.Lon, TZ: cfg.Timezone}
	provider := schedulinginfra.NewPrayerProvider(loc, cfg.Madhab, cfg.Method)
	scheduler := schedulingdomain.NewScheduler(provider, schedulingdomain.DefaultPolicy(), loc)
	scheduleUC := schedulingapp.NewScheduleReminder(scheduler, reminderRepo)

	reminders := []schedulingapp.ScheduleReminderInput{
		{OrgID: org.ID(), Type: "call", ClientID: clientIDs[0], ContractID: seededContracts[0].ID(), Note: "Подтвердить готовность к оплате", DesiredAt: todayAt(cfg.Timezone, 9, 30), DurationMinutes: 15},
		{OrgID: org.ID(), Type: "payment_followup", ClientID: clientIDs[1], ContractID: seededContracts[1].ID(), Note: "Напомнить о просроченном платеже", DesiredAt: todayAt(cfg.Timezone, 13, 0), DurationMinutes: 20},
		{OrgID: org.ID(), Type: "delivery", ClientID: clientIDs[2], ContractID: seededContracts[2].ID(), Note: "Согласовать доставку холодильника", DesiredAt: todayAt(cfg.Timezone, 16, 5), DurationMinutes: 45},
		{OrgID: org.ID(), Type: "call", ClientID: clientIDs[3], ContractID: seededContracts[3].ID(), Note: "Проверить поступление перевода", DesiredAt: todayAt(cfg.Timezone, 18, 10), DurationMinutes: 10},
	}
	for i, in := range reminders {
		if _, err := scheduleUC.Execute(ctx, in); err != nil {
			return seedErr(fmt.Sprintf("reminder %d", i+1), err)
		}
	}

	log.Info("seed: demo data created",
		"organization", org.Name(),
		"owner", OwnerEmail, "owner_password", OwnerPassword,
		"manager", ManagerEmail, "manager_password", ManagerPass,
		"demo_api_key", DemoAPIKey,
		"clients", len(clientIDs), "products", len(productIDs), "contracts", len(seededContracts), "reminders", len(reminders), "expenses", len(expenses), "chat_messages", len(demoChats),
	)
	return nil
}

func seedHistoricInstallmentPayments(ctx context.Context, tx financingdomain.TxManager, repo financingdomain.ContractRepository, contract *financingdomain.Contract, paidInstallments int) error {
	if paidInstallments <= 0 {
		return nil
	}
	schedule := contract.Schedule()
	if paidInstallments > len(schedule) {
		paidInstallments = len(schedule)
	}
	for i := 0; i < paidInstallments; i++ {
		amount := schedule[i].Amount()
		paidAt := schedule[i].DueDate().Add(24 * time.Hour)
		if err := tx.WithinTx(ctx, func(ctx context.Context) error {
			if err := contract.RegisterPayment(uuid.NewString(), amount, paidAt); err != nil {
				return err
			}
			payments := contract.Payments()
			if err := repo.AddPayment(ctx, contract.ID(), payments[len(payments)-1]); err != nil {
				return err
			}
			return repo.SaveState(ctx, contract)
		}); err != nil {
			return err
		}
	}
	return nil
}

func newUser(orgID, name, email, password string, role iamdomain.Role, hasher *iaminfra.BcryptHasher) (iamdomain.User, error) {
	hash, err := hasher.Hash(password)
	if err != nil {
		return iamdomain.User{}, err
	}
	return iamdomain.NewUser(uuid.NewString(), orgID, name, email, hash, role)
}

func rub(s string) money.Money {
	m, _ := money.FromString(s, "RUB")
	return m
}

func markup(s string) financingdomain.Markup {
	mk, _ := financingdomain.NewMarkup(rub(s))
	return mk
}

func monthsFrom(t time.Time, n int) time.Time {
	y, m, d := t.AddDate(0, n, 0).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func weekDate(t time.Time, weekday time.Weekday) time.Time {
	start := seedWeekStart(t)
	offset := (int(weekday) + 6) % 7
	return start.AddDate(0, 0, offset)
}

func seedWeekStart(t time.Time) time.Time {
	u := t.UTC()
	day := time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
	delta := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -delta)
}

func tomorrowAt(tz *time.Location, hour, min int) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	t := time.Now().In(tz).AddDate(0, 0, 1)
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, tz)
}

func todayAt(tz *time.Location, hour, min int) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	t := time.Now().In(tz)
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, tz)
}

func nextFriday(tz *time.Location, hour, min int) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	t := time.Now().In(tz)
	for i := 1; i <= 7; i++ {
		d := t.AddDate(0, 0, i)
		if d.Weekday() == time.Friday {
			return time.Date(d.Year(), d.Month(), d.Day(), hour, min, 0, 0, tz)
		}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), hour, min, 0, 0, tz)
}

func seedErr(what string, err error) error { return fmt.Errorf("seed: %s: %w", what, err) }
