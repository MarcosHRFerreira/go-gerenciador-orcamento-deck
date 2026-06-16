package budgetimport

import "testing"

func TestTroxImportLayoutShouldParseNormalizedRow(t *testing.T) {
	layout := newTroxImportLayout()

	row, err := layout.ParseNormalizedRow(2, []string{
		"477139",
		"0",
		"09/06/2026",
		"Consulta de preco",
		"Informado",
		"ELDER J. BONETTI",
		"FILTROS",
		"BR1007854",
		"ABECON ENGENHARIA E CLIMATIZACAO LT",
		"DIVERSOS DE JUNHO",
		"DECK - EMANUEL FERRI",
		"ABECON ENGENHARIA E CLIMATIZACAO LT",
		"65515.83",
		"0.8",
	})
	if err != nil {
		t.Fatalf("expected row parse without error, got %v", err)
	}

	if row.rowNumber != 2 {
		t.Fatalf("expected row number 2, got %d", row.rowNumber)
	}
	if row.budgetNumber != "477139" {
		t.Fatalf("expected budget number 477139, got %s", row.budgetNumber)
	}
	if row.yearBudget != 2026 {
		t.Fatalf("expected year budget 2026, got %d", row.yearBudget)
	}
	if row.revision != 0 {
		t.Fatalf("expected revision 0, got %d", row.revision)
	}
	if row.projectName != "Diversos De Junho" {
		t.Fatalf("expected project name Diversos De Junho, got %s", row.projectName)
	}
	if row.salespersonName != "Emanuel Ferri" {
		t.Fatalf("expected normalized salesperson Emanuel Ferri, got %s", row.salespersonName)
	}
	if row.currentFollowUp != "Informado" {
		t.Fatalf("expected current follow-up Informado, got %s", row.currentFollowUp)
	}
	if row.statusName != notInformedName {
		t.Fatalf("expected status name %s, got %s", notInformedName, row.statusName)
	}
}

func TestTroxImportLayoutShouldValidateHeaderAndEmptyRows(t *testing.T) {
	layout := newTroxImportLayout()

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
		t.Fatal("expected Trox header to be accepted")
	}

	if !layout.IsRowEmpty(make([]string, 14)) {
		t.Fatal("expected empty row to be considered empty")
	}

	if layout.IsRowEmpty([]string{"477139"}) {
		t.Fatal("expected row with content to not be considered empty")
	}
}

func TestTroxImportLayoutShouldRejectInvalidRows(t *testing.T) {
	layout := newTroxImportLayout()

	_, err := layout.ParseNormalizedRow(2, []string{
		"477139",
		"ABC",
		"09/06/2026",
		"Consulta de preco",
		"Informado",
		"Contato",
		"FILTROS",
		"BR1007854",
		"Cliente X",
		"Obra X",
		"DECK - VENDEDOR",
		"Instalador X",
		"65515.83",
	})
	if err == nil {
		t.Fatal("expected invalid row parse error")
	}
	if err.Error() != "Revisao invalida." {
		t.Fatalf("expected invalid revision error, got %v", err)
	}
}
