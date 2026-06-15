package infra

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// BcryptHasher implements domain.PasswordHasher with bcrypt.
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

var _ domain.PasswordHasher = (*BcryptHasher)(nil)

func (h *BcryptHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare returns nil on match, domain.ErrInvalidCredentials on mismatch.
func (h *BcryptHasher) Compare(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return domain.ErrInvalidCredentials
	}
	return nil
}
