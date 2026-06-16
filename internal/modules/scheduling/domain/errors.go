package domain

import "errors"

var (
	ErrReminderIDRequired    = errors.New("reminder id is required")
	ErrOrgIDRequired         = errors.New("organization id is required")
	ErrInvalidReminderType   = errors.New("invalid reminder type: must be call, delivery or payment_followup")
	ErrReminderNotFound      = errors.New("reminder not found")
	ErrInvalidReminderStatus = errors.New("invalid reminder status")
	ErrReminderAlreadyDone   = errors.New("reminder is already completed")
	ErrReminderCancelled     = errors.New("reminder is cancelled")
)
