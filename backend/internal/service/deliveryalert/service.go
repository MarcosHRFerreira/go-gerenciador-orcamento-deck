package deliveryalert

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	conversationrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/conversation"
	deliveryalertrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/deliveryalert"
	noticerepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/notice"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

const runTimeout = 30 * time.Second

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	config            *config.Config
	conversationRepo  conversationrepository.Repository
	deliveryAlertRepo deliveryalertrepository.Repository
	noticeRepo        noticerepository.Repository
	userRepo          userrepository.Repository
	now               func() time.Time
}

func NewService(
	cfg *config.Config,
	conversationRepo conversationrepository.Repository,
	deliveryAlertRepo deliveryalertrepository.Repository,
	noticeRepo noticerepository.Repository,
	userRepo userrepository.Repository,
) Service {
	return &service{
		config:            cfg,
		conversationRepo:  conversationRepo,
		deliveryAlertRepo: deliveryAlertRepo,
		noticeRepo:        noticeRepo,
		userRepo:          userRepo,
		now:               time.Now,
	}
}

func (s *service) Run(ctx context.Context) error {
	runCtx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	referenceDate := truncateToDate(s.now())
	senderUser, err := s.resolveSenderUser(runCtx)
	if err != nil {
		return err
	}

	candidates, err := s.deliveryAlertRepo.ListDueInTwoDaysCandidates(runCtx, referenceDate)
	if err != nil {
		return fmt.Errorf("list delivery alert candidates: %w", err)
	}

	if len(candidates) == 0 {
		logInfo(runCtx, "Nenhum alerta de entrega elegivel nesta execucao", slog.Int("delivery_alert_candidates", 0))
		return nil
	}

	sentCount := 0
	for _, candidate := range candidates {
		if err := s.dispatchCandidate(runCtx, senderUser, candidate); err != nil {
			logError(
				runCtx,
				"Falha ao disparar alerta de entrega",
				err,
				slog.Int64("budget_id", candidate.BudgetID),
				slog.Int64("recipient_user_id", candidate.RecipientUserID),
			)
			continue
		}

		sentCount++
	}

	logInfo(
		runCtx,
		"Execucao do monitor de entrega concluida",
		slog.Int("delivery_alert_candidates", len(candidates)),
		slog.Int("delivery_alert_sent", sentCount),
	)

	return nil
}

func (s *service) dispatchCandidate(ctx context.Context, senderUser *model.UserModel, candidate model.DeliveryAlertCandidateModel) error {
	sentAt := s.now()
	messageBody := buildDeliveryMessage(candidate)
	noticeID, err := s.createAdminNotice(ctx, senderUser, candidate, sentAt)
	if err != nil {
		return fmt.Errorf("create admin notice: %w", err)
	}

	conversationID, messageID, err := s.conversationRepo.CreateOrAppendDirect(
		ctx,
		senderUser.ID,
		candidate.RecipientUserID,
		candidate.ProjectID,
		messageBody,
		sentAt,
	)
	if err != nil {
		return fmt.Errorf("create conversation message: %w", err)
	}

	_, err = s.deliveryAlertRepo.CreateEvent(ctx, &model.DeliveryAlertEventModel{
		BudgetID:        candidate.BudgetID,
		RecipientUserID: candidate.RecipientUserID,
		ConversationID:  conversationID,
		MessageID:       messageID,
		AlertType:       model.DeliveryAlertTypeDueInTwoDays,
		DeliveryDate:    truncateToDate(candidate.DeliveryDate),
		SentAt:          sentAt,
		CreatedAt:       sentAt,
		UpdatedAt:       sentAt,
	})
	if err != nil {
		if deliveryalertrepository.IsDuplicateEventError(err) {
			logInfo(
				ctx,
				"Alerta de entrega ignorado por duplicidade",
				slog.Int64("budget_id", candidate.BudgetID),
				slog.Int64("recipient_user_id", candidate.RecipientUserID),
			)
			return nil
		}

		return fmt.Errorf("persist delivery alert event: %w", err)
	}

	logInfo(
		ctx,
		"Alerta de entrega enviado com sucesso",
		slog.Int64("budget_id", candidate.BudgetID),
		slog.Int64("conversation_id", conversationID),
		slog.Int64("message_id", messageID),
		slog.Int64("notice_id", noticeID),
		slog.Int64("recipient_user_id", candidate.RecipientUserID),
	)

	return nil
}

func (s *service) createAdminNotice(
	ctx context.Context,
	senderUser *model.UserModel,
	candidate model.DeliveryAlertCandidateModel,
	sentAt time.Time,
) (int64, error) {
	adminRecipientIDs, err := s.resolveAdminRecipientUserIDs(ctx)
	if err != nil {
		return 0, err
	}

	noticeID, err := s.noticeRepo.Create(ctx, &model.NoticeModel{
		Title:           buildDeliveryNoticeTitle(candidate),
		Body:            buildDeliveryNoticeBody(candidate),
		ScopeType:       model.NoticeScopeUsers,
		Priority:        model.NoticePriorityWarning,
		Pinned:          false,
		ExpiresAt:       nil,
		CreatedByUserID: senderUser.ID,
		CreatedAt:       sentAt,
		UpdatedAt:       sentAt,
	}, adminRecipientIDs)
	if err != nil {
		return 0, err
	}

	return noticeID, nil
}

func (s *service) resolveAdminRecipientUserIDs(ctx context.Context) ([]int64, error) {
	users, err := s.userRepo.ListActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active admin users: %w", err)
	}

	adminRecipientIDs := make([]int64, 0, len(users))
	for _, user := range users {
		if user.Active && user.Role == model.RoleAdmin {
			adminRecipientIDs = append(adminRecipientIDs, user.ID)
		}
	}

	if len(adminRecipientIDs) == 0 {
		return nil, sql.ErrNoRows
	}

	return adminRecipientIDs, nil
}

func (s *service) resolveSenderUser(ctx context.Context) (*model.UserModel, error) {
	if s.config != nil && strings.TrimSpace(s.config.DeliveryAlertSenderUsername) != "" {
		user, err := s.userRepo.GetUserByUsername(ctx, strings.TrimSpace(s.config.DeliveryAlertSenderUsername))
		if err != nil {
			return nil, fmt.Errorf("load configured delivery alert sender: %w", err)
		}
		if user != nil && user.Active && user.Role == model.RoleAdmin {
			return user, nil
		}
	}

	users, err := s.userRepo.ListActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active users: %w", err)
	}

	for _, user := range users {
		if user.Active && user.Role == model.RoleAdmin {
			userCopy := user
			return &userCopy, nil
		}
	}

	return nil, sql.ErrNoRows
}

func buildDeliveryMessage(candidate model.DeliveryAlertCandidateModel) string {
	projectLabel := buildDeliveryProjectLabel(candidate)
	constructionCompany := buildDeliveryConstructionCompany(candidate)

	return fmt.Sprintf(
		"Alerta automatico do sistema: o orcamento %s da obra %s esta com entrega prevista para %s, faltando %d dias. Construtora/empresa: %s. Favor acompanhar o pedido.",
		candidate.BudgetNumber,
		projectLabel,
		truncateToDate(candidate.DeliveryDate).Format("02/01/2006"),
		candidate.DaysUntilDelivery,
		constructionCompany,
	)
}

func buildDeliveryNoticeTitle(candidate model.DeliveryAlertCandidateModel) string {
	return fmt.Sprintf("Alerta automatico de entrega do orcamento %s", candidate.BudgetNumber)
}

func buildDeliveryNoticeBody(candidate model.DeliveryAlertCandidateModel) string {
	projectLabel := buildDeliveryProjectLabel(candidate)
	constructionCompany := buildDeliveryConstructionCompany(candidate)

	return fmt.Sprintf(
		"Alerta automatico do sistema para acompanhamento administrativo: o orcamento %s da obra %s esta com entrega prevista para %s, faltando %d dias. Vendedor responsavel: %s (@%s). Construtora/empresa: %s.",
		candidate.BudgetNumber,
		projectLabel,
		truncateToDate(candidate.DeliveryDate).Format("02/01/2006"),
		candidate.DaysUntilDelivery,
		candidate.RecipientUserName,
		candidate.RecipientUsername,
		constructionCompany,
	)
}

func buildDeliveryProjectLabel(candidate model.DeliveryAlertCandidateModel) string {
	projectLabel := "obra nao informada"
	switch {
	case candidate.ProjectCode != nil && candidate.ProjectName != nil:
		projectLabel = fmt.Sprintf("%s - %s", *candidate.ProjectCode, *candidate.ProjectName)
	case candidate.ProjectName != nil:
		projectLabel = *candidate.ProjectName
	case candidate.ProjectCode != nil:
		projectLabel = *candidate.ProjectCode
	}

	return projectLabel
}

func buildDeliveryConstructionCompany(candidate model.DeliveryAlertCandidateModel) string {
	constructionCompany := strings.TrimSpace(candidate.ConstructionCompany)
	if constructionCompany == "" {
		constructionCompany = "empresa nao informada"
	}

	return constructionCompany
}

func truncateToDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func logInfo(ctx context.Context, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, message, attrs...)
}

func logError(ctx context.Context, message string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	logger.FromContext(ctx).LogAttrs(ctx, slog.LevelError, message, attrs...)
}
