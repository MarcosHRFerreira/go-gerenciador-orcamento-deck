package budgetpriority

import "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"

type Definition struct {
	Code   string
	Name   string
	Weight int
}

var definitions = []Definition{
	{
		Code:   "faixa_0_a_50k",
		Name:   "Faixa 0 a 50k",
		Weight: 10,
	},
	{
		Code:   "faixa_50k_a_250k",
		Name:   "Faixa 50k a 250k",
		Weight: 20,
	},
	{
		Code:   "faixa_acima_de_250k",
		Name:   "Faixa acima de 250k",
		Weight: 30,
	},
}

func ResolveByGrossValue(grossValue float64) Definition {
	switch {
	case grossValue <= 50000:
		return definitions[0]
	case grossValue <= 250000:
		return definitions[1]
	default:
		return definitions[2]
	}
}

func Definitions() []Definition {
	result := make([]Definition, len(definitions))
	copy(result, definitions)

	return result
}

func ToPriorityModel(definition Definition) model.PriorityModel {
	return model.PriorityModel{
		Code:   definition.Code,
		Name:   definition.Name,
		Weight: definition.Weight,
	}
}
