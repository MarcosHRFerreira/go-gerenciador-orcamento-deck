package budget

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateBudgetRequest) (int64, error)
	List(ctx context.Context, filters *dto.ListBudgetsFilters) (*dto.ListBudgetsResponse, error)
	GetByID(ctx context.Context, budgetID int64) (*dto.BudgetResponse, error)
	Update(ctx context.Context, budgetID int64, req *dto.UpdateBudgetRequest) error
	Delete(ctx context.Context, budgetID int64) error
}

type service struct {
	repo budgetrepository.Repository
}

func NewService(repo budgetrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateBudgetRequest) (int64, error) {
	budgetNumber := strings.TrimSpace(req.BudgetNumber)
	if budgetNumber == "" {
		return 0, apperror.BadRequest("budget_number is required")
	}

	if req.YearBudget <= 0 {
		return 0, apperror.BadRequest("year_budget is required")
	}

	if req.SentAt.IsZero() {
		return 0, apperror.BadRequest("sent_at is required")
	}

	if req.GrossValue <= 0 {
		return 0, apperror.BadRequest("gross_value must be greater than zero")
	}

	if req.StatusID <= 0 {
		return 0, apperror.BadRequest("status_id is required")
	}

	exists, err := s.repo.ExistsByNumberAndYear(ctx, budgetNumber, req.YearBudget)
	if err != nil {
		return 0, apperror.Internal("failed to check budget uniqueness", err)
	}
	if exists {
		return 0, apperror.Conflict("budget already exists for informed budget_number and year_budget")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.BudgetModel{
		BudgetNumber:         budgetNumber,
		YearBudget:           req.YearBudget,
		Revision:             req.Revision,
		SentAt:               req.SentAt,
		GrossValue:           req.GrossValue,
		CommissionValue:      req.CommissionValue,
		AreaM2:               req.AreaM2,
		StatusID:             req.StatusID,
		PriorityID:           newNullInt64(req.PriorityID),
		InstallerID:          newNullInt64(req.InstallerID),
		ProjectID:            newNullInt64(req.ProjectID),
		SalespersonID:        newNullInt64(req.SalespersonID),
		ContactID:            newNullInt64(req.ContactID),
		LossReasonID:         newNullInt64(req.LossReasonID),
		CompetitorName:       strings.TrimSpace(req.CompetitorName),
		CompetitorPrice:      newNullFloat64(req.CompetitorPrice),
		DesignerName:         strings.TrimSpace(req.DesignerName),
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

func (s *service) List(ctx context.Context, filters *dto.ListBudgetsFilters) (*dto.ListBudgetsResponse, error) {
	normalizedFilters, err := normalizeListFilters(filters)
	if err != nil {
		return nil, err
	}

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

func (s *service) GetByID(ctx context.Context, budgetID int64) (*dto.BudgetResponse, error) {
	if budgetID <= 0 {
		return nil, apperror.BadRequest("budget_id is required")
	}

	item, err := s.repo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, apperror.Internal("failed to get budget", err)
	}
	if item == nil {
		return nil, apperror.NotFound("budget not found")
	}

	response := mapBudgetResponse(item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, budgetID int64, req *dto.UpdateBudgetRequest) error {
	if budgetID <= 0 {
		return apperror.BadRequest("budget_id is required")
	}

	currentBudget, err := s.repo.GetByID(ctx, budgetID)
	if err != nil {
		return apperror.Internal("failed to check budget", err)
	}
	if currentBudget == nil {
		return apperror.NotFound("budget not found")
	}

	budgetNumber := strings.TrimSpace(req.BudgetNumber)
	if budgetNumber == "" {
		return apperror.BadRequest("budget_number is required")
	}

	if req.YearBudget <= 0 {
		return apperror.BadRequest("year_budget is required")
	}

	if req.SentAt.IsZero() {
		return apperror.BadRequest("sent_at is required")
	}

	if req.GrossValue <= 0 {
		return apperror.BadRequest("gross_value must be greater than zero")
	}

	if req.StatusID <= 0 {
		return apperror.BadRequest("status_id is required")
	}

	if currentBudget.BudgetNumber != budgetNumber || currentBudget.YearBudget != req.YearBudget {
		exists, existsErr := s.repo.ExistsByNumberAndYear(ctx, budgetNumber, req.YearBudget)
		if existsErr != nil {
			return apperror.Internal("failed to check budget uniqueness", existsErr)
		}
		if exists {
			return apperror.Conflict("budget already exists for informed budget_number and year_budget")
		}
	}

	err = s.repo.Update(ctx, &model.BudgetModel{
		ID:                   budgetID,
		BudgetNumber:         budgetNumber,
		YearBudget:           req.YearBudget,
		Revision:             req.Revision,
		SentAt:               req.SentAt,
		GrossValue:           req.GrossValue,
		CommissionValue:      req.CommissionValue,
		AreaM2:               req.AreaM2,
		StatusID:             req.StatusID,
		PriorityID:           newNullInt64(req.PriorityID),
		InstallerID:          newNullInt64(req.InstallerID),
		ProjectID:            newNullInt64(req.ProjectID),
		SalespersonID:        newNullInt64(req.SalespersonID),
		ContactID:            newNullInt64(req.ContactID),
		LossReasonID:         newNullInt64(req.LossReasonID),
		CompetitorName:       strings.TrimSpace(req.CompetitorName),
		CompetitorPrice:      newNullFloat64(req.CompetitorPrice),
		DesignerName:         strings.TrimSpace(req.DesignerName),
		SpecificationDetails: strings.TrimSpace(req.SpecificationDetails),
		CurrentFollowUp:      strings.TrimSpace(req.CurrentFollowUp),
		UpdatedAt:            time.Now(),
	})
	if err != nil {
		return mapBudgetPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, budgetID int64) error {
	if budgetID <= 0 {
		return apperror.BadRequest("budget_id is required")
	}

	item, err := s.repo.GetByID(ctx, budgetID)
	if err != nil {
		return apperror.Internal("failed to check budget", err)
	}
	if item == nil {
		return apperror.NotFound("budget not found")
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
			return apperror.Conflict("budget already exists for informed budget_number and year_budget")
		case "fk_budgets_status_id":
			return apperror.BadRequest("budget status not found")
		case "fk_budgets_priority_id":
			return apperror.BadRequest("priority not found")
		case "fk_budgets_installer_id":
			return apperror.BadRequest("installer not found")
		case "fk_budgets_project_id":
			return apperror.BadRequest("project not found")
		case "fk_budgets_salesperson_id":
			return apperror.BadRequest("salesperson not found")
		case "fk_budgets_contact_id":
			return apperror.BadRequest("contact not found")
		case "fk_budgets_loss_reason_id":
			return apperror.BadRequest("loss reason not found")
		}

		if pgError.Code == "23503" {
			return apperror.BadRequest("invalid related entity reference")
		}
		if pgError.Code == "23505" {
			return apperror.Conflict("budget already exists for informed budget_number and year_budget")
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
	normalized.DesignerName = strings.TrimSpace(filters.DesignerName)
	normalized.CompetitorName = strings.TrimSpace(filters.CompetitorName)
	normalized.SortBy = strings.TrimSpace(strings.ToLower(filters.SortBy))
	normalized.SortOrder = strings.TrimSpace(strings.ToLower(filters.SortOrder))

	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 20
	}
	if normalized.PageSize > 100 {
		return nil, apperror.BadRequest("page_size cannot be greater than 100")
	}
	if normalized.SortBy == "" {
		normalized.SortBy = "sent_at"
	}
	if normalized.SortOrder == "" {
		normalized.SortOrder = "desc"
	}

	if normalized.YearBudget != nil && *normalized.YearBudget <= 0 {
		return nil, apperror.BadRequest("year_budget must be greater than zero")
	}
	if normalized.StatusID != nil && *normalized.StatusID <= 0 {
		return nil, apperror.BadRequest("status_id must be greater than zero")
	}
	if normalized.SalespersonID != nil && *normalized.SalespersonID <= 0 {
		return nil, apperror.BadRequest("salesperson_id must be greater than zero")
	}
	if normalized.InstallerID != nil && *normalized.InstallerID <= 0 {
		return nil, apperror.BadRequest("installer_id must be greater than zero")
	}
	if normalized.PriorityID != nil && *normalized.PriorityID <= 0 {
		return nil, apperror.BadRequest("priority_id must be greater than zero")
	}
	if normalized.ProjectTypeID != nil && *normalized.ProjectTypeID <= 0 {
		return nil, apperror.BadRequest("project_type_id must be greater than zero")
	}
	if normalized.GrossValueMin != nil && *normalized.GrossValueMin < 0 {
		return nil, apperror.BadRequest("gross_value_min must be greater than or equal to zero")
	}
	if normalized.GrossValueMax != nil && *normalized.GrossValueMax < 0 {
		return nil, apperror.BadRequest("gross_value_max must be greater than or equal to zero")
	}
	if normalized.SentAtFrom != nil && normalized.SentAtTo != nil && normalized.SentAtFrom.After(*normalized.SentAtTo) {
		return nil, apperror.BadRequest("sent_at_from cannot be greater than sent_at_to")
	}
	if normalized.GrossValueMin != nil && normalized.GrossValueMax != nil && *normalized.GrossValueMin > *normalized.GrossValueMax {
		return nil, apperror.BadRequest("gross_value_min cannot be greater than gross_value_max")
	}
	if normalized.SortOrder != "asc" && normalized.SortOrder != "desc" {
		return nil, apperror.BadRequest("sort_order must be asc or desc")
	}
	if !isAllowedSortBy(normalized.SortBy) {
		return nil, apperror.BadRequest("sort_by is invalid")
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

func mapBudgetResponse(item *model.BudgetModel) dto.BudgetResponse {
	return dto.BudgetResponse{
		ID:                   item.ID,
		BudgetNumber:         item.BudgetNumber,
		YearBudget:           item.YearBudget,
		Revision:             item.Revision,
		SentAt:               item.SentAt,
		GrossValue:           item.GrossValue,
		CommissionValue:      item.CommissionValue,
		AreaM2:               item.AreaM2,
		StatusID:             item.StatusID,
		PriorityID:           nullableInt64Pointer(item.PriorityID),
		InstallerID:          nullableInt64Pointer(item.InstallerID),
		ProjectID:            nullableInt64Pointer(item.ProjectID),
		SalespersonID:        nullableInt64Pointer(item.SalespersonID),
		ContactID:            nullableInt64Pointer(item.ContactID),
		LossReasonID:         nullableInt64Pointer(item.LossReasonID),
		CompetitorName:       item.CompetitorName,
		CompetitorPrice:      nullableFloat64Pointer(item.CompetitorPrice),
		DesignerName:         item.DesignerName,
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

func nullableFloat64Pointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}
