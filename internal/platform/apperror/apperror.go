// Package apperror is the single place where errors are classified and mapped
// to HTTP responses. Domain packages return typed errors; the HTTP layer wraps
// or classifies them here so the mapping lives in exactly one location
// (CLAUDE.md §3: "маппинг доменной ошибки → HTTP-код в одном месте").
package apperror

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Kind is a transport-agnostic error category.
type Kind int

const (
	KindInternal Kind = iota
	KindInvalid
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
)

// Error is a classified application error carrying a machine-readable code and
// a human-readable message. It wraps an optional underlying error.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func newErr(kind Kind, code, msg string, err error) *Error {
	return &Error{Kind: kind, Code: code, Message: msg, Err: err}
}

// Constructors for each kind.
func Internal(msg string, err error) *Error { return newErr(KindInternal, "internal_error", msg, err) }
func Invalid(code, msg string) *Error       { return newErr(KindInvalid, code, msg, nil) }
func Unauthorized(code, msg string) *Error  { return newErr(KindUnauthorized, code, msg, nil) }
func Forbidden(code, msg string) *Error     { return newErr(KindForbidden, code, msg, nil) }
func NotFound(code, msg string) *Error      { return newErr(KindNotFound, code, msg, nil) }
func Conflict(code, msg string) *Error      { return newErr(KindConflict, code, msg, nil) }

func kindToStatus(k Kind) int {
	switch k {
	case KindInvalid:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// HTTPStatus resolves an error (possibly wrapped) to an HTTP status code.
// Unclassified errors default to 500.
func HTTPStatus(err error) int {
	var ae *Error
	if errors.As(err, &ae) {
		return kindToStatus(ae.Kind)
	}
	return http.StatusInternalServerError
}

type responseBody struct {
	Error responseError `json:"error"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Write serializes err as a JSON error response and sets the proper status.
// 5xx responses are logged (with the underlying detail) but the detail is never
// leaked to the client.
func Write(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	msg := "internal error"

	var ae *Error
	if errors.As(err, &ae) {
		status = kindToStatus(ae.Kind)
		code = ae.Code
		// Internal errors must not leak detail to clients.
		if ae.Kind != KindInternal {
			msg = ae.Message
		}
	}

	if status >= http.StatusInternalServerError && log != nil {
		log.Error("request failed",
			"error", err.Error(),
			"path", r.URL.Path,
			"status", status,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(responseBody{Error: responseError{Code: code, Message: msg}})
}
