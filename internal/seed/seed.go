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

	catalogdomain "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/domain"
	cataloginfra "github.com/astralis-s/hakaton-ansar/internal/modules/catalog/infra"
	crmdomain "github.com/astralis-s/hakaton-ansar/internal/modules/crm/domain"
	crminfra "github.com/astralis-s/hakaton-ansar/internal/modules/crm/infra"
	financingapp "github.com/astralis-s/hakaton-ansar/internal/modules/financing/app"
	financingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/financing/domain"
	financinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra"
	iamdomain "github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
	iaminfra "github.com/astralis-s/hakaton-ansar/internal/modules/iam/infra"
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
	clientRepo := crminfra.NewClientRepository(pool)
	contractRepo := financinginfra.NewContractRepository(pool)
	charityRepo := financinginfra.NewCharityRepository(pool)
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

	// --- Products (mixed halal statuses) -------------------------------------
	products := []struct {
		name, category, cost string
		status               catalogdomain.HalalStatus
	}{
		{"Диван угловой «Кавказ»", "Мебель", "85000.00", catalogdomain.HalalStatusHalal},
		{"Кухонный гарнитур «Беркат»", "Мебель", "120000.00", catalogdomain.HalalStatusHalal},
		{"Холодильник Bosch", "Техника", "65000.00", catalogdomain.HalalStatusHalal},
		{"Подарочный набор (сомнительный поставщик)", "Прочее", "5000.00", catalogdomain.HalalStatusDoubtful},
		{"Вино столовое", "Напитки", "3000.00", catalogdomain.HalalStatusHaram},
	}
	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		cost, _ := money.FromString(p.cost, "RUB")
		prod, err := catalogdomain.NewProduct(uuid.NewString(), org.ID(), p.name, p.category, cost, p.status)
		if err != nil {
			return seedErr("product", err)
		}
		created, err := productRepo.Create(ctx, prod)
		if err != nil {
			return seedErr("create product", err)
		}
		productIDs = append(productIDs, created.ID())
	}

	// --- Contracts (different states) ----------------------------------------
	productReader := financinginfra.NewProductReader(productRepo)
	clientReader := financinginfra.NewClientReader(clientRepo)
	createUC := financingapp.NewCreateContract(contractRepo, productReader, clientReader, tx)
	payUC := financingapp.NewRegisterPayment(contractRepo, tx)
	charityUC := financingapp.NewAccrueLateCharity(contractRepo, charityRepo)

	now := time.Now()

	// 1) Active, partially paid (down 30000 + one 20000 payment).
	c1, err := createUC.Execute(ctx, financingapp.CreateContractInput{
		OrgID: org.ID(), ClientID: clientIDs[0], ProductID: productIDs[0],
		CostPrice: rub("85000.00"), Markup: markup("17000.00"), DownPayment: rub("25000.00"),
		Installments: 6, Cadence: financingdomain.CadenceMonthly, StartDate: monthsFrom(now, 1),
	})
	if err != nil {
		return seedErr("contract 1", err)
	}
	if _, err := payUC.Execute(ctx, financingapp.RegisterPaymentInput{OrgID: org.ID(), ContractID: c1.ID(), Amount: rub("20000.00")}); err != nil {
		return seedErr("contract 1 payment", err)
	}

	// 2) Overdue (started 8 months ago, unpaid) with an accrued sadaqa charge.
	c2, err := createUC.Execute(ctx, financingapp.CreateContractInput{
		OrgID: org.ID(), ClientID: clientIDs[1], ProductID: productIDs[1],
		CostPrice: rub("120000.00"), Markup: markup("24000.00"), DownPayment: rub("0"),
		Installments: 12, Cadence: financingdomain.CadenceMonthly, StartDate: monthsFrom(now, -8),
	})
	if err != nil {
		return seedErr("contract 2", err)
	}
	if _, err := charityUC.Execute(ctx, financingapp.AccrueLateCharityInput{
		OrgID: org.ID(), ContractID: c2.ID(), Amount: rub("500.00"),
		Note: "Просрочка платежа — садака на благотворительность", CreatedBy: owner.ID(),
	}); err != nil {
		return seedErr("contract 2 charity", err)
	}

	// 3) Fresh active contract, no payments yet.
	if _, err := createUC.Execute(ctx, financingapp.CreateContractInput{
		OrgID: org.ID(), ClientID: clientIDs[2], ProductID: productIDs[2],
		CostPrice: rub("65000.00"), Markup: markup("9750.00"), DownPayment: rub("15000.00"),
		Installments: 5, Cadence: financingdomain.CadenceMonthly, StartDate: monthsFrom(now, 1),
	}); err != nil {
		return seedErr("contract 3", err)
	}

	// --- Reminders (namaz-aware) ---------------------------------------------
	loc := schedulingdomain.Location{Lat: cfg.Lat, Lon: cfg.Lon, TZ: cfg.Timezone}
	provider := schedulinginfra.NewPrayerProvider(loc, cfg.Madhab, cfg.Method)
	scheduler := schedulingdomain.NewScheduler(provider, schedulingdomain.DefaultPolicy(), loc)
	scheduleUC := schedulingapp.NewScheduleReminder(scheduler, reminderRepo)

	reminders := []schedulingapp.ScheduleReminderInput{
		{OrgID: org.ID(), Type: "delivery", ClientID: clientIDs[0], ContractID: c1.ID(), Note: "Доставка дивана", DesiredAt: nextFriday(cfg.Timezone, 13, 0), DurationMinutes: 90},
		{OrgID: org.ID(), Type: "payment_followup", ClientID: clientIDs[1], ContractID: c2.ID(), Note: "Напомнить о платеже", DesiredAt: tomorrowAt(cfg.Timezone, 16, 5), DurationMinutes: 15},
		{OrgID: org.ID(), Type: "call", ClientID: clientIDs[2], Note: "Уточнить адрес доставки", DesiredAt: tomorrowAt(cfg.Timezone, 10, 0), DurationMinutes: 0},
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
		"clients", len(clientIDs), "products", len(productIDs), "contracts", 3, "reminders", len(reminders),
	)
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

func tomorrowAt(tz *time.Location, hour, min int) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	t := time.Now().In(tz).AddDate(0, 0, 1)
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
