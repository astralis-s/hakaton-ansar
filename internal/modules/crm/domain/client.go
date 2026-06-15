package domain

import (
	"strings"
	"time"
)

// Client is a customer of the organization. Intentionally simple (no sales
// funnel). The identification document is needed when a contract is drawn up.
type Client struct {
	id        string
	orgID     string
	fullName  string
	phone     string
	document  string
	createdAt time.Time
}

// NewClient validates invariants and creates a fresh client.
func NewClient(id, orgID, fullName, phone, document string) (Client, error) {
	if id == "" {
		return Client{}, ErrClientIDRequired
	}
	if orgID == "" {
		return Client{}, ErrOrgIDRequired
	}
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return Client{}, ErrClientNameRequired
	}
	return Client{
		id:        id,
		orgID:     orgID,
		fullName:  fullName,
		phone:     strings.TrimSpace(phone),
		document:  strings.TrimSpace(document),
		createdAt: time.Now().UTC(),
	}, nil
}

// RehydrateClient rebuilds a client from trusted storage.
func RehydrateClient(id, orgID, fullName, phone, document string, createdAt time.Time) Client {
	return Client{
		id:        id,
		orgID:     orgID,
		fullName:  fullName,
		phone:     phone,
		document:  document,
		createdAt: createdAt,
	}
}

// Update returns a validated copy with new mutable fields, preserving identity.
func (c Client) Update(fullName, phone, document string) (Client, error) {
	updated, err := NewClient(c.id, c.orgID, fullName, phone, document)
	if err != nil {
		return Client{}, err
	}
	updated.createdAt = c.createdAt
	return updated, nil
}

func (c Client) ID() string           { return c.id }
func (c Client) OrgID() string        { return c.orgID }
func (c Client) FullName() string     { return c.fullName }
func (c Client) Phone() string        { return c.phone }
func (c Client) Document() string     { return c.document }
func (c Client) CreatedAt() time.Time { return c.createdAt }
