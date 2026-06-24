package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

type Repository interface {
	GetSummary(ctx context.Context, filters *dto.DashboardSalespeopleFilters) (*dto.DashboardSummaryResponse, error)
	GetGrossValueRange(ctx context.Context, filters *dto.DashboardSalespeopleFilters) (*dto.DashboardGrossValueRangeResponse, error)
	ListSalespersonSummaries(ctx context.Context, filters *dto.DashboardSalespeopleFilters) ([]dto.DashboardSalespersonSummaryResponse, error)
	ListEstimatorSummaries(ctx context.Context, filters *dto.DashboardSalespeopleFilters) ([]dto.DashboardEstimatorSummaryResponse, error)
	ListStaleBudgets(ctx context.Context, filters *dto.DashboardSalespeopleFilters, limit int) ([]dto.DashboardStaleBudgetResponse, error)
	ListMonthlyEvolution(ctx context.Context, filters *dto.DashboardSalespeopleFilters, limit int) ([]dto.DashboardMonthlyEvolutionResponse, error)
	ListConstructionCompanyPerformance(ctx context.Context, filters *dto.DashboardSalespeopleFilters, limit int) ([]dto.DashboardEntityPerformanceResponse, error)
	ListProjectPerformance(ctx context.Context, filters *dto.DashboardSalespeopleFilters, limit int) ([]dto.DashboardEntityPerformanceResponse, error)
	ListLossReasonSummaries(ctx context.Context, filters *dto.DashboardSalespeopleFilters, limit int) ([]dto.DashboardLossReasonSummaryResponse, error)
	ListClosingTimeSummaries(ctx context.Context, filters *dto.DashboardSalespeopleFilters) ([]dto.DashboardClosingTimeSummaryResponse, error)
	ListAvailableYears(ctx context.Context, filters *dto.DashboardSalespeopleFilters) ([]int, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetGrossValueRange(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) (*dto.DashboardGrossValueRangeResponse, error) {
	query, args := buildFilteredBudgetsCTE(withoutGrossValueFilters(filters), `
		SELECT
			COALESCE(MIN(gross_value), 0)::double precision,
			COALESCE(MAX(gross_value), 0)::double precision
		FROM filtered_budgets
	`)

	row := r.db.QueryRowContext(ctx, query, args...)
	response := &dto.DashboardGrossValueRangeResponse{}
	if err := row.Scan(&response.Min, &response.Max); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *repository) GetSummary(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) (*dto.DashboardSummaryResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, `
		SELECT
			COUNT(DISTINCT salesperson_label)::int AS active_salespeople,
			COUNT(*)::int AS total_budgets,
			COALESCE(SUM(gross_value), 0)::double precision AS total_gross_value,
			COALESCE(AVG(gross_value), 0)::double precision AS average_ticket,
			COALESCE(SUM(CASE WHEN status_category = 'negotiation' THEN gross_value ELSE 0 END), 0)::double precision AS total_negotiation_gross_value,
			COALESCE((COUNT(*) FILTER (WHERE status_category = 'won')::double precision / NULLIF(COUNT(*), 0)) * 100, 0)::double precision AS conversion_rate,
			COALESCE((SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END) / NULLIF(SUM(gross_value), 0)) * 100, 0)::double precision AS value_conversion_rate,
			COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budgets,
			COUNT(*) FILTER (WHERE status_category = 'negotiation')::int AS negotiation_budgets,
			COUNT(*) FILTER (WHERE status_category = 'lost')::int AS lost_budgets,
			COUNT(*) FILTER (
				WHERE status_category = 'negotiation'
				  AND last_activity_at <= CURRENT_TIMESTAMP - INTERVAL '7 days'
			)::int AS stalled_budgets_count
		FROM filtered_budgets
	`)

	row := r.db.QueryRowContext(ctx, query, args...)
	var item dto.DashboardSummaryResponse
	if err := row.Scan(
		&item.ActiveSalespeople,
		&item.TotalBudgets,
		&item.TotalGrossValue,
		&item.AverageTicket,
		&item.TotalNegotiationGrossValue,
		&item.ConversionRate,
		&item.ValueConversionRate,
		&item.WonBudgets,
		&item.NegotiationBudgets,
		&item.LostBudgets,
		&item.StalledBudgetsCount,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *repository) ListSalespersonSummaries(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) ([]dto.DashboardSalespersonSummaryResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, `
		SELECT
			salesperson_label,
			COUNT(*)::int AS budget_count,
			COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
			COUNT(*) FILTER (WHERE status_category = 'negotiation')::int AS negotiation_budget_count,
			COALESCE(SUM(CASE WHEN status_category = 'negotiation' THEN gross_value ELSE 0 END), 0)::double precision AS negotiation_gross_value,
			COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budget_count,
			COUNT(*) FILTER (
				WHERE status_category = 'negotiation'
				  AND last_activity_at <= CURRENT_TIMESTAMP - INTERVAL '7 days'
			)::int AS stalled_budget_count,
			COALESCE(AVG(gross_value), 0)::double precision AS average_ticket,
			COALESCE((COUNT(*) FILTER (WHERE status_category = 'won')::double precision / NULLIF(COUNT(*), 0)) * 100, 0)::double precision AS conversion_rate,
			MAX(last_activity_at) AS last_activity_at
		FROM filtered_budgets
		GROUP BY salesperson_label
	`)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardSalespersonSummaryResponse, 0)
	for rows.Next() {
		var item dto.DashboardSalespersonSummaryResponse
		var lastActivityAt sql.NullTime
		if err := rows.Scan(
			&item.Label,
			&item.BudgetCount,
			&item.GrossValue,
			&item.NegotiationBudgetCount,
			&item.NegotiationGrossValue,
			&item.WonBudgetCount,
			&item.StalledBudgetCount,
			&item.AverageTicket,
			&item.ConversionRate,
			&lastActivityAt,
		); err != nil {
			return nil, err
		}

		if lastActivityAt.Valid {
			value := lastActivityAt.Time
			item.LastActivityAt = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListEstimatorSummaries(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) ([]dto.DashboardEstimatorSummaryResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, `
		SELECT
			estimator_label,
			COUNT(*)::int AS budget_count,
			COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
			COUNT(*) FILTER (WHERE status_category = 'negotiation')::int AS negotiation_budget_count,
			COALESCE(SUM(CASE WHEN status_category = 'negotiation' THEN gross_value ELSE 0 END), 0)::double precision AS negotiation_gross_value,
			COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budget_count,
			COUNT(*) FILTER (
				WHERE status_category = 'negotiation'
				  AND last_activity_at <= CURRENT_TIMESTAMP - INTERVAL '7 days'
			)::int AS stalled_budget_count,
			COALESCE(AVG(gross_value), 0)::double precision AS average_ticket,
			COALESCE((COUNT(*) FILTER (WHERE status_category = 'won')::double precision / NULLIF(COUNT(*), 0)) * 100, 0)::double precision AS conversion_rate,
			MAX(last_activity_at) AS last_activity_at
		FROM filtered_budgets
		WHERE estimator_id IS NOT NULL
		GROUP BY estimator_label
	`)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardEstimatorSummaryResponse, 0)
	for rows.Next() {
		var item dto.DashboardEstimatorSummaryResponse
		var lastActivityAt sql.NullTime
		if err := rows.Scan(
			&item.Label,
			&item.BudgetCount,
			&item.GrossValue,
			&item.NegotiationBudgetCount,
			&item.NegotiationGrossValue,
			&item.WonBudgetCount,
			&item.StalledBudgetCount,
			&item.AverageTicket,
			&item.ConversionRate,
			&lastActivityAt,
		); err != nil {
			return nil, err
		}

		if lastActivityAt.Valid {
			value := lastActivityAt.Time
			item.LastActivityAt = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListStaleBudgets(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	limit int,
) ([]dto.DashboardStaleBudgetResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, fmt.Sprintf(`
		SELECT
			id,
			budget_number,
			salesperson_label,
			project_label,
			status_label,
			construction_company_label,
			gross_value,
			last_activity_at,
			GREATEST(
				0,
				FLOOR(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - last_activity_at)) / 86400)
			)::int AS stalled_days
		FROM filtered_budgets
		WHERE status_category = 'negotiation'
		  AND last_activity_at <= CURRENT_TIMESTAMP - INTERVAL '7 days'
		ORDER BY stalled_days DESC, gross_value DESC
		LIMIT %d
	`, limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardStaleBudgetResponse, 0)
	for rows.Next() {
		var item dto.DashboardStaleBudgetResponse
		if err := rows.Scan(
			&item.ID,
			&item.BudgetNumber,
			&item.SalespersonLabel,
			&item.ProjectLabel,
			&item.StatusLabel,
			&item.ConstructionCompanyLabel,
			&item.GrossValue,
			&item.LastActivityAt,
			&item.StalledDays,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListMonthlyEvolution(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	limit int,
) ([]dto.DashboardMonthlyEvolutionResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, fmt.Sprintf(`
		, monthly AS (
			SELECT
				EXTRACT(YEAR FROM sent_at)::int AS year_value,
				EXTRACT(MONTH FROM sent_at)::int AS month_value,
				COUNT(*)::int AS budget_count,
				COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
				COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budget_count,
				COALESCE(SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END), 0)::double precision AS won_gross_value
			FROM filtered_budgets
			GROUP BY
				EXTRACT(YEAR FROM sent_at)::int,
				EXTRACT(MONTH FROM sent_at)::int
		),
		limited_monthly AS (
			SELECT
				year_value,
				month_value,
				budget_count,
				gross_value,
				won_budget_count,
				won_gross_value
			FROM monthly
			ORDER BY year_value DESC, month_value DESC
			LIMIT %d
		)
		SELECT
			year_value,
			month_value,
			budget_count,
			gross_value,
			won_budget_count,
			won_gross_value
		FROM limited_monthly
		ORDER BY year_value ASC, month_value ASC
	`, limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardMonthlyEvolutionResponse, 0)
	for rows.Next() {
		var yearValue int
		var monthValue int
		var item dto.DashboardMonthlyEvolutionResponse
		if err := rows.Scan(
			&yearValue,
			&monthValue,
			&item.BudgetCount,
			&item.GrossValue,
			&item.WonBudgetCount,
			&item.WonGrossValue,
		); err != nil {
			return nil, err
		}

		item.MonthKey = fmt.Sprintf("%04d-%02d", yearValue, monthValue)
		item.MonthLabel = formatMonthLabel(yearValue, monthValue)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListConstructionCompanyPerformance(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	limit int,
) ([]dto.DashboardEntityPerformanceResponse, error) {
	return r.listEntityPerformance(ctx, filters, "construction_company_label", limit)
}

func (r *repository) ListProjectPerformance(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	limit int,
) ([]dto.DashboardEntityPerformanceResponse, error) {
	return r.listEntityPerformance(ctx, filters, "project_label", limit)
}

func (r *repository) listEntityPerformance(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	groupField string,
	limit int,
) ([]dto.DashboardEntityPerformanceResponse, error) {
	queryBody := fmt.Sprintf(`
		SELECT
			%[1]s AS label,
			NULL::bigint AS project_id,
			COUNT(*)::int AS budget_count,
			COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budget_count,
			COUNT(*) FILTER (WHERE status_category = 'lost')::int AS lost_budget_count,
			COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
			COALESCE(SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END), 0)::double precision AS won_gross_value,
			COALESCE((COUNT(*) FILTER (WHERE status_category = 'won')::double precision / NULLIF(COUNT(*), 0)) * 100, 0)::double precision AS conversion_rate,
			COALESCE((SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END) / NULLIF(SUM(gross_value), 0)) * 100, 0)::double precision AS value_conversion_rate,
			MAX(last_activity_at) AS last_activity_at
		FROM filtered_budgets
		GROUP BY %[1]s
		ORDER BY gross_value DESC, budget_count DESC
		LIMIT %d
	`, groupField, limit)
	if groupField == "project_label" {
		queryBody = fmt.Sprintf(`
			SELECT
				project_label AS label,
				project_id,
				COUNT(*)::int AS budget_count,
				COUNT(*) FILTER (WHERE status_category = 'won')::int AS won_budget_count,
				COUNT(*) FILTER (WHERE status_category = 'lost')::int AS lost_budget_count,
				COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
				COALESCE(SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END), 0)::double precision AS won_gross_value,
				COALESCE((COUNT(*) FILTER (WHERE status_category = 'won')::double precision / NULLIF(COUNT(*), 0)) * 100, 0)::double precision AS conversion_rate,
				COALESCE((SUM(CASE WHEN status_category = 'won' THEN gross_value ELSE 0 END) / NULLIF(SUM(gross_value), 0)) * 100, 0)::double precision AS value_conversion_rate,
				MAX(last_activity_at) AS last_activity_at
			FROM filtered_budgets
			GROUP BY project_id, project_label
			ORDER BY gross_value DESC, budget_count DESC
			LIMIT %d
		`, limit)
	}
	query, args := buildFilteredBudgetsCTE(filters, queryBody)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardEntityPerformanceResponse, 0)
	for rows.Next() {
		var item dto.DashboardEntityPerformanceResponse
		var lastActivityAt sql.NullTime
		var projectID sql.NullInt64
		if err := rows.Scan(
			&item.Label,
			&projectID,
			&item.BudgetCount,
			&item.WonBudgetCount,
			&item.LostBudgetCount,
			&item.GrossValue,
			&item.WonGrossValue,
			&item.ConversionRate,
			&item.ValueConversionRate,
			&lastActivityAt,
		); err != nil {
			return nil, err
		}

		if lastActivityAt.Valid {
			value := lastActivityAt.Time
			item.LastActivityAt = &value
		}
		if projectID.Valid {
			value := projectID.Int64
			item.ProjectID = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListLossReasonSummaries(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	limit int,
) ([]dto.DashboardLossReasonSummaryResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, fmt.Sprintf(`
		SELECT
			loss_reason_label,
			COUNT(*)::int AS lost_budget_count,
			COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
			COALESCE(AVG(gross_value), 0)::double precision AS average_ticket
		FROM filtered_budgets
		WHERE status_category = 'lost'
		GROUP BY loss_reason_label
		ORDER BY gross_value DESC, lost_budget_count DESC
		LIMIT %d
	`, limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardLossReasonSummaryResponse, 0)
	for rows.Next() {
		var item dto.DashboardLossReasonSummaryResponse
		if err := rows.Scan(
			&item.Label,
			&item.LostBudgetCount,
			&item.GrossValue,
			&item.AverageTicket,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListClosingTimeSummaries(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) ([]dto.DashboardClosingTimeSummaryResponse, error) {
	query, args := buildFilteredBudgetsCTE(filters, `
		, closed_budgets AS (
			SELECT
				status_category,
				gross_value,
				GREATEST(
					0,
					EXTRACT(EPOCH FROM (COALESCE(closed_at, last_activity_at) - sent_at)) / 86400
				)::double precision AS closing_days
			FROM filtered_budgets
			WHERE status_category IN ('won', 'lost')
		),
		closing_summary AS (
			SELECT
				'Geral' AS label,
				COUNT(*)::int AS budget_count,
				COALESCE(AVG(closing_days), 0)::double precision AS average_closing_days,
				COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
				0 AS sort_order
			FROM closed_budgets
			UNION ALL
			SELECT
				'Fechados' AS label,
				COUNT(*)::int AS budget_count,
				COALESCE(AVG(closing_days), 0)::double precision AS average_closing_days,
				COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
				1 AS sort_order
			FROM closed_budgets
			WHERE status_category = 'won'
			UNION ALL
			SELECT
				'Cancelados' AS label,
				COUNT(*)::int AS budget_count,
				COALESCE(AVG(closing_days), 0)::double precision AS average_closing_days,
				COALESCE(SUM(gross_value), 0)::double precision AS gross_value,
				2 AS sort_order
			FROM closed_budgets
			WHERE status_category = 'lost'
		)
		SELECT
			label,
			budget_count,
			average_closing_days,
			gross_value
		FROM closing_summary
		WHERE budget_count > 0
		ORDER BY sort_order
	`)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.DashboardClosingTimeSummaryResponse, 0)
	for rows.Next() {
		var item dto.DashboardClosingTimeSummaryResponse
		if err := rows.Scan(
			&item.Label,
			&item.BudgetCount,
			&item.AverageClosingDays,
			&item.GrossValue,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListAvailableYears(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
) ([]int, error) {
	const baseQuery = `
		SELECT DISTINCT b.year_budget
		FROM budgets b
	`

	whereClause, args := buildWhereClause(withoutPeriodFilters(filters))
	query := baseQuery + whereClause + "\nORDER BY b.year_budget DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	years := make([]int, 0)
	for rows.Next() {
		var year int
		if err := rows.Scan(&year); err != nil {
			return nil, err
		}

		years = append(years, year)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return years, nil
}

func withoutPeriodFilters(filters *dto.DashboardSalespeopleFilters) *dto.DashboardSalespeopleFilters {
	if filters == nil {
		return nil
	}

	return &dto.DashboardSalespeopleFilters{
		InstallerID:             filters.InstallerID,
		StatusID:                filters.StatusID,
		SourceCompany:           filters.SourceCompany,
		SalespersonID:           filters.SalespersonID,
		RestrictedSalespersonID: filters.RestrictedSalespersonID,
		GrossValueMin:           filters.GrossValueMin,
		GrossValueMax:           filters.GrossValueMax,
	}
}

func withoutGrossValueFilters(filters *dto.DashboardSalespeopleFilters) *dto.DashboardSalespeopleFilters {
	if filters == nil {
		return nil
	}

	clonedFilters := *filters
	clonedFilters.GrossValueMin = nil
	clonedFilters.GrossValueMax = nil

	return &clonedFilters
}

func buildWhereClause(filters *dto.DashboardSalespeopleFilters) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if filters != nil {
		if filters.SourceCompany != "" {
			args = append(args, strings.TrimSpace(filters.SourceCompany))
			conditions = append(
				conditions,
				fmt.Sprintf("lower(b.source_company) = lower($%d)", len(args)),
			)
		}
		if filters.RestrictedSalespersonID != nil {
			args = append(args, *filters.RestrictedSalespersonID)
			conditions = append(
				conditions,
				fmt.Sprintf("b.salesperson_id = $%d", len(args)),
			)
		}
		if filters.SalespersonID != nil {
			args = append(args, *filters.SalespersonID)
			conditions = append(
				conditions,
				fmt.Sprintf("b.salesperson_id = $%d", len(args)),
			)
		}
		if filters.InstallerID != nil {
			args = append(args, *filters.InstallerID)
			conditions = append(
				conditions,
				fmt.Sprintf("b.installer_id = $%d", len(args)),
			)
		}
		if filters.StatusID != nil {
			args = append(args, *filters.StatusID)
			conditions = append(
				conditions,
				fmt.Sprintf("b.status_id = $%d", len(args)),
			)
		}
		if filters.Year != nil {
			args = append(args, *filters.Year)
			conditions = append(
				conditions,
				fmt.Sprintf("b.year_budget = $%d", len(args)),
			)
		}
		if filters.Month != nil {
			args = append(args, *filters.Month)
			conditions = append(
				conditions,
				fmt.Sprintf("EXTRACT(MONTH FROM b.sent_at) = $%d", len(args)),
			)
		}
		if filters.GrossValueMin != nil {
			args = append(args, *filters.GrossValueMin)
			conditions = append(
				conditions,
				fmt.Sprintf("b.gross_value >= $%d", len(args)),
			)
		}
		if filters.GrossValueMax != nil {
			args = append(args, *filters.GrossValueMax)
			conditions = append(
				conditions,
				fmt.Sprintf("b.gross_value <= $%d", len(args)),
			)
		}
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "\nWHERE " + strings.Join(conditions, "\n  AND "), args
}

func buildFilteredBudgetsCTE(
	filters *dto.DashboardSalespeopleFilters,
	body string,
) (string, []interface{}) {
	whereClause, args := buildWhereClause(filters)

	return `
		WITH budget_follow_up_activity AS (
			SELECT
				budget_id,
				MAX(follow_up_at) AS last_follow_up_at
			FROM budget_follow_ups
			GROUP BY budget_id
		),
		budget_final_activity AS (
			SELECT
				bsh.budget_id,
				MAX(bsh.changed_at) FILTER (
					WHERE lower(TRIM(COALESCE(target_status.name, ''))) IN ('pedido', 'cancelado')
					   OR target_status.is_final = TRUE
				) AS closed_at
			FROM budget_status_history bsh
			LEFT JOIN budget_statuses target_status ON target_status.id = bsh.to_status_id
			GROUP BY bsh.budget_id
		),
		budget_status_activity AS (
			SELECT
				budget_id,
				MAX(changed_at) AS last_status_changed_at
			FROM budget_status_history
			GROUP BY budget_id
		),
		filtered_budgets AS (
			SELECT
				b.id,
				b.budget_number,
				b.sent_at,
				b.year_budget,
				b.gross_value,
				b.estimator_id,
				b.project_id,
				COALESCE(NULLIF(TRIM(b.construction_company), ''), 'Construtora nao informada') AS construction_company_label,
				COALESCE(NULLIF(TRIM(lr.name), ''), 'Motivo nao informado') AS loss_reason_label,
				GREATEST(
					b.sent_at,
					COALESCE(follow_up_activity.last_follow_up_at, b.sent_at),
					COALESCE(status_activity.last_status_changed_at, b.sent_at)
				) AS last_activity_at,
				COALESCE(
					final_activity.closed_at,
					CASE
						WHEN lower(TRIM(COALESCE(bs.name, ''))) IN ('pedido', 'cancelado') OR bs.is_final = TRUE THEN GREATEST(
							b.sent_at,
							COALESCE(follow_up_activity.last_follow_up_at, b.sent_at),
							COALESCE(status_activity.last_status_changed_at, b.sent_at)
						)
						ELSE NULL
					END
				) AS closed_at,
				COALESCE(NULLIF(TRIM(bs.name), ''), 'Status nao informado') AS status_label,
				COALESCE(NULLIF(TRIM(p.name), ''), 'Sem obra vinculada') AS project_label,
				COALESCE(NULLIF(TRIM(s.name), ''), 'Sem vendedor') AS salesperson_label,
				COALESCE(NULLIF(TRIM(e.name), ''), 'Orcamentista sem nome') AS estimator_label,
				CASE
					WHEN lower(TRIM(COALESCE(bs.name, ''))) = 'pedido' THEN 'won'
					WHEN lower(TRIM(COALESCE(bs.name, ''))) = 'cancelado' THEN 'lost'
					ELSE 'negotiation'
				END AS status_category
			FROM budgets b
			LEFT JOIN budget_statuses bs ON bs.id = b.status_id
			LEFT JOIN loss_reasons lr ON lr.id = b.loss_reason_id
			LEFT JOIN projects p ON p.id = b.project_id
			LEFT JOIN salespeople s ON s.id = b.salesperson_id
			LEFT JOIN estimators e ON e.id = b.estimator_id
			LEFT JOIN budget_follow_up_activity follow_up_activity ON follow_up_activity.budget_id = b.id
			LEFT JOIN budget_final_activity final_activity ON final_activity.budget_id = b.id
			LEFT JOIN budget_status_activity status_activity ON status_activity.budget_id = b.id
			` + whereClause + `
		)
	` + body, args
}

func formatMonthLabel(yearValue int, monthValue int) string {
	dateValue := time.Date(yearValue, time.Month(monthValue), 1, 0, 0, 0, 0, time.UTC)
	monthLabels := [...]string{
		"Jan",
		"Fev",
		"Mar",
		"Abr",
		"Mai",
		"Jun",
		"Jul",
		"Ago",
		"Set",
		"Out",
		"Nov",
		"Dez",
	}

	return monthLabels[dateValue.Month()-1] + "/" + dateValue.Format("2006")
}
