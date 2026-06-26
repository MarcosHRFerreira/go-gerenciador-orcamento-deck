package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
)

func TestNoticesAdminShouldCreateGeneralNoticeListAndMarkRead(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Aviso Geral",
		"user.notice.general@local.dev",
		"user_notice_general",
		"user",
		"salesperson",
	)

	createNoticeResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/notices",
		adminToken,
		`{"title":"Parada programada","body":"O sistema ficara indisponivel entre 18h e 19h.","scope_type":"all","priority":"warning","pinned":true,"recipient_user_ids":[]}`,
	)
	if createNoticeResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createNoticeResponse.Code)
	}

	createNoticePayload := decodeJSONResponse[dto.CreateNoticeResponse](t, createNoticeResponse.Body)

	listUnreadResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices?status=unread",
		userSession.token,
		"",
	)
	if listUnreadResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listUnreadResponse.Code)
	}

	unreadPayload := decodeJSONResponse[[]dto.NoticeResponse](t, listUnreadResponse.Body)
	if len(unreadPayload) != 1 {
		t.Fatalf("expected 1 unread notice, got %d", len(unreadPayload))
	}
	if unreadPayload[0].ID != createNoticePayload.ID {
		t.Fatalf("expected notice id %d, got %d", createNoticePayload.ID, unreadPayload[0].ID)
	}
	if unreadPayload[0].CreatedByUserName != "Admin Local" {
		t.Fatalf("expected author Admin Local, got %s", unreadPayload[0].CreatedByUserName)
	}
	if unreadPayload[0].ReadAt != nil {
		t.Fatal("expected unread notice to have nil read_at")
	}

	unreadCountResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices/unread-count",
		userSession.token,
		"",
	)
	if unreadCountResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, unreadCountResponse.Code)
	}

	unreadCountPayload := decodeJSONResponse[dto.NoticeUnreadCountResponse](t, unreadCountResponse.Body)
	if unreadCountPayload.Count != 1 {
		t.Fatalf("expected unread count 1, got %d", unreadCountPayload.Count)
	}

	markReadResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/notices/%d/read", createNoticePayload.ID),
		userSession.token,
		"",
	)
	if markReadResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, markReadResponse.Code)
	}

	listReadResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices?status=read",
		userSession.token,
		"",
	)
	if listReadResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listReadResponse.Code)
	}

	readPayload := decodeJSONResponse[[]dto.NoticeResponse](t, listReadResponse.Body)
	if len(readPayload) != 1 {
		t.Fatalf("expected 1 read notice, got %d", len(readPayload))
	}
	if readPayload[0].ReadAt == nil {
		t.Fatal("expected read notice to have read_at")
	}

	unreadCountAfterReadResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices/unread-count",
		userSession.token,
		"",
	)
	if unreadCountAfterReadResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, unreadCountAfterReadResponse.Code)
	}

	unreadCountAfterReadPayload := decodeJSONResponse[dto.NoticeUnreadCountResponse](t, unreadCountAfterReadResponse.Body)
	if unreadCountAfterReadPayload.Count != 0 {
		t.Fatalf("expected unread count 0, got %d", unreadCountAfterReadPayload.Count)
	}
}

func TestNoticesAdminShouldCreateDirectedNoticeOnlyForTargetUser(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	targetUserSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Destino",
		"user.notice.target@local.dev",
		"user_notice_target",
		"user",
		"salesperson",
	)
	otherUserSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Fora",
		"user.notice.other@local.dev",
		"user_notice_other",
		"user",
		"salesperson",
	)

	createNoticeResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/notices",
		adminToken,
		fmt.Sprintf(`{"title":"Revisar obra 90","body":"Favor revisar o grupo da obra 90 ainda hoje.","scope_type":"users","priority":"critical","pinned":false,"recipient_user_ids":[%d]}`, targetUserSession.userID),
	)
	if createNoticeResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createNoticeResponse.Code)
	}

	targetListResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices",
		targetUserSession.token,
		"",
	)
	if targetListResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, targetListResponse.Code)
	}

	targetPayload := decodeJSONResponse[[]dto.NoticeResponse](t, targetListResponse.Body)
	if len(targetPayload) != 1 {
		t.Fatalf("expected 1 targeted notice, got %d", len(targetPayload))
	}
	if targetPayload[0].Title != "Revisar obra 90" {
		t.Fatalf("expected targeted notice title, got %s", targetPayload[0].Title)
	}

	otherListResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/notices",
		otherUserSession.token,
		"",
	)
	if otherListResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, otherListResponse.Code)
	}

	otherPayload := decodeJSONResponse[[]dto.NoticeResponse](t, otherListResponse.Body)
	if len(otherPayload) != 0 {
		t.Fatalf("expected 0 notices for unrelated user, got %d", len(otherPayload))
	}
}

func TestNoticesUserShouldNotCreateNotice(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userToken := env.createUserTokenWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Sem Permissao",
		"user.notice.blocked@local.dev",
		"user_notice_blocked",
		"user",
		"salesperson",
	)

	createNoticeResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/notices",
		userToken,
		`{"title":"Tentativa invalida","body":"Usuario comum nao pode criar aviso.","scope_type":"all","priority":"info","pinned":false,"recipient_user_ids":[]}`,
	)
	if createNoticeResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, createNoticeResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, createNoticeResponse.Body)
	if errorPayload.Message != "Permissoes insuficientes" {
		t.Fatalf("expected forbidden message, got %s", errorPayload.Message)
	}
}

func TestNoticesShouldShowAutomaticBudgetClosingNoticeForSalespersonAndAdmins(t *testing.T) {
	env := newIntegrationTestEnv(t)
	bootstrapAdminToken := env.createAdminToken(t)
	suffix := uniqueSuffix()
	seed := env.seedBudgetData(t, suffix)

	actorAdminSession := env.createUserSessionWithCredentialsAndKind(
		t,
		bootstrapAdminToken,
		"Admin Fechamento",
		fmt.Sprintf("admin.budget.actor.%s@local.dev", suffix),
		fmt.Sprintf("admin_budget_actor_%s", suffix),
		"admin",
		"",
	)
	otherAdminSession := env.createUserSessionWithCredentialsAndKind(
		t,
		bootstrapAdminToken,
		"Admin Apoio",
		fmt.Sprintf("admin.budget.other.%s@local.dev", suffix),
		fmt.Sprintf("admin_budget_other_%s", suffix),
		"admin",
		"",
	)
	salespersonSession := env.createUserSessionWithCredentialsAndKind(
		t,
		bootstrapAdminToken,
		"Vendedor Fechamento",
		fmt.Sprintf("sales.%s@local.dev", suffix),
		fmt.Sprintf("sales_budget_%s", suffix),
		"user",
		"salesperson",
	)
	unrelatedUserSession := env.createUserSessionWithCredentialsAndKind(
		t,
		bootstrapAdminToken,
		"Usuario Fora",
		fmt.Sprintf("user.budget.other.%s@local.dev", suffix),
		fmt.Sprintf("user_budget_other_%s", suffix),
		"user",
		"salesperson",
	)

	createBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/budgets",
		actorAdminSession.token,
		buildBudgetRequestBody(
			"BGT-NTC-001",
			2026,
			time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
			4200,
			seed,
			"Projetista Aviso",
			"Concorrente Aviso",
		),
	)
	if createBudgetResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createBudgetResponse.Code)
	}

	createBudgetPayload := decodeJSONResponse[createResourceResponse](t, createBudgetResponse.Body)
	pedidoStatusID := insertNamedBudgetStatus(t, env, "PEDIDO", "Pedido "+uniqueSuffix(), true, 2)

	updateBudgetResponse := env.doJSONRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/budgets/%d", createBudgetPayload.ID),
		actorAdminSession.token,
		fmt.Sprintf(`{
			"budget_number":"BGT-NTC-001",
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
			"competitor_name":"Concorrente Aviso",
			"competitor_price":900.00,
			"projetista_name":"Projetista Aviso",
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
	if updateBudgetResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, updateBudgetResponse.Code)
	}

	salespersonNoticesResponse := env.doJSONRequest(t, http.MethodGet, "/notices", salespersonSession.token, "")
	if salespersonNoticesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, salespersonNoticesResponse.Code)
	}

	salespersonPayload := decodeJSONResponse[[]dto.NoticeResponse](t, salespersonNoticesResponse.Body)
	if len(salespersonPayload) != 1 {
		t.Fatalf("expected 1 notice for salesperson, got %d", len(salespersonPayload))
	}
	if salespersonPayload[0].Title != "Fechamento registrado no orcamento BGT-NTC-001/2026" {
		t.Fatalf("unexpected automatic notice title: %s", salespersonPayload[0].Title)
	}
	if salespersonPayload[0].CreatedByUserName != "Admin Fechamento" {
		t.Fatalf("expected automatic notice author Admin Fechamento, got %s", salespersonPayload[0].CreatedByUserName)
	}

	actorAdminNoticesResponse := env.doJSONRequest(t, http.MethodGet, "/notices", actorAdminSession.token, "")
	if actorAdminNoticesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, actorAdminNoticesResponse.Code)
	}

	actorAdminPayload := decodeJSONResponse[[]dto.NoticeResponse](t, actorAdminNoticesResponse.Body)
	if len(actorAdminPayload) != 1 {
		t.Fatalf("expected 1 notice for actor admin, got %d", len(actorAdminPayload))
	}

	otherAdminNoticesResponse := env.doJSONRequest(t, http.MethodGet, "/notices", otherAdminSession.token, "")
	if otherAdminNoticesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, otherAdminNoticesResponse.Code)
	}

	otherAdminPayload := decodeJSONResponse[[]dto.NoticeResponse](t, otherAdminNoticesResponse.Body)
	if len(otherAdminPayload) != 1 {
		t.Fatalf("expected 1 notice for second admin, got %d", len(otherAdminPayload))
	}

	unrelatedUserNoticesResponse := env.doJSONRequest(t, http.MethodGet, "/notices", unrelatedUserSession.token, "")
	if unrelatedUserNoticesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, unrelatedUserNoticesResponse.Code)
	}

	unrelatedUserPayload := decodeJSONResponse[[]dto.NoticeResponse](t, unrelatedUserNoticesResponse.Body)
	if len(unrelatedUserPayload) != 0 {
		t.Fatalf("expected 0 notices for unrelated user, got %d", len(unrelatedUserPayload))
	}
}
