package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/server"
	"github.com/go-playground/validator/v10"
)

type healthCheckerStub struct {
	pingErr error
}

func (h *healthCheckerStub) PingContext(_ context.Context) error {
	return h.pingErr
}

func TestCheckHealthShouldReturnOKWhenDatabaseIsAvailable(t *testing.T) {
	t.Parallel()

	router := server.NewRouter(validator.New(), server.Dependencies{
		HealthChecker: &healthCheckerStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/check-health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	expectedBody := "{\"message\":\"service is healthy\"}"
	if res.Body.String() != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, res.Body.String())
	}
}

func TestCheckHealthShouldReturnServiceUnavailableWhenDatabaseIsUnavailable(t *testing.T) {
	t.Parallel()

	router := server.NewRouter(validator.New(), server.Dependencies{
		HealthChecker: &healthCheckerStub{pingErr: errors.New("database unavailable")},
	})

	req := httptest.NewRequest(http.MethodGet, "/check-health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, res.Code)
	}

	expectedBody := "{\"message\":\"database unavailable\"}"
	if res.Body.String() != expectedBody {
		t.Fatalf("expected body %s, got %s", expectedBody, res.Body.String())
	}
}
