package unit

import (
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/budgetpriority"
)

func TestBudgetPriorityResolveByGrossValueShouldReturnFaixa0a50k(t *testing.T) {
	definition := budgetpriority.ResolveByGrossValue(50000)

	if definition.Code != "faixa_0_a_50k" {
		t.Fatalf("expected faixa_0_a_50k, got %s", definition.Code)
	}
	if definition.Name != "Faixa 0 a 50k" {
		t.Fatalf("expected label Faixa 0 a 50k, got %s", definition.Name)
	}
}

func TestBudgetPriorityResolveByGrossValueShouldReturnFaixa50ka250k(t *testing.T) {
	definition := budgetpriority.ResolveByGrossValue(50000.01)

	if definition.Code != "faixa_50k_a_250k" {
		t.Fatalf("expected faixa_50k_a_250k, got %s", definition.Code)
	}
}

func TestBudgetPriorityResolveByGrossValueShouldReturnFaixaAcima250k(t *testing.T) {
	definition := budgetpriority.ResolveByGrossValue(250000.01)

	if definition.Code != "faixa_acima_de_250k" {
		t.Fatalf("expected faixa_acima_de_250k, got %s", definition.Code)
	}
}
