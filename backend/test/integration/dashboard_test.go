package integration

import (
	"context"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

func TestDashboardSalespeopleShouldRequireAdminRole(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userToken := env.createUserToken(t, adminToken, uniqueSuffix(), "user")

	userResponse := env.doJSONRequest(t, http.MethodGet, "/dashboard/salespeople", userToken, "")
	if userResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for user role, got %d", http.StatusForbidden, userResponse.Code)
	}

	adminResponse := env.doJSONRequest(t, http.MethodGet, "/dashboard/salespeople", adminToken, "")
	if adminResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d for admin role, got %d", http.StatusOK, adminResponse.Code)
	}
}

func TestDashboardSalespeopleShouldUseCommercialActivityInsteadOfTechnicalUpdate(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	sentAt := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC)
	createBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"DASH-001",
			2026,
			sentAt,
			3200,
			seed,
			"Projetista Dashboard",
			"Concorrente Dashboard",
		),
	)
	if createBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponse.Code)
	}

	createBudgetPayload := decodeJSONResponse[createResourceResponse](t, createBudgetResponse.Body)
	ctx := context.Background()
	var adminUserID int64
	if err := env.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'admin'`).Scan(&adminUserID); err != nil {
		t.Fatalf("failed to load admin user id: %v", err)
	}

	commercialFollowUpAt := time.Date(2026, time.April, 3, 9, 0, 0, 0, time.UTC)
	commercialStatusAt := time.Date(2026, time.April, 5, 15, 30, 0, 0, time.UTC)
	technicalUpdateAt := time.Date(2026, time.April, 18, 8, 45, 0, 0, time.UTC)
	secondStatusID := insertNamedBudgetStatus(t, env, "STATUS_DASH_"+uniqueSuffix(), "Status Dashboard", false, 2)

	env.insertReturningID(
		t,
		ctx,
		`INSERT INTO budget_follow_ups (budget_id, created_by_user_id, notes, follow_up_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		createBudgetPayload.ID,
		adminUserID,
		"Cliente pediu retorno",
		commercialFollowUpAt,
		commercialFollowUpAt,
		commercialFollowUpAt,
	)
	env.insertReturningID(
		t,
		ctx,
		`INSERT INTO budget_status_history (budget_id, from_status_id, to_status_id, changed_by_user_id, notes, changed_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		createBudgetPayload.ID,
		seed.statusID,
		secondStatusID,
		adminUserID,
		"Status alterado apos reuniao",
		commercialStatusAt,
		commercialStatusAt,
		commercialStatusAt,
	)
	if _, err := env.db.ExecContext(ctx, `UPDATE budgets SET updated_at = $2 WHERE id = $1`, createBudgetPayload.ID, technicalUpdateAt); err != nil {
		t.Fatalf("failed to simulate technical update: %v", err)
	}

	dashboardResponse := env.doJSONRequest(t, http.MethodGet, "/dashboard/salespeople", adminToken, "")
	if dashboardResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, dashboardResponse.Code)
	}

	dashboardPayload := decodeJSONResponse[dto.SalespeopleDashboardResponse](t, dashboardResponse.Body)
	if len(dashboardPayload.RecentSalespeople) != 1 {
		t.Fatalf("expected 1 recent salesperson, got %d", len(dashboardPayload.RecentSalespeople))
	}
	if dashboardPayload.RecentSalespeople[0].LastActivityAt == nil {
		t.Fatal("expected last_activity_at to be returned")
	}
	if !dashboardPayload.RecentSalespeople[0].LastActivityAt.Equal(commercialStatusAt) {
		t.Fatalf(
			"expected commercial activity at %s, got %s",
			commercialStatusAt.Format(time.RFC3339),
			dashboardPayload.RecentSalespeople[0].LastActivityAt.Format(time.RFC3339),
		)
	}
	if dashboardPayload.RecentSalespeople[0].LastActivityAt.Equal(technicalUpdateAt) {
		t.Fatalf("expected technical updated_at %s to be ignored", technicalUpdateAt.Format(time.RFC3339))
	}
}

func TestDashboardSalespeopleShouldReturnStrategicMetrics(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	ctx := context.Background()
	var adminUserID int64
	if err := env.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'admin'`).Scan(&adminUserID); err != nil {
		t.Fatalf("failed to load admin user id: %v", err)
	}

	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO_"+uniqueSuffix(), "Pedido", true, 2)
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO_"+uniqueSuffix(), "Cancelado", true, 3)

	wonSentAt := time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC)
	lostSentAt := time.Date(2026, time.April, 2, 10, 0, 0, 0, time.UTC)
	wonClosedAt := time.Date(2026, time.April, 5, 10, 0, 0, 0, time.UTC)
	lostClosedAt := time.Date(2026, time.April, 8, 10, 0, 0, 0, time.UTC)

	createWonResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"DASH-STRAT-001",
			2026,
			wonSentAt,
			1000,
			seed,
			"Projetista Estrategico A",
			"Concorrente Estrategico A",
		),
	)
	if createWonResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createWonResponse.Code)
	}
	wonPayload := decodeJSONResponse[createResourceResponse](t, createWonResponse.Body)

	createLostResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"DASH-STRAT-002",
			2026,
			lostSentAt,
			2000,
			seed,
			"Projetista Estrategico B",
			"Concorrente Estrategico B",
		),
	)
	if createLostResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createLostResponse.Code)
	}
	lostPayload := decodeJSONResponse[createResourceResponse](t, createLostResponse.Body)

	env.insertReturningID(
		t,
		ctx,
		`INSERT INTO budget_status_history (budget_id, from_status_id, to_status_id, changed_by_user_id, notes, changed_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		wonPayload.ID,
		seed.statusID,
		pedidoStatusID,
		adminUserID,
		"Fechamento como pedido",
		wonClosedAt,
		wonClosedAt,
		wonClosedAt,
	)
	env.insertReturningID(
		t,
		ctx,
		`INSERT INTO budget_status_history (budget_id, from_status_id, to_status_id, changed_by_user_id, notes, changed_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		lostPayload.ID,
		seed.statusID,
		canceladoStatusID,
		adminUserID,
		"Fechamento como cancelado",
		lostClosedAt,
		lostClosedAt,
		lostClosedAt,
	)

	if _, err := env.db.ExecContext(
		ctx,
		`UPDATE budgets SET status_id = $2 WHERE id = $1`,
		wonPayload.ID,
		pedidoStatusID,
	); err != nil {
		t.Fatalf("failed to update won budget status: %v", err)
	}
	if _, err := env.db.ExecContext(
		ctx,
		`UPDATE budgets SET status_id = $2 WHERE id = $1`,
		lostPayload.ID,
		canceladoStatusID,
	); err != nil {
		t.Fatalf("failed to update lost budget status: %v", err)
	}

	dashboardResponse := env.doJSONRequest(t, http.MethodGet, "/dashboard/salespeople", adminToken, "")
	if dashboardResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, dashboardResponse.Code)
	}

	dashboardPayload := decodeJSONResponse[dto.SalespeopleDashboardResponse](t, dashboardResponse.Body)
	if math.Abs(dashboardPayload.Summary.ValueConversionRate-33.3333333333) > 0.01 {
		t.Fatalf("expected value conversion rate close to 33.33, got %f", dashboardPayload.Summary.ValueConversionRate)
	}
	if len(dashboardPayload.TopConstructionCompanies) != 1 {
		t.Fatalf("expected 1 construction company item, got %d", len(dashboardPayload.TopConstructionCompanies))
	}
	if dashboardPayload.TopConstructionCompanies[0].Label != "Construtora Teste" {
		t.Fatalf("expected construction company label Construtora Teste, got %s", dashboardPayload.TopConstructionCompanies[0].Label)
	}
	if len(dashboardPayload.TopProjects) != 1 {
		t.Fatalf("expected 1 project item, got %d", len(dashboardPayload.TopProjects))
	}
	if dashboardPayload.TopProjects[0].Label != seed.projectName {
		t.Fatalf("expected project label %s, got %s", seed.projectName, dashboardPayload.TopProjects[0].Label)
	}
	if len(dashboardPayload.TopLossReasons) != 1 {
		t.Fatalf("expected 1 loss reason item, got %d", len(dashboardPayload.TopLossReasons))
	}
	if dashboardPayload.TopLossReasons[0].Label == "" {
		t.Fatal("expected loss reason label to be returned")
	}
	if len(dashboardPayload.AverageClosingTimes) != 3 {
		t.Fatalf("expected 3 closing time rows, got %d", len(dashboardPayload.AverageClosingTimes))
	}
	if dashboardPayload.AverageClosingTimes[0].Label != "Geral" {
		t.Fatalf("expected first closing time label Geral, got %s", dashboardPayload.AverageClosingTimes[0].Label)
	}
	if math.Abs(dashboardPayload.AverageClosingTimes[0].AverageClosingDays-5) > 0.01 {
		t.Fatalf("expected average closing days close to 5, got %f", dashboardPayload.AverageClosingTimes[0].AverageClosingDays)
	}
}

func TestDashboardSalespeopleShouldReturnTechnicalOverviewByEstimator(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	now := time.Now()
	ctx := context.Background()
	firstEstimatorID := env.insertReturningID(
		t,
		ctx,
		`INSERT INTO estimators (code, name, email, phone, active, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		"EST-DASH-001",
		"Orcamentista Dashboard A",
		"dash.a@local.dev",
		"11999990001",
		true,
		"orcamentista tecnico A",
		now,
		now,
	)
	secondEstimatorID := env.insertReturningID(
		t,
		ctx,
		`INSERT INTO estimators (code, name, email, phone, active, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		"EST-DASH-002",
		"Orcamentista Dashboard B",
		"dash.b@local.dev",
		"11999990002",
		true,
		"orcamentista tecnico B",
		now,
		now,
	)

	firstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"DASH-TEC-001",
			2026,
			time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			4500,
			seed,
			"Projetista Tecnico A",
			"Concorrente Tecnico A",
		),
	)
	if firstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstBudgetResponse.Code)
	}
	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, firstBudgetResponse.Body)

	secondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"DASH-TEC-002",
			2026,
			time.Date(2026, time.May, 3, 10, 0, 0, 0, time.UTC),
			1800,
			seed,
			"Projetista Tecnico B",
			"Concorrente Tecnico B",
		),
	)
	if secondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, secondBudgetResponse.Code)
	}
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, secondBudgetResponse.Body)

	if _, err := env.db.ExecContext(
		ctx,
		`UPDATE budgets SET estimator_id = $2 WHERE id = $1`,
		firstBudgetPayload.ID,
		firstEstimatorID,
	); err != nil {
		t.Fatalf("failed to assign first estimator: %v", err)
	}
	if _, err := env.db.ExecContext(
		ctx,
		`UPDATE budgets SET estimator_id = $2 WHERE id = $1`,
		secondBudgetPayload.ID,
		secondEstimatorID,
	); err != nil {
		t.Fatalf("failed to assign second estimator: %v", err)
	}

	dashboardResponse := env.doJSONRequest(t, http.MethodGet, "/dashboard/salespeople", adminToken, "")
	if dashboardResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, dashboardResponse.Code)
	}

	dashboardPayload := decodeJSONResponse[dto.SalespeopleDashboardResponse](t, dashboardResponse.Body)
	if dashboardPayload.TechnicalOverview.Summary.ActiveEstimators != 2 {
		t.Fatalf("expected 2 active estimators, got %d", dashboardPayload.TechnicalOverview.Summary.ActiveEstimators)
	}
	if dashboardPayload.TechnicalOverview.Summary.BudgetsWithEstimator != 2 {
		t.Fatalf("expected 2 budgets with estimator, got %d", dashboardPayload.TechnicalOverview.Summary.BudgetsWithEstimator)
	}
	if len(dashboardPayload.TechnicalOverview.TopEstimatorsByValue) != 2 {
		t.Fatalf("expected 2 estimator ranking items, got %d", len(dashboardPayload.TechnicalOverview.TopEstimatorsByValue))
	}
	if dashboardPayload.TechnicalOverview.TopEstimatorsByValue[0].Label != "Orcamentista Dashboard A" {
		t.Fatalf(
			"expected first estimator ranking label Orcamentista Dashboard A, got %s",
			dashboardPayload.TechnicalOverview.TopEstimatorsByValue[0].Label,
		)
	}
}
