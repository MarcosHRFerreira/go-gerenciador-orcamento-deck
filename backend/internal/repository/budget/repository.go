package budget

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.BudgetModel) (int64, error)
	List(ctx context.Context, filters *dto.ListBudgetsFilters) ([]model.BudgetModel, int64, error)
	ExistsByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (bool, error)
	GetByID(ctx context.Context, budgetID int64) (*model.BudgetModel, error)
	Update(ctx context.Context, item *model.BudgetModel) error
	Delete(ctx context.Context, budgetID int64) error
	UpdateCurrentFollowUp(ctx context.Context, budgetID int64, currentFollowUp string, updatedAt time.Time) error
	UpdateStatus(ctx context.Context, budgetID int64, statusID int64, updatedAt time.Time) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.BudgetModel) (int64, error) {
	const query = `
		INSERT INTO budgets (
			budget_number,
			year_budget,
			revision,
			sent_at,
			gross_value,
			commission_value,
			area_m2,
			status_id,
			priority_id,
			installer_id,
			project_id,
			salesperson_id,
			contact_id,
			loss_reason_id,
			competitor_name,
			competitor_price,
			designer_name,
			specification_details,
			current_follow_up,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.BudgetNumber,
		item.YearBudget,
		item.Revision,
		item.SentAt,
		item.GrossValue,
		item.CommissionValue,
		item.AreaM2,
		item.StatusID,
		nullableInt64(item.PriorityID),
		nullableInt64(item.InstallerID),
		nullableInt64(item.ProjectID),
		nullableInt64(item.SalespersonID),
		nullableInt64(item.ContactID),
		nullableInt64(item.LossReasonID),
		item.CompetitorName,
		nullableFloat64(item.CompetitorPrice),
		item.DesignerName,
		item.SpecificationDetails,
		item.CurrentFollowUp,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context, filters *dto.ListBudgetsFilters) ([]model.BudgetModel, int64, error) {
	baseQuery := `
		SELECT
			b.id,
			b.budget_number,
			b.year_budget,
			b.revision,
			b.sent_at,
			b.gross_value,
			b.commission_value,
			b.area_m2,
			b.status_id,
			b.priority_id,
			b.installer_id,
			b.project_id,
			b.salesperson_id,
			b.contact_id,
			b.loss_reason_id,
			b.competitor_name,
			b.competitor_price,
			b.designer_name,
			b.specification_details,
			b.current_follow_up,
			b.created_at,
			b.updated_at
		FROM budgets b
		LEFT JOIN projects p ON p.id = b.project_id
	`

	whereClause, whereArgs := buildListWhereClause(filters)
	countQuery := "SELECT COUNT(*) FROM budgets b LEFT JOIN projects p ON p.id = b.project_id" + whereClause

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query, args := buildListQuery(baseQuery, filters, whereClause, whereArgs)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.BudgetModel, 0)
	for rows.Next() {
		var item model.BudgetModel
		if err := rows.Scan(
			&item.ID,
			&item.BudgetNumber,
			&item.YearBudget,
			&item.Revision,
			&item.SentAt,
			&item.GrossValue,
			&item.CommissionValue,
			&item.AreaM2,
			&item.StatusID,
			&item.PriorityID,
			&item.InstallerID,
			&item.ProjectID,
			&item.SalespersonID,
			&item.ContactID,
			&item.LossReasonID,
			&item.CompetitorName,
			&item.CompetitorPrice,
			&item.DesignerName,
			&item.SpecificationDetails,
			&item.CurrentFollowUp,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func buildListQuery(baseQuery string, filters *dto.ListBudgetsFilters, whereClause string, whereArgs []interface{}) (string, []interface{}) {
	args := append([]interface{}{}, whereArgs...)
	builder := strings.Builder{}
	builder.WriteString(baseQuery)
	builder.WriteString(whereClause)
	builder.WriteString("\nORDER BY ")
	builder.WriteString(orderByClause(filters))

	args = append(args, filters.PageSize)
	limitIndex := len(args)
	args = append(args, (filters.Page-1)*filters.PageSize)
	offsetIndex := len(args)
	builder.WriteString(fmt.Sprintf("\nLIMIT $%d OFFSET $%d", limitIndex, offsetIndex))

	return builder.String(), args
}

func buildListWhereClause(filters *dto.ListBudgetsFilters) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if filters != nil {
		if filters.BudgetNumber != "" {
			args = append(args, "%"+filters.BudgetNumber+"%")
			conditions = append(conditions, fmt.Sprintf("b.budget_number ILIKE $%d", len(args)))
		}
		if filters.YearBudget != nil {
			args = append(args, *filters.YearBudget)
			conditions = append(conditions, fmt.Sprintf("b.year_budget = $%d", len(args)))
		}
		if filters.StatusID != nil {
			args = append(args, *filters.StatusID)
			conditions = append(conditions, fmt.Sprintf("b.status_id = $%d", len(args)))
		}
		if filters.SalespersonID != nil {
			args = append(args, *filters.SalespersonID)
			conditions = append(conditions, fmt.Sprintf("b.salesperson_id = $%d", len(args)))
		}
		if filters.InstallerID != nil {
			args = append(args, *filters.InstallerID)
			conditions = append(conditions, fmt.Sprintf("b.installer_id = $%d", len(args)))
		}
		if filters.PriorityID != nil {
			args = append(args, *filters.PriorityID)
			conditions = append(conditions, fmt.Sprintf("b.priority_id = $%d", len(args)))
		}
		if filters.ProjectTypeID != nil {
			args = append(args, *filters.ProjectTypeID)
			conditions = append(conditions, fmt.Sprintf("p.project_type_id = $%d", len(args)))
		}
		if filters.DesignerName != "" {
			args = append(args, "%"+filters.DesignerName+"%")
			conditions = append(conditions, fmt.Sprintf("b.designer_name ILIKE $%d", len(args)))
		}
		if filters.CompetitorName != "" {
			args = append(args, "%"+filters.CompetitorName+"%")
			conditions = append(conditions, fmt.Sprintf("b.competitor_name ILIKE $%d", len(args)))
		}
		if filters.SentAtFrom != nil {
			args = append(args, *filters.SentAtFrom)
			conditions = append(conditions, fmt.Sprintf("b.sent_at >= $%d", len(args)))
		}
		if filters.SentAtTo != nil {
			args = append(args, *filters.SentAtTo)
			conditions = append(conditions, fmt.Sprintf("b.sent_at <= $%d", len(args)))
		}
		if filters.GrossValueMin != nil {
			args = append(args, *filters.GrossValueMin)
			conditions = append(conditions, fmt.Sprintf("b.gross_value >= $%d", len(args)))
		}
		if filters.GrossValueMax != nil {
			args = append(args, *filters.GrossValueMax)
			conditions = append(conditions, fmt.Sprintf("b.gross_value <= $%d", len(args)))
		}
	}

	builder := strings.Builder{}
	if len(conditions) > 0 {
		builder.WriteString("\nWHERE ")
		builder.WriteString(strings.Join(conditions, "\n  AND "))
	}

	return builder.String(), args
}

func orderByClause(filters *dto.ListBudgetsFilters) string {
	sortByMap := map[string]string{
		"sent_at":       "b.sent_at",
		"gross_value":   "b.gross_value",
		"created_at":    "b.created_at",
		"updated_at":    "b.updated_at",
		"year_budget":   "b.year_budget",
		"budget_number": "b.budget_number",
	}

	column, exists := sortByMap[filters.SortBy]
	if !exists {
		column = "b.sent_at"
	}

	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	return fmt.Sprintf("%s %s, b.id DESC", column, sortOrder)
}

func (r *repository) ExistsByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM budgets
			WHERE budget_number = $1 AND year_budget = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, budgetNumber, yearBudget).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) GetByID(ctx context.Context, budgetID int64) (*model.BudgetModel, error) {
	const query = `
		SELECT
			id,
			budget_number,
			year_budget,
			revision,
			sent_at,
			gross_value,
			commission_value,
			area_m2,
			status_id,
			priority_id,
			installer_id,
			project_id,
			salesperson_id,
			contact_id,
			loss_reason_id,
			competitor_name,
			competitor_price,
			designer_name,
			specification_details,
			current_follow_up,
			created_at,
			updated_at
		FROM budgets
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, budgetID)

	var item model.BudgetModel
	err := row.Scan(
		&item.ID,
		&item.BudgetNumber,
		&item.YearBudget,
		&item.Revision,
		&item.SentAt,
		&item.GrossValue,
		&item.CommissionValue,
		&item.AreaM2,
		&item.StatusID,
		&item.PriorityID,
		&item.InstallerID,
		&item.ProjectID,
		&item.SalespersonID,
		&item.ContactID,
		&item.LossReasonID,
		&item.CompetitorName,
		&item.CompetitorPrice,
		&item.DesignerName,
		&item.SpecificationDetails,
		&item.CurrentFollowUp,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) Update(ctx context.Context, item *model.BudgetModel) error {
	const query = `
		UPDATE budgets
		SET
			budget_number = $2,
			year_budget = $3,
			revision = $4,
			sent_at = $5,
			gross_value = $6,
			commission_value = $7,
			area_m2 = $8,
			status_id = $9,
			priority_id = $10,
			installer_id = $11,
			project_id = $12,
			salesperson_id = $13,
			contact_id = $14,
			loss_reason_id = $15,
			competitor_name = $16,
			competitor_price = $17,
			designer_name = $18,
			specification_details = $19,
			current_follow_up = $20,
			updated_at = $21
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.BudgetNumber,
		item.YearBudget,
		item.Revision,
		item.SentAt,
		item.GrossValue,
		item.CommissionValue,
		item.AreaM2,
		item.StatusID,
		nullableInt64(item.PriorityID),
		nullableInt64(item.InstallerID),
		nullableInt64(item.ProjectID),
		nullableInt64(item.SalespersonID),
		nullableInt64(item.ContactID),
		nullableInt64(item.LossReasonID),
		item.CompetitorName,
		nullableFloat64(item.CompetitorPrice),
		item.DesignerName,
		item.SpecificationDetails,
		item.CurrentFollowUp,
		item.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, budgetID int64) error {
	const query = `
		DELETE FROM budgets
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, budgetID)
	return err
}

func (r *repository) UpdateCurrentFollowUp(ctx context.Context, budgetID int64, currentFollowUp string, updatedAt time.Time) error {
	const query = `
		UPDATE budgets
		SET current_follow_up = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, budgetID, currentFollowUp, updatedAt)
	return err
}

func (r *repository) UpdateStatus(ctx context.Context, budgetID int64, statusID int64, updatedAt time.Time) error {
	const query = `
		UPDATE budgets
		SET status_id = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, budgetID, statusID, updatedAt)
	return err
}

func nullableInt64(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Int64
}

func nullableFloat64(value sql.NullFloat64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Float64
}
