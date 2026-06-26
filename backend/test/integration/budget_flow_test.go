package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
)

const automaticProjectCancellationNote = "Cancelado automaticamente porque outro orcamento da obra foi marcado como Fechado"
const automaticProjectRestorationNote = "Status restaurado automaticamente para permitir a troca do vencedor da obra"

func TestBudgetFollowUpsShouldCreateListAndSyncCurrentFollowUp(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	createBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"FUP-001",
			2026,
			time.Date(2026, time.April, 1, 10, 0, 0, 0, time.UTC),
			2100,
			seed,
			"Projetista Follow Up",
			"Concorrente Follow Up",
		),
	)
	if createBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponse.Code)
	}

	createBudgetPayload := decodeJSONResponse[createResourceResponse](t, createBudgetResponse.Body)
	followUpAt := time.Date(2026, time.April, 2, 14, 30, 0, 0, time.UTC)
	createFollowUpBody := fmt.Sprintf(`{
		"notes":"Cliente pediu retorno na sexta",
		"follow_up_at":"%s"
	}`, followUpAt.Format(time.RFC3339))

	createFollowUpResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/budgets/%d/follow-ups", createBudgetPayload.ID),
		token,
		createFollowUpBody,
	)
	if createFollowUpResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFollowUpResponse.Code)
	}

	listFollowUpsResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/follow-ups", createBudgetPayload.ID),
		token,
		"",
	)
	if listFollowUpsResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listFollowUpsResponse.Code)
	}

	followUpsPayload := decodeJSONResponse[[]dto.BudgetFollowUpResponse](t, listFollowUpsResponse.Body)
	if len(followUpsPayload) != 1 {
		t.Fatalf("expected 1 follow up, got %d", len(followUpsPayload))
	}
	if followUpsPayload[0].Notes != "Cliente pediu retorno na sexta" {
		t.Fatalf("expected follow up notes to be synced, got %s", followUpsPayload[0].Notes)
	}

	getBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", createBudgetPayload.ID),
		token,
		"",
	)
	if getBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getBudgetResponse.Code)
	}

	budgetPayload := decodeJSONResponse[dto.BudgetResponse](t, getBudgetResponse.Body)
	if budgetPayload.CurrentFollowUp != "Cliente pediu retorno na sexta" {
		t.Fatalf("expected current follow up to be updated, got %s", budgetPayload.CurrentFollowUp)
	}
}

func TestBudgetFollowUpsShouldReturnNotFoundForMissingBudget(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)

	response := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets/999999/follow-ups",
		token,
		`{"notes":"Tentativa sem budget"}`,
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, response.Body)
	if errorPayload.Message != "Orcamento nao encontrado" {
		t.Fatalf("expected Orcamento nao encontrado message, got %s", errorPayload.Message)
	}
}

func TestBudgetStatusHistoryShouldChangeStatusListAndSyncBudgetStatus(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	createBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"STS-001",
			2026,
			time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			2500,
			seed,
			"Projetista Status",
			"Concorrente Status",
		),
	)
	if createBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponse.Code)
	}

	createBudgetPayload := decodeJSONResponse[createResourceResponse](t, createBudgetResponse.Body)
	secondStatusID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO budget_statuses (code, name, description, is_final, sort_order, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"STATUS_CHANGED_"+uniqueSuffix(),
		"Status Alterado",
		"status alternativo para teste",
		false,
		2,
		time.Now(),
		time.Now(),
	)

	changeStatusResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", createBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Status alterado apos contato"}`, secondStatusID),
	)
	if changeStatusResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, changeStatusResponse.Code)
	}

	listHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", createBudgetPayload.ID),
		token,
		"",
	)
	if listHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listHistoryResponse.Code)
	}

	historyPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, listHistoryResponse.Body)
	if len(historyPayload) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(historyPayload))
	}
	if historyPayload[0].FromStatusID == nil || *historyPayload[0].FromStatusID != seed.statusID {
		t.Fatalf("expected from_status_id %d, got %v", seed.statusID, historyPayload[0].FromStatusID)
	}
	if historyPayload[0].ToStatusID != secondStatusID {
		t.Fatalf("expected to_status_id %d, got %d", secondStatusID, historyPayload[0].ToStatusID)
	}
	if historyPayload[0].Notes != "Status alterado apos contato" {
		t.Fatalf("expected history notes to be persisted, got %s", historyPayload[0].Notes)
	}

	getBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", createBudgetPayload.ID),
		token,
		"",
	)
	if getBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getBudgetResponse.Code)
	}

	budgetPayload := decodeJSONResponse[dto.BudgetResponse](t, getBudgetResponse.Body)
	if budgetPayload.StatusID != secondStatusID {
		t.Fatalf("expected budget status id %d, got %d", secondStatusID, budgetPayload.StatusID)
	}
}

func TestBudgetStatusHistoryShouldRejectSameStatus(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())

	createBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"STS-002",
			2026,
			time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC),
			2600,
			seed,
			"Projetista Status Igual",
			"Concorrente Status Igual",
		),
	)
	if createBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponse.Code)
	}

	createBudgetPayload := decodeJSONResponse[createResourceResponse](t, createBudgetResponse.Body)
	changeStatusResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", createBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Tentativa repetida"}`, seed.statusID),
	)
	if changeStatusResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, changeStatusResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, changeStatusResponse.Body)
	if errorPayload.Message != "O orcamento ja possui o status informado" {
		t.Fatalf("expected same status message, got %s", errorPayload.Message)
	}
}

func TestBudgetStatusHistoryShouldNotCancelOtherProjectBudgetsWhenOneBecomesPedido(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	orcamentoStatusID := insertNamedBudgetStatus(t, env, "ORCAMENTO", "ORCAMENTO", false, 1)
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "PEDIDO", true, 2)
	seed.statusID = orcamentoStatusID
	otherInstallerSeed.statusID = orcamentoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-001",
			2026,
			time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC),
			3200,
			seed,
			"Projetista Grupo 1",
			"Concorrente Grupo 1",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-002",
			2026,
			time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC),
			3300,
			seed,
			"Projetista Grupo 2",
			"Concorrente Grupo 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	createThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-003",
			2026,
			time.Date(2026, time.July, 3, 10, 0, 0, 0, time.UTC),
			3400,
			otherInstallerSeed,
			"Projetista Grupo 3",
			"Concorrente Grupo 3",
		),
	)
	if createThirdBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createThirdBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)
	thirdBudgetPayload := decodeJSONResponse[createResourceResponse](t, createThirdBudgetResponse.Body)

	changeStatusResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Obra aprovada"}`, pedidoStatusID),
	)
	if changeStatusResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, changeStatusResponse.Code)
	}

	getFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		"",
	)
	if getFirstBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	getThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", thirdBudgetPayload.ID),
		token,
		"",
	)
	if getThirdBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getThirdBudgetResponse.Code)
	}

	firstBudget := decodeJSONResponse[dto.BudgetResponse](t, getFirstBudgetResponse.Body)
	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	thirdBudget := decodeJSONResponse[dto.BudgetResponse](t, getThirdBudgetResponse.Body)
	if firstBudget.StatusID != pedidoStatusID {
		t.Fatalf("expected first budget status id %d, got %d", pedidoStatusID, firstBudget.StatusID)
	}
	if secondBudget.StatusID != orcamentoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", orcamentoStatusID, secondBudget.StatusID)
	}
	if thirdBudget.StatusID != orcamentoStatusID {
		t.Fatalf("expected third budget status id %d, got %d", orcamentoStatusID, thirdBudget.StatusID)
	}

	firstHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", firstBudgetPayload.ID),
		token,
		"",
	)
	if firstHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, firstHistoryResponse.Code)
	}

	secondHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", secondBudgetPayload.ID),
		token,
		"",
	)
	if secondHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, secondHistoryResponse.Code)
	}

	thirdHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", thirdBudgetPayload.ID),
		token,
		"",
	)
	if thirdHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, thirdHistoryResponse.Code)
	}

	firstHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, firstHistoryResponse.Body)
	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	thirdHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, thirdHistoryResponse.Body)
	if len(firstHistoryPayload) != 1 {
		t.Fatalf("expected 1 first history item, got %d", len(firstHistoryPayload))
	}
	if len(secondHistoryPayload) != 0 {
		t.Fatalf("expected 0 second history items for same installer budget, got %d", len(secondHistoryPayload))
	}
	if len(thirdHistoryPayload) != 0 {
		t.Fatalf("expected 0 third history items for isolated status-history flow, got %d", len(thirdHistoryPayload))
	}
	if firstHistoryPayload[0].FromStatusID == nil || *firstHistoryPayload[0].FromStatusID != orcamentoStatusID {
		t.Fatalf("expected first budget from_status_id %d, got %v", orcamentoStatusID, firstHistoryPayload[0].FromStatusID)
	}
	if firstHistoryPayload[0].ToStatusID != pedidoStatusID {
		t.Fatalf("expected first budget to_status_id %d, got %d", pedidoStatusID, firstHistoryPayload[0].ToStatusID)
	}
}

func TestBudgetUpdateShouldCancelOnlyOtherInstallerProjectBudgetsWhenOneBecomesPedido(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	emNegociacaoStatusID := getBudgetStatusIDByName(t, env, "Em Negociacao")
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO", "Cancelado "+uniqueSuffix(), true, 3)
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+uniqueSuffix(), true, 2)
	seed.statusID = emNegociacaoStatusID
	otherInstallerSeed.statusID = emNegociacaoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			4200,
			seed,
			"Projetista Grupo Update 1",
			"Concorrente Grupo Update 1",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-002",
			2026,
			time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC),
			4300,
			seed,
			"Projetista Grupo Update 2",
			"Concorrente Grupo Update 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	createThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-003",
			2026,
			time.Date(2026, time.September, 3, 10, 0, 0, 0, time.UTC),
			4400,
			otherInstallerSeed,
			"Projetista Grupo Update 3",
			"Concorrente Grupo Update 3",
		),
	)
	if createThirdBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createThirdBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)
	thirdBudgetPayload := decodeJSONResponse[createResourceResponse](t, createThirdBudgetResponse.Body)

	updateFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{
			"budget_number":"GRP-UPD-001",
			"year_budget":2026,
			"revision":0,
			"sent_at":"%s",
			"gross_value":4200.00,
			"commission_value":100.00,
			"area_m2":35.00,
			"status_id":%d,
			"priority_id":%d,
			"installer_id":%d,
			"project_id":%d,
			"salesperson_id":%d,
			"contact_id":%d,
			"loss_reason_id":%d,
			"construction_company":"Construtora Teste",
			"competitor_name":"Concorrente Grupo Update 1",
			"competitor_price":900.00,
			"projetista_name":"Projetista Grupo Update 1",
			"specification_details":"Especificacao de teste",
			"current_follow_up":"Follow up inicial"
		}`,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			pedidoStatusID,
			seed.priorityID,
			seed.installerID,
			seed.projectID,
			seed.salespersonID,
			seed.contactID,
			seed.lossReasonID,
		),
	)
	if updateFirstBudgetResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateFirstBudgetResponse.Code)
	}

	getFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		"",
	)
	if getFirstBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	firstBudget := decodeJSONResponse[dto.BudgetResponse](t, getFirstBudgetResponse.Body)
	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	if firstBudget.StatusID != pedidoStatusID {
		t.Fatalf("expected first budget status id %d, got %d", pedidoStatusID, firstBudget.StatusID)
	}
	if secondBudget.StatusID != emNegociacaoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", emNegociacaoStatusID, secondBudget.StatusID)
	}

	getThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", thirdBudgetPayload.ID),
		token,
		"",
	)
	if getThirdBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getThirdBudgetResponse.Code)
	}

	thirdBudget := decodeJSONResponse[dto.BudgetResponse](t, getThirdBudgetResponse.Body)
	if thirdBudget.StatusID != canceladoStatusID {
		t.Fatalf("expected third budget status id %d, got %d", canceladoStatusID, thirdBudget.StatusID)
	}

	secondHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", secondBudgetPayload.ID),
		token,
		"",
	)
	if secondHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, secondHistoryResponse.Code)
	}

	thirdHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", thirdBudgetPayload.ID),
		token,
		"",
	)
	if thirdHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, thirdHistoryResponse.Code)
	}

	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	thirdHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, thirdHistoryResponse.Body)
	if len(secondHistoryPayload) != 0 {
		t.Fatalf("expected no automatic history for same installer budget, got %d", len(secondHistoryPayload))
	}
	if len(thirdHistoryPayload) != 1 {
		t.Fatalf("expected 1 automatic cancellation history for other installer budget, got %d", len(thirdHistoryPayload))
	}
	if thirdHistoryPayload[0].Notes != automaticProjectCancellationNote {
		t.Fatalf("expected automatic cancellation note, got %s", thirdHistoryPayload[0].Notes)
	}
}

func TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWithoutInstallerWhenOneBecomesPedido(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	emNegociacaoStatusID := getBudgetStatusIDByName(t, env, "Em Negociacao")
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+uniqueSuffix(), true, 2)
	seed.statusID = emNegociacaoStatusID
	otherInstallerSeed.statusID = emNegociacaoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-NOINSTALLER-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			4200,
			seed,
			"Projetista Grupo Update Sem Instalador 1",
			"Concorrente Grupo Update Sem Instalador 1",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-NOINSTALLER-002",
			2026,
			time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC),
			4300,
			otherInstallerSeed,
			"Projetista Grupo Update Sem Instalador 2",
			"Concorrente Grupo Update Sem Instalador 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)

	updateFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{
			"budget_number":"GRP-UPD-NOINSTALLER-001",
			"year_budget":2026,
			"revision":0,
			"sent_at":"%s",
			"gross_value":4200.00,
			"commission_value":100.00,
			"area_m2":35.00,
			"status_id":%d,
			"priority_id":%d,
			"installer_id":null,
			"project_id":%d,
			"salesperson_id":%d,
			"contact_id":null,
			"loss_reason_id":%d,
			"construction_company":"Construtora Teste",
			"competitor_name":"Concorrente Grupo Update Sem Instalador 1",
			"competitor_price":900.00,
			"projetista_name":"Projetista Grupo Update Sem Instalador 1",
			"specification_details":"Especificacao de teste",
			"current_follow_up":"Follow up inicial"
		}`,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			pedidoStatusID,
			seed.priorityID,
			seed.projectID,
			seed.salespersonID,
			seed.lossReasonID,
		),
	)
	if updateFirstBudgetResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	if secondBudget.StatusID != emNegociacaoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", emNegociacaoStatusID, secondBudget.StatusID)
	}
}

func TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWithoutProjectWhenOneBecomesPedido(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	emNegociacaoStatusID := getBudgetStatusIDByName(t, env, "Em Negociacao")
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+uniqueSuffix(), true, 2)
	seed.statusID = emNegociacaoStatusID
	otherInstallerSeed.statusID = emNegociacaoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-NOPROJECT-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			4200,
			seed,
			"Projetista Grupo Update Sem Obra 1",
			"Concorrente Grupo Update Sem Obra 1",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-UPD-NOPROJECT-002",
			2026,
			time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC),
			4300,
			otherInstallerSeed,
			"Projetista Grupo Update Sem Obra 2",
			"Concorrente Grupo Update Sem Obra 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)

	updateFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{
			"budget_number":"GRP-UPD-NOPROJECT-001",
			"year_budget":2026,
			"revision":0,
			"sent_at":"%s",
			"gross_value":4200.00,
			"commission_value":100.00,
			"area_m2":35.00,
			"status_id":%d,
			"priority_id":%d,
			"installer_id":%d,
			"project_id":null,
			"salesperson_id":%d,
			"contact_id":%d,
			"loss_reason_id":%d,
			"construction_company":"Construtora Teste",
			"competitor_name":"Concorrente Grupo Update Sem Obra 1",
			"competitor_price":900.00,
			"projetista_name":"Projetista Grupo Update Sem Obra 1",
			"specification_details":"Especificacao de teste",
			"current_follow_up":"Follow up inicial"
		}`,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			pedidoStatusID,
			seed.priorityID,
			seed.installerID,
			seed.salespersonID,
			seed.contactID,
			seed.lossReasonID,
		),
	)
	if updateFirstBudgetResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	if secondBudget.StatusID != emNegociacaoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", emNegociacaoStatusID, secondBudget.StatusID)
	}
}

func TestBudgetElectWinnerShouldCancelOnlyOtherInstallerProjectBudgets(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	statusSuffix := uniqueSuffix()
	emNegociacaoStatusID := getBudgetStatusIDByName(t, env, "Em Negociacao")
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+statusSuffix, true, 2)
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO", "Cancelado "+statusSuffix, true, 3)
	seed.statusID = emNegociacaoStatusID
	otherInstallerSeed.statusID = emNegociacaoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-WIN-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			4200,
			seed,
			"Projetista Grupo Winner 1",
			"Concorrente Grupo Winner 1",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-WIN-002",
			2026,
			time.Date(2026, time.September, 2, 10, 0, 0, 0, time.UTC),
			4300,
			seed,
			"Projetista Grupo Winner 2",
			"Concorrente Grupo Winner 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	createThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-WIN-003",
			2026,
			time.Date(2026, time.September, 3, 10, 0, 0, 0, time.UTC),
			4400,
			otherInstallerSeed,
			"Projetista Grupo Winner 3",
			"Concorrente Grupo Winner 3",
		),
	)
	if createThirdBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createThirdBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)
	thirdBudgetPayload := decodeJSONResponse[createResourceResponse](t, createThirdBudgetResponse.Body)

	electWinnerResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/budgets/%d/elect-winner", firstBudgetPayload.ID),
		token,
		`{"notes":"Definido como vencedor da obra"}`,
	)
	if electWinnerResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, electWinnerResponse.Code)
	}

	getFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		"",
	)
	if getFirstBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	firstBudget := decodeJSONResponse[dto.BudgetResponse](t, getFirstBudgetResponse.Body)
	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	if firstBudget.StatusID != pedidoStatusID {
		t.Fatalf("expected first budget status id %d, got %d", pedidoStatusID, firstBudget.StatusID)
	}
	if secondBudget.StatusID != emNegociacaoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", emNegociacaoStatusID, secondBudget.StatusID)
	}

	getThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", thirdBudgetPayload.ID),
		token,
		"",
	)
	if getThirdBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getThirdBudgetResponse.Code)
	}

	thirdBudget := decodeJSONResponse[dto.BudgetResponse](t, getThirdBudgetResponse.Body)
	if thirdBudget.StatusID != canceladoStatusID {
		t.Fatalf("expected third budget status id %d, got %d", canceladoStatusID, thirdBudget.StatusID)
	}

	secondHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", secondBudgetPayload.ID),
		token,
		"",
	)
	if secondHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, secondHistoryResponse.Code)
	}

	thirdHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", thirdBudgetPayload.ID),
		token,
		"",
	)
	if thirdHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, thirdHistoryResponse.Code)
	}

	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	thirdHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, thirdHistoryResponse.Body)
	if len(secondHistoryPayload) != 0 {
		t.Fatalf("expected no history item for same installer complementary budget, got %d", len(secondHistoryPayload))
	}
	if len(thirdHistoryPayload) != 1 {
		t.Fatalf("expected 1 history item for other installer automatic cancellation, got %d", len(thirdHistoryPayload))
	}
	if thirdHistoryPayload[0].Notes != automaticProjectCancellationNote {
		t.Fatalf("expected automatic cancellation note, got %s", thirdHistoryPayload[0].Notes)
	}
}

func TestBudgetElectWinnerShouldReplacePreviousWinnerFromDifferentInstallerAndRestoreCurrentInstallerBudgets(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	otherInstallerSeed := env.seedBudgetData(t, uniqueSuffix())
	statusSuffix := uniqueSuffix()
	emNegociacaoStatusID := getBudgetStatusIDByName(t, env, "Em Negociacao")
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+statusSuffix, true, 2)
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO", "Cancelado "+statusSuffix, true, 3)
	seed.statusID = emNegociacaoStatusID
	otherInstallerSeed.statusID = emNegociacaoStatusID
	otherInstallerSeed.projectID = seed.projectID

	createFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-003",
			2026,
			time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC),
			3400,
			seed,
			"Projetista Grupo 3",
			"Concorrente Grupo 3",
		),
	)
	if createFirstBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createFirstBudgetResponse.Code)
	}

	createSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-004",
			2026,
			time.Date(2026, time.August, 2, 10, 0, 0, 0, time.UTC),
			3500,
			otherInstallerSeed,
			"Projetista Grupo 4",
			"Concorrente Grupo 4",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	createThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		token,
		buildBudgetRequestBody(
			"GRP-005",
			2026,
			time.Date(2026, time.August, 3, 10, 0, 0, 0, time.UTC),
			3600,
			otherInstallerSeed,
			"Projetista Grupo 5",
			"Concorrente Grupo 5",
		),
	)
	if createThirdBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createThirdBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)
	thirdBudgetPayload := decodeJSONResponse[createResourceResponse](t, createThirdBudgetResponse.Body)

	firstWinnerResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/budgets/%d/elect-winner", firstBudgetPayload.ID),
		token,
		`{"notes":"Primeira eleicao do vencedor"}`,
	)
	if firstWinnerResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, firstWinnerResponse.Code)
	}

	secondWinnerResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/budgets/%d/elect-winner", thirdBudgetPayload.ID),
		token,
		`{"notes":"Troca do vencedor da obra"}`,
	)
	if secondWinnerResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, secondWinnerResponse.Code)
	}

	getFirstBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", firstBudgetPayload.ID),
		token,
		"",
	)
	if getFirstBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getFirstBudgetResponse.Code)
	}

	getSecondBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", secondBudgetPayload.ID),
		token,
		"",
	)
	if getSecondBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getSecondBudgetResponse.Code)
	}

	getThirdBudgetResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d", thirdBudgetPayload.ID),
		token,
		"",
	)
	if getThirdBudgetResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getThirdBudgetResponse.Code)
	}

	firstBudget := decodeJSONResponse[dto.BudgetResponse](t, getFirstBudgetResponse.Body)
	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	thirdBudget := decodeJSONResponse[dto.BudgetResponse](t, getThirdBudgetResponse.Body)
	if firstBudget.StatusID != canceladoStatusID {
		t.Fatalf("expected first budget status id %d, got %d", canceladoStatusID, firstBudget.StatusID)
	}
	if secondBudget.StatusID != emNegociacaoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", emNegociacaoStatusID, secondBudget.StatusID)
	}
	if thirdBudget.StatusID != pedidoStatusID {
		t.Fatalf("expected third budget status id %d, got %d", pedidoStatusID, thirdBudget.StatusID)
	}

	firstHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", firstBudgetPayload.ID),
		token,
		"",
	)
	if firstHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, firstHistoryResponse.Code)
	}

	secondHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", secondBudgetPayload.ID),
		token,
		"",
	)
	if secondHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, secondHistoryResponse.Code)
	}

	thirdHistoryResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/budgets/%d/status-history", thirdBudgetPayload.ID),
		token,
		"",
	)
	if thirdHistoryResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, thirdHistoryResponse.Code)
	}

	firstHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, firstHistoryResponse.Body)
	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	thirdHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, thirdHistoryResponse.Body)
	if len(firstHistoryPayload) < 3 {
		t.Fatalf("expected at least 3 history items for previous winner, got %d", len(firstHistoryPayload))
	}
	if len(secondHistoryPayload) < 2 {
		t.Fatalf("expected at least 2 history items for restored current installer budget, got %d", len(secondHistoryPayload))
	}
	if len(thirdHistoryPayload) < 2 {
		t.Fatalf("expected at least 2 history items for new winner, got %d", len(thirdHistoryPayload))
	}
	if firstHistoryPayload[0].Notes != automaticProjectCancellationNote {
		t.Fatalf("expected latest first budget history note to be automatic cancellation, got %s", firstHistoryPayload[0].Notes)
	}
	if secondHistoryPayload[0].Notes != automaticProjectRestorationNote {
		t.Fatalf("expected latest second budget history note to be automatic restoration, got %s", secondHistoryPayload[0].Notes)
	}
	if thirdHistoryPayload[0].ToStatusID != pedidoStatusID {
		t.Fatalf("expected latest third budget to_status_id %d, got %d", pedidoStatusID, thirdHistoryPayload[0].ToStatusID)
	}
}

func insertNamedBudgetStatus(t *testing.T, env *integrationTestEnv, code string, name string, isFinal bool, sortOrder int) int64 {
	t.Helper()

	now := time.Now()
	return env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO budget_statuses (code, name, description, is_final, sort_order, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		code,
		name,
		"status semantico de teste",
		isFinal,
		sortOrder,
		now,
		now,
	)
}

func getBudgetStatusIDByName(t *testing.T, env *integrationTestEnv, name string) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), integrationRequestTimeout)
	defer cancel()

	var id int64
	err := env.db.QueryRowContext(
		ctx,
		`SELECT id FROM budget_statuses WHERE name = $1 ORDER BY id ASC LIMIT 1`,
		name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to query budget status %s: %v", name, err)
	}

	return id
}
