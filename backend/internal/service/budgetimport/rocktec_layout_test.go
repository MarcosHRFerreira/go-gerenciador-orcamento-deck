package budgetimport

import "testing"

func TestRocktecImportLayoutShouldParseNormalizedRow(t *testing.T) {
	layout := newRocktecImportLayout()

	row, err := layout.ParseNormalizedRow(11, []string{
		"45660",
		"1001",
		"R2",
		"Instalador A",
		"Obra XPTO",
		"Industrial",
		"Vendedor A",
		"Contato A",
		"1234.56",
		"0.05",
		"10",
		"FECHADO",
		"-",
		"-",
		"-",
		"1000",
		"-",
		"-",
	})
	if err != nil {
		t.Fatalf("expected row parse without error, got %v", err)
	}

	if row.rowNumber != 11 {
		t.Fatalf("expected row number 11, got %d", row.rowNumber)
	}
	if row.budgetNumber != "1001" {
		t.Fatalf("expected budget number 1001, got %s", row.budgetNumber)
	}
	if row.yearBudget != 2025 {
		t.Fatalf("expected year budget 2025, got %d", row.yearBudget)
	}
	if row.revision != 2 {
		t.Fatalf("expected revision 2, got %d", row.revision)
	}
	if row.projectName != "Obra Xpto" {
		t.Fatalf("expected project name Obra Xpto, got %s", row.projectName)
	}
	if row.priorityName != notInformedName {
		t.Fatalf("expected default priority name %s, got %s", notInformedName, row.priorityName)
	}
	if len(row.warnings) != 4 {
		t.Fatalf("expected 4 warnings for missing optional text fields, got %d", len(row.warnings))
	}
}

func TestRocktecImportLayoutShouldValidateHeaderAndEmptyRows(t *testing.T) {
	layout := newRocktecImportLayout()

	if !layout.HasExpectedHeader([]string{
		"DATA",
		"Nº DE ORCA",
		"REV.",
		"INSTALADOR",
		"NOME DA OBRA",
		"TIPO DE OBRA",
		"VENDEDOR",
		"CONTATO",
		"VALOR BRUTO",
		"COMISSAO",
		"M2",
		"PRIORIDADE",
		"STATUS",
		"CONCORRENTE",
		"MOTIVO",
		"VALOR CONCORRENTE",
		"PROJETISTA",
		"ESPECIFICACOES",
	}) {
		t.Fatal("expected Rocktec header to be accepted")
	}

	if !layout.IsRowEmpty(make([]string, 18)) {
		t.Fatal("expected row with empty cells to be considered empty")
	}

	if layout.IsRowEmpty([]string{"45660"}) {
		t.Fatal("expected row with content to not be considered empty")
	}
}

func TestRocktecImportLayoutShouldRejectInvalidRows(t *testing.T) {
	layout := newRocktecImportLayout()

	_, err := layout.ParseNormalizedRow(11, []string{
		"invalid-date",
		"1001",
	})
	if err == nil {
		t.Fatal("expected invalid row parse error")
	}
	if err.Error() != "Data do orcamento invalida." {
		t.Fatalf("expected invalid date error, got %v", err)
	}
}
