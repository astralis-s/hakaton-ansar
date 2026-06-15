// Package web holds tiny HTTP helpers shared by every module's handler layer:
// JSON encoding, and request decode+validate that yields a classified
// apperror on failure (so handlers stay "decode → validate → app → map").
package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/astralis-s/hakaton-ansar/internal/platform/apperror"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var validate = validator.New(validator.WithRequiredStructEnabled())

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// DecodeAndValidate reads a JSON body into dst and runs struct validation.
// On any failure it returns an *apperror.Error of kind Invalid (HTTP 400).
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return apperror.Invalid("empty_body", "request body is required")
		}
		return apperror.Invalid("invalid_json", "invalid request body: "+err.Error())
	}

	if err := validate.Struct(dst); err != nil {
		return apperror.Invalid("validation_failed", validationMessage(err))
	}
	return nil
}

func validationMessage(err error) string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return "validation failed"
	}
	parts := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		parts = append(parts, fe.Field()+" failed on '"+fe.Tag()+"'")
	}
	return strings.Join(parts, "; ")
}
