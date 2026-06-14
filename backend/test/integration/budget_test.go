package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
)

type createResourceResponse struct {
	ID int64 `json:"id"`
}

func TestBudgetsCRUDFlow(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	sentAt := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	competitorPrice := 990.5
	createBody := fmt.Sprintf(`{
		"budget_number":"BGT-100",
		"year_budget":2026,
		"revision":0,
		"sent_at":"%s",
		"gross_value":1500.75,
		"commission_value":125.25,
		"area_m2":45.5,
		"status_id":%d,
		"priority_id":%d,
		"installer_id":%d,
		"project_id":%d,
		"salesperson_id":%d,
		"contact_id":%d,
		"loss_reason_id":%d,
		"competitor_name":"Concorrente A",
		"competitor_price":%.2f,
		"designer_name":"Designer A",
		"specification_details":"Especificacao inicial",
		"current_follow_up":"Primeiro contato"
	}`,
		sentAt.Format(time.RFC3339),
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
		competitorPrice,
	)

	createResponse := env.doJSONRequest(t, http.MethodPost, "/budgets", token, createBody)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createResponse.Code)
	}

	createPayload := decodeJSONResponse[createResourceResponse](t, createResponse.Body)
	if createPayload.ID <= 0 {
		t.Fatalf("expected budget id greater than zero, got %d", createPayload.ID)
	}

	getResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", createPayload.ID), token, "")
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getResponse.Code)
	}

	getPayload := decodeJSONResponse[dto.BudgetResponse](t, getResponse.Body)
	if getPayload.BudgetNumber != "BGT-100" {
		t.Fatalf("expected budget number BGT-100, got %s", getPayload.BudgetNumber)
	}
	if getPayload.ProjectID == nil || *getPayload.ProjectID != seed.projectID {
		t.Fatalf("expected project id %d, got %v", seed.projectID, getPayload.ProjectID)
	}
	if getPayload.CompetitorPrice == nil || *getPayload.CompetitorPrice != competitorPrice {
		t.Fatalf("expected competitor price %.2f, got %v", competitorPrice, getPayload.CompetitorPrice)
	}

	updatedSentAt := time.Date(2026, time.February, 15, 15, 30, 0, 0, time.UTC)
	updateBody := fmt.Sprintf(`{
		"budget_number":"BGT-101",
		"year_budget":2027,
		"revision":2,
		"sent_at":"%s",
		"gross_value":2750.90,
		"commission_value":200.10,
		"area_m2":60.25,
		"status_id":%d,
		"priority_id":null,
		"installer_id":null,
		"project_id":null,
		"salesperson_id":%d,
		"contact_id":null,
		"loss_reason_id":null,
		"competitor_name":"Concorrente B",
		"competitor_price":null,
		"designer_name":"Designer B",
		"specification_details":"Especificacao atualizada",
		"current_follow_up":"Retorno enviado"
	}`,
		updatedSentAt.Format(time.RFC3339),
		seed.statusID,
		seed.salespersonID,
	)

	updateResponse := env.doJSONRequest(t, http.MethodPut, fmt.Sprintf("/budgets/%d", createPayload.ID), token, updateBody)
	if updateResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateResponse.Code)
	}

	getUpdatedResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", createPayload.ID), token, "")
	if getUpdatedResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getUpdatedResponse.Code)
	}

	updatedPayload := decodeJSONResponse[dto.BudgetResponse](t, getUpdatedResponse.Body)
	if updatedPayload.BudgetNumber != "BGT-101" {
		t.Fatalf("expected updated budget number BGT-101, got %s", updatedPayload.BudgetNumber)
	}
	if updatedPayload.YearBudget != 2027 {
		t.Fatalf("expected year budget 2027, got %d", updatedPayload.YearBudget)
	}
	if updatedPayload.PriorityID != nil {
		t.Fatalf("expected priority id to be nil, got %v", updatedPayload.PriorityID)
	}
	if updatedPayload.ProjectID != nil {
		t.Fatalf("expected project id to be nil, got %v", updatedPayload.ProjectID)
	}
	if updatedPayload.CompetitorPrice != nil {
		t.Fatalf("expected competitor price to be nil, got %v", updatedPayload.CompetitorPrice)
	}

	deleteResponse := env.doJSONRequest(t, http.MethodDelete, fmt.Sprintf("/budgets/%d", createPayload.ID), token, "")
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}

	getAfterDeleteResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", createPayload.ID), token, "")
	if getAfterDeleteResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getAfterDeleteResponse.Code)
	}
}

func TestBudgetsListShouldSupportFiltersPaginationAndSorting(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seedA := env.seedBudgetData(t, uniqueSuffix())
	seedB := env.seedBudgetData(t, uniqueSuffix())

	createBudgetResponseA1 := env.doJSONRequest(t, http.MethodPost, "/budgets", token, buildBudgetRequestBody(
		"AAA-001",
		2026,
		time.Date(2026, time.January, 5, 10, 0, 0, 0, time.UTC),
		1200,
		seedA,
		"Designer Alfa",
		"Concorrente Alfa",
	))
	if createBudgetResponseA1.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponseA1.Code)
	}

	createBudgetResponseA2 := env.doJSONRequest(t, http.MethodPost, "/budgets", token, buildBudgetRequestBody(
		"AAA-002",
		2026,
		time.Date(2026, time.January, 6, 10, 0, 0, 0, time.UTC),
		1800,
		seedA,
		"Designer Alfa",
		"Concorrente Alfa",
	))
	if createBudgetResponseA2.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponseA2.Code)
	}

	createBudgetResponseB := env.doJSONRequest(t, http.MethodPost, "/budgets", token, buildBudgetRequestBody(
		"BBB-001",
		2026,
		time.Date(2026, time.February, 1, 10, 0, 0, 0, time.UTC),
		3000,
		seedB,
		"Designer Beta",
		"Concorrente Beta",
	))
	if createBudgetResponseB.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponseB.Code)
	}

	listPath := fmt.Sprintf(
		"/budgets?year_budget=2026&status_id=%d&project_type_id=%d&designer_name=Designer%%20Alfa&page=1&page_size=1&sort_by=budget_number&sort_order=asc",
		seedA.statusID,
		seedA.projectTypeID,
	)
	listResponse := env.doJSONRequest(t, http.MethodGet, listPath, token, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	listPayload := decodeJSONResponse[dto.ListBudgetsResponse](t, listResponse.Body)
	if listPayload.Total != 2 {
		t.Fatalf("expected total 2, got %d", listPayload.Total)
	}
	if listPayload.Page != 1 {
		t.Fatalf("expected page 1, got %d", listPayload.Page)
	}
	if listPayload.PageSize != 1 {
		t.Fatalf("expected page size 1, got %d", listPayload.PageSize)
	}
	if len(listPayload.Items) != 1 {
		t.Fatalf("expected 1 listed item, got %d", len(listPayload.Items))
	}
	if listPayload.Items[0].BudgetNumber != "AAA-001" {
		t.Fatalf("expected first listed budget AAA-001, got %s", listPayload.Items[0].BudgetNumber)
	}
}

func TestBudgetsCreateShouldRejectDuplicateNumberAndYear(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	requestBody := buildBudgetRequestBody(
		"DUP-001",
		2026,
		time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC),
		1900,
		seed,
		"Designer Duplicado",
		"Concorrente Duplicado",
	)

	firstCreateResponse := env.doJSONRequest(t, http.MethodPost, "/budgets", token, requestBody)
	if firstCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstCreateResponse.Code)
	}

	secondCreateResponse := env.doJSONRequest(t, http.MethodPost, "/budgets", token, requestBody)
	if secondCreateResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondCreateResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, secondCreateResponse.Body)
	if errorPayload.Message != "budget already exists for informed budget_number and year_budget" {
		t.Fatalf("expected duplicate budget message, got %s", errorPayload.Message)
	}
}

func TestBudgetsWriteRoutesShouldRespectAdminAuthorization(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userToken := env.createUserToken(t, adminToken, uniqueSuffix(), "user")
	seed := env.seedBudgetData(t, uniqueSuffix())

	createBody := buildBudgetRequestBody(
		"AUTH-001",
		2026,
		time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC),
		1800,
		seed,
		"Designer Auth",
		"Concorrente Auth",
	)

	forbiddenCreateResponse := env.doJSONRequest(t, http.MethodPost, "/budgets", userToken, createBody)
	if forbiddenCreateResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenCreateResponse.Code)
	}

	forbiddenCreatePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenCreateResponse.Body)
	if forbiddenCreatePayload.Message != "insufficient permissions" {
		t.Fatalf("expected forbidden message, got %s", forbiddenCreatePayload.Message)
	}

	adminCreateResponse := env.doJSONRequest(t, http.MethodPost, "/budgets", adminToken, createBody)
	if adminCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, adminCreateResponse.Code)
	}

	createdBudget := decodeJSONResponse[createResourceResponse](t, adminCreateResponse.Body)
	updateBody := buildBudgetRequestBody(
		"AUTH-002",
		2026,
		time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC),
		1900,
		seed,
		"Designer Auth Atualizado",
		"Concorrente Auth Atualizado",
	)

	forbiddenUpdateResponse := env.doJSONRequest(t, http.MethodPut, fmt.Sprintf("/budgets/%d", createdBudget.ID), userToken, updateBody)
	if forbiddenUpdateResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenUpdateResponse.Code)
	}

	forbiddenUpdatePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenUpdateResponse.Body)
	if forbiddenUpdatePayload.Message != "insufficient permissions" {
		t.Fatalf("expected forbidden message, got %s", forbiddenUpdatePayload.Message)
	}

	forbiddenDeleteResponse := env.doJSONRequest(t, http.MethodDelete, fmt.Sprintf("/budgets/%d", createdBudget.ID), userToken, "")
	if forbiddenDeleteResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, forbiddenDeleteResponse.Code)
	}

	forbiddenDeletePayload := decodeJSONResponse[httpresponse.ErrorResponse](t, forbiddenDeleteResponse.Body)
	if forbiddenDeletePayload.Message != "insufficient permissions" {
		t.Fatalf("expected forbidden message, got %s", forbiddenDeletePayload.Message)
	}
}

func TestBudgetsReadRoutesShouldRestrictUserToOwnSalespersonScope(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	ownSuffix := uniqueSuffix()
	otherSuffix := uniqueSuffix()
	seedOwn := env.seedBudgetData(t, ownSuffix)
	seedOther := env.seedBudgetData(t, otherSuffix)

	ownCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"SCOPE-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			2100,
			seedOwn,
			"Designer Scope",
			"Concorrente Scope",
		),
	)
	if ownCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, ownCreateResponse.Code)
	}
	ownBudget := decodeJSONResponse[createResourceResponse](t, ownCreateResponse.Body)

	otherCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"SCOPE-002",
			2026,
			time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC),
			2200,
			seedOther,
			"Designer Scope B",
			"Concorrente Scope B",
		),
	)
	if otherCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, otherCreateResponse.Code)
	}
	otherBudget := decodeJSONResponse[createResourceResponse](t, otherCreateResponse.Body)

	normalizedOwnSuffix := strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(ownSuffix))
	userToken := env.createUserTokenWithCredentials(
		t,
		adminToken,
		"Vendedor Escopo",
		fmt.Sprintf("scope.user.%s@local.dev", normalizedOwnSuffix),
		fmt.Sprintf("sales.%s", normalizedOwnSuffix),
		"user",
	)

	listResponse := env.doJSONRequest(t, http.MethodGet, "/budgets", userToken, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	listPayload := decodeJSONResponse[dto.ListBudgetsResponse](t, listResponse.Body)
	if listPayload.Total != 1 {
		t.Fatalf("expected total 1, got %d", listPayload.Total)
	}
	if len(listPayload.Items) != 1 {
		t.Fatalf("expected 1 listed item, got %d", len(listPayload.Items))
	}
	if listPayload.Items[0].ID != ownBudget.ID {
		t.Fatalf("expected only own budget id %d, got %d", ownBudget.ID, listPayload.Items[0].ID)
	}

	ownGetResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", ownBudget.ID), userToken, "")
	if ownGetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ownGetResponse.Code)
	}

	otherGetResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", otherBudget.ID), userToken, "")
	if otherGetResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, otherGetResponse.Code)
	}
}

func TestBudgetsReadRoutesShouldRestrictUserToOwnSalespersonScopeBySalespersonName(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	ownSuffix := uniqueSuffix()
	otherSuffix := uniqueSuffix()
	seedOwn := env.seedBudgetData(t, ownSuffix)
	seedOther := env.seedBudgetData(t, otherSuffix)

	ownCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"SCOPE-NAME-001",
			2026,
			time.Date(2026, time.October, 1, 10, 0, 0, 0, time.UTC),
			2300,
			seedOwn,
			"Designer Scope Name",
			"Concorrente Scope Name",
		),
	)
	if ownCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, ownCreateResponse.Code)
	}
	ownBudget := decodeJSONResponse[createResourceResponse](t, ownCreateResponse.Body)

	otherCreateResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		adminToken,
		buildBudgetRequestBody(
			"SCOPE-NAME-002",
			2026,
			time.Date(2026, time.October, 2, 10, 0, 0, 0, time.UTC),
			2400,
			seedOther,
			"Designer Scope Name B",
			"Concorrente Scope Name B",
		),
	)
	if otherCreateResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, otherCreateResponse.Code)
	}
	otherBudget := decodeJSONResponse[createResourceResponse](t, otherCreateResponse.Body)

	userToken := env.createUserTokenWithCredentials(
		t,
		adminToken,
		"Usuario Guilherme",
		fmt.Sprintf("scope.name.user.%s@local.dev", strings.ToLower(ownSuffix)),
		fmt.Sprintf("Vendedor %s", ownSuffix),
		"user",
	)

	listResponse := env.doJSONRequest(t, http.MethodGet, "/budgets", userToken, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	listPayload := decodeJSONResponse[dto.ListBudgetsResponse](t, listResponse.Body)
	if listPayload.Total != 1 {
		t.Fatalf("expected total 1, got %d", listPayload.Total)
	}
	if len(listPayload.Items) != 1 {
		t.Fatalf("expected 1 listed item, got %d", len(listPayload.Items))
	}
	if listPayload.Items[0].ID != ownBudget.ID {
		t.Fatalf("expected only own budget id %d, got %d", ownBudget.ID, listPayload.Items[0].ID)
	}

	otherGetResponse := env.doJSONRequest(t, http.MethodGet, fmt.Sprintf("/budgets/%d", otherBudget.ID), userToken, "")
	if otherGetResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, otherGetResponse.Code)
	}
}

func buildBudgetRequestBody(
	budgetNumber string,
	yearBudget int,
	sentAt time.Time,
	grossValue float64,
	seed budgetSeedData,
	designerName string,
	competitorName string,
) string {
	return fmt.Sprintf(`{
		"budget_number":"%s",
		"year_budget":%d,
		"revision":0,
		"sent_at":"%s",
		"gross_value":%.2f,
		"commission_value":100.00,
		"area_m2":35.00,
		"status_id":%d,
		"priority_id":%d,
		"installer_id":%d,
		"project_id":%d,
		"salesperson_id":%d,
		"contact_id":%d,
		"loss_reason_id":%d,
		"competitor_name":"%s",
		"competitor_price":900.00,
		"designer_name":"%s",
		"specification_details":"Especificacao de teste",
		"current_follow_up":"Follow up inicial"
	}`,
		budgetNumber,
		yearBudget,
		sentAt.Format(time.RFC3339),
		grossValue,
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
		competitorName,
		designerName,
	)
}
