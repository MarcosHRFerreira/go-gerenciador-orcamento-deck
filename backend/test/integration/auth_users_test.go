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

	registerBody := fmt.Sprintf(`{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"%s","password_confirm":"%s"}`, integrationStrongPassword, integrationStrongPassword)
	registerResponse := env.doAuthRegisterRequest(t, registerBody)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerResponse.Code)
	}

	registerPayload := decodeJSONResponse[dto.RegisterResponse](t, registerResponse.Body)
	if registerPayload.ID <= 0 {
		t.Fatalf("expected created user id to be greater than zero, got %d", registerPayload.ID)
	}

	loginBody := fmt.Sprintf(`{"email":"admin@local.dev","password":"%s"}`, integrationStrongPassword)
	loginResponse := env.doJSONRequest(t, http.MethodPost, "/auth/login", "", loginBody)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)
	if loginPayload.Token == "" {
		t.Fatal("expected access token to be returned")
	}
	refreshCookie := env.requireRefreshCookie(t, loginResponse)

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
	if mePayload.MustChangePassword {
		t.Fatal("expected first registered admin to access the system without forced password change")
	}

	refreshResponse := env.doJSONRequestWithOptions(t, http.MethodPost, "/auth/refresh", jsonRequestOptions{
		cookies: []*http.Cookie{refreshCookie},
	})
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, refreshResponse.Code)
	}

	refreshPayload := decodeJSONResponse[dto.RefreshTokenResponse](t, refreshResponse.Body)
	if refreshPayload.Token == "" {
		t.Fatal("expected refreshed access token to be returned")
	}
	refreshedCookie := env.requireRefreshCookie(t, refreshResponse)

	logoutResponse := env.doJSONRequestWithOptions(t, http.MethodPost, "/auth/logout", jsonRequestOptions{
		cookies: []*http.Cookie{refreshedCookie},
	})
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, logoutResponse.Code)
	}

	refreshAfterLogoutResponse := env.doJSONRequestWithOptions(t, http.MethodPost, "/auth/refresh", jsonRequestOptions{
		cookies: []*http.Cookie{refreshedCookie},
	})
	if refreshAfterLogoutResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, refreshAfterLogoutResponse.Code)
	}

	refreshAfterLogoutPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, refreshAfterLogoutResponse.Body)
	if refreshAfterLogoutPayload.Message != "Refresh token expirado" {
		t.Fatalf("expected refresh token expired message, got %s", refreshAfterLogoutPayload.Message)
	}
}

func TestAuthRegisterShouldBeBlockedAfterFirstUser(t *testing.T) {
	env := newIntegrationTestEnv(t)

	firstRegisterBody := fmt.Sprintf(`{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"%s","password_confirm":"%s"}`, integrationStrongPassword, integrationStrongPassword)
	firstRegisterResponse := env.doAuthRegisterRequest(t, firstRegisterBody)
	if firstRegisterResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstRegisterResponse.Code)
	}

	secondRegisterBody := fmt.Sprintf(`{"name":"Outro Admin","email":"other@local.dev","username":"other","password":"%s","password_confirm":"%s"}`, integrationStrongPassword, integrationStrongPassword)
	secondRegisterResponse := env.doAuthRegisterRequest(t, secondRegisterBody)
	if secondRegisterResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, secondRegisterResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, secondRegisterResponse.Body)
	if errorPayload.Message != "Cadastro publico nao esta mais disponivel" {
		t.Fatalf("expected forbidden message, got %s", errorPayload.Message)
	}
}

func TestAuthAndUsersShouldRejectWeakPasswords(t *testing.T) {
	env := newIntegrationTestEnv(t)

	registerResponse := env.doAuthRegisterRequest(
		t,
		fmt.Sprintf(`{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"%s","password_confirm":"%s"}`, integrationWeakPassword, integrationWeakPassword),
	)
	if registerResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, registerResponse.Code)
	}

	registerPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, registerResponse.Body)
	if registerPayload.Message != "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial" {
		t.Fatalf("expected strong password message, got %s", registerPayload.Message)
	}

	adminToken := env.createAdminToken(t)

	createUserResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/users",
		adminToken,
		fmt.Sprintf(`{"name":"Usuario Fraco","email":"weak.user@local.dev","username":"weak_user","password":"%s","password_confirm":"%s","role":"user"}`, integrationWeakPassword, integrationWeakPassword),
	)
	if createUserResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, createUserResponse.Code)
	}

	createUserPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, createUserResponse.Body)
	if createUserPayload.Message != "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial" {
		t.Fatalf("expected strong password message, got %s", createUserPayload.Message)
	}
}

func TestUsersRoutesShouldRespectAdminAuthorization(t *testing.T) {
	env := newIntegrationTestEnv(t)

	adminRegisterBody := fmt.Sprintf(`{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"%s","password_confirm":"%s"}`, integrationStrongPassword, integrationStrongPassword)
	adminRegisterResponse := env.doAuthRegisterRequest(t, adminRegisterBody)
	if adminRegisterResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, adminRegisterResponse.Code)
	}

	adminLoginBody := fmt.Sprintf(`{"email":"admin@local.dev","password":"%s"}`, integrationStrongPassword)
	adminLoginResponse := env.doJSONRequest(t, http.MethodPost, "/auth/login", "", adminLoginBody)
	if adminLoginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminLoginResponse.Code)
	}

	adminLoginPayload := decodeJSONResponse[dto.LoginResponse](t, adminLoginResponse.Body)

	createUserBody := fmt.Sprintf(`{"name":"Usuario Comum","email":"user@local.dev","username":"user","password":"%s","password_confirm":"%s","role":"user"}`, integrationStrongPassword, integrationStrongPassword)
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
	if !listUsersPayload[1].MustChangePassword {
		t.Fatal("expected admin-created user to require password change on first access")
	}

	userToken := env.createUserToken(t, adminLoginPayload.Token, uniqueSuffix(), "user")

	forbiddenListResponse := env.doJSONRequest(t, http.MethodGet, "/users", userToken, "")
	if forbiddenListResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenListResponse.Code)
	}

	forbiddenPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenListResponse.Body)
	if forbiddenPayload.Message != "Permissoes insuficientes" {
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
	if forbiddenPayload.Message != "Permissoes insuficientes" {
		t.Fatalf("expected forbidden message, got %s", forbiddenPayload.Message)
	}
}

func TestUsersAdminShouldUpdateRoleAndActive(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)

	createUserResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/users",
		adminToken,
		fmt.Sprintf(`{"name":"Usuario Comum","email":"user@local.dev","username":"user","password":"%s","password_confirm":"%s","role":"user"}`, integrationStrongPassword, integrationStrongPassword),
	)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	createUserPayload := decodeJSONResponse[dto.CreateUserResponse](t, createUserResponse.Body)

	updateRoleResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/users/%d/role", createUserPayload.ID),
		adminToken,
		`{"role":"admin"}`,
	)
	if updateRoleResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateRoleResponse.Code)
	}

	updateActiveResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/users/%d/active", createUserPayload.ID),
		adminToken,
		`{"active":false}`,
	)
	if updateActiveResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateActiveResponse.Code)
	}

	listUsersResponse := env.doJSONRequest(t, http.MethodGet, "/users", adminToken, "")
	if listUsersResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listUsersResponse.Code)
	}

	listUsersPayload := decodeJSONResponse[[]dto.UserResponse](t, listUsersResponse.Body)
	if len(listUsersPayload) != 2 {
		t.Fatalf("expected 2 users, got %d", len(listUsersPayload))
	}

	var updatedUser dto.UserResponse
	for _, item := range listUsersPayload {
		if item.ID == createUserPayload.ID {
			updatedUser = item
			break
		}
	}

	if updatedUser.Role != "admin" {
		t.Fatalf("expected updated role admin, got %s", updatedUser.Role)
	}
	if updatedUser.Active {
		t.Fatal("expected updated user to be inactive")
	}

	loginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"user@local.dev","password":"%s"}`, integrationStrongPassword),
	)
	if loginResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, loginResponse.Body)
	if loginPayload.Message != "Usuario desativado" {
		t.Fatalf("expected inactive user message, got %s", loginPayload.Message)
	}
}

func TestUsersAdminShouldResetPasswordAndRequireChangeOnNextLogin(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)

	createUserResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/users",
		adminToken,
		fmt.Sprintf(`{"name":"Usuario Reset","email":"user.reset@local.dev","username":"user_reset","password":"%s","password_confirm":"%s","role":"user"}`, integrationStrongPassword, integrationStrongPassword),
	)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	createUserPayload := decodeJSONResponse[dto.CreateUserResponse](t, createUserResponse.Body)

	changePasswordResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"user.reset@local.dev","password":"%s"}`, integrationStrongPassword),
	)
	if changePasswordResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, changePasswordResponse.Code)
	}

	firstLoginPayload := decodeJSONResponse[dto.LoginResponse](t, changePasswordResponse.Body)

	firstAccessChangeResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/auth/change-password",
		firstLoginPayload.Token,
		fmt.Sprintf(`{"current_password":"%s","new_password":"%s","new_password_confirm":"%s"}`, integrationStrongPassword, integrationUpdatedPassword, integrationUpdatedPassword),
	)
	if firstAccessChangeResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, firstAccessChangeResponse.Code)
	}

	resetPasswordResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/users/%d/reset-password", createUserPayload.ID),
		adminToken,
		fmt.Sprintf(`{"password":"%s","password_confirm":"%s"}`, integrationResetPassword, integrationResetPassword),
	)
	if resetPasswordResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resetPasswordResponse.Code)
	}

	oldPasswordLoginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"user.reset@local.dev","password":"%s"}`, integrationUpdatedPassword),
	)
	if oldPasswordLoginResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, oldPasswordLoginResponse.Code)
	}

	resetLoginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"user.reset@local.dev","password":"%s"}`, integrationResetPassword),
	)
	if resetLoginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resetLoginResponse.Code)
	}

	resetLoginPayload := decodeJSONResponse[dto.LoginResponse](t, resetLoginResponse.Body)

	meResponse := env.doJSONRequest(t, http.MethodGet, "/users/me", resetLoginPayload.Token, "")
	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, meResponse.Code)
	}

	mePayload := decodeJSONResponse[dto.UserResponse](t, meResponse.Body)
	if !mePayload.MustChangePassword {
		t.Fatal("expected password reset to require a new change on next access")
	}
}

func TestAuthChangePasswordAndResetPasswordShouldRejectWeakPassword(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)

	createUserResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/users",
		adminToken,
		fmt.Sprintf(`{"name":"Usuario Seguro","email":"secure.user@local.dev","username":"secure_user","password":"%s","password_confirm":"%s","role":"user"}`, integrationStrongPassword, integrationStrongPassword),
	)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	createUserPayload := decodeJSONResponse[dto.CreateUserResponse](t, createUserResponse.Body)

	loginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"secure.user@local.dev","password":"%s"}`, integrationStrongPassword),
	)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)

	changePasswordResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/auth/change-password",
		loginPayload.Token,
		fmt.Sprintf(`{"current_password":"%s","new_password":"%s","new_password_confirm":"%s"}`, integrationStrongPassword, integrationWeakPassword, integrationWeakPassword),
	)
	if changePasswordResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, changePasswordResponse.Code)
	}

	changePasswordPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, changePasswordResponse.Body)
	if changePasswordPayload.Message != "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial" {
		t.Fatalf("expected strong password message, got %s", changePasswordPayload.Message)
	}

	resetPasswordResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/users/%d/reset-password", createUserPayload.ID),
		adminToken,
		fmt.Sprintf(`{"password":"%s","password_confirm":"%s"}`, integrationWeakPassword, integrationWeakPassword),
	)
	if resetPasswordResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resetPasswordResponse.Code)
	}

	resetPasswordPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, resetPasswordResponse.Body)
	if resetPasswordPayload.Message != "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial" {
		t.Fatalf("expected strong password message, got %s", resetPasswordPayload.Message)
	}
}

func TestAuthChangePasswordShouldUnlockFirstAccessUser(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)

	createUserResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/users",
		adminToken,
		fmt.Sprintf(`{"name":"Primeiro Acesso","email":"first.access@local.dev","username":"first_access","password":"%s","password_confirm":"%s","role":"user"}`, integrationStrongPassword, integrationStrongPassword),
	)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	loginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"first.access@local.dev","password":"%s"}`, integrationStrongPassword),
	)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)

	meResponse := env.doJSONRequest(t, http.MethodGet, "/users/me", loginPayload.Token, "")
	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, meResponse.Code)
	}

	mePayload := decodeJSONResponse[dto.UserResponse](t, meResponse.Body)
	if !mePayload.MustChangePassword {
		t.Fatal("expected first access user to require password change")
	}

	budgetsResponse := env.doJSONRequest(t, http.MethodGet, "/budgets", loginPayload.Token, "")
	if budgetsResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, budgetsResponse.Code)
	}

	budgetsErrorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, budgetsResponse.Body)
	if budgetsErrorPayload.Message != "Troca de senha obrigatoria antes de acessar o sistema" {
		t.Fatalf("expected password change required message, got %s", budgetsErrorPayload.Message)
	}

	changePasswordResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/auth/change-password",
		loginPayload.Token,
		fmt.Sprintf(`{"current_password":"%s","new_password":"%s","new_password_confirm":"%s"}`, integrationStrongPassword, integrationResetPassword, integrationResetPassword),
	)
	if changePasswordResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, changePasswordResponse.Code)
	}

	changePasswordPayload := decodeJSONResponse[dto.ChangePasswordResponse](t, changePasswordResponse.Body)
	if changePasswordPayload.Token == "" {
		t.Fatal("expected updated access token after changing password")
	}
	env.requireRefreshCookie(t, changePasswordResponse)

	updatedMeResponse := env.doJSONRequest(t, http.MethodGet, "/users/me", changePasswordPayload.Token, "")
	if updatedMeResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updatedMeResponse.Code)
	}

	updatedMePayload := decodeJSONResponse[dto.UserResponse](t, updatedMeResponse.Body)
	if updatedMePayload.MustChangePassword {
		t.Fatal("expected password change requirement to be cleared")
	}

	oldLoginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"first.access@local.dev","password":"%s"}`, integrationStrongPassword),
	)
	if oldLoginResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, oldLoginResponse.Code)
	}

	newLoginResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/auth/login",
		"",
		fmt.Sprintf(`{"email":"first.access@local.dev","password":"%s"}`, integrationResetPassword),
	)
	if newLoginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, newLoginResponse.Code)
	}
}

func TestUsersAdminUpdateRoutesShouldProtectLastAdminAndRespectAuthorization(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userToken := env.createUserToken(t, adminToken, uniqueSuffix(), "user")

	forbiddenRoleResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/users/1/role",
		userToken,
		`{"role":"admin"}`,
	)
	if forbiddenRoleResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenRoleResponse.Code)
	}

	forbiddenRolePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenRoleResponse.Body)
	if forbiddenRolePayload.Message != "Permissoes insuficientes" {
		t.Fatalf("expected forbidden message, got %s", forbiddenRolePayload.Message)
	}

	lastAdminRoleResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/users/1/role",
		adminToken,
		`{"role":"user"}`,
	)
	if lastAdminRoleResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, lastAdminRoleResponse.Code)
	}

	lastAdminRolePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, lastAdminRoleResponse.Body)
	if lastAdminRolePayload.Message != "Nao e permitido alterar o proprio perfil" {
		t.Fatalf("expected self role protection message, got %s", lastAdminRolePayload.Message)
	}

	lastAdminActiveResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/users/1/active",
		adminToken,
		`{"active":false}`,
	)
	if lastAdminActiveResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, lastAdminActiveResponse.Code)
	}

	lastAdminActivePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, lastAdminActiveResponse.Body)
	if lastAdminActivePayload.Message != "Nao e permitido desativar o proprio usuario" {
		t.Fatalf("expected self active protection message, got %s", lastAdminActivePayload.Message)
	}

	resetOwnPasswordResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		"/users/1/reset-password",
		adminToken,
		fmt.Sprintf(`{"password":"%s","password_confirm":"%s"}`, integrationResetPassword, integrationResetPassword),
	)
	if resetOwnPasswordResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resetOwnPasswordResponse.Code)
	}

	resetOwnPasswordPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, resetOwnPasswordResponse.Body)
	if resetOwnPasswordPayload.Message != "Nao e permitido resetar a propria senha" {
		t.Fatalf("expected self reset protection message, got %s", resetOwnPasswordPayload.Message)
	}
}
