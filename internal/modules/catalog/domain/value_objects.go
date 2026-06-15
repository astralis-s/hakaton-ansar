package domain

import "strings"

// HalalStatus is the mandatory permissibility status of a product.
type HalalStatus string

const (
	HalalStatusHalal    HalalStatus = "halal"
	HalalStatusHaram    HalalStatus = "haram"
	HalalStatusDoubtful HalalStatus = "doubtful" // mushbooh
)

// ParseHalalStatus validates and normalizes a halal-status string.
func ParseHalalStatus(s string) (HalalStatus, error) {
	switch HalalStatus(strings.ToLower(strings.TrimSpace(s))) {
	case HalalStatusHalal:
		return HalalStatusHalal, nil
	case HalalStatusHaram:
		return HalalStatusHaram, nil
	case HalalStatusDoubtful:
		return HalalStatusDoubtful, nil
	default:
		return "", ErrInvalidHalalStatus
	}
}

func (h HalalStatus) Valid() bool {
	return h == HalalStatusHalal || h == HalalStatusHaram || h == HalalStatusDoubtful
}

func (h HalalStatus) String() string { return string(h) }
