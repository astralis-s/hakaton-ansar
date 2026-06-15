package infra

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/iam/domain"
)

// JWTService implements domain.TokenService with HMAC-SHA256 signed JWTs for the
// internal /api/app surface.
type JWTService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewJWTService(secret string, ttl time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), ttl: ttl, now: time.Now}
}

var _ domain.TokenService = (*JWTService)(nil)

// claims embeds standard JWT claims plus the org and role.
type claims struct {
	OrgID string `json:"org_id"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTService) Issue(u domain.User) (string, time.Time, error) {
	now := s.now()
	expiresAt := now.Add(s.ttl)
	c := claims{
		OrgID: u.OrgID(),
		Role:  u.Role().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *JWTService) Parse(token string) (domain.Principal, error) {
	var c claims
	parsed, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return domain.Principal{}, domain.ErrInvalidCredentials
	}
	role, err := domain.ParseRole(c.Role)
	if err != nil {
		return domain.Principal{}, domain.ErrInvalidCredentials
	}
	return domain.Principal{
		UserID: c.Subject,
		OrgID:  c.OrgID,
		Role:   role,
	}, nil
}
