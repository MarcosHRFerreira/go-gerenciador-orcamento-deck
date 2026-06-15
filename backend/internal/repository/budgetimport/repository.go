package budgetimport

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	CreateBatch(ctx context.Context, item *model.BudgetImportBatchModel) (int64, error)
	UpdateBatch(ctx context.Context, item *model.BudgetImportBatchModel) error
	CreateRowRaw(ctx context.Context, item *model.BudgetImportRowRawModel) (int64, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBatch(ctx context.Context, item *model.BudgetImportBatchModel) (int64, error) {
	const query = `
		INSERT INTO budget_import_batches (
			preview_id,
			file_name,
			source_company,
			source_layout,
			status,
			executed_by_user_id,
			started_at,
			finished_at,
			rows_processed,
			budgets_created,
			budgets_updated,
			budgets_ignored,
			rows_failed,
			catalogs_created,
			result_message,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.PreviewID,
		item.FileName,
		item.SourceCompany,
		item.SourceLayout,
		item.Status,
		nullableInt64(item.ExecutedByUserID),
		item.StartedAt,
		nullableTime(item.FinishedAt),
		item.RowsProcessed,
		item.BudgetsCreated,
		item.BudgetsUpdated,
		item.BudgetsIgnored,
		item.RowsFailed,
		item.CatalogsCreated,
		item.ResultMessage,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) UpdateBatch(ctx context.Context, item *model.BudgetImportBatchModel) error {
	const query = `
		UPDATE budget_import_batches
		SET
			status = $2,
			finished_at = $3,
			rows_processed = $4,
			budgets_created = $5,
			budgets_updated = $6,
			budgets_ignored = $7,
			rows_failed = $8,
			catalogs_created = $9,
			result_message = $10,
			updated_at = $11
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.Status,
		nullableTime(item.FinishedAt),
		item.RowsProcessed,
		item.BudgetsCreated,
		item.BudgetsUpdated,
		item.BudgetsIgnored,
		item.RowsFailed,
		item.CatalogsCreated,
		item.ResultMessage,
		item.UpdatedAt,
	)
	return err
}

func (r *repository) CreateRowRaw(ctx context.Context, item *model.BudgetImportRowRawModel) (int64, error) {
	const query = `
		INSERT INTO budget_import_rows_raw (
			import_batch_id,
			row_number,
			budget_number,
			status,
			action,
			raw_row_data,
			normalized_row_data,
			messages,
			budget_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8::jsonb, $9, $10)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ImportBatchID,
		item.RowNumber,
		item.BudgetNumber,
		item.Status,
		item.Action,
		string(item.RawRowData),
		nullableJSON(item.NormalizedRowData),
		string(item.Messages),
		nullableInt64(item.BudgetID),
		item.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func nullableInt64(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Int64
}

func nullableTime(value sql.NullTime) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Time
}

func nullableJSON(value []byte) interface{} {
	if len(value) == 0 {
		return nil
	}

	return string(value)
}
