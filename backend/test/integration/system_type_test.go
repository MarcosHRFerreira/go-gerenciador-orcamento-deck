package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

func TestSystemTypesShouldListDefaultsAndBeReturnedInBudget(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	systemTypeID := env.requireSystemTypeIDByCode(t, "VRF")

	listResponse := env.doJSONRequest(t, http.MethodGet, "/system-types", token, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	systemTypes := decodeJSONResponse[[]dto.SystemTypeResponse](t, listResponse.Body)
	if len(systemTypes) < 3 {
		t.Fatalf("expected at least 3 system types, got %d", len(systemTypes))
	}

	foundVRF := false
	for _, item := range systemTypes {
		if item.Code == "VRF" && item.Name == "VRF" {
			foundVRF = true
			break
		}
	}
	if !foundVRF {
		t.Fatal("expected VRF system type to be listed")
	}

	sentAt := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	createBody := fmt.Sprintf(`{
		"budget_number":"BGT-SYSTEM-001",
		"year_budget":2026,
		"revision":0,
		"sent_at":"%s",
		"gross_value":1800.50,
		"commission_value":150.00,
		"area_m2":55.2,
		"status_id":%d,
		"priority_id":%d,
		"installer_id":%d,
		"system_type_id":%d,
		"project_id":%d,
		"salesperson_id":%d,
		"contact_id":%d,
		"loss_reason_id":%d,
		"competitor_name":"Concorrente Sistema",
		"competitor_price":999.99,
		"projetista_name":"Projetista Sistema",
		"specification_details":"Detalhes do sistema",
		"current_follow_up":"Aguardando retorno"
	}`,
		sentAt.Format(time.RFC3339),
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		systemTypeID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
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
	if getPayload.SystemTypeID == nil || *getPayload.SystemTypeID != systemTypeID {
		t.Fatalf("expected system type id %d, got %v", systemTypeID, getPayload.SystemTypeID)
	}
	if getPayload.SystemTypeCode == nil || *getPayload.SystemTypeCode != "VRF" {
		t.Fatalf("expected system type code VRF, got %v", getPayload.SystemTypeCode)
	}
	if getPayload.SystemTypeName == nil || *getPayload.SystemTypeName != "VRF" {
		t.Fatalf("expected system type name VRF, got %v", getPayload.SystemTypeName)
	}
}

func TestBudgetsShouldFilterBySystemType(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	vrfSystemTypeID := env.requireSystemTypeIDByCode(t, "VRF")
	exaustaoSystemTypeID := env.requireSystemTypeIDByCode(t, "EXAUSTAO")

	firstBudgetID := env.createBudgetWithSystemType(
		t,
		token,
		seed,
		"BGT-FILTER-SYSTEM-001",
		vrfSystemTypeID,
	)
	_ = env.createBudgetWithSystemType(
		t,
		token,
		seed,
		"BGT-FILTER-SYSTEM-002",
		exaustaoSystemTypeID,
	)

	listResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets?system_type_id=%d", vrfSystemTypeID),
		token,
		"",
	)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	listPayload := decodeJSONResponse[dto.ListBudgetsResponse](t, listResponse.Body)
	if listPayload.Total != 1 {
		t.Fatalf("expected total 1, got %d", listPayload.Total)
	}
	if len(listPayload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(listPayload.Items))
	}

	firstItem := listPayload.Items[0]
	if firstItem.ID != firstBudgetID {
		t.Fatalf("expected budget id %d, got %d", firstBudgetID, firstItem.ID)
	}
	if firstItem.SystemTypeID == nil || *firstItem.SystemTypeID != vrfSystemTypeID {
		t.Fatalf("expected system type id %d, got %v", vrfSystemTypeID, firstItem.SystemTypeID)
	}
	if firstItem.SystemTypeCode == nil || *firstItem.SystemTypeCode != "VRF" {
		t.Fatalf("expected system type code VRF, got %v", firstItem.SystemTypeCode)
	}
}

func TestSystemTypesShouldSupportAdminCrud(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)

	createBody := `{
		"code":"SISTEMA_TESTE",
		"name":"Sistema Teste",
		"description":"Descricao inicial"
	}`

	createResponse := env.doJSONRequest(t, http.MethodPost, "/system-types", token, createBody)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createResponse.Code)
	}

	createPayload := decodeJSONResponse[createResourceResponse](t, createResponse.Body)
	if createPayload.ID <= 0 {
		t.Fatalf("expected system type id greater than zero, got %d", createPayload.ID)
	}

	getResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/system-types/%d", createPayload.ID),
		token,
		"",
	)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getResponse.Code)
	}

	getPayload := decodeJSONResponse[dto.SystemTypeResponse](t, getResponse.Body)
	if getPayload.Code != "SISTEMA_TESTE" {
		t.Fatalf("expected code SISTEMA_TESTE, got %s", getPayload.Code)
	}
	if getPayload.Name != "Sistema Teste" {
		t.Fatalf("expected name Sistema Teste, got %s", getPayload.Name)
	}

	updateBody := `{
		"code":"SISTEMA_TESTE_ATUALIZADO",
		"name":"Sistema Teste Atualizado",
		"description":"Descricao atualizada"
	}`

	updateResponse := env.doJSONRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/system-types/%d", createPayload.ID),
		token,
		updateBody,
	)
	if updateResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateResponse.Code)
	}

	updatedResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/system-types/%d", createPayload.ID),
		token,
		"",
	)
	if updatedResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updatedResponse.Code)
	}

	updatedPayload := decodeJSONResponse[dto.SystemTypeResponse](t, updatedResponse.Body)
	if updatedPayload.Code != "SISTEMA_TESTE_ATUALIZADO" {
		t.Fatalf("expected updated code, got %s", updatedPayload.Code)
	}
	if updatedPayload.Name != "Sistema Teste Atualizado" {
		t.Fatalf("expected updated name, got %s", updatedPayload.Name)
	}

	deleteResponse := env.doJSONRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/system-types/%d", createPayload.ID),
		token,
		"",
	)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}

	notFoundResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/system-types/%d", createPayload.ID),
		token,
		"",
	)
	if notFoundResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, notFoundResponse.Code)
	}
}

func (e *integrationTestEnv) createBudgetWithSystemType(
	t *testing.T,
	token string,
	seed budgetSeedData,
	budgetNumber string,
	systemTypeID int64,
) int64 {
	t.Helper()

	sentAt := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	createBody := fmt.Sprintf(`{
		"budget_number":"%s",
		"year_budget":2026,
		"revision":0,
		"sent_at":"%s",
		"gross_value":1800.50,
		"commission_value":150.00,
		"area_m2":55.2,
		"status_id":%d,
		"priority_id":%d,
		"installer_id":%d,
		"system_type_id":%d,
		"project_id":%d,
		"salesperson_id":%d,
		"contact_id":%d,
		"loss_reason_id":%d,
		"competitor_name":"Concorrente Sistema",
		"competitor_price":999.99,
		"projetista_name":"Projetista Sistema",
		"specification_details":"Detalhes do sistema",
		"current_follow_up":"Aguardando retorno"
	}`,
		budgetNumber,
		sentAt.Format(time.RFC3339),
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		systemTypeID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
	)

	createResponse := e.doJSONRequest(t, http.MethodPost, "/budgets", token, createBody)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createResponse.Code)
	}

	createPayload := decodeJSONResponse[createResourceResponse](t, createResponse.Body)
	if createPayload.ID <= 0 {
		t.Fatalf("expected budget id greater than zero, got %d", createPayload.ID)
	}

	return createPayload.ID
}

func (e *integrationTestEnv) requireSystemTypeIDByCode(t *testing.T, code string) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), integrationRequestTimeout)
	defer cancel()

	var systemTypeID int64
	err := e.db.QueryRowContext(ctx, "SELECT id FROM system_types WHERE code = $1", code).Scan(&systemTypeID)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("expected system type with code %s to exist", code)
		}

		t.Fatalf("failed to load system type %s: %v", code, err)
	}

	return systemTypeID
}
