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
			designer_name,
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
			designer_name,
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
