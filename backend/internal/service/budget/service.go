package budget

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/accessscope"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, role model.UserRole, username string, req *dto.CreateBudgetRequest) (int64, error)
	List(ctx context.Context, filters *dto.ListBudgetsFilters, role model.UserRole, username string) (*dto.ListBudgetsResponse, error)
	ListDeliveryMonitor(ctx context.Context, filters *dto.ListBudgetDeliveryMonitorFilters, role model.UserRole, username string) (*dto.ListBudgetDeliveryMonitorResponse, error)
	GetByID(ctx context.Context, budgetID int64, role model.UserRole, username string) (*dto.BudgetResponse, error)
	Update(ctx context.Context, budgetID int64, role model.UserRole, username string, req *dto.UpdateBudgetRequest) error
	ElectProjectWinner(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.ElectBudgetWinnerRequest) error
	Delete(ctx context.Context, budgetID int64, role model.UserRole, username string) error
}

type service struct {
	repo             budgetrepository.Repository
	budgetStatusRepo budgetstatusrepository.Repository
	userRepo         userrepository.Repository
	salespersonRepo  salespersonrepository.Repository
	estimatorRepo    estimatorrepository.Repository
}

func NewService(
	repo budgetrepository.Repository,
	budgetStatusRepo budgetstatusrepository.Repository,
	userRepo userrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
	estimatorRepo estimatorrepository.Repository,
) Service {
	return &service{
		repo:             repo,
		budgetStatusRepo: budgetStatusRepo,
		userRepo:         userRepo,
		salespersonRepo:  salespersonRepo,
		estimatorRepo:    estimatorRepo,
	}
}

func (s *service) Create(ctx context.Context, role model.UserRole, username string, req *dto.CreateBudgetRequest) (int64, error) {
	budgetNumber := strings.TrimSpace(req.BudgetNumber)
	if budgetNumber == "" {
		return 0, apperror.BadRequest("budget_number e obrigatorio")
	}

	if req.YearBudget <= 0 {
		return 0, apperror.BadRequest("year_budget e obrigatorio")
	}

	if req.SentAt.IsZero() {
		return 0, apperror.BadRequest("sent_at e obrigatorio")
	}

	if req.GrossValue <= 0 {
		return 0, apperror.BadRequest("gross_value deve ser maior que zero")
	}

	if req.StatusID <= 0 {
		return 0, apperror.BadRequest("status_id e obrigatorio")
	}

	deliveryDate, err := parseOptionalDeliveryDate(req.DeliveryDate)
	if err != nil {
		return 0, err
	}

	exists, err := s.repo.ExistsByNumberAndYear(ctx, budgetNumber, req.YearBudget)
	if err != nil {
		return 0, apperror.Internal("failed to check budget uniqueness", err)
	}
	if exists {
		return 0, apperror.Conflict("Ja existe um orcamento para o budget_number e year_budget informados")
	}

	salespersonID, estimatorID, err := s.resolveCreateAndUpdateAssignments(
		ctx,
		role,
		username,
		req.SalespersonID,
		req.EstimatorID,
		true,
	)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.BudgetModel{
		BudgetNumber:         budgetNumber,
		YearBudget:           req.YearBudget,
		Revision:             req.Revision,
		SentAt:               req.SentAt,
		DeliveryDate:         deliveryDate,
		GrossValue:           req.GrossValue,
		CommissionValue:      req.CommissionValue,
		AreaM2:               req.AreaM2,
		StatusID:             req.StatusID,
		PriorityID:           newNullInt64(req.PriorityID),
		InstallerID:          newNullInt64(req.InstallerID),
		ProductLineID:        newNullInt64(req.ProductLineID),
		SystemTypeID:         newNullInt64(req.SystemTypeID),
		ProjectID:            newNullInt64(req.ProjectID),
		SalespersonID:        newNullInt64(salespersonID),
		EstimatorID:          newNullInt64(estimatorID),
		ContactID:            newNullInt64(req.ContactID),
		LossReasonID:         newNullInt64(req.LossReasonID),
		ConstructionCompany:  strings.TrimSpace(req.ConstructionCompany),
		CompetitorName:       strings.TrimSpace(req.CompetitorName),
		CompetitorPrice:      newNullFloat64(req.CompetitorPrice),
		ProjetistaName:       strings.TrimSpace(req.ProjetistaName),
		SpecificationDetails: strings.TrimSpace(req.SpecificationDetails),
		CurrentFollowUp:      strings.TrimSpace(req.CurrentFollowUp),
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return 0, mapBudgetPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context, filters *dto.ListBudgetsFilters, role model.UserRole, username string) (*dto.ListBudgetsResponse, error) {
	normalizedFilters, err := normalizeListFilters(filters)
	if err != nil {
		return nil, err
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return nil, err
	}
	normalizedFilters.RestrictedSalespersonID = scope.RestrictedSalespersonID
	normalizedFilters.RestrictedEstimatorID = scope.RestrictedEstimatorID

	items, total, err := s.repo.List(ctx, normalizedFilters)
	if err != nil {
		return nil, apperror.Internal("failed to list budgets", err)
	}

	return &dto.ListBudgetsResponse{
		Items:    mapBudgetResponses(items),
		Page:     normalizedFilters.Page,
		PageSize: normalizedFilters.PageSize,
		Total:    total,
	}, nil
}

func (s *service) GetByID(ctx context.Context, budgetID int64, role model.UserRole, username string) (*dto.BudgetResponse, error) {
	if budgetID <= 0 {
		return nil, apperror.BadRequest("budget_id e obrigatorio")
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
	if err != nil {
		return nil, apperror.Internal("failed to get budget", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Orcamento nao encontrado")
	}

	response := mapBudgetResponse(item)
	return &response, nil
}

func (s *service) ListDeliveryMonitor(
	ctx context.Context,
	filters *dto.ListBudgetDeliveryMonitorFilters,
	role model.UserRole,
	username string,
) (*dto.ListBudgetDeliveryMonitorResponse, error) {
	normalizedFilters, err := normalizeDeliveryMonitorFilters(filters)
	if err != nil {
		return nil, err
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return nil, err
	}
	normalizedFilters.RestrictedSalespersonID = scope.RestrictedSalespersonID
	normalizedFilters.RestrictedEstimatorID = scope.RestrictedEstimatorID

	items, total, summary, err := s.repo.ListDeliveryMonitor(ctx, normalizedFilters)
	if err != nil {
		return nil, apperror.Internal("failed to list budget delivery monitor", err)
	}

	return &dto.ListBudgetDeliveryMonitorResponse{
		Items:    mapBudgetDeliveryMonitorResponses(items),
		Summary:  *summary,
		Page:     normalizedFilters.Page,
		PageSize: normalizedFilters.PageSize,
		Total:    total,
	}, nil
}

func (s *service) Update(ctx context.Context, budgetID int64, role model.UserRole, username string, req *dto.UpdateBudgetRequest) error {
	if budgetID <= 0 {
		return apperror.BadRequest("budget_id e obrigatorio")
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return err
	}

	currentBudget, err := s.repo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
	if err != nil {
		return apperror.Internal("failed to check budget", err)
	}
	if currentBudget == nil {
		return apperror.NotFound("Orcamento nao encontrado")
	}

	budgetNumber := strings.TrimSpace(req.BudgetNumber)
	if budgetNumber == "" {
		return apperror.BadRequest("budget_number e obrigatorio")
	}

	if req.YearBudget <= 0 {
		return apperror.BadRequest("year_budget e obrigatorio")
	}

	if req.SentAt.IsZero() {
		return apperror.BadRequest("sent_at e obrigatorio")
	}

	if req.GrossValue <= 0 {
		return apperror.BadRequest("gross_value deve ser maior que zero")
	}

	if req.StatusID <= 0 {
		return apperror.BadRequest("status_id e obrigatorio")
	}

	deliveryDate, err := parseOptionalDeliveryDate(req.DeliveryDate)
	if err != nil {
		return err
	}

	if currentBudget.BudgetNumber != budgetNumber || currentBudget.YearBudget != req.YearBudget {
		exists, existsErr := s.repo.ExistsByNumberAndYear(ctx, budgetNumber, req.YearBudget)
		if existsErr != nil {
			return apperror.Internal("failed to check budget uniqueness", existsErr)
		}
		if exists {
			return apperror.Conflict("Ja existe um orcamento para o budget_number e year_budget informados")
		}
	}

	salespersonID, estimatorID, err := s.resolveCreateAndUpdateAssignments(
		ctx,
		role,
		username,
		req.SalespersonID,
		req.EstimatorID,
		false,
	)
	if err != nil {
		return err
	}

	updatedAt := time.Now()
	updateItem := &model.BudgetModel{
		ID:                   budgetID,
		BudgetNumber:         budgetNumber,
		YearBudget:           req.YearBudget,
		Revision:             req.Revision,
		SentAt:               req.SentAt,
		DeliveryDate:         deliveryDate,
		GrossValue:           req.GrossValue,
		CommissionValue:      req.CommissionValue,
		AreaM2:               req.AreaM2,
		StatusID:             req.StatusID,
		PriorityID:           newNullInt64(req.PriorityID),
		InstallerID:          newNullInt64(req.InstallerID),
		ProductLineID:        newNullInt64(req.ProductLineID),
		SystemTypeID:         newNullInt64(req.SystemTypeID),
		ProjectID:            newNullInt64(req.ProjectID),
		SalespersonID:        newNullInt64(salespersonID),
		EstimatorID:          newNullInt64(estimatorID),
		ContactID:            newNullInt64(req.ContactID),
		LossReasonID:         newNullInt64(req.LossReasonID),
		ConstructionCompany:  strings.TrimSpace(req.ConstructionCompany),
		CompetitorName:       strings.TrimSpace(req.CompetitorName),
		CompetitorPrice:      newNullFloat64(req.CompetitorPrice),
		ProjetistaName:       strings.TrimSpace(req.ProjetistaName),
		SpecificationDetails: strings.TrimSpace(req.SpecificationDetails),
		CurrentFollowUp:      strings.TrimSpace(req.CurrentFollowUp),
		UpdatedAt:            updatedAt,
	}

	if currentBudget.StatusID == req.StatusID {
		err = s.repo.Update(ctx, updateItem)
		if err != nil {
			return mapBudgetPersistenceError("update", err)
		}

		return nil
	}

	changeStatusParams, err := s.buildStatusChangeParams(ctx, username, budgetID, req.StatusID, updatedAt)
	if err != nil {
		return err
	}

	err = s.repo.UpdateAndChangeStatus(ctx, updateItem, changeStatusParams)
	if err != nil {
		return mapBudgetPersistenceError("update", err)
	}

	return nil
}

func (s *service) ElectProjectWinner(
	ctx context.Context,
	budgetID int64,
	userID int64,
	role model.UserRole,
	username string,
	req *dto.ElectBudgetWinnerRequest,
) error {
	if budgetID <= 0 {
		return apperror.BadRequest("budget_id e obrigatorio")
	}
	if userID <= 0 {
		return apperror.Unauthorized("Usuario autenticado obrigatorio")
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return err
	}

	currentBudget, err := s.repo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
	if err != nil {
		return apperror.Internal("failed to check budget", err)
	}
	if currentBudget == nil {
		return apperror.NotFound("Orcamento nao encontrado")
	}
	if !currentBudget.ProjectID.Valid {
		return apperror.BadRequest("O orcamento precisa estar vinculado a uma obra para definir vencedor")
	}

	pedidoStatus, err := s.ensureBudgetStatus(
		ctx,
		"PEDIDO",
		"Pedido",
		"Orcamento vencedor da obra",
		true,
		90,
	)
	if err != nil {
		return err
	}

	cancelledStatus, err := s.ensureBudgetStatus(
		ctx,
		"CANCELADO",
		"Cancelado",
		"Orcamento encerrado automaticamente apos definicao do vencedor da obra",
		true,
		100,
	)
	if err != nil {
		return err
	}

	notes := "Orcamento definido como vencedor da obra"
	if req != nil {
		trimmedNotes := strings.TrimSpace(req.Notes)
		if trimmedNotes != "" {
			notes = trimmedNotes
		}
	}

	err = s.repo.ElectProjectWinner(ctx, &budgetrepository.ElectProjectWinnerParams{
		BudgetID:          budgetID,
		PedidoStatusID:    pedidoStatus.ID,
		CancelledStatusID: cancelledStatus.ID,
		UserID:            userID,
		Notes:             notes,
		ChangedAt:         time.Now(),
	})
	if err != nil {
		if errors.Is(err, budgetrepository.ErrProjectAlreadyHasPedido) {
			return apperror.Conflict("Ja existe outro orcamento da obra marcado como PEDIDO")
		}

		return apperror.Internal("failed to elect project winner", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, budgetID int64, role model.UserRole, username string) error {
	if budgetID <= 0 {
		return apperror.BadRequest("budget_id e obrigatorio")
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return err
	}
	if role == model.RoleUser && scope.UserKind == model.UserKindSalesperson {
		return apperror.Forbidden("Perfil comercial nao pode excluir orcamentos")
	}

	item, err := s.repo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
	if err != nil {
		return apperror.Internal("failed to check budget", err)
	}
	if item == nil {
		return apperror.NotFound("Orcamento nao encontrado")
	}

	if err := s.repo.Delete(ctx, budgetID); err != nil {
		return apperror.Internal("failed to delete budget", err)
	}

	return nil
}

func mapBudgetPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "uq_budgets_budget_number_year":
			return apperror.Conflict("Ja existe um orcamento para o budget_number e year_budget informados")
		case "fk_budgets_status_id":
			return apperror.BadRequest("Status de orcamento nao encontrado")
		case "fk_budgets_priority_id":
			return apperror.BadRequest("Prioridade nao encontrada")
		case "fk_budgets_installer_id":
			return apperror.BadRequest("Instalador nao encontrado")
		case "fk_budgets_product_line_id":
			return apperror.BadRequest("Linha de produto nao encontrada")
		case "fk_budgets_system_type_id":
			return apperror.BadRequest("Tipo de sistema nao encontrado")
		case "fk_budgets_project_id":
			return apperror.BadRequest("Obra nao encontrada")
		case "fk_budgets_salesperson_id":
			return apperror.BadRequest("Vendedor nao encontrado")
		case "fk_budgets_estimator_id":
			return apperror.BadRequest("Orcamentista nao encontrado")
		case "fk_budgets_contact_id":
			return apperror.BadRequest("Contato nao encontrado")
		case "fk_budgets_loss_reason_id":
			return apperror.BadRequest("Motivo de perda nao encontrado")
		}

		if pgError.Code == "23503" {
			return apperror.BadRequest("Referencia de entidade relacionada invalida")
		}
		if pgError.Code == "23505" {
			return apperror.Conflict("Ja existe um orcamento para o budget_number e year_budget informados")
		}
	}

	return apperror.Internal("failed to "+action+" budget", err)
}

func normalizeListFilters(filters *dto.ListBudgetsFilters) (*dto.ListBudgetsFilters, error) {
	if filters == nil {
		return &dto.ListBudgetsFilters{}, nil
	}

	normalized := *filters
	normalized.BudgetNumber = strings.TrimSpace(filters.BudgetNumber)
	normalized.SourceCompany = strings.TrimSpace(filters.SourceCompany)
	normalized.ProjectCode = strings.TrimSpace(filters.ProjectCode)
	normalized.ProjectName = strings.TrimSpace(filters.ProjectName)
	normalized.ProjetistaName = strings.TrimSpace(filters.ProjetistaName)
	normalized.CompetitorName = strings.TrimSpace(filters.CompetitorName)
	normalized.SortBy = strings.TrimSpace(strings.ToLower(filters.SortBy))
	normalized.SortOrder = strings.TrimSpace(strings.ToLower(filters.SortOrder))

	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 50
	}
	if normalized.PageSize > 100 {
		return nil, apperror.BadRequest("page_size nao pode ser maior que 100")
	}
	if normalized.SortBy == "" {
		normalized.SortBy = "sent_at"
	}
	if normalized.SortOrder == "" {
		normalized.SortOrder = "desc"
	}

	if normalized.YearBudget != nil && *normalized.YearBudget <= 0 {
		return nil, apperror.BadRequest("year_budget deve ser maior que zero")
	}
	if normalized.StatusID != nil && *normalized.StatusID <= 0 {
		return nil, apperror.BadRequest("status_id deve ser maior que zero")
	}
	if normalized.SalespersonID != nil && *normalized.SalespersonID <= 0 {
		return nil, apperror.BadRequest("salesperson_id deve ser maior que zero")
	}
	if normalized.EstimatorID != nil && *normalized.EstimatorID <= 0 {
		return nil, apperror.BadRequest("estimator_id deve ser maior que zero")
	}
	if normalized.InstallerID != nil && *normalized.InstallerID <= 0 {
		return nil, apperror.BadRequest("installer_id deve ser maior que zero")
	}
	if normalized.SystemTypeID != nil && *normalized.SystemTypeID <= 0 {
		return nil, apperror.BadRequest("system_type_id deve ser maior que zero")
	}
	if normalized.PriorityID != nil && *normalized.PriorityID <= 0 {
		return nil, apperror.BadRequest("priority_id deve ser maior que zero")
	}
	if normalized.ProjectID != nil && *normalized.ProjectID <= 0 {
		return nil, apperror.BadRequest("project_id deve ser maior que zero")
	}
	if normalized.ProjectTypeID != nil && *normalized.ProjectTypeID <= 0 {
		return nil, apperror.BadRequest("project_type_id deve ser maior que zero")
	}
	if normalized.GrossValueMin != nil && *normalized.GrossValueMin < 0 {
		return nil, apperror.BadRequest("gross_value_min deve ser maior ou igual a zero")
	}
	if normalized.GrossValueMax != nil && *normalized.GrossValueMax < 0 {
		return nil, apperror.BadRequest("gross_value_max deve ser maior ou igual a zero")
	}
	if normalized.SentAtFrom != nil && normalized.SentAtTo != nil && normalized.SentAtFrom.After(*normalized.SentAtTo) {
		return nil, apperror.BadRequest("sent_at_from nao pode ser maior que sent_at_to")
	}
	if normalized.GrossValueMin != nil && normalized.GrossValueMax != nil && *normalized.GrossValueMin > *normalized.GrossValueMax {
		return nil, apperror.BadRequest("gross_value_min nao pode ser maior que gross_value_max")
	}
	if normalized.SortOrder != "asc" && normalized.SortOrder != "desc" {
		return nil, apperror.BadRequest("sort_order deve ser asc ou desc")
	}
	if !isAllowedSortBy(normalized.SortBy) {
		return nil, apperror.BadRequest("sort_by e invalido")
	}

	return &normalized, nil
}

func normalizeDeliveryMonitorFilters(filters *dto.ListBudgetDeliveryMonitorFilters) (*dto.ListBudgetDeliveryMonitorFilters, error) {
	if filters == nil {
		return &dto.ListBudgetDeliveryMonitorFilters{
			Page:     1,
			PageSize: 25,
		}, nil
	}

	normalized := *filters
	normalized.BudgetNumber = strings.TrimSpace(filters.BudgetNumber)
	normalized.ProjectName = strings.TrimSpace(filters.ProjectName)
	normalized.DeliveryStatus = strings.TrimSpace(strings.ToLower(filters.DeliveryStatus))

	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 25
	}
	if normalized.PageSize > 100 {
		return nil, apperror.BadRequest("page_size nao pode ser maior que 100")
	}
	if normalized.SalespersonID != nil && *normalized.SalespersonID <= 0 {
		return nil, apperror.BadRequest("salesperson_id deve ser maior que zero")
	}
	if normalized.StatusID != nil && *normalized.StatusID <= 0 {
		return nil, apperror.BadRequest("status_id deve ser maior que zero")
	}
	if normalized.DeliveryDateFrom != nil && normalized.DeliveryDateTo != nil && normalized.DeliveryDateFrom.After(*normalized.DeliveryDateTo) {
		return nil, apperror.BadRequest("delivery_date_from nao pode ser maior que delivery_date_to")
	}

	switch normalized.DeliveryStatus {
	case "", "overdue", "due_today", "due_in_1_day", "due_in_2_days", "future", "missing_delivery_date":
	default:
		return nil, apperror.BadRequest("delivery_status e invalido")
	}

	return &normalized, nil
}

func isAllowedSortBy(value string) bool {
	allowedValues := map[string]struct{}{
		"sent_at":       {},
		"gross_value":   {},
		"created_at":    {},
		"updated_at":    {},
		"year_budget":   {},
		"budget_number": {},
	}

	_, exists := allowedValues[value]
	return exists
}

func mapBudgetResponses(items []model.BudgetModel) []dto.BudgetResponse {
	response := make([]dto.BudgetResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapBudgetResponse(&item))
	}

	return response
}

func mapBudgetDeliveryMonitorResponses(items []model.BudgetDeliveryMonitorModel) []dto.BudgetDeliveryMonitorItemResponse {
	response := make([]dto.BudgetDeliveryMonitorItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.BudgetDeliveryMonitorItemResponse{
			ID:                  item.ID,
			BudgetNumber:        item.BudgetNumber,
			ProjectID:           nullableInt64Pointer(item.ProjectID),
			ProjectCode:         nullableStringPointer(item.ProjectCode),
			ProjectName:         nullableStringPointer(item.ProjectName),
			ConstructionCompany: item.ConstructionCompany,
			SalespersonID:       nullableInt64Pointer(item.SalespersonID),
			SalespersonName:     nullableStringPointer(item.SalespersonName),
			StatusID:            item.StatusID,
			StatusName:          nullableStringPointer(item.StatusName),
			DeliveryDate:        nullableDatePointer(item.DeliveryDate),
			DaysUntilDelivery:   nullableInt64Pointer(item.DaysUntilDelivery),
			DeliveryStatus:      item.DeliveryStatus,
			DeliveryStatusLabel: mapDeliveryStatusLabel(item.DeliveryStatus),
			UpdatedAt:           item.UpdatedAt,
		})
	}

	return response
}

func mapBudgetResponse(item *model.BudgetModel) dto.BudgetResponse {
	return dto.BudgetResponse{
		ID:                   item.ID,
		BudgetNumber:         item.BudgetNumber,
		YearBudget:           item.YearBudget,
		Revision:             item.Revision,
		SentAt:               item.SentAt,
		DeliveryDate:         nullableDatePointer(item.DeliveryDate),
		GrossValue:           item.GrossValue,
		CommissionValue:      item.CommissionValue,
		AreaM2:               item.AreaM2,
		StatusID:             item.StatusID,
		PriorityID:           nullableInt64Pointer(item.PriorityID),
		InstallerID:          nullableInt64Pointer(item.InstallerID),
		ProductLineID:        nullableInt64Pointer(item.ProductLineID),
		SystemTypeID:         nullableInt64Pointer(item.SystemTypeID),
		ProjectID:            nullableInt64Pointer(item.ProjectID),
		SalespersonID:        nullableInt64Pointer(item.SalespersonID),
		EstimatorID:          nullableInt64Pointer(item.EstimatorID),
		ContactID:            nullableInt64Pointer(item.ContactID),
		LossReasonID:         nullableInt64Pointer(item.LossReasonID),
		ConstructionCompany:  item.ConstructionCompany,
		CompetitorName:       item.CompetitorName,
		CompetitorPrice:      nullableFloat64Pointer(item.CompetitorPrice),
		ProjetistaName:       item.ProjetistaName,
		SourceCompany:        item.SourceCompany,
		StatusName:           nullableStringPointer(item.StatusName),
		PriorityName:         nullableStringPointer(item.PriorityName),
		InstallerName:        nullableStringPointer(item.InstallerName),
		ProductLineCode:      nullableStringPointer(item.ProductLineCode),
		ProductLineName:      nullableStringPointer(item.ProductLineName),
		SystemTypeCode:       nullableStringPointer(item.SystemTypeCode),
		SystemTypeName:       nullableStringPointer(item.SystemTypeName),
		ProjectCode:          nullableStringPointer(item.ProjectCode),
		ProjectName:          nullableStringPointer(item.ProjectName),
		SalespersonName:      nullableStringPointer(item.SalespersonName),
		EstimatorName:        nullableStringPointer(item.EstimatorName),
		ContactName:          nullableStringPointer(item.ContactName),
		LossReasonName:       nullableStringPointer(item.LossReasonName),
		SpecificationDetails: item.SpecificationDetails,
		CurrentFollowUp:      item.CurrentFollowUp,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}

func newNullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: *value,
		Valid: true,
	}
}

func newNullFloat64(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{}
	}

	return sql.NullFloat64{
		Float64: *value,
		Valid:   true,
	}
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}

func mapDeliveryStatusLabel(value string) string {
	switch value {
	case "overdue":
		return "Atrasado"
	case "due_today":
		return "Entrega hoje"
	case "due_in_1_day":
		return "Entrega em 1 dia"
	case "due_in_2_days":
		return "Entrega em 2 dias"
	case "future":
		return "Entrega futura"
	case "missing_delivery_date":
		return "Pedido sem data de entrega"
	default:
		return "Nao informado"
	}
}

func nullableFloat64Pointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func nullableDatePointer(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	formattedValue := value.Time.Format("2006-01-02")
	return &formattedValue
}

func parseOptionalDeliveryDate(value *string) (sql.NullTime, error) {
	if value == nil {
		return sql.NullTime{}, nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return sql.NullTime{}, nil
	}

	parsedValue, err := time.Parse("2006-01-02", trimmedValue)
	if err != nil {
		return sql.NullTime{}, apperror.BadRequest("delivery_date invalida")
	}

	return sql.NullTime{
		Time:  parsedValue,
		Valid: true,
	}, nil
}

func (s *service) resolveCreateAndUpdateAssignments(
	ctx context.Context,
	role model.UserRole,
	username string,
	requestedSalespersonID *int64,
	requestedEstimatorID *int64,
	isCreate bool,
) (*int64, *int64, error) {
	if role == model.RoleAdmin {
		return requestedSalespersonID, requestedEstimatorID, nil
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return nil, nil, err
	}

	switch scope.UserKind {
	case model.UserKindEstimator:
		return requestedSalespersonID, requestedEstimatorID, nil
	case model.UserKindSalesperson:
		if isCreate {
			return nil, nil, apperror.Forbidden("Perfil comercial nao pode criar orcamentos")
		}

		if scope.RestrictedSalespersonID == nil || *scope.RestrictedSalespersonID <= 0 {
			return nil, nil, apperror.Forbidden("Usuario operacional nao possui vinculo ativo com vendedor")
		}

		return scope.RestrictedSalespersonID, requestedEstimatorID, nil
	default:
		return nil, nil, apperror.Forbidden("Usuario operacional sem tipo funcional valido")
	}
}

func (s *service) buildStatusChangeParams(
	ctx context.Context,
	username string,
	budgetID int64,
	statusID int64,
	changedAt time.Time,
) (*budgetrepository.ChangeStatusParams, error) {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return nil, apperror.Unauthorized("Usuario autenticado obrigatorio")
	}

	user, err := s.userRepo.GetUserByUsername(ctx, normalizedUsername)
	if err != nil {
		return nil, apperror.Internal("failed to resolve authenticated user", err)
	}
	if user == nil || !user.Active {
		return nil, apperror.Unauthorized("Usuario autenticado obrigatorio")
	}

	status, err := s.budgetStatusRepo.GetByID(ctx, statusID)
	if err != nil {
		return nil, apperror.Internal("failed to check budget status", err)
	}
	if status == nil {
		return nil, apperror.BadRequest("Status de orcamento nao encontrado")
	}

	params := &budgetrepository.ChangeStatusParams{
		BudgetID:  budgetID,
		StatusID:  statusID,
		UserID:    user.ID,
		ChangedAt: changedAt,
	}

	return params, nil
}

func (s *service) ensureBudgetStatus(
	ctx context.Context,
	code string,
	name string,
	description string,
	isFinal bool,
	sortOrder int,
) (*model.BudgetStatusModel, error) {
	existingStatus, err := s.budgetStatusRepo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return nil, apperror.Internal("failed to check required budget status", err)
	}
	if existingStatus != nil {
		return existingStatus, nil
	}

	now := time.Now()
	statusID, err := s.budgetStatusRepo.Create(ctx, &model.BudgetStatusModel{
		Code:        code,
		Name:        name,
		Description: description,
		IsFinal:     isFinal,
		SortOrder:   sortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		existingStatus, lookupErr := s.budgetStatusRepo.GetByCodeOrName(ctx, code, name)
		if lookupErr == nil && existingStatus != nil {
			return existingStatus, nil
		}

		return nil, apperror.Internal("failed to ensure required budget status", err)
	}

	return &model.BudgetStatusModel{
		ID:          statusID,
		Code:        code,
		Name:        name,
		Description: description,
		IsFinal:     isFinal,
		SortOrder:   sortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
