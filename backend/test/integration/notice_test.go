package integration

import (
	"fmt"
	"net/http"
	"testing"

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
