package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
)

func TestAuthRegisterLoginRefreshAndGetMeFlow(t *testing.T) {
	env := newIntegrationTestEnv(t)

	registerBody := `{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"123456","password_confirm":"123456"}`
	registerResponse := env.doJSONRequest(t, http.MethodPost, "/auth/register", "", registerBody)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerResponse.Code)
	}

	registerPayload := decodeJSONResponse[dto.RegisterResponse](t, registerResponse.Body)
	if registerPayload.ID <= 0 {
		t.Fatalf("expected created user id to be greater than zero, got %d", registerPayload.ID)
	}

	loginBody := `{"email":"admin@local.dev","password":"123456"}`
	loginResponse := env.doJSONRequest(t, http.MethodPost, "/auth/login", "", loginBody)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)
	if loginPayload.Token == "" {
		t.Fatal("expected access token to be returned")
	}
	if loginPayload.RefreshToken == "" {
		t.Fatal("expected refresh token to be returned")
	}

	meResponse := env.doJSONRequest(t, http.MethodGet, "/users/me", loginPayload.Token, "")
	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, meResponse.Code)
	}

	mePayload := decodeJSONResponse[dto.UserResponse](t, meResponse.Body)
	if mePayload.Email != "admin@local.dev" {
		t.Fatalf("expected email admin@local.dev, got %s", mePayload.Email)
	}
	if mePayload.Role != "admin" {
		t.Fatalf("expected role admin, got %s", mePayload.Role)
	}

	refreshBody := fmt.Sprintf(`{"refresh_token":"%s"}`, loginPayload.RefreshToken)
	refreshResponse := env.doJSONRequest(t, http.MethodPost, "/auth/refresh", loginPayload.Token, refreshBody)
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, refreshResponse.Code)
	}

	refreshPayload := decodeJSONResponse[dto.RefreshTokenResponse](t, refreshResponse.Body)
	if refreshPayload.Token == "" {
		t.Fatal("expected refreshed access token to be returned")
	}
	if refreshPayload.RefreshToken == "" {
		t.Fatal("expected refreshed refresh token to be returned")
	}
}

func TestAuthRegisterShouldBeBlockedAfterFirstUser(t *testing.T) {
	env := newIntegrationTestEnv(t)

	firstRegisterBody := `{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"123456","password_confirm":"123456"}`
	firstRegisterResponse := env.doJSONRequest(t, http.MethodPost, "/auth/register", "", firstRegisterBody)
	if firstRegisterResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstRegisterResponse.Code)
	}

	secondRegisterBody := `{"name":"Outro Admin","email":"other@local.dev","username":"other","password":"123456","password_confirm":"123456"}`
	secondRegisterResponse := env.doJSONRequest(t, http.MethodPost, "/auth/register", "", secondRegisterBody)
	if secondRegisterResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, secondRegisterResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, secondRegisterResponse.Body)
	if errorPayload.Message != "public registration is no longer available" {
		t.Fatalf("expected forbidden message, got %s", errorPayload.Message)
	}
}

func TestUsersRoutesShouldRespectAdminAuthorization(t *testing.T) {
	env := newIntegrationTestEnv(t)

	adminRegisterBody := `{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"123456","password_confirm":"123456"}`
	adminRegisterResponse := env.doJSONRequest(t, http.MethodPost, "/auth/register", "", adminRegisterBody)
	if adminRegisterResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, adminRegisterResponse.Code)
	}

	adminLoginBody := `{"email":"admin@local.dev","password":"123456"}`
	adminLoginResponse := env.doJSONRequest(t, http.MethodPost, "/auth/login", "", adminLoginBody)
	if adminLoginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminLoginResponse.Code)
	}

	adminLoginPayload := decodeJSONResponse[dto.LoginResponse](t, adminLoginResponse.Body)

	createUserBody := `{"name":"Usuario Comum","email":"user@local.dev","username":"user","password":"123456","password_confirm":"123456","role":"user"}`
	createUserResponse := env.doJSONRequest(t, http.MethodPost, "/users", adminLoginPayload.Token, createUserBody)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	listUsersResponse := env.doJSONRequest(t, http.MethodGet, "/users", adminLoginPayload.Token, "")
	if listUsersResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listUsersResponse.Code)
	}

	listUsersPayload := decodeJSONResponse[[]dto.UserResponse](t, listUsersResponse.Body)
	if len(listUsersPayload) != 2 {
		t.Fatalf("expected 2 users, got %d", len(listUsersPayload))
	}

	userToken := env.createUserToken(t, adminLoginPayload.Token, uniqueSuffix(), "user")

	forbiddenListResponse := env.doJSONRequest(t, http.MethodGet, "/users", userToken, "")
	if forbiddenListResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenListResponse.Code)
	}

	forbiddenPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenListResponse.Body)
	if forbiddenPayload.Message != "insufficient permissions" {
		t.Fatalf("expected forbidden message, got %s", forbiddenPayload.Message)
	}
}

func TestProjectsRoutesShouldRespectAdminAuthorization(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userToken := env.createUserToken(t, adminToken, uniqueSuffix(), "user")

	adminCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/projects",
		adminToken,
		`{"name":"Projeto Admin","city":"Campinas","state":"SP","notes":"projeto liberado para admin"}`,
	)
	if adminCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, adminCreateResponse.Code)
	}

	forbiddenCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/projects",
		userToken,
		`{"name":"Projeto User","city":"Campinas","state":"SP","notes":"projeto bloqueado para user"}`,
	)
	if forbiddenCreateResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenCreateResponse.Code)
	}

	forbiddenPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenCreateResponse.Body)
	if forbiddenPayload.Message != "insufficient permissions" {
		t.Fatalf("expected forbidden message, got %s", forbiddenPayload.Message)
	}
}
