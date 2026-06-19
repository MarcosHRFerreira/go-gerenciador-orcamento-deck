package conversation

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	conversationrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/conversation"
	projectrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/project"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

type Service interface {
	Create(ctx context.Context, actorUserID int64, req *dto.CreateConversationRequest) (int64, error)
	ListByUser(ctx context.Context, userID int64) ([]dto.ConversationListItemResponse, error)
	ListMessagesByConversation(ctx context.Context, userID int64, conversationID int64) ([]dto.ConversationMessageResponse, error)
	ListAvailableUsers(ctx context.Context, actorUserID int64) ([]dto.ConversationParticipantResponse, error)
	CountUnreadByUser(ctx context.Context, userID int64) (int64, error)
	SendMessage(ctx context.Context, userID int64, conversationID int64, req *dto.SendConversationMessageRequest) (int64, error)
	MarkAsRead(ctx context.Context, userID int64, conversationID int64) error
}

type service struct {
	conversationRepo conversationrepository.Repository
	projectRepo      projectrepository.Repository
	userRepo         userrepository.Repository
}

func NewService(
	conversationRepo conversationrepository.Repository,
	projectRepo projectrepository.Repository,
	userRepo userrepository.Repository,
) Service {
	return &service{
		conversationRepo: conversationRepo,
		projectRepo:      projectRepo,
		userRepo:         userRepo,
	}
}

func (s *service) Create(ctx context.Context, actorUserID int64, req *dto.CreateConversationRequest) (int64, error) {
	if actorUserID <= 0 {
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}

	actorUser, err := s.loadActiveUser(ctx, actorUserID)
	if err != nil {
		return 0, err
	}

	participantUserID := req.ParticipantUserID
	if participantUserID <= 0 {
		return 0, apperror.BadRequest("participant_user_id e obrigatorio")
	}
	if participantUserID == actorUserID {
		return 0, apperror.BadRequest("Nao e permitido iniciar conversa com o proprio usuario")
	}

	participantUser, err := s.loadActiveUser(ctx, participantUserID)
	if err != nil {
		return 0, err
	}

	if _, err := s.resolveConversationProject(ctx, req.ProjectID); err != nil {
		return 0, err
	}

	if err := validateConversationPermission(actorUser, participantUser); err != nil {
		logConversationWarn(
			ctx,
			"Criacao de conversa bloqueada por regra de permissao",
			slog.String("conversation_action", "create_conversation"),
			slog.Int64("actor_user_id", actorUserID),
			slog.Int64("participant_user_id", participantUserID),
		)
		return 0, err
	}

	conversationID, _, err := s.conversationRepo.CreateOrAppendDirect(
		ctx,
		actorUserID,
		participantUserID,
		req.ProjectID,
		strings.TrimSpace(req.InitialMessage),
		time.Now(),
	)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao criar ou reutilizar conversa",
			err,
			slog.String("conversation_action", "create_conversation"),
			slog.Int64("actor_user_id", actorUserID),
			slog.Int64("participant_user_id", participantUserID),
		)
		return 0, apperror.Internal("failed to create conversation", err)
	}

	return conversationID, nil
}

func (s *service) ListByUser(ctx context.Context, userID int64) ([]dto.ConversationListItemResponse, error) {
	if userID <= 0 {
		return nil, apperror.Forbidden("Usuario autenticado invalido")
	}

	items, err := s.conversationRepo.ListByUser(ctx, userID)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao listar conversas",
			err,
			slog.String("conversation_action", "list_conversations"),
			slog.Int64("user_id", userID),
		)
		return nil, apperror.Internal("failed to list conversations", err)
	}

	response := make([]dto.ConversationListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toConversationListItemResponse(item))
	}

	return response, nil
}

func (s *service) ListMessagesByConversation(
	ctx context.Context,
	userID int64,
	conversationID int64,
) ([]dto.ConversationMessageResponse, error) {
	if userID <= 0 {
		return nil, apperror.Forbidden("Usuario autenticado invalido")
	}
	if conversationID <= 0 {
		return nil, apperror.BadRequest("conversation_id e obrigatorio")
	}

	items, err := s.conversationRepo.ListMessagesByConversation(ctx, conversationID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Conversa nao encontrada")
		}

		logConversationError(
			ctx,
			"Falha ao listar mensagens da conversa",
			err,
			slog.String("conversation_action", "list_conversation_messages"),
			slog.Int64("user_id", userID),
			slog.Int64("conversation_id", conversationID),
		)
		return nil, apperror.Internal("failed to list conversation messages", err)
	}

	response := make([]dto.ConversationMessageResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toConversationMessageResponse(item))
	}

	return response, nil
}

func (s *service) ListAvailableUsers(ctx context.Context, actorUserID int64) ([]dto.ConversationParticipantResponse, error) {
	if actorUserID <= 0 {
		return nil, apperror.Forbidden("Usuario autenticado invalido")
	}

	actorUser, err := s.loadActiveUser(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.userRepo.ListActiveUsers(ctx)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao listar usuarios ativos para conversa",
			err,
			slog.String("conversation_action", "list_available_users"),
			slog.Int64("actor_user_id", actorUserID),
		)
		return nil, apperror.Internal("failed to list available conversation users", err)
	}

	response := make([]dto.ConversationParticipantResponse, 0, len(activeUsers))
	for _, item := range activeUsers {
		if item.ID == actorUser.ID {
			continue
		}
		if !canStartConversationWith(*actorUser, item) {
			continue
		}

		response = append(response, dto.ConversationParticipantResponse{
			ID:       item.ID,
			Name:     item.Name,
			Username: item.Username,
			Role:     string(item.Role),
		})
	}

	return response, nil
}

func (s *service) CountUnreadByUser(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}

	count, err := s.conversationRepo.CountUnreadByUser(ctx, userID)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao contar conversas nao lidas",
			err,
			slog.String("conversation_action", "count_unread_conversations"),
			slog.Int64("user_id", userID),
		)
		return 0, apperror.Internal("failed to count unread conversations", err)
	}

	return count, nil
}

func (s *service) SendMessage(
	ctx context.Context,
	userID int64,
	conversationID int64,
	req *dto.SendConversationMessageRequest,
) (int64, error) {
	if userID <= 0 {
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}
	if conversationID <= 0 {
		return 0, apperror.BadRequest("conversation_id e obrigatorio")
	}

	messageID, err := s.conversationRepo.SendMessage(
		ctx,
		conversationID,
		userID,
		strings.TrimSpace(req.Body),
		time.Now(),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, apperror.NotFound("Conversa nao encontrada")
		}

		logConversationError(
			ctx,
			"Falha ao enviar mensagem da conversa",
			err,
			slog.String("conversation_action", "send_conversation_message"),
			slog.Int64("user_id", userID),
			slog.Int64("conversation_id", conversationID),
		)
		return 0, apperror.Internal("failed to send conversation message", err)
	}

	return messageID, nil
}

func (s *service) MarkAsRead(ctx context.Context, userID int64, conversationID int64) error {
	if userID <= 0 {
		return apperror.Forbidden("Usuario autenticado invalido")
	}
	if conversationID <= 0 {
		return apperror.BadRequest("conversation_id e obrigatorio")
	}

	if err := s.conversationRepo.MarkAsRead(ctx, conversationID, userID, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFound("Conversa nao encontrada")
		}

		logConversationError(
			ctx,
			"Falha ao marcar conversa como lida",
			err,
			slog.String("conversation_action", "mark_conversation_read"),
			slog.Int64("user_id", userID),
			slog.Int64("conversation_id", conversationID),
		)
		return apperror.Internal("failed to mark conversation as read", err)
	}

	return nil
}

func (s *service) loadActiveUser(ctx context.Context, userID int64) (*model.UserModel, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao carregar usuario da conversa",
			err,
			slog.String("conversation_action", "load_conversation_user"),
			slog.Int64("user_id", userID),
		)
		return nil, apperror.Internal("failed to load conversation user", err)
	}
	if user == nil || !user.Active {
		return nil, apperror.BadRequest("Usuario da conversa nao encontrado ou inativo")
	}

	return user, nil
}

func (s *service) resolveConversationProject(ctx context.Context, projectID *int64) (*model.ProjectModel, error) {
	if projectID == nil {
		return nil, nil
	}
	if *projectID <= 0 {
		return nil, apperror.BadRequest("project_id deve ser maior que zero")
	}

	project, err := s.projectRepo.GetByID(ctx, *projectID)
	if err != nil {
		logConversationError(
			ctx,
			"Falha ao carregar obra da conversa",
			err,
			slog.String("conversation_action", "load_conversation_project"),
			slog.Int64("project_id", *projectID),
		)
		return nil, apperror.Internal("failed to load conversation project", err)
	}
	if project == nil {
		return nil, apperror.BadRequest("Obra da conversa nao encontrada")
	}

	return project, nil
}

func validateConversationPermission(actorUser *model.UserModel, participantUser *model.UserModel) error {
	if actorUser == nil || participantUser == nil {
		return apperror.BadRequest("Usuario da conversa nao encontrado ou inativo")
	}
	if canStartConversationWith(*actorUser, *participantUser) {
		return nil
	}

	return apperror.Forbidden("Permissoes insuficientes para iniciar esta conversa")
}

func canStartConversationWith(actorUser model.UserModel, participantUser model.UserModel) bool {
	if actorUser.Role == model.RoleAdmin {
		return true
	}

	return actorUser.Role == model.RoleUser && participantUser.Role == model.RoleAdmin
}

func toConversationListItemResponse(item model.ConversationListItemModel) dto.ConversationListItemResponse {
	var project *dto.ConversationProjectResponse
	if item.ProjectID != nil && item.ProjectCode != nil && item.ProjectName != nil {
		project = &dto.ConversationProjectResponse{
			ID:   *item.ProjectID,
			Code: *item.ProjectCode,
			Name: *item.ProjectName,
		}
	}

	return dto.ConversationListItemResponse{
		ID:        item.ConversationID,
		Type:      string(item.Type),
		UpdatedAt: item.UpdatedAt,
		Project:   project,
		Participant: dto.ConversationParticipantResponse{
			ID:       item.ParticipantUserID,
			Name:     item.ParticipantName,
			Username: item.ParticipantUsername,
			Role:     string(item.ParticipantRole),
		},
		LastMessageID:     item.LastMessageID,
		LastMessageBody:   item.LastMessageBody,
		LastMessageAt:     item.LastMessageAt,
		LastMessageSender: item.LastMessageSenderID,
		UnreadCount:       item.UnreadCount,
	}
}

func toConversationMessageResponse(item model.ConversationMessageDetailsModel) dto.ConversationMessageResponse {
	return dto.ConversationMessageResponse{
		ID:             item.MessageID,
		ConversationID: item.ConversationID,
		Sender: dto.ConversationParticipantResponse{
			ID:       item.SenderUserID,
			Name:     item.SenderName,
			Username: item.SenderUsername,
			Role:     string(item.SenderRole),
		},
		Body:      item.Body,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func logConversationWarn(ctx context.Context, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelWarn, message, attrs...)
}

func logConversationError(ctx context.Context, message string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelError, message, attrs...)
}
