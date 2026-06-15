package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog"
	"github.com/astralis-s/hakaton-ansar/internal/modules/crm"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing"
	financinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling"
	schedulingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	schedulinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/infra"
	publicapiv1 "github.com/astralis-s/hakaton-ansar/internal/publicapi/v1"
)

// newTestRouter builds the real route tree. Modules are constructed with a nil
// pool: the routes exercised here (health, swagger, and the auth-rejection
// paths) never reach the database.
func newTestRouter() chi.Router {
	iamModule := iam.New(iam.Deps{
		Pool:      nil,
		Tx:        nil,
		Log:       nil,
		JWTSecret: "test-secret",
		JWTTTL:    0,
	})
	catalogModule := catalog.New(catalog.Deps{Pool: nil, Log: nil})
	crmModule := crm.New(crm.Deps{Pool: nil, Log: nil})
	financingModule := financing.New(financing.Deps{
		Pool:                  nil,
		Tx:                    nil,
		Log:                   nil,
		ComparisonRatePercent: decimal.NewFromInt(28),
		Products:              financinginfra.NewProductReader(catalogModule.Products()),
		Clients:               financinginfra.NewClientReader(crmModule.Clients()),
		OwnerOnly:             iamModule.OwnerMiddleware(),
	})
	prayerLoc := schedulingdomain.Location{Lat: 43.3178, Lon: 45.6949, TZ: time.UTC}
	schedulingModule := scheduling.New(scheduling.Deps{
		Pool:     nil,
		Log:      nil,
		Provider: schedulinginfra.NewPrayerProvider(prayerLoc, "shafii", "MWL"),
		Policy:   schedulingdomain.DefaultPolicy(),
		Location: prayerLoc,
	})
	publicAPI := publicapiv1.New(publicapiv1.Deps{
		CreateContract: financingModule.CreateContractUseCase(),
		GetContract:    financingModule.GetContractUseCase(),
		Log:            nil,
	})
	r := chi.NewRouter()
	mountRoutes(r, iamModule, catalogModule, crmModule, financingModule, schedulingModule, publicAPI)
	return r
}

func TestPublicRoutes(t *testing.T) {
	r := newTestRouter()

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health", http.MethodGet, "/health", http.StatusOK},
		{"swagger doc", http.MethodGet, "/swagger/doc.json", http.StatusOK},
		{"protected app route without token → 401", http.MethodGet, "/api/app/auth/me", http.StatusUnauthorized},
		{"public api route without key → 401", http.MethodGet, "/api/v1/contracts/00000000-0000-0000-0000-000000000000/payments", http.StatusUnauthorized},
		{"unknown route → 404", http.MethodGet, "/nope", http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("%s %s = %d, want %d (body: %s)", tc.method, tc.path, rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHealthBody(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %q, want \"ok\"", body["status"])
	}
}
