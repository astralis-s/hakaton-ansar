package infra

import (
	"context"

	crmapp "github.com/astralis-s/hakaton-ansar/internal/modules/crm/app"
	"github.com/astralis-s/hakaton-ansar/internal/modules/telegrambot/domain"
)

// ClientDirectory implements domain.ClientDirectory over the crm CreateClient
// use-case, so a registered Telegram user becomes a real CRM client (with the
// full name shown in the manager inbox).
type ClientDirectory struct{ create *crmapp.CreateClient }

func NewClientDirectory(create *crmapp.CreateClient) *ClientDirectory {
	return &ClientDirectory{create: create}
}

var _ domain.ClientDirectory = (*ClientDirectory)(nil)

func (d *ClientDirectory) CreateClient(ctx context.Context, orgID, fullName, phone string) (string, error) {
	client, err := d.create.Execute(ctx, crmapp.CreateClientInput{
		OrgID:    orgID,
		FullName: fullName,
		Phone:    phone,
		Document: "", // passport data is collected later, when a contract is drawn up
	})
	if err != nil {
		return "", err
	}
	return client.ID(), nil
}
