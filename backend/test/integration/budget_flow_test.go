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

const automaticProjectCancellationNote = "Cancelado automaticamente porque outro orcamento do projeto foi marcado como PEDIDO"

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
			"Designer Follow Up",
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
			"Designer Status",
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
			"Designer Status Igual",
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

func TestBudgetStatusHistoryShouldCancelOtherProjectBudgetsWhenOneBecomesPedido(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	orcamentoStatusID := insertNamedBudgetStatus(t, env, "ORCAMENTO", "ORCAMENTO", false, 1)
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "PEDIDO", true, 2)
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO", "CANCELADO", true, 3)
	seed.statusID = orcamentoStatusID

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
			"Designer Grupo 1",
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
			"Designer Grupo 2",
			"Concorrente Grupo 2",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)

	changeStatusResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Projeto aprovado"}`, pedidoStatusID),
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

	firstBudget := decodeJSONResponse[dto.BudgetResponse](t, getFirstBudgetResponse.Body)
	secondBudget := decodeJSONResponse[dto.BudgetResponse](t, getSecondBudgetResponse.Body)
	if firstBudget.StatusID != pedidoStatusID {
		t.Fatalf("expected first budget status id %d, got %d", pedidoStatusID, firstBudget.StatusID)
	}
	if secondBudget.StatusID != canceladoStatusID {
		t.Fatalf("expected second budget status id %d, got %d", canceladoStatusID, secondBudget.StatusID)
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

	firstHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, firstHistoryResponse.Body)
	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	if len(firstHistoryPayload) != 1 {
		t.Fatalf("expected 1 first history item, got %d", len(firstHistoryPayload))
	}
	if len(secondHistoryPayload) != 1 {
		t.Fatalf("expected 1 second history item, got %d", len(secondHistoryPayload))
	}
	if firstHistoryPayload[0].FromStatusID == nil || *firstHistoryPayload[0].FromStatusID != orcamentoStatusID {
		t.Fatalf("expected first budget from_status_id %d, got %v", orcamentoStatusID, firstHistoryPayload[0].FromStatusID)
	}
	if firstHistoryPayload[0].ToStatusID != pedidoStatusID {
		t.Fatalf("expected first budget to_status_id %d, got %d", pedidoStatusID, firstHistoryPayload[0].ToStatusID)
	}
	if secondHistoryPayload[0].FromStatusID == nil || *secondHistoryPayload[0].FromStatusID != orcamentoStatusID {
		t.Fatalf("expected second budget from_status_id %d, got %v", orcamentoStatusID, secondHistoryPayload[0].FromStatusID)
	}
	if secondHistoryPayload[0].ToStatusID != canceladoStatusID {
		t.Fatalf("expected second budget to_status_id %d, got %d", canceladoStatusID, secondHistoryPayload[0].ToStatusID)
	}
	if secondHistoryPayload[0].Notes != automaticProjectCancellationNote {
		t.Fatalf("expected automatic cancellation note, got %s", secondHistoryPayload[0].Notes)
	}
}

func TestBudgetStatusHistoryShouldRejectSecondPedidoFromSameProject(t *testing.T) {
	env := newIntegrationTestEnv(t)
	token := env.createAdminToken(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	orcamentoStatusID := insertNamedBudgetStatus(t, env, "ORCAMENTO", "ORCAMENTO", false, 1)
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "PEDIDO", true, 2)
	canceladoStatusID := insertNamedBudgetStatus(t, env, "CANCELADO", "CANCELADO", true, 3)
	seed.statusID = orcamentoStatusID

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
			"Designer Grupo 3",
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
			seed,
			"Designer Grupo 4",
			"Concorrente Grupo 4",
		),
	)
	if createSecondBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createSecondBudgetResponse.Code)
	}

	firstBudgetPayload := decodeJSONResponse[createResourceResponse](t, createFirstBudgetResponse.Body)
	secondBudgetPayload := decodeJSONResponse[createResourceResponse](t, createSecondBudgetResponse.Body)

	firstPedidoResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", firstBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Primeiro pedido do projeto"}`, pedidoStatusID),
	)
	if firstPedidoResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstPedidoResponse.Code)
	}

	secondPedidoResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/budgets/%d/status", secondBudgetPayload.ID),
		token,
		fmt.Sprintf(`{"status_id":%d,"notes":"Tentativa de segundo pedido"}`, pedidoStatusID),
	)
	if secondPedidoResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondPedidoResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, secondPedidoResponse.Body)
	if errorPayload.Message != "Ja existe outro orcamento do projeto marcado como PEDIDO" {
		t.Fatalf("expected conflict message, got %s", errorPayload.Message)
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
	if secondBudget.StatusID != canceladoStatusID {
		t.Fatalf("expected second budget to remain with status id %d, got %d", canceladoStatusID, secondBudget.StatusID)
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

	secondHistoryPayload := decodeJSONResponse[[]dto.BudgetStatusHistoryResponse](t, secondHistoryResponse.Body)
	if len(secondHistoryPayload) != 1 {
		t.Fatalf("expected 1 history item for automatic cancellation, got %d", len(secondHistoryPayload))
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
