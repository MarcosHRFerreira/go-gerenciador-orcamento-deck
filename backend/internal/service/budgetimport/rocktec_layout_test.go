package budgetimport

import "testing"

func TestRocktecImportLayoutShouldParseNormalizedRow(t *testing.T) {
	layout := newRocktecImportLayout()

	row, err := layout.ParseNormalizedRow(2, []string{
		"700001",
		"2",
		"09/06/2026",
		"Consulta de preco",
		"Informado",
		"Contato A",
		"Filtros",
		"BR100001",
		"Cliente A",
		"Obra XPTO",
		"DECK - Vendedor A",
		"Instalador A",
		"1234.56",
		"0.8",
	})
	if err != nil {
		t.Fatalf("expected row parse without error, got %v", err)
	}

	if row.rowNumber != 2 {
		t.Fatalf("expected row number 2, got %d", row.rowNumber)
	}
	if row.budgetNumber != "700001" {
		t.Fatalf("expected budget number 700001, got %s", row.budgetNumber)
	}
	if row.yearBudget != 2026 {
		t.Fatalf("expected year budget 2026, got %d", row.yearBudget)
	}
	if row.revision != 2 {
		t.Fatalf("expected revision 2, got %d", row.revision)
	}
	if row.projectName != "Obra Xpto" {
		t.Fatalf("expected project name Obra Xpto, got %s", row.projectName)
	}
	if row.productLineName != "Filtros" {
		t.Fatalf("expected product line Filtros, got %s", row.productLineName)
	}
	if row.constructionCompany != "Cliente A" {
		t.Fatalf("expected construction company Cliente A, got %s", row.constructionCompany)
	}
	if row.salespersonName != "Vendedor A" {
		t.Fatalf("expected salesperson Vendedor A, got %s", row.salespersonName)
	}
	if row.priorityName != notInformedName {
		t.Fatalf("expected default priority name %s, got %s", notInformedName, row.priorityName)
	}
	if row.statusName != "Em Negociacao" {
		t.Fatalf("expected status name Em Negociacao, got %s", row.statusName)
	}
	if len(row.warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(row.warnings))
	}
}

func TestRocktecImportLayoutShouldValidateHeaderAndEmptyRows(t *testing.T) {
	layout := newRocktecImportLayout()

	if !layout.HasExpectedHeader([]string{
		"Or\u00e7amento",
		"Revis\u00e3o",
		"Data de Emiss\u00e3o",
		"Tipo",
		"Status",
		"Contato",
		"Linha de produtos",
		"C\u00f3digo Cliente",
		"Nome Cliente",
		"Obra",
		"Vendedor",
		"Instalador",
		"Total do or\u00e7amento",
		"Fator M\u00e9dio",
	}) {
		t.Fatal("expected Rocktec header to be accepted")
	}

	if !layout.IsRowEmpty(make([]string, 14)) {
		t.Fatal("expected row with empty cells to be considered empty")
	}

	if layout.IsRowEmpty([]string{"45660"}) {
		t.Fatal("expected row with content to not be considered empty")
	}
}

func TestRocktecImportLayoutShouldRejectInvalidRows(t *testing.T) {
	layout := newRocktecImportLayout()

	_, err := layout.ParseNormalizedRow(2, []string{
		"invalid-date",
		"0",
	})
	if err == nil {
		t.Fatal("expected invalid row parse error")
	}
	if err.Error() != "Data do orcamento invalida." {
		t.Fatalf("expected invalid date error, got %v", err)
	}
}
