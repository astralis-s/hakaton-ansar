// Package scheduling is the composition root of the scheduling bounded context
// (namaz-aware reminders).
package scheduling

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	schedulinghttp "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/http"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/infra"
)

// Deps are the external dependencies of the scheduling module.
type Deps struct {
	Pool     *pgxpool.Pool
	Log      *slog.Logger
	Provider domain.PrayerTimesProvider
	Policy   domain.Policy
	Location domain.Location
}

// Module is the assembled scheduling module.
type Module struct {
	handler *schedulinghttp.Handler
}

// New wires the scheduling module.
func New(d Deps) *Module {
	scheduler := domain.NewScheduler(d.Provider, d.Policy, d.Location)
	reminders := infra.NewReminderRepository(d.Pool)

	handler := schedulinghttp.NewHandler(schedulinghttp.HandlerDeps{
		Schedule: app.NewScheduleReminder(scheduler, reminders),
		Preview:  app.NewPreviewSlot(scheduler),
		List:     app.NewListReminders(reminders),
		Get:      app.NewGetReminder(reminders),
		Update:   app.NewUpdateReminder(scheduler, reminders),
		Complete: app.NewCompleteReminder(reminders),
		Cancel:   app.NewCancelReminder(reminders),
		Log:      d.Log,
	})
	return &Module{handler: handler}
}

// RegisterRoutes mounts the scheduling routes onto a JWT-protected router.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
