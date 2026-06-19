package budget

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

var ErrProjectAlreadyHasPedido = errors.New("project already has pedido budget")

const automaticProjectCancellationNote = "Cancelado automaticamente porque outro orcamento da obra foi marcado como PEDIDO"
const automaticProjectRestorationNote = "Status restaurado automaticamente para permitir a troca do vencedor da obra"
const automaticProjectWinnerReplacementNote = "Vencedor anterior restaurado automaticamente porque outro orcamento da obra foi definido como PEDIDO"

type ChangeStatusParams struct {
	BudgetID                 int64
	StatusID                 int64
	UserID                   int64
	Notes                    string
	ChangedAt                time.Time
	EnforceProjectWinnerRule bool
	CancelledStatusID        int64
}

type ElectProjectWinnerParams struct {
	BudgetID          int64
	PedidoStatusID    int64
	CancelledStatusID int64
	UserID            int64
	Notes             string
	ChangedAt         time.Time
}

type Repository interface {
	Create(ctx context.Context, item *model.BudgetModel) (int64, error)
	List(ctx context.Context, filters *dto.ListBudgetsFilters) ([]model.BudgetModel, int64, error)
	ExistsByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (bool, error)
	ExistsBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (bool, error)
	GetByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (*model.BudgetModel, error)
	GetBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (*model.BudgetModel, error)
	GetByID(ctx context.Context, budgetID int64) (*model.BudgetModel, error)
	GetByIDScoped(ctx context.Context, budgetID int64, restrictedSalespersonID *int64, restrictedEstimatorID *int64) (*model.BudgetModel, error)
	Update(ctx context.Context, item *model.BudgetModel) error
	UpdateAndChangeStatus(ctx context.Context, item *model.BudgetModel, params *ChangeStatusParams) error
	Delete(ctx context.Context, budgetID int64) error
	UpdateCurrentFollowUp(ctx context.Context, budgetID int64, currentFollowUp string, updatedAt time.Time) error
	UpdateStatus(ctx context.Context, budgetID int64, statusID int64, updatedAt time.Time) error
	ChangeStatus(ctx context.Context, params *ChangeStatusParams) (int64, error)
	ElectProjectWinner(ctx context.Context, params *ElectProjectWinnerParams) error
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
			product_line_id,
			system_type_id,
			project_id,
			salesperson_id,
			estimator_id,
			contact_id,
			loss_reason_id,
			construction_company,
			competitor_name,
			competitor_price,
			projetista_name,
			specification_details,
			current_follow_up,
			source_company,
			source_layout,
			import_batch_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
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
		nullableInt64(item.ProductLineID),
		nullableInt64(item.SystemTypeID),
		nullableInt64(item.ProjectID),
		nullableInt64(item.SalespersonID),
		nullableInt64(item.EstimatorID),
		nullableInt64(item.ContactID),
		nullableInt64(item.LossReasonID),
		item.ConstructionCompany,
		item.CompetitorName,
		nullableFloat64(item.CompetitorPrice),
		item.ProjetistaName,
		item.SpecificationDetails,
		item.CurrentFollowUp,
		item.SourceCompany,
		item.SourceLayout,
		nullableInt64(item.ImportBatchID),
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
			b.product_line_id,
			b.system_type_id,
			b.project_id,
			b.salesperson_id,
			b.estimator_id,
			b.contact_id,
			b.loss_reason_id,
			b.construction_company,
			b.competitor_name,
			b.competitor_price,
			b.projetista_name,
			b.source_company,
			bs.name AS status_name,
			pr.name AS priority_name,
			i.name AS installer_name,
			pl.code AS product_line_code,
			pl.name AS product_line_name,
			st.code AS system_type_code,
			st.name AS system_type_name,
			p.code AS project_code,
			p.name AS project_name,
			s.name AS salesperson_name,
			e.name AS estimator_name,
			c.name AS contact_name,
			lr.name AS loss_reason_name,
			b.specification_details,
			b.current_follow_up,
			b.created_at,
			b.updated_at
		FROM budgets b
		LEFT JOIN budget_statuses bs ON bs.id = b.status_id
		LEFT JOIN priorities pr ON pr.id = b.priority_id
		LEFT JOIN installers i ON i.id = b.installer_id
		LEFT JOIN product_lines pl ON pl.id = b.product_line_id
		LEFT JOIN system_types st ON st.id = b.system_type_id
		LEFT JOIN projects p ON p.id = b.project_id
		LEFT JOIN salespeople s ON s.id = b.salesperson_id
		LEFT JOIN estimators e ON e.id = b.estimator_id
		LEFT JOIN contacts c ON c.id = b.contact_id
		LEFT JOIN loss_reasons lr ON lr.id = b.loss_reason_id
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
			&item.ProductLineID,
			&item.SystemTypeID,
			&item.ProjectID,
			&item.SalespersonID,
			&item.EstimatorID,
			&item.ContactID,
			&item.LossReasonID,
			&item.ConstructionCompany,
			&item.CompetitorName,
			&item.CompetitorPrice,
			&item.ProjetistaName,
			&item.SourceCompany,
			&item.StatusName,
			&item.PriorityName,
			&item.InstallerName,
			&item.ProductLineCode,
			&item.ProductLineName,
			&item.SystemTypeCode,
			&item.SystemTypeName,
			&item.ProjectCode,
			&item.ProjectName,
			&item.SalespersonName,
			&item.EstimatorName,
			&item.ContactName,
			&item.LossReasonName,
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
		if filters.RestrictedSalespersonID != nil {
			args = append(args, *filters.RestrictedSalespersonID)
			conditions = append(conditions, fmt.Sprintf("b.salesperson_id = $%d", len(args)))
		}
		if filters.SalespersonID != nil {
			args = append(args, *filters.SalespersonID)
			conditions = append(conditions, fmt.Sprintf("b.salesperson_id = $%d", len(args)))
		}
		if filters.RestrictedEstimatorID != nil {
			args = append(args, *filters.RestrictedEstimatorID)
			conditions = append(conditions, fmt.Sprintf("b.estimator_id = $%d", len(args)))
		}
		if filters.EstimatorID != nil {
			args = append(args, *filters.EstimatorID)
			conditions = append(conditions, fmt.Sprintf("b.estimator_id = $%d", len(args)))
		}
		if filters.InstallerID != nil {
			args = append(args, *filters.InstallerID)
			conditions = append(conditions, fmt.Sprintf("b.installer_id = $%d", len(args)))
		}
		if filters.SystemTypeID != nil {
			args = append(args, *filters.SystemTypeID)
			conditions = append(conditions, fmt.Sprintf("b.system_type_id = $%d", len(args)))
		}
		if filters.PriorityID != nil {
			args = append(args, *filters.PriorityID)
			conditions = append(conditions, fmt.Sprintf("b.priority_id = $%d", len(args)))
		}
		if filters.ProjectID != nil {
			args = append(args, *filters.ProjectID)
			conditions = append(conditions, fmt.Sprintf("b.project_id = $%d", len(args)))
		}
		if filters.ProjectCode != "" {
			args = append(args, "%"+strings.TrimSpace(filters.ProjectCode)+"%")
			conditions = append(conditions, fmt.Sprintf("p.code ILIKE $%d", len(args)))
		}
		if filters.SourceCompany != "" {
			args = append(args, strings.TrimSpace(filters.SourceCompany))
			conditions = append(
				conditions,
				fmt.Sprintf("lower(b.source_company) = lower($%d)", len(args)),
			)
		}
		if filters.ProjectTypeID != nil {
			args = append(args, *filters.ProjectTypeID)
			conditions = append(conditions, fmt.Sprintf("p.project_type_id = $%d", len(args)))
		}
		if filters.ProjetistaName != "" {
			args = append(args, "%"+filters.ProjetistaName+"%")
			conditions = append(conditions, fmt.Sprintf("b.projetista_name ILIKE $%d", len(args)))
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

func (r *repository) ExistsBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM budgets
			WHERE budget_number = $1
			  AND year_budget = $2
			  AND (source_company = $3 OR source_company = '')
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, budgetNumber, yearBudget, strings.TrimSpace(sourceCompany)).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) GetByNumberAndYear(ctx context.Context, budgetNumber string, yearBudget int) (*model.BudgetModel, error) {
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
			product_line_id,
			project_id,
			salesperson_id,
			estimator_id,
			contact_id,
			loss_reason_id,
			construction_company,
			competitor_name,
			competitor_price,
			projetista_name,
			NULL::text AS product_line_code,
			NULL::text AS product_line_name,
			NULL::text AS project_code,
			NULL::text AS project_name,
			NULL::text AS salesperson_name,
			NULL::text AS estimator_name,
			NULL::text AS contact_name,
			specification_details,
			current_follow_up,
			created_at,
			updated_at
		FROM budgets
		WHERE budget_number = $1 AND year_budget = $2
	`

	row := r.db.QueryRowContext(ctx, query, budgetNumber, yearBudget)

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
		&item.ProductLineID,
		&item.ProjectID,
		&item.SalespersonID,
		&item.EstimatorID,
		&item.ContactID,
		&item.LossReasonID,
		&item.ConstructionCompany,
		&item.CompetitorName,
		&item.CompetitorPrice,
		&item.ProjetistaName,
		&item.ProductLineCode,
		&item.ProductLineName,
		&item.ProjectCode,
		&item.ProjectName,
		&item.SalespersonName,
		&item.EstimatorName,
		&item.ContactName,
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

func (r *repository) GetBySourceAndNumberAndYear(ctx context.Context, sourceCompany string, budgetNumber string, yearBudget int) (*model.BudgetModel, error) {
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
			product_line_id,
			project_id,
			salesperson_id,
			estimator_id,
			contact_id,
			loss_reason_id,
			construction_company,
			competitor_name,
			competitor_price,
			projetista_name,
			NULL::text AS product_line_code,
			NULL::text AS product_line_name,
			NULL::text AS project_code,
			NULL::text AS project_name,
			NULL::text AS salesperson_name,
			NULL::text AS estimator_name,
			NULL::text AS contact_name,
			specification_details,
			current_follow_up,
			source_company,
			source_layout,
			import_batch_id,
			created_at,
			updated_at
		FROM budgets
		WHERE budget_number = $1
		  AND year_budget = $2
		  AND (source_company = $3 OR source_company = '')
		ORDER BY CASE WHEN source_company = $3 THEN 0 ELSE 1 END, id DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, budgetNumber, yearBudget, strings.TrimSpace(sourceCompany))

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
		&item.ProductLineID,
		&item.ProjectID,
		&item.SalespersonID,
		&item.EstimatorID,
		&item.ContactID,
		&item.LossReasonID,
		&item.ConstructionCompany,
		&item.CompetitorName,
		&item.CompetitorPrice,
		&item.ProjetistaName,
		&item.ProductLineCode,
		&item.ProductLineName,
		&item.ProjectCode,
		&item.ProjectName,
		&item.SalespersonName,
		&item.EstimatorName,
		&item.ContactName,
		&item.SpecificationDetails,
		&item.CurrentFollowUp,
		&item.SourceCompany,
		&item.SourceLayout,
		&item.ImportBatchID,
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

func (r *repository) GetByID(ctx context.Context, budgetID int64) (*model.BudgetModel, error) {
	return r.GetByIDScoped(ctx, budgetID, nil, nil)
}

func (r *repository) GetByIDScoped(ctx context.Context, budgetID int64, restrictedSalespersonID *int64, restrictedEstimatorID *int64) (*model.BudgetModel, error) {
	const query = `
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
			b.product_line_id,
			b.system_type_id,
			b.project_id,
			b.salesperson_id,
			b.estimator_id,
			b.contact_id,
			b.loss_reason_id,
			b.construction_company,
			b.competitor_name,
			b.competitor_price,
			b.projetista_name,
			bs.name AS status_name,
			pr.name AS priority_name,
			i.name AS installer_name,
			pl.code AS product_line_code,
			pl.name AS product_line_name,
			st.code AS system_type_code,
			st.name AS system_type_name,
			p.code AS project_code,
			p.name AS project_name,
			s.name AS salesperson_name,
			e.name AS estimator_name,
			c.name AS contact_name,
			lr.name AS loss_reason_name,
			b.specification_details,
			b.current_follow_up,
			b.created_at,
			b.updated_at
		FROM budgets b
		LEFT JOIN budget_statuses bs ON bs.id = b.status_id
		LEFT JOIN priorities pr ON pr.id = b.priority_id
		LEFT JOIN installers i ON i.id = b.installer_id
		LEFT JOIN product_lines pl ON pl.id = b.product_line_id
		LEFT JOIN system_types st ON st.id = b.system_type_id
		LEFT JOIN projects p ON p.id = b.project_id
		LEFT JOIN salespeople s ON s.id = b.salesperson_id
		LEFT JOIN estimators e ON e.id = b.estimator_id
		LEFT JOIN contacts c ON c.id = b.contact_id
		LEFT JOIN loss_reasons lr ON lr.id = b.loss_reason_id
		WHERE b.id = $1
	`
	args := []interface{}{budgetID}
	finalQuery := query
	if restrictedSalespersonID != nil {
		args = append(args, *restrictedSalespersonID)
		finalQuery += fmt.Sprintf(" AND b.salesperson_id = $%d", len(args))
	}
	if restrictedEstimatorID != nil {
		args = append(args, *restrictedEstimatorID)
		finalQuery += fmt.Sprintf(" AND b.estimator_id = $%d", len(args))
	}

	row := r.db.QueryRowContext(ctx, finalQuery, args...)

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
		&item.ProductLineID,
		&item.SystemTypeID,
		&item.ProjectID,
		&item.SalespersonID,
		&item.EstimatorID,
		&item.ContactID,
		&item.LossReasonID,
		&item.ConstructionCompany,
		&item.CompetitorName,
		&item.CompetitorPrice,
		&item.ProjetistaName,
		&item.StatusName,
		&item.PriorityName,
		&item.InstallerName,
		&item.ProductLineCode,
		&item.ProductLineName,
		&item.SystemTypeCode,
		&item.SystemTypeName,
		&item.ProjectCode,
		&item.ProjectName,
		&item.SalespersonName,
		&item.EstimatorName,
		&item.ContactName,
		&item.LossReasonName,
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
	return updateBudgetExecutor(ctx, r.db, item, true)
}

func (r *repository) UpdateAndChangeStatus(ctx context.Context, item *model.BudgetModel, params *ChangeStatusParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := updateBudgetExecutor(ctx, tx, item, false); err != nil {
		return err
	}

	if _, err := changeBudgetStatusExecutor(ctx, tx, params); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true
	return nil
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

func (r *repository) ChangeStatus(ctx context.Context, params *ChangeStatusParams) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	historyID, err := changeBudgetStatusExecutor(ctx, tx, params)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	committed = true
	return historyID, nil
}

func (r *repository) ElectProjectWinner(ctx context.Context, params *ElectProjectWinnerParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := electProjectWinnerExecutor(ctx, tx, params); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true
	return nil
}

func changeBudgetStatusExecutor(ctx context.Context, db executor, params *ChangeStatusParams) (int64, error) {
	if params == nil {
		return 0, errors.New("change status params is required")
	}

	budgetSnapshot, err := getBudgetStatusSnapshotForUpdate(ctx, db, params.BudgetID)
	if err != nil {
		return 0, err
	}

	if params.EnforceProjectWinnerRule && budgetSnapshot.ProjectID.Valid {
		if params.CancelledStatusID <= 0 {
			return 0, errors.New("cancelled status id is required when enforcing project winner rule")
		}

		hasPedidoWinner, err := projectHasOtherBudgetWithStatus(ctx, db, budgetSnapshot.ProjectID.Int64, params.BudgetID, params.StatusID)
		if err != nil {
			return 0, err
		}
		if hasPedidoWinner {
			return 0, ErrProjectAlreadyHasPedido
		}
	}

	historyID, err := insertStatusHistory(ctx, db, &model.BudgetStatusHistoryModel{
		BudgetID:        params.BudgetID,
		FromStatusID:    sql.NullInt64{Int64: budgetSnapshot.StatusID, Valid: true},
		ToStatusID:      params.StatusID,
		ChangedByUserID: params.UserID,
		Notes:           params.Notes,
		ChangedAt:       params.ChangedAt,
		CreatedAt:       params.ChangedAt,
		UpdatedAt:       params.ChangedAt,
	})
	if err != nil {
		return 0, err
	}

	if err := updateBudgetStatusExecutor(ctx, db, params.BudgetID, params.StatusID, params.ChangedAt); err != nil {
		return 0, err
	}

	if params.EnforceProjectWinnerRule && budgetSnapshot.ProjectID.Valid {
		if err := cancelOtherProjectBudgets(ctx, db, budgetSnapshot.ProjectID.Int64, params.BudgetID, params.CancelledStatusID, params.UserID, params.ChangedAt); err != nil {
			return 0, err
		}
	}

	return historyID, nil
}

func electProjectWinnerExecutor(ctx context.Context, db executor, params *ElectProjectWinnerParams) error {
	if params == nil {
		return errors.New("elect winner params is required")
	}
	if params.PedidoStatusID <= 0 {
		return errors.New("pedido status id is required")
	}
	if params.CancelledStatusID <= 0 {
		return errors.New("cancelled status id is required")
	}

	budgetSnapshot, err := getBudgetStatusSnapshotForUpdate(ctx, db, params.BudgetID)
	if err != nil {
		return err
	}
	if !budgetSnapshot.ProjectID.Valid {
		return errors.New("budget project id is required")
	}

	if err := restoreOtherProjectWinners(
		ctx,
		db,
		budgetSnapshot.ProjectID.Int64,
		params.BudgetID,
		params.PedidoStatusID,
		params.CancelledStatusID,
		params.UserID,
		params.ChangedAt,
	); err != nil {
		return err
	}

	if err := restoreAutomaticallyCancelledProjectBudgets(
		ctx,
		db,
		budgetSnapshot.ProjectID.Int64,
		params.BudgetID,
		params.CancelledStatusID,
		params.UserID,
		params.ChangedAt,
	); err != nil {
		return err
	}

	if budgetSnapshot.StatusID != params.PedidoStatusID {
		if _, err := insertStatusHistory(ctx, db, &model.BudgetStatusHistoryModel{
			BudgetID:        params.BudgetID,
			FromStatusID:    sql.NullInt64{Int64: budgetSnapshot.StatusID, Valid: true},
			ToStatusID:      params.PedidoStatusID,
			ChangedByUserID: params.UserID,
			Notes:           params.Notes,
			ChangedAt:       params.ChangedAt,
			CreatedAt:       params.ChangedAt,
			UpdatedAt:       params.ChangedAt,
		}); err != nil {
			return err
		}

		if err := updateBudgetStatusExecutor(ctx, db, params.BudgetID, params.PedidoStatusID, params.ChangedAt); err != nil {
			return err
		}
	}

	if err := cancelOtherProjectBudgets(
		ctx,
		db,
		budgetSnapshot.ProjectID.Int64,
		params.BudgetID,
		params.CancelledStatusID,
		params.UserID,
		params.ChangedAt,
	); err != nil {
		return err
	}

	return nil
}

func restoreOtherProjectWinners(
	ctx context.Context,
	db executor,
	projectID int64,
	currentBudgetID int64,
	pedidoStatusID int64,
	cancelledStatusID int64,
	userID int64,
	changedAt time.Time,
) error {
	const query = `
		SELECT id, status_id
		FROM budgets
		WHERE project_id = $1
			AND id <> $2
			AND status_id = $3
		FOR UPDATE
	`

	rows, err := db.QueryContext(ctx, query, projectID, currentBudgetID, pedidoStatusID)
	if err != nil {
		return err
	}
	defer rows.Close()

	winners := make([]budgetStatusChangeSnapshot, 0)
	for rows.Next() {
		var item budgetStatusChangeSnapshot
		if err := rows.Scan(&item.ID, &item.FromStatus); err != nil {
			return err
		}

		winners = append(winners, item)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, winner := range winners {
		restoreStatusID := cancelledStatusID
		latestHistory, err := getLatestStatusHistoryForCurrentStatus(ctx, db, winner.ID, pedidoStatusID)
		if err != nil {
			return err
		}
		if latestHistory != nil && latestHistory.FromStatusID.Valid && latestHistory.FromStatusID.Int64 > 0 {
			restoreStatusID = latestHistory.FromStatusID.Int64
		}

		if _, err := insertStatusHistory(ctx, db, &model.BudgetStatusHistoryModel{
			BudgetID:        winner.ID,
			FromStatusID:    sql.NullInt64{Int64: pedidoStatusID, Valid: true},
			ToStatusID:      restoreStatusID,
			ChangedByUserID: userID,
			Notes:           automaticProjectWinnerReplacementNote,
			ChangedAt:       changedAt,
			CreatedAt:       changedAt,
			UpdatedAt:       changedAt,
		}); err != nil {
			return err
		}

		if err := updateBudgetStatusExecutor(ctx, db, winner.ID, restoreStatusID, changedAt); err != nil {
			return err
		}
	}

	return nil
}

func restoreAutomaticallyCancelledProjectBudgets(
	ctx context.Context,
	db executor,
	projectID int64,
	currentBudgetID int64,
	cancelledStatusID int64,
	userID int64,
	changedAt time.Time,
) error {
	const query = `
		SELECT b.id, h.from_status_id
		FROM budgets b
		JOIN LATERAL (
			SELECT id, from_status_id, to_status_id, notes
			FROM budget_status_history
			WHERE budget_id = b.id
			ORDER BY changed_at DESC, id DESC
			LIMIT 1
		) h ON TRUE
		WHERE b.project_id = $1
			AND b.id <> $2
			AND b.status_id = $3
			AND h.to_status_id = $3
			AND h.notes = $4
			AND h.from_status_id IS NOT NULL
		FOR UPDATE OF b
	`

	rows, err := db.QueryContext(
		ctx,
		query,
		projectID,
		currentBudgetID,
		cancelledStatusID,
		automaticProjectCancellationNote,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	candidateBudgets := make([]budgetStatusChangeSnapshot, 0)
	for rows.Next() {
		var item budgetStatusChangeSnapshot
		if err := rows.Scan(&item.ID, &item.FromStatus); err != nil {
			return err
		}

		candidateBudgets = append(candidateBudgets, item)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, candidateBudget := range candidateBudgets {
		if _, err := insertStatusHistory(ctx, db, &model.BudgetStatusHistoryModel{
			BudgetID:        candidateBudget.ID,
			FromStatusID:    sql.NullInt64{Int64: cancelledStatusID, Valid: true},
			ToStatusID:      candidateBudget.FromStatus,
			ChangedByUserID: userID,
			Notes:           automaticProjectRestorationNote,
			ChangedAt:       changedAt,
			CreatedAt:       changedAt,
			UpdatedAt:       changedAt,
		}); err != nil {
			return err
		}

		if err := updateBudgetStatusExecutor(ctx, db, candidateBudget.ID, candidateBudget.FromStatus, changedAt); err != nil {
			return err
		}
	}

	return nil
}

type budgetStatusSnapshot struct {
	ID        int64
	StatusID  int64
	ProjectID sql.NullInt64
}

type budgetStatusChangeSnapshot struct {
	ID         int64
	FromStatus int64
}

type budgetLatestStatusHistorySnapshot struct {
	ID           int64
	BudgetID     int64
	FromStatusID sql.NullInt64
	ToStatusID   int64
	Notes        string
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func getBudgetStatusSnapshotForUpdate(ctx context.Context, db executor, budgetID int64) (*budgetStatusSnapshot, error) {
	const query = `
		SELECT id, status_id, project_id
		FROM budgets
		WHERE id = $1
		FOR UPDATE
	`

	row := db.QueryRowContext(ctx, query, budgetID)

	var item budgetStatusSnapshot
	err := row.Scan(&item.ID, &item.StatusID, &item.ProjectID)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func projectHasOtherBudgetWithStatus(ctx context.Context, db executor, projectID int64, budgetID int64, statusID int64) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM budgets
			WHERE project_id = $1
				AND id <> $2
				AND status_id = $3
		)
	`

	var exists bool
	if err := db.QueryRowContext(ctx, query, projectID, budgetID, statusID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func cancelOtherProjectBudgets(ctx context.Context, db executor, projectID int64, currentBudgetID int64, cancelledStatusID int64, userID int64, changedAt time.Time) error {
	const query = `
		SELECT id, status_id
		FROM budgets
		WHERE project_id = $1
			AND id <> $2
			AND status_id <> $3
		FOR UPDATE
	`

	rows, err := db.QueryContext(ctx, query, projectID, currentBudgetID, cancelledStatusID)
	if err != nil {
		return err
	}
	defer rows.Close()

	candidateBudgets := make([]budgetStatusChangeSnapshot, 0)
	for rows.Next() {
		var item budgetStatusChangeSnapshot
		if err := rows.Scan(&item.ID, &item.FromStatus); err != nil {
			return err
		}

		candidateBudgets = append(candidateBudgets, item)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, candidateBudget := range candidateBudgets {
		if _, err := insertStatusHistory(ctx, db, &model.BudgetStatusHistoryModel{
			BudgetID:        candidateBudget.ID,
			FromStatusID:    sql.NullInt64{Int64: candidateBudget.FromStatus, Valid: true},
			ToStatusID:      cancelledStatusID,
			ChangedByUserID: userID,
			Notes:           automaticProjectCancellationNote,
			ChangedAt:       changedAt,
			CreatedAt:       changedAt,
			UpdatedAt:       changedAt,
		}); err != nil {
			return err
		}
	}

	for _, candidateBudget := range candidateBudgets {
		if err := updateBudgetStatusExecutor(ctx, db, candidateBudget.ID, cancelledStatusID, changedAt); err != nil {
			return err
		}
	}

	return nil
}

func getLatestStatusHistoryForCurrentStatus(ctx context.Context, db executor, budgetID int64, currentStatusID int64) (*budgetLatestStatusHistorySnapshot, error) {
	const query = `
		SELECT id, budget_id, from_status_id, to_status_id, notes
		FROM budget_status_history
		WHERE budget_id = $1
			AND to_status_id = $2
		ORDER BY changed_at DESC, id DESC
		LIMIT 1
	`

	row := db.QueryRowContext(ctx, query, budgetID, currentStatusID)

	var item budgetLatestStatusHistorySnapshot
	err := row.Scan(
		&item.ID,
		&item.BudgetID,
		&item.FromStatusID,
		&item.ToStatusID,
		&item.Notes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func insertStatusHistory(ctx context.Context, db executor, item *model.BudgetStatusHistoryModel) (int64, error) {
	const query = `
		INSERT INTO budget_status_history (
			budget_id,
			from_status_id,
			to_status_id,
			changed_by_user_id,
			notes,
			changed_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := db.QueryRowContext(
		ctx,
		query,
		item.BudgetID,
		nullableInt64(item.FromStatusID),
		item.ToStatusID,
		item.ChangedByUserID,
		item.Notes,
		item.ChangedAt,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func updateBudgetStatusExecutor(ctx context.Context, db executor, budgetID int64, statusID int64, updatedAt time.Time) error {
	const query = `
		UPDATE budgets
		SET status_id = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := db.ExecContext(ctx, query, budgetID, statusID, updatedAt)
	return err
}

func updateBudgetExecutor(ctx context.Context, db executor, item *model.BudgetModel, includeStatus bool) error {
	if includeStatus {
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
				product_line_id = $12,
				system_type_id = $13,
				project_id = $14,
				salesperson_id = $15,
				estimator_id = $16,
				contact_id = $17,
				loss_reason_id = $18,
				construction_company = $19,
				competitor_name = $20,
				competitor_price = $21,
				projetista_name = $22,
				specification_details = $23,
				current_follow_up = $24,
				updated_at = $25,
				source_company = COALESCE(NULLIF($26, ''), source_company),
				source_layout = COALESCE(NULLIF($27, ''), source_layout),
				import_batch_id = COALESCE($28, import_batch_id)
			WHERE id = $1
		`

		_, err := db.ExecContext(
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
			nullableInt64(item.ProductLineID),
			nullableInt64(item.SystemTypeID),
			nullableInt64(item.ProjectID),
			nullableInt64(item.SalespersonID),
			nullableInt64(item.EstimatorID),
			nullableInt64(item.ContactID),
			nullableInt64(item.LossReasonID),
			item.ConstructionCompany,
			item.CompetitorName,
			nullableFloat64(item.CompetitorPrice),
			item.ProjetistaName,
			item.SpecificationDetails,
			item.CurrentFollowUp,
			item.UpdatedAt,
			item.SourceCompany,
			item.SourceLayout,
			nullableInt64(item.ImportBatchID),
		)

		return err
	}

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
			priority_id = $9,
			installer_id = $10,
			product_line_id = $11,
			system_type_id = $12,
			project_id = $13,
			salesperson_id = $14,
			estimator_id = $15,
			contact_id = $16,
			loss_reason_id = $17,
			construction_company = $18,
			competitor_name = $19,
			competitor_price = $20,
			projetista_name = $21,
			specification_details = $22,
			current_follow_up = $23,
			updated_at = $24,
			source_company = COALESCE(NULLIF($25, ''), source_company),
			source_layout = COALESCE(NULLIF($26, ''), source_layout),
			import_batch_id = COALESCE($27, import_batch_id)
		WHERE id = $1
	`

	_, err := db.ExecContext(
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
		nullableInt64(item.PriorityID),
		nullableInt64(item.InstallerID),
		nullableInt64(item.ProductLineID),
		nullableInt64(item.SystemTypeID),
		nullableInt64(item.ProjectID),
		nullableInt64(item.SalespersonID),
		nullableInt64(item.EstimatorID),
		nullableInt64(item.ContactID),
		nullableInt64(item.LossReasonID),
		item.ConstructionCompany,
		item.CompetitorName,
		nullableFloat64(item.CompetitorPrice),
		item.ProjetistaName,
		item.SpecificationDetails,
		item.CurrentFollowUp,
		item.UpdatedAt,
		item.SourceCompany,
		item.SourceLayout,
		nullableInt64(item.ImportBatchID),
	)

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
