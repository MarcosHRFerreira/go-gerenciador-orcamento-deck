package notice

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	noticerepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/notice"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

type Service interface {
	Create(ctx context.Context, actorUserID int64, req *dto.CreateNoticeRequest) (int64, error)
	ListByUser(ctx context.Context, userID int64, statusFilter string) ([]dto.NoticeResponse, error)
	GetByIDForUser(ctx context.Context, userID int64, noticeID int64) (*dto.NoticeResponse, error)
	CountUnreadByUser(ctx context.Context, userID int64) (int64, error)
	MarkAsRead(ctx context.Context, userID int64, noticeID int64) error
}

type service struct {
	noticeRepo noticerepository.Repository
	userRepo   userrepository.Repository
}

func NewService(noticeRepo noticerepository.Repository, userRepo userrepository.Repository) Service {
	return &service{
		noticeRepo: noticeRepo,
		userRepo:   userRepo,
	}
}

func (s *service) Create(ctx context.Context, actorUserID int64, req *dto.CreateNoticeRequest) (int64, error) {
	if actorUserID <= 0 {
		logNoticeWarn(ctx, "Criacao de aviso bloqueada por usuario autenticado invalido", slog.String("notice_action", "create_notice"), slog.String("reason", "invalid_actor"))
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}

	actorUser, err := s.userRepo.GetUserByID(ctx, actorUserID)
	if err != nil {
		logNoticeError(ctx, "Falha ao carregar autor do aviso", err, slog.String("notice_action", "create_notice"), slog.Int64("actor_user_id", actorUserID))
		return 0, apperror.Internal("failed to load actor user", err)
	}
	if actorUser == nil || !actorUser.Active {
		logNoticeWarn(ctx, "Criacao de aviso bloqueada por autor inexistente ou inativo", slog.String("notice_action", "create_notice"), slog.Int64("actor_user_id", actorUserID), slog.String("reason", "actor_not_available"))
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}

	if actorUser.Role != model.RoleAdmin {
		logNoticeWarn(ctx, "Criacao de aviso bloqueada por perfil sem permissao", slog.String("notice_action", "create_notice"), slog.Int64("actor_user_id", actorUserID), slog.String("reason", "insufficient_permissions"))
		return 0, apperror.Forbidden("Permissoes insuficientes para publicar avisos")
	}

	scopeType, err := normalizeNoticeScopeType(req.ScopeType)
	if err != nil {
		return 0, err
	}

	priority, err := normalizeNoticePriority(req.Priority)
	if err != nil {
		return 0, err
	}

	recipientUserIDs, err := s.resolveRecipientUserIDs(ctx, scopeType, req.RecipientUserIDs)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	noticeID, err := s.noticeRepo.Create(ctx, &model.NoticeModel{
		Title:           strings.TrimSpace(req.Title),
		Body:            strings.TrimSpace(req.Body),
		ScopeType:       scopeType,
		Priority:        priority,
		Pinned:          req.Pinned,
		ExpiresAt:       req.ExpiresAt,
		CreatedByUserID: actorUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, recipientUserIDs)
	if err != nil {
		logNoticeError(ctx, "Falha ao criar aviso", err, slog.String("notice_action", "create_notice"), slog.Int64("actor_user_id", actorUserID))
		return 0, apperror.Internal("failed to create notice", err)
	}

	logNoticeInfo(ctx, "Aviso criado com sucesso", slog.String("notice_action", "create_notice"), slog.Int64("actor_user_id", actorUserID), slog.Int64("notice_id", noticeID), slog.String("scope_type", string(scopeType)), slog.Int("recipients_count", len(recipientUserIDs)))
	return noticeID, nil
}

func (s *service) ListByUser(ctx context.Context, userID int64, statusFilter string) ([]dto.NoticeResponse, error) {
	if userID <= 0 {
		return nil, apperror.Forbidden("Usuario autenticado invalido")
	}

	items, err := s.noticeRepo.ListByUser(ctx, userID, statusFilter, time.Now())
	if err != nil {
		logNoticeError(ctx, "Falha ao listar avisos", err, slog.String("notice_action", "list_notices"), slog.Int64("user_id", userID))
		return nil, apperror.Internal("failed to list notices", err)
	}

	response := make([]dto.NoticeResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toNoticeResponse(item))
	}

	return response, nil
}

func (s *service) GetByIDForUser(ctx context.Context, userID int64, noticeID int64) (*dto.NoticeResponse, error) {
	if userID <= 0 {
		return nil, apperror.Forbidden("Usuario autenticado invalido")
	}
	if noticeID <= 0 {
		return nil, apperror.BadRequest("notice_id e obrigatorio")
	}

	item, err := s.noticeRepo.GetByIDForUser(ctx, noticeID, userID, time.Now())
	if err != nil {
		logNoticeError(ctx, "Falha ao carregar aviso", err, slog.String("notice_action", "get_notice"), slog.Int64("user_id", userID), slog.Int64("notice_id", noticeID))
		return nil, apperror.Internal("failed to load notice", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Aviso nao encontrado")
	}

	response := toNoticeResponse(*item)
	return &response, nil
}

func (s *service) CountUnreadByUser(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, apperror.Forbidden("Usuario autenticado invalido")
	}

	count, err := s.noticeRepo.CountUnreadByUser(ctx, userID, time.Now())
	if err != nil {
		logNoticeError(ctx, "Falha ao contar avisos nao lidos", err, slog.String("notice_action", "count_unread_notices"), slog.Int64("user_id", userID))
		return 0, apperror.Internal("failed to count unread notices", err)
	}

	return count, nil
}

func (s *service) MarkAsRead(ctx context.Context, userID int64, noticeID int64) error {
	if userID <= 0 {
		return apperror.Forbidden("Usuario autenticado invalido")
	}
	if noticeID <= 0 {
		return apperror.BadRequest("notice_id e obrigatorio")
	}

	item, err := s.noticeRepo.GetByIDForUser(ctx, noticeID, userID, time.Now())
	if err != nil {
		logNoticeError(ctx, "Falha ao verificar aviso antes da leitura", err, slog.String("notice_action", "mark_notice_read"), slog.Int64("user_id", userID), slog.Int64("notice_id", noticeID))
		return apperror.Internal("failed to check notice", err)
	}
	if item == nil {
		return apperror.NotFound("Aviso nao encontrado")
	}
	if item.ReadAt != nil {
		return nil
	}

	if err := s.noticeRepo.MarkAsRead(ctx, noticeID, userID, time.Now()); err != nil {
		logNoticeError(ctx, "Falha ao marcar aviso como lido", err, slog.String("notice_action", "mark_notice_read"), slog.Int64("user_id", userID), slog.Int64("notice_id", noticeID))
		return apperror.Internal("failed to mark notice as read", err)
	}

	return nil
}

func (s *service) resolveRecipientUserIDs(ctx context.Context, scopeType model.NoticeScopeType, rawRecipientUserIDs []int64) ([]int64, error) {
	if scopeType == model.NoticeScopeAllUsers {
		activeUsers, err := s.userRepo.ListActiveUsers(ctx)
		if err != nil {
			logNoticeError(ctx, "Falha ao listar usuarios ativos para aviso geral", err, slog.String("notice_action", "create_notice"))
			return nil, apperror.Internal("failed to list active users", err)
		}

		recipientUserIDs := make([]int64, 0, len(activeUsers))
		for _, user := range activeUsers {
			recipientUserIDs = append(recipientUserIDs, user.ID)
		}

		if len(recipientUserIDs) == 0 {
			return nil, apperror.Conflict("Nao existem usuarios ativos para receber o aviso")
		}

		return recipientUserIDs, nil
	}

	recipientUserIDs := normalizeRecipientUserIDs(rawRecipientUserIDs)
	if len(recipientUserIDs) == 0 {
		return nil, apperror.BadRequest("recipient_user_ids e obrigatorio para aviso direcionado")
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, recipientUserIDs)
	if err != nil {
		logNoticeError(ctx, "Falha ao carregar destinatarios do aviso", err, slog.String("notice_action", "create_notice"))
		return nil, apperror.Internal("failed to load notice recipients", err)
	}
	if len(users) != len(recipientUserIDs) {
		return nil, apperror.BadRequest("Um ou mais destinatarios do aviso nao foram encontrados")
	}

	for _, user := range users {
		if !user.Active {
			return nil, apperror.BadRequest("Um ou mais destinatarios do aviso estao inativos")
		}
	}

	return recipientUserIDs, nil
}

func normalizeRecipientUserIDs(values []int64) []int64 {
	uniqueValues := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if slices.Contains(uniqueValues, value) {
			continue
		}

		uniqueValues = append(uniqueValues, value)
	}

	return uniqueValues
}

func normalizeNoticeScopeType(value string) (model.NoticeScopeType, error) {
	scopeType := model.NoticeScopeType(strings.TrimSpace(strings.ToLower(value)))
	if scopeType != model.NoticeScopeAllUsers && scopeType != model.NoticeScopeUsers {
		return "", apperror.BadRequest("Tipo de destinatario do aviso invalido")
	}

	return scopeType, nil
}

func normalizeNoticePriority(value string) (model.NoticePriority, error) {
	priority := model.NoticePriority(strings.TrimSpace(strings.ToLower(value)))
	if priority != model.NoticePriorityInfo && priority != model.NoticePriorityWarning && priority != model.NoticePriorityCritical {
		return "", apperror.BadRequest("Prioridade do aviso invalida")
	}

	return priority, nil
}

func toNoticeResponse(item model.NoticeListItemModel) dto.NoticeResponse {
	return dto.NoticeResponse{
		ID:                item.ID,
		Title:             item.Title,
		Body:              item.Body,
		ScopeType:         string(item.ScopeType),
		Priority:          string(item.Priority),
		Pinned:            item.Pinned,
		ExpiresAt:         item.ExpiresAt,
		CreatedByUserID:   item.CreatedByUserID,
		CreatedByUserName: item.CreatedByUserName,
		ReadAt:            item.ReadAt,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func logNoticeInfo(ctx context.Context, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, message, attrs...)
}

func logNoticeWarn(ctx context.Context, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelWarn, message, attrs...)
}

func logNoticeError(ctx context.Context, message string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelError, message, attrs...)
}
