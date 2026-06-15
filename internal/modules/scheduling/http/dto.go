package http

import (
	"time"

	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
)

type createReminderRequest struct {
	Type            string `json:"type" validate:"required,oneof=call delivery payment_followup"`
	ClientID        string `json:"client_id" validate:"omitempty,uuid"`
	ContractID      string `json:"contract_id" validate:"omitempty,uuid"`
	Note            string `json:"note"`
	DesiredAt       string `json:"desired_at" validate:"required"` // RFC3339
	DurationMinutes int    `json:"duration_minutes" validate:"min=0"`
}

type previewSlotRequest struct {
	DesiredAt       string `json:"desired_at" validate:"required"` // RFC3339
	DurationMinutes int    `json:"duration_minutes" validate:"min=0"`
}

type reminderResponse struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"org_id"`
	Type            string    `json:"type"`
	ClientID        string    `json:"client_id,omitempty"`
	ContractID      string    `json:"contract_id,omitempty"`
	Note            string    `json:"note"`
	DesiredAt       time.Time `json:"desired_at"`
	DurationMinutes int       `json:"duration_minutes"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	WasShifted      bool      `json:"was_shifted"`
	Reason          string    `json:"reason"`
	CreatedAt       time.Time `json:"created_at"`
}

type slotResponse struct {
	ScheduledAt time.Time `json:"scheduled_at"`
	WasShifted  bool      `json:"was_shifted"`
	Reason      string    `json:"reason"`
}

func toReminderResponse(r domain.Reminder) reminderResponse {
	return reminderResponse{
		ID:              r.ID(),
		OrgID:           r.OrgID(),
		Type:            r.Type().String(),
		ClientID:        r.ClientID(),
		ContractID:      r.ContractID(),
		Note:            r.Note(),
		DesiredAt:       r.DesiredAt(),
		DurationMinutes: int(r.Duration() / time.Minute),
		ScheduledAt:     r.ScheduledAt(),
		WasShifted:      r.WasShifted(),
		Reason:          r.Reason(),
		CreatedAt:       r.CreatedAt(),
	}
}

func toSlotResponse(s domain.ScheduledTime) slotResponse {
	return slotResponse{ScheduledAt: s.Time, WasShifted: s.WasShifted, Reason: s.Reason}
}
