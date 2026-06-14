package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

	requestID := strings.TrimSpace(res.Header().Get("X-Request-Id"))
	if requestID == "" {
		t.Fatal("expected X-Request-Id header to be set")
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

func TestCheckHealthShouldPreserveIncomingRequestIDHeader(t *testing.T) {
	t.Parallel()

	router := server.NewRouter(validator.New(), server.Dependencies{
		HealthChecker: &healthCheckerStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/check-health", nil)
	req.Header.Set("X-Request-Id", "req_teste_123")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if res.Header().Get("X-Request-Id") != "req_teste_123" {
		t.Fatalf("expected X-Request-Id header to be preserved, got %s", res.Header().Get("X-Request-Id"))
	}
}
