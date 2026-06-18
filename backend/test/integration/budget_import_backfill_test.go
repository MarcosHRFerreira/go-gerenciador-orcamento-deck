package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLegacyBudgetSourceBackfillMigrationShouldPopulateRocktecMetadata(t *testing.T) {
	env := newIntegrationTestEnv(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	now := time.Date(2026, time.June, 15, 10, 0, 0, 0, time.UTC)

	legacyBudgetID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO budgets (
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
			projetista_name,
			specification_details,
			current_follow_up,
			source_company,
			source_layout,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id`,
		"LEGACY-ROCKTEC-001",
		2026,
		0,
		now,
		1500.00,
		100.00,
		35.00,
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
		"",
		nil,
		"",
		"",
		"",
		"",
		"",
		now,
		now,
	)

	troxBudgetID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO budgets (
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
			projetista_name,
			specification_details,
			current_follow_up,
			source_company,
			source_layout,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id`,
		"TROX-001",
		2026,
		0,
		now,
		2500.00,
		180.00,
		45.00,
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		seed.projectID,
		seed.salespersonID,
		seed.contactID,
		seed.lossReasonID,
		"",
		nil,
		"",
		"",
		"",
		"Trox",
		"trox",
		now,
		now,
	)

	legacySourceCompany, legacySourceLayout := env.requireBudgetSourceMetadata(t, legacyBudgetID)
	if legacySourceCompany != "" || legacySourceLayout != "" {
		t.Fatalf("expected legacy budget to start blank, got company=%q layout=%q", legacySourceCompany, legacySourceLayout)
	}

	troxSourceCompany, troxSourceLayout := env.requireBudgetSourceMetadata(t, troxBudgetID)
	if troxSourceCompany != "Trox" || troxSourceLayout != "trox" {
		t.Fatalf("expected Trox budget metadata to remain seeded, got company=%q layout=%q", troxSourceCompany, troxSourceLayout)
	}

	env.applyMigrationFile(t, "20260615190000_backfill_legacy_budget_sources.sql")

	legacySourceCompany, legacySourceLayout = env.requireBudgetSourceMetadata(t, legacyBudgetID)
	if legacySourceCompany != "Rocktec" || legacySourceLayout != "rocktec" {
		t.Fatalf("expected legacy budget metadata to be backfilled to Rocktec/rocktec, got company=%q layout=%q", legacySourceCompany, legacySourceLayout)
	}

	troxSourceCompany, troxSourceLayout = env.requireBudgetSourceMetadata(t, troxBudgetID)
	if troxSourceCompany != "Trox" || troxSourceLayout != "trox" {
		t.Fatalf("expected Trox budget metadata to remain unchanged, got company=%q layout=%q", troxSourceCompany, troxSourceLayout)
	}
}

func TestBudgetSalespersonCanonicalMigrationShouldRelinkBudgetsToSingleFirstNameSalesperson(t *testing.T) {
	env := newIntegrationTestEnv(t)
	seed := env.seedBudgetData(t, uniqueSuffix())
	now := time.Date(2026, time.June, 16, 10, 0, 0, 0, time.UTC)

	canonicalSalespersonID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO salespeople (name, email, phone, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		"Guilherme",
		"guilherme.canonical@local.dev",
		"11999999999",
		true,
		now,
		now,
	)

	duplicateSalespersonID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO salespeople (name, email, phone, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		"Guilherme Oliveira",
		"guilherme.oliveira@local.dev",
		"11888888888",
		true,
		now,
		now,
	)

	budgetID := env.insertReturningID(
		t,
		context.Background(),
		`INSERT INTO budgets (
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
			projetista_name,
			specification_details,
			current_follow_up,
			source_company,
			source_layout,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id`,
		"TROX-CANONICAL-001",
		2026,
		0,
		now,
		3500.00,
		200.00,
		40.00,
		seed.statusID,
		seed.priorityID,
		seed.installerID,
		seed.projectID,
		duplicateSalespersonID,
		seed.contactID,
		seed.lossReasonID,
		"",
		nil,
		"",
		"",
		"",
		"Trox",
		"trox",
		now,
		now,
	)

	beforeSalespersonID := env.requireBudgetSalespersonID(t, budgetID)
	if beforeSalespersonID != duplicateSalespersonID {
		t.Fatalf("expected budget to start with duplicate salesperson id %d, got %d", duplicateSalespersonID, beforeSalespersonID)
	}

	env.applyMigrationFile(t, "20260616090000_relink_budget_salespeople_to_canonical_first_name.sql")

	afterSalespersonID := env.requireBudgetSalespersonID(t, budgetID)
	if afterSalespersonID != canonicalSalespersonID {
		t.Fatalf("expected budget salesperson to be relinked to canonical id %d, got %d", canonicalSalespersonID, afterSalespersonID)
	}
}

func (e *integrationTestEnv) requireBudgetSourceMetadata(t *testing.T, budgetID int64) (string, string) {
	t.Helper()

	row := e.db.QueryRowContext(
		context.Background(),
		`SELECT source_company, source_layout FROM budgets WHERE id = $1`,
		budgetID,
	)

	var sourceCompany string
	var sourceLayout string
	if err := row.Scan(&sourceCompany, &sourceLayout); err != nil {
		t.Fatalf("failed to query budget source metadata: %v", err)
	}

	return sourceCompany, sourceLayout
}

func (e *integrationTestEnv) requireBudgetSalespersonID(t *testing.T, budgetID int64) int64 {
	t.Helper()

	row := e.db.QueryRowContext(
		context.Background(),
		`SELECT salesperson_id FROM budgets WHERE id = $1`,
		budgetID,
	)

	var salespersonID int64
	if err := row.Scan(&salespersonID); err != nil {
		t.Fatalf("failed to query budget salesperson id: %v", err)
	}

	return salespersonID
}

func (e *integrationTestEnv) applyMigrationFile(t *testing.T, fileName string) {
	t.Helper()

	projectRoot, err := projectRootDir()
	if err != nil {
		t.Fatalf("failed to resolve project root: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, "db", "migrations", fileName))
	if err != nil {
		t.Fatalf("failed to read migration file %s: %v", fileName, err)
	}

	upSQL := extractUpMigration(string(content))
	if upSQL == "" {
		t.Fatalf("expected migration %s to have up sql", fileName)
	}

	if _, err := e.db.ExecContext(context.Background(), upSQL); err != nil {
		t.Fatalf("failed to apply migration %s: %v", fileName, err)
	}
}
