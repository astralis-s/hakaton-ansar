package infra

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// BcryptHasher implements domain.Hasher with bcrypt.
type BcryptHasher struct{ cost int }

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

var _ domain.Hasher = (*BcryptHasher)(nil)

func (h *BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
