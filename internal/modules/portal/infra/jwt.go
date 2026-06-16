package infra

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/astralis-s/hakaton-ansar/internal/modules/portal/domain"
)

// ClientJWTService implements domain.ClientTokenService with HMAC-SHA256 signed
// JWTs for the client portal (/api/portal). The "kind":"client" claim keeps
// portal tokens distinct from staff tokens even though they share the secret.
type ClientJWTService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewClientJWTService(secret string, ttl time.Duration) *ClientJWTService {
	return &ClientJWTService{secret: []byte(secret), ttl: ttl, now: time.Now}
}

var _ domain.ClientTokenService = (*ClientJWTService)(nil)

const clientTokenKind = "client"

type clientClaims struct {
	OrgID string `json:"org_id"`
	Kind  string `json:"kind"`
	jwt.RegisteredClaims
}

func (s *ClientJWTService) Issue(orgID, clientID string) (string, time.Time, error) {
	now := s.now()
	expiresAt := now.Add(s.ttl)
	c := clientClaims{
		OrgID: orgID,
		Kind:  clientTokenKind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   clientID,
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

func (s *ClientJWTService) Parse(token string) (domain.ClientPrincipal, error) {
	var c clientClaims
	parsed, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid || c.Kind != clientTokenKind {
		return domain.ClientPrincipal{}, domain.ErrInvalidCredentials
	}
	return domain.ClientPrincipal{OrgID: c.OrgID, ClientID: c.Subject}, nil
}
