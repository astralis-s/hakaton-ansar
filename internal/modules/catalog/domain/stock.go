package domain

import (
	"strings"
	"time"
)

// StockReason is why a product's stock changed.
type StockReason string

const (
	StockReceipt    StockReason = "receipt"    // поступление на склад
	StockSale       StockReason = "sale"       // списание при продаже (договор)
	StockAdjustment StockReason = "adjustment" // ручная корректировка
	StockWriteoff   StockReason = "writeoff"   // списание (брак/потеря)
)

func ParseStockReason(s string) (StockReason, error) {
	switch StockReason(strings.ToLower(strings.TrimSpace(s))) {
	case StockReceipt:
		return StockReceipt, nil
	case StockSale:
		return StockSale, nil
	case StockAdjustment:
		return StockAdjustment, nil
	case StockWriteoff:
		return StockWriteoff, nil
	default:
		return "", ErrInvalidStockReason
	}
}

func (r StockReason) Valid() bool {
	return r == StockReceipt || r == StockSale || r == StockAdjustment || r == StockWriteoff
}

func (r StockReason) String() string { return string(r) }

// StockMovement is one logged change to a product's stock balance.
type StockMovement struct {
	id           string
	orgID        string
	productID    string
	delta        int
	reason       StockReason
	note         string
	balanceAfter int
	createdAt    time.Time
}

// NewStockMovement validates and creates a movement record.
func NewStockMovement(id, orgID, productID string, delta int, reason StockReason, note string, balanceAfter int) (StockMovement, error) {
	if id == "" || orgID == "" || productID == "" {
		return StockMovement{}, ErrProductIDRequired
	}
	if delta == 0 {
		return StockMovement{}, ErrStockDeltaZero
	}
	if !reason.Valid() {
		return StockMovement{}, ErrInvalidStockReason
	}
	return StockMovement{
		id: id, orgID: orgID, productID: productID, delta: delta,
		reason: reason, note: strings.TrimSpace(note), balanceAfter: balanceAfter,
		createdAt: time.Now().UTC(),
	}, nil
}

// RehydrateStockMovement rebuilds a movement from trusted storage.
func RehydrateStockMovement(id, orgID, productID string, delta int, reason StockReason, note string, balanceAfter int, createdAt time.Time) StockMovement {
	return StockMovement{
		id: id, orgID: orgID, productID: productID, delta: delta,
		reason: reason, note: note, balanceAfter: balanceAfter, createdAt: createdAt,
	}
}

func (m StockMovement) ID() string           { return m.id }
func (m StockMovement) OrgID() string         { return m.orgID }
func (m StockMovement) ProductID() string     { return m.productID }
func (m StockMovement) Delta() int            { return m.delta }
func (m StockMovement) Reason() StockReason   { return m.reason }
func (m StockMovement) Note() string          { return m.note }
func (m StockMovement) BalanceAfter() int     { return m.balanceAfter }
func (m StockMovement) CreatedAt() time.Time  { return m.createdAt }
