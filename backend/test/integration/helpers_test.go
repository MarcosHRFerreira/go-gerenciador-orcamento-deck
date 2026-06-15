package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/server"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const integrationRequestTimeout = 5 * time.Second
const integrationSetupToken = "integration-bootstrap-token"

const (
	integrationStrongPassword  = "Strong@123"
	integrationUpdatedPassword = "Updated@123"
	integrationResetPassword   = "Reseted@123"
	integrationWeakPassword    = "12345678"
)

var integrationUniqueCounter atomic.Uint64

type integrationTestEnv struct {
	db     *sql.DB
	router http.Handler
	config *config.Config
}

type jsonRequestOptions struct {
	token   string
	body    string
	headers map[string]string
	cookies []*http.Cookie
}

type budgetSeedData struct {
	statusID      int64
	priorityID    int64
	installerID   int64
	contactID     int64
	projectTypeID int64
	projectID     int64
	projectName   string
	salespersonID int64
	lossReasonID  int64
}

func newIntegrationTestEnv(t *testing.T) *integrationTestEnv {
	t.Helper()

	projectRoot, err := projectRootDir()
	if err != nil {
		t.Fatalf("failed to resolve project root: %v", err)
	}

	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		t.Fatalf("failed to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	databaseName := buildTestDatabaseName(t.Name())
	adminDatabaseURL, err := replaceDatabaseName(cfg.DatabaseURL, "postgres")
	if err != nil {
		t.Fatalf("failed to build admin database url: %v", err)
	}

	adminDB, err := sql.Open("pgx", adminDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open admin database connection: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), integrationRequestTimeout)
	defer cancel()

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", databaseName)); err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	testDatabaseURL, err := replaceDatabaseName(cfg.DatabaseURL, databaseName)
	if err != nil {
		t.Fatalf("failed to build test database url: %v", err)
	}

	testDB, err := sql.Open("pgx", testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test database connection: %v", err)
	}

	if err := testDB.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	if err := applyTestMigrations(ctx, testDB); err != nil {
		t.Fatalf("failed to apply test migrations: %v", err)
	}

	testConfig := *cfg
	testConfig.DatabaseURL = testDatabaseURL
	testConfig.InitialAdminSetupToken = integrationSetupToken

	env := &integrationTestEnv{
		db:     testDB,
		router: server.NewRouter(validator.New(), server.Dependencies{DB: testDB, HealthChecker: testDB, Config: &testConfig}),
		config: &testConfig,
	}

	t.Cleanup(func() {
		_ = testDB.Close()
		dropTestDatabase(t, adminDatabaseURL, databaseName)
	})

	return env
}

func (e *integrationTestEnv) doJSONRequest(t *testing.T, method string, path string, token string, body string) *httptest.ResponseRecorder {
	t.Helper()

	return e.doJSONRequestWithOptions(t, method, path, jsonRequestOptions{
		token: token,
		body:  body,
	})
}

func (e *integrationTestEnv) doJSONRequestWithOptions(t *testing.T, method string, path string, options jsonRequestOptions) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(options.body))
	if options.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if options.token != "" {
		req.Header.Set("Authorization", "Bearer "+options.token)
	}
	for key, value := range options.headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range options.cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	e.router.ServeHTTP(recorder, req)

	return recorder
}

func (e *integrationTestEnv) doAuthRegisterRequest(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	return e.doJSONRequestWithOptions(t, http.MethodPost, "/auth/register", jsonRequestOptions{
		body: body,
		headers: map[string]string{
			"X-Setup-Token": e.config.InitialAdminSetupToken,
		},
	})
}

func (e *integrationTestEnv) requireRefreshCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	response := recorder.Result()
	defer response.Body.Close()

	for _, cookie := range response.Cookies() {
		if cookie.Name == e.config.RefreshCookieName {
			return cookie
		}
	}

	t.Fatalf("expected refresh cookie %s to be set", e.config.RefreshCookieName)
	return nil
}

func decodeJSONResponse[T any](t *testing.T, body io.Reader) T {
	t.Helper()

	var payload T
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}

	return payload
}

func (e *integrationTestEnv) createAdminToken(t *testing.T) string {
	t.Helper()

	registerBody := fmt.Sprintf(`{"name":"Admin Local","email":"admin@local.dev","username":"admin","password":"%s","password_confirm":"%s"}`, integrationStrongPassword, integrationStrongPassword)
	registerResponse := e.doAuthRegisterRequest(t, registerBody)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, registerResponse.Code)
	}

	loginBody := fmt.Sprintf(`{"email":"admin@local.dev","password":"%s"}`, integrationStrongPassword)
	loginResponse := e.doJSONRequest(t, http.MethodPost, "/auth/login", "", loginBody)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)
	if loginPayload.Token == "" {
		t.Fatal("expected access token to be returned")
	}

	return loginPayload.Token
}

func (e *integrationTestEnv) createUserToken(t *testing.T, adminToken string, suffix string, role string) string {
	t.Helper()

	emailSuffix := strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(suffix))
	return e.createUserTokenWithCredentials(
		t,
		adminToken,
		fmt.Sprintf("User %s", suffix),
		fmt.Sprintf("user.%s@local.dev", emailSuffix),
		fmt.Sprintf("user_%s", emailSuffix),
		role,
	)
}

func (e *integrationTestEnv) createUserTokenWithCredentials(
	t *testing.T,
	adminToken string,
	name string,
	email string,
	username string,
	role string,
) string {
	t.Helper()

	createUserBody := fmt.Sprintf(
		`{"name":"%s","email":"%s","username":"%s","password":"%s","password_confirm":"%s","role":"%s"}`,
		name,
		email,
		username,
		integrationStrongPassword,
		integrationStrongPassword,
		role,
	)
	createUserResponse := e.doJSONRequest(t, http.MethodPost, "/users", adminToken, createUserBody)
	if createUserResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createUserResponse.Code)
	}

	loginBody := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, integrationStrongPassword)
	loginResponse := e.doJSONRequest(t, http.MethodPost, "/auth/login", "", loginBody)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, loginResponse.Code)
	}

	loginPayload := decodeJSONResponse[dto.LoginResponse](t, loginResponse.Body)
	if loginPayload.Token == "" {
		t.Fatal("expected access token to be returned")
	}

	changePasswordResponse := e.doJSONRequest(
		t,
		http.MethodPatch,
		"/auth/change-password",
		loginPayload.Token,
		fmt.Sprintf(`{"current_password":"%s","new_password":"%s","new_password_confirm":"%s"}`, integrationStrongPassword, integrationUpdatedPassword, integrationUpdatedPassword),
	)
	if changePasswordResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, changePasswordResponse.Code)
	}

	changePasswordPayload := decodeJSONResponse[dto.ChangePasswordResponse](t, changePasswordResponse.Body)
	if changePasswordPayload.Token == "" {
		t.Fatal("expected updated access token to be returned")
	}

	return changePasswordPayload.Token
}

func (e *integrationTestEnv) seedBudgetData(t *testing.T, suffix string) budgetSeedData {
	t.Helper()

	now := time.Now()
	phoneSuffix := safePhoneSuffix(suffix)
	ctx, cancel := context.WithTimeout(context.Background(), integrationRequestTimeout)
	defer cancel()

	statusID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO budget_statuses (code, name, description, is_final, sort_order, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"STATUS_"+suffix,
		"Status "+suffix,
		"status de teste",
		false,
		1,
		now,
		now,
	)
	priorityID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO priorities (code, name, weight, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		"PRIORITY_"+suffix,
		"Priority "+suffix,
		10,
		now,
		now,
	)
	installerID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO installers (name, document, email, phone, city, state, notes, active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		"Installer "+suffix,
		"DOC-"+suffix,
		"installer."+strings.ToLower(suffix)+"@local.dev",
		"1190000"+phoneSuffix,
		"Sao Paulo",
		"SP",
		"installer de teste",
		true,
		now,
		now,
	)
	contactID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO contacts (installer_id, name, email, phone, role, is_primary, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		installerID,
		"Contato "+suffix,
		"contact."+strings.ToLower(suffix)+"@local.dev",
		"1180000"+phoneSuffix,
		"Compras",
		true,
		now,
		now,
	)
	projectTypeID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO project_types (code, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		"TYPE_"+suffix,
		"Tipo "+suffix,
		"tipo de projeto de teste",
		now,
		now,
	)
	projectID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO projects (name, project_type_id, city, state, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Projeto "+suffix,
		projectTypeID,
		"Campinas",
		"SP",
		"projeto de teste",
		now,
		now,
	)
	salespersonID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO salespeople (name, email, phone, active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"Vendedor "+suffix,
		"sales."+strings.ToLower(suffix)+"@local.dev",
		"1170000"+phoneSuffix,
		true,
		now,
		now,
	)
	lossReasonID := e.insertReturningID(
		t,
		ctx,
		`INSERT INTO loss_reasons (code, name, description, active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		"LOSS_"+suffix,
		"Perda "+suffix,
		"motivo de perda de teste",
		true,
		now,
		now,
	)

	return budgetSeedData{
		statusID:      statusID,
		priorityID:    priorityID,
		installerID:   installerID,
		contactID:     contactID,
		projectTypeID: projectTypeID,
		projectID:     projectID,
		projectName:   "Projeto " + suffix,
		salespersonID: salespersonID,
		lossReasonID:  lossReasonID,
	}
}

func (e *integrationTestEnv) insertReturningID(t *testing.T, ctx context.Context, query string, args ...interface{}) int64 {
	t.Helper()

	var id int64
	if err := e.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		t.Fatalf("failed to insert seed data: %v", err)
	}

	return id
}

func uniqueSuffix() string {
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), integrationUniqueCounter.Add(1))
}

func safePhoneSuffix(value string) string {
	cleaned := strings.NewReplacer("-", "", "_", "", " ", "").Replace(value)
	if len(cleaned) >= 4 {
		return cleaned[len(cleaned)-4:]
	}

	return strings.Repeat("0", 4-len(cleaned)) + cleaned
}

func buildTestDatabaseName(testName string) string {
	normalized := strings.ToLower(testName)
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "-", "_")
	normalized = replacer.Replace(normalized)

	return fmt.Sprintf("budget_management_test_%d_%s", time.Now().UnixNano(), normalized)
}

func replaceDatabaseName(databaseURL string, databaseName string) (string, error) {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path = "/" + databaseName
	return parsedURL.String(), nil
}

func applyTestMigrations(ctx context.Context, db *sql.DB) error {
	projectRoot, err := projectRootDir()
	if err != nil {
		return err
	}

	migrationFiles, err := filepath.Glob(filepath.Join(projectRoot, "db", "migrations", "*.sql"))
	if err != nil {
		return err
	}

	for _, filePath := range migrationFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		upSQL := extractUpMigration(string(content))
		if strings.TrimSpace(upSQL) == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, upSQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(filePath), err)
		}
	}

	return nil
}

func projectRootDir() (string, error) {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve current test file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", "..")), nil
}

func extractUpMigration(content string) string {
	parts := strings.Split(content, "-- migrate:down")
	if len(parts) == 0 {
		return ""
	}

	upPart := strings.Replace(parts[0], "-- migrate:up", "", 1)
	return strings.TrimSpace(upPart)
}

func dropTestDatabase(t *testing.T, adminDatabaseURL string, databaseName string) {
	t.Helper()

	adminDB, err := sql.Open("pgx", adminDatabaseURL)
	if err != nil {
		t.Fatalf("failed to reopen admin database connection: %v", err)
	}
	defer func() {
		_ = adminDB.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), integrationRequestTimeout)
	defer cancel()

	terminateConnectionsQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	if _, err := adminDB.ExecContext(ctx, terminateConnectionsQuery, databaseName); err != nil {
		t.Fatalf("failed to terminate test database connections: %v", err)
	}

	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", databaseName)); err != nil {
		t.Fatalf("failed to drop test database: %v", err)
	}
}
