// Package pgconv holds small, reusable conversions between pgx/pgtype values and
// plain Go/domain types (UUID strings, timestamps, decimal money), plus a helper
// to detect unique-constraint violations. Shared by every module's repository.
package pgconv

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// UUID converts a string UUID into pgtype.UUID.
func UUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

// StrUUID renders a pgtype.UUID as its canonical string form ("" if null).
func StrUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

// NullableUUID converts a string into pgtype.UUID, treating "" as SQL NULL.
func NullableUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return UUID(s)
}

// Timestamp wraps a time.Time as a valid pgtype.Timestamptz.
func Timestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// NullableTimestamp wraps an optional time as pgtype.Timestamptz.
func NullableTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return Timestamp(*t)
}

// TimeValue returns the time.Time of a pgtype.Timestamptz (zero if null).
func TimeValue(ts pgtype.Timestamptz) time.Time { return ts.Time }

// TimePtr returns a *time.Time, nil when the timestamp is null.
func TimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

// Date wraps a time.Time as a valid pgtype.Date (date-only column).
func Date(t time.Time) pgtype.Date { return pgtype.Date{Time: t, Valid: true} }

// DateValue returns the time.Time of a pgtype.Date (zero if null).
func DateValue(d pgtype.Date) time.Time { return d.Time }

// Numeric converts a decimal into pgtype.Numeric (via its canonical string).
func Numeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(d.String()); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("encode numeric: %w", err)
	}
	return n, nil
}

// DecimalFromNumeric converts a pgtype.Numeric back into a decimal.
func DecimalFromNumeric(n pgtype.Numeric) (decimal.Decimal, error) {
	if !n.Valid {
		return decimal.Zero, nil
	}
	v, err := n.Value()
	if err != nil {
		return decimal.Zero, fmt.Errorf("decode numeric: %w", err)
	}
	s, ok := v.(string)
	if !ok {
		return decimal.Zero, fmt.Errorf("unexpected numeric value type %T", v)
	}
	return decimal.NewFromString(s)
}

// IsUniqueViolation reports whether err is a Postgres unique-constraint error
// (SQLSTATE 23505), optionally for a specific constraint name.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}
