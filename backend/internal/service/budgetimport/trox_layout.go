package budgetimport

import (
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

const (
	troxImportLayoutKey       = "trox"
	troxImportSheetName       = "Capa"
	troxImportHeaderRowNumber = 1
)

type troxImportLayout struct{}

func newTroxImportLayout() importLayout {
	return troxImportLayout{}
}

func (troxImportLayout) Key() string {
	return troxImportLayoutKey
}

func (troxImportLayout) Name() string {
	return "Trox"
}

func (troxImportLayout) SourceCompany() string {
	return "Trox"
}

func (troxImportLayout) Description() string {
	return "Layout resumido da Trox com aba Capa, cabecalho na linha 1 e campos comerciais normalizados para o dominio atual."
}

func (troxImportLayout) SheetName() string {
	return troxImportSheetName
}

func (troxImportLayout) HeaderRowNumber() int {
	return troxImportHeaderRowNumber
}

func (troxImportLayout) PreviewWarnings() []dto.BudgetImportPreviewMessage {
	return []dto.BudgetImportPreviewMessage{
		{
			Code:    "TROX_STATUS_AS_FOLLOW_UP",
			Message: "A coluna Status da Trox foi tratada como follow-up atual, mantendo o status principal como Nao informado.",
		},
		{
			Code:    "TROX_ORIGIN_FIELDS_IGNORED",
			Message: "As colunas Tipo, Linha de produtos, Codigo Cliente, Nome Cliente e Fator Medio nao entram no dominio principal nesta fase.",
		},
	}
}

func (troxImportLayout) FieldGroups() []dto.BudgetImportPreviewFieldGroup {
	return []dto.BudgetImportPreviewFieldGroup{
		{
			Key:         "domain",
			Title:       "Campos do dominio principal",
			Description: "Entram no cadastro principal apos normalizacao do layout Trox.",
			Fields: []string{
				"Orcamento",
				"Revisao",
				"Data de Emissao",
				"Obra",
				"Vendedor",
				"Instalador",
				"Contato",
				"Total do orcamento",
				"Status como follow-up atual",
			},
		},
		{
			Key:         "tracking",
			Title:       "Campos preservados so para rastreabilidade",
			Description: "Ficam gravados nas linhas brutas/normalizadas do lote, sem entrar no dominio principal nesta fase.",
			Fields: []string{
				"Tipo",
				"Linha de produtos",
				"Codigo Cliente",
				"Nome Cliente",
				"Fator Medio",
			},
		},
		{
			Key:         "business_notes",
			Title:       "Regras aplicadas no parser",
			Description: "Transformacoes e convencoes especificas usadas para a Trox.",
			Fields: []string{
				"Prefixo DECK - removido do vendedor",
				"Status principal definido como Nao informado",
				"Status da Trox enviado para follow-up atual",
				"Tipo e Linha de produtos nao viram tipo de obra",
			},
		},
	}
}

func (troxImportLayout) Governance() dto.BudgetImportPreviewGovernance {
	return dto.BudgetImportPreviewGovernance{
		DuplicateScope:      "source_company + budget_number + year_budget",
		DuplicatePolicy:     "A Trox concilia duplicidade pela origem Trox, numero do orcamento e ano. Registros legados sem origem definida ainda podem ser conciliados para evitar duplicacao na migracao.",
		MissingValuePolicy:  "Campos sem aderencia ao dominio principal usam Nao informado ou seguem apenas para rastreabilidade, conforme o mapeamento aprovado da Trox.",
		DefaultCatalogs:     []string{"Status", "Prioridade", "Instalador", "Projeto", "Tipo de obra", "Vendedor", "Contato", "Motivo de perda"},
		LegacyMatchingScope: "Registros sem source_company continuam elegiveis como correspondencia legado durante a transicao.",
	}
}

func (troxImportLayout) HasExpectedHeader(header []string) bool {
	return troxHasExpectedHeader(header)
}

func (troxImportLayout) IsRowEmpty(rowValues []string) bool {
	return troxIsRowEmpty(rowValues)
}

func (troxImportLayout) ParseNormalizedRow(rowNumber int, rowValues []string) (normalizedBudgetImportRow, error) {
	row := normalizedBudgetImportRow{
		rowNumber: rowNumber,
		warnings:  []string{},
	}

	row.budgetNumber = normalizeCellText(getCell(rowValues, 0))
	if row.budgetNumber == "" {
		return row, fmt.Errorf("Numero do orcamento nao informado.")
	}

	rawRevision := getCell(rowValues, 1)
	if isMissingValue(rawRevision) {
		row.revision = 0
		row.warnings = append(row.warnings, "Revisao nao informada, sera usada revisao 0.")
	} else {
		row.revision = extractRevision(rawRevision)
		if row.revision == 0 && normalizeCellText(rawRevision) != "0" {
			return row, fmt.Errorf("Revisao invalida.")
		}
	}

	sentAt, err := parseDateBR(getCell(rowValues, 2))
	if err != nil {
		return row, fmt.Errorf("Data do orcamento invalida.")
	}
	row.sentAt = sentAt
	row.yearBudget = sentAt.Year()

	row.grossValue, err = parseOptionalNumber(getCell(rowValues, 12), true)
	if err != nil {
		return row, fmt.Errorf("Valor bruto invalido.")
	}

	row.statusName = notInformedName
	row.priorityName = notInformedName
	row.installerName = fallbackName(getCell(rowValues, 11))
	row.projectName = fallbackName(getCell(rowValues, 9))
	row.projectTypeName = notInformedName
	row.salespersonName = fallbackName(normalizeTroxSalespersonName(getCell(rowValues, 10)))
	row.contactName = fallbackName(getCell(rowValues, 5))
	row.lossReasonName = notInformedName
	row.competitorName = notInformedName
	row.designerName = notInformedName
	row.specification = notInformedName
	row.currentFollowUp = fallbackName(getCell(rowValues, 4))

	return row, nil
}

func troxHasExpectedHeader(header []string) bool {
	if len(header) < 14 {
		return false
	}

	return normalizeLookupKey(getCell(header, 0)) == "or\u00e7amento" &&
		normalizeLookupKey(getCell(header, 1)) == "revis\u00e3o" &&
		normalizeLookupKey(getCell(header, 2)) == "data de emiss\u00e3o" &&
		normalizeLookupKey(getCell(header, 9)) == "obra" &&
		normalizeLookupKey(getCell(header, 10)) == "vendedor" &&
		normalizeLookupKey(getCell(header, 11)) == "instalador" &&
		normalizeLookupKey(getCell(header, 12)) == "total do or\u00e7amento"
}

func troxIsRowEmpty(values []string) bool {
	for column := 0; column < 14; column++ {
		if !isMissingValue(getCell(values, column)) {
			return false
		}
	}

	return true
}

func parseDateBR(raw string) (time.Time, error) {
	normalized := normalizeCellText(raw)
	if normalized == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	parsed, err := time.Parse("02/01/2006", normalized)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}

func normalizeTroxSalespersonName(raw string) string {
	normalized := normalizeCellText(raw)
	if normalized == "" {
		return normalized
	}

	const prefix = "deck - "
	if strings.HasPrefix(strings.ToLower(normalized), prefix) {
		return normalizeCellText(normalized[len(prefix):])
	}

	return normalized
}
