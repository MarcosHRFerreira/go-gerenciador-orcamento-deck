package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
)

func TestConversationsAdminShouldCreateListReplyAndMarkRead(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Conversa",
		"user.conversation@local.dev",
		"user_conversation",
		"user",
		"salesperson",
	)

	createConversationResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/conversations",
		adminToken,
		fmt.Sprintf(`{"participant_user_id":%d,"initial_message":"Bom dia, vamos alinhar a obra 90?"}`, userSession.userID),
	)
	if createConversationResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createConversationResponse.Code)
	}

	createConversationPayload := decodeJSONResponse[dto.CreateConversationResponse](t, createConversationResponse.Body)
	if createConversationPayload.ID <= 0 {
		t.Fatalf("expected conversation id greater than zero, got %d", createConversationPayload.ID)
	}

	userListResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations",
		userSession.token,
		"",
	)
	if userListResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userListResponse.Code)
	}

	userListPayload := decodeJSONResponse[[]dto.ConversationListItemResponse](t, userListResponse.Body)
	if len(userListPayload) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(userListPayload))
	}
	if userListPayload[0].ID != createConversationPayload.ID {
		t.Fatalf("expected conversation id %d, got %d", createConversationPayload.ID, userListPayload[0].ID)
	}
	if userListPayload[0].Participant.Name != "Admin Local" {
		t.Fatalf("expected participant Admin Local, got %s", userListPayload[0].Participant.Name)
	}
	if userListPayload[0].UnreadCount != 1 {
		t.Fatalf("expected unread count 1, got %d", userListPayload[0].UnreadCount)
	}
	if userListPayload[0].LastMessageBody == nil || *userListPayload[0].LastMessageBody != "Bom dia, vamos alinhar a obra 90?" {
		t.Fatal("expected last message body to match initial message")
	}

	userUnreadCountResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations/unread-count",
		userSession.token,
		"",
	)
	if userUnreadCountResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userUnreadCountResponse.Code)
	}

	userUnreadCountPayload := decodeJSONResponse[dto.ConversationUnreadCountResponse](t, userUnreadCountResponse.Body)
	if userUnreadCountPayload.Count != 1 {
		t.Fatalf("expected unread count 1, got %d", userUnreadCountPayload.Count)
	}

	userMessagesResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/conversations/%d/messages", createConversationPayload.ID),
		userSession.token,
		"",
	)
	if userMessagesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userMessagesResponse.Code)
	}

	userMessagesPayload := decodeJSONResponse[[]dto.ConversationMessageResponse](t, userMessagesResponse.Body)
	if len(userMessagesPayload) != 1 {
		t.Fatalf("expected 1 message, got %d", len(userMessagesPayload))
	}
	if userMessagesPayload[0].Sender.Name != "Admin Local" {
		t.Fatalf("expected sender Admin Local, got %s", userMessagesPayload[0].Sender.Name)
	}

	markReadResponse := env.doJSONRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/conversations/%d/read", createConversationPayload.ID),
		userSession.token,
		"",
	)
	if markReadResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, markReadResponse.Code)
	}

	userUnreadCountAfterReadResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations/unread-count",
		userSession.token,
		"",
	)
	if userUnreadCountAfterReadResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userUnreadCountAfterReadResponse.Code)
	}

	userUnreadCountAfterReadPayload := decodeJSONResponse[dto.ConversationUnreadCountResponse](t, userUnreadCountAfterReadResponse.Body)
	if userUnreadCountAfterReadPayload.Count != 0 {
		t.Fatalf("expected unread count 0, got %d", userUnreadCountAfterReadPayload.Count)
	}

	replyResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/conversations/%d/messages", createConversationPayload.ID),
		userSession.token,
		`{"body":"Recebido. Vou revisar ainda hoje."}`,
	)
	if replyResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, replyResponse.Code)
	}

	adminUnreadCountResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations/unread-count",
		adminToken,
		"",
	)
	if adminUnreadCountResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminUnreadCountResponse.Code)
	}

	adminUnreadCountPayload := decodeJSONResponse[dto.ConversationUnreadCountResponse](t, adminUnreadCountResponse.Body)
	if adminUnreadCountPayload.Count != 1 {
		t.Fatalf("expected unread count 1, got %d", adminUnreadCountPayload.Count)
	}

	adminListResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations",
		adminToken,
		"",
	)
	if adminListResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminListResponse.Code)
	}

	adminListPayload := decodeJSONResponse[[]dto.ConversationListItemResponse](t, adminListResponse.Body)
	if len(adminListPayload) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(adminListPayload))
	}
	if adminListPayload[0].UnreadCount != 1 {
		t.Fatalf("expected unread count 1 for admin, got %d", adminListPayload[0].UnreadCount)
	}
	if adminListPayload[0].LastMessageBody == nil || *adminListPayload[0].LastMessageBody != "Recebido. Vou revisar ainda hoje." {
		t.Fatal("expected admin to see latest reply as last message")
	}
}

func TestConversationsShouldSupportProjectContextAndReuseThreadByProject(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	userSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Projeto",
		"user.conversation.project@local.dev",
		"user_conversation_project",
		"user",
		"salesperson",
	)

	createProjectResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/projects",
		adminToken,
		`{"code":"OBR-009999","name":"Projeto Contextual","city":"Sao Paulo","state":"SP","notes":"Obra para conversa contextual"}`,
	)
	if createProjectResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createProjectResponse.Code)
	}

	createProjectPayload := decodeJSONResponse[struct {
		ID int64 `json:"id"`
	}](t, createProjectResponse.Body)
	if createProjectPayload.ID <= 0 {
		t.Fatalf("expected project id greater than zero, got %d", createProjectPayload.ID)
	}

	firstCreateConversationResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/conversations",
		adminToken,
		fmt.Sprintf(
			`{"participant_user_id":%d,"project_id":%d,"initial_message":"Vamos tratar desta obra."}`,
			userSession.userID,
			createProjectPayload.ID,
		),
	)
	if firstCreateConversationResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstCreateConversationResponse.Code)
	}

	firstCreateConversationPayload := decodeJSONResponse[dto.CreateConversationResponse](t, firstCreateConversationResponse.Body)

	secondCreateConversationResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/conversations",
		adminToken,
		fmt.Sprintf(
			`{"participant_user_id":%d,"project_id":%d,"initial_message":"Atualizando o contexto da mesma obra."}`,
			userSession.userID,
			createProjectPayload.ID,
		),
	)
	if secondCreateConversationResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, secondCreateConversationResponse.Code)
	}

	secondCreateConversationPayload := decodeJSONResponse[dto.CreateConversationResponse](t, secondCreateConversationResponse.Body)
	if firstCreateConversationPayload.ID != secondCreateConversationPayload.ID {
		t.Fatalf("expected same contextual conversation id, got %d and %d", firstCreateConversationPayload.ID, secondCreateConversationPayload.ID)
	}

	userListResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations",
		userSession.token,
		"",
	)
	if userListResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userListResponse.Code)
	}

	userListPayload := decodeJSONResponse[[]dto.ConversationListItemResponse](t, userListResponse.Body)
	if len(userListPayload) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(userListPayload))
	}
	if userListPayload[0].Project == nil {
		t.Fatal("expected project context to be present in conversation list item")
	}
	if userListPayload[0].Project.ID != createProjectPayload.ID {
		t.Fatalf("expected project id %d, got %d", createProjectPayload.ID, userListPayload[0].Project.ID)
	}
	if userListPayload[0].Project.Code != "OBR-009999" {
		t.Fatalf("expected project code OBR-009999, got %s", userListPayload[0].Project.Code)
	}
	if userListPayload[0].Project.Name != "Projeto Contextual" {
		t.Fatalf("expected project name Projeto Contextual, got %s", userListPayload[0].Project.Name)
	}

	userMessagesResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/conversations/%d/messages", firstCreateConversationPayload.ID),
		userSession.token,
		"",
	)
	if userMessagesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userMessagesResponse.Code)
	}

	userMessagesPayload := decodeJSONResponse[[]dto.ConversationMessageResponse](t, userMessagesResponse.Body)
	if len(userMessagesPayload) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(userMessagesPayload))
	}
}

func TestConversationsUserShouldStartConversationWithAdmin(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	adminMeResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/users/me",
		adminToken,
		"",
	)
	if adminMeResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminMeResponse.Code)
	}

	adminMePayload := decodeJSONResponse[dto.UserResponse](t, adminMeResponse.Body)
	userSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario Origem",
		"user.conversation.origin@local.dev",
		"user_conversation_origin",
		"user",
		"salesperson",
	)

	createConversationResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/conversations",
		userSession.token,
		fmt.Sprintf(`{"participant_user_id":%d,"initial_message":"Preciso de ajuda com a obra 120."}`, adminMePayload.ID),
	)
	if createConversationResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createConversationResponse.Code)
	}

	adminUnreadCountResponse := env.doJSONRequest(
		t,
		http.MethodGet,
		"/conversations/unread-count",
		adminToken,
		"",
	)
	if adminUnreadCountResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, adminUnreadCountResponse.Code)
	}

	adminUnreadCountPayload := decodeJSONResponse[dto.ConversationUnreadCountResponse](t, adminUnreadCountResponse.Body)
	if adminUnreadCountPayload.Count != 1 {
		t.Fatalf("expected unread count 1, got %d", adminUnreadCountPayload.Count)
	}
}

func TestConversationsUserShouldNotStartConversationWithAnotherUser(t *testing.T) {
	env := newIntegrationTestEnv(t)
	adminToken := env.createAdminToken(t)
	firstUserSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario A",
		"user.conversation.a@local.dev",
		"user_conversation_a",
		"user",
		"salesperson",
	)
	secondUserSession := env.createUserSessionWithCredentialsAndKind(
		t,
		adminToken,
		"Usuario B",
		"user.conversation.b@local.dev",
		"user_conversation_b",
		"user",
		"salesperson",
	)

	createConversationResponse := env.doJSONRequest(
		t,
		http.MethodPost,
		"/conversations",
		firstUserSession.token,
		fmt.Sprintf(`{"participant_user_id":%d,"initial_message":"Tentativa invalida entre usuarios."}`, secondUserSession.userID),
	)
	if createConversationResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, createConversationResponse.Code)
	}

	errorPayload := decodeJSONResponse[httpresponse.ErrorResponse](t, createConversationResponse.Body)
	if errorPayload.Message != "Permissoes insuficientes para iniciar esta conversa" {
		t.Fatalf("expected forbidden message, got %s", errorPayload.Message)
	}
}
