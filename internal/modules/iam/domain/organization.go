package domain

import (
	"strings"
	"time"
)

// Organization is the tenant root. Amana is multi-organization (SaaS).
type Organization struct {
	id        string
	name      string
	currency  string
	createdAt time.Time
}

// NewOrganization validates and creates a fresh organization. currency defaults
// to RUB when empty.
func NewOrganization(id, name, currency string) (Organization, error) {
	name = strings.TrimSpace(name)
	if id == "" {
		return Organization{}, ErrOrgIDRequired
	}
	if name == "" {
		return Organization{}, ErrOrgNameRequired
	}
	currency = strings.TrimSpace(currency)
	if currency == "" {
		currency = "RUB"
	}
	return Organization{
		id:        id,
		name:      name,
		currency:  currency,
		createdAt: time.Now().UTC(),
	}, nil
}

// RehydrateOrganization rebuilds an organization from trusted storage.
func RehydrateOrganization(id, name, currency string, createdAt time.Time) Organization {
	return Organization{id: id, name: name, currency: currency, createdAt: createdAt}
}

func (o Organization) ID() string           { return o.id }
func (o Organization) Name() string         { return o.name }
func (o Organization) Currency() string     { return o.currency }
func (o Organization) CreatedAt() time.Time { return o.createdAt }
