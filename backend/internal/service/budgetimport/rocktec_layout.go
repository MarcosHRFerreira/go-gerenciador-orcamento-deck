package budgetimport

import (
	"fmt"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

const (
	rocktecImportLayoutKey       = "rocktec"
	rocktecImportSheetName       = "Rocktec"
	rocktecImportHeaderRowNumber = 1
)

type normalizedBudgetImportRow struct {
	rowNumber           int
	budgetNumber        string
	yearBudget          int
	revision            int
	sentAt              time.Time
	grossValue          float64
	commissionValue     float64
	areaM2              float64
	statusName          string
	priorityName        string
	installerName       string
	productLineName     string
	projectName         string
	projectTypeName     string
	salespersonName     string
	contactName         string
	lossReasonName      string
	constructionCompany string
	competitorName      string
	competitorPrice     *float64
	projetistaName      string
	specification       string
	currentFollowUp     string
	warnings            []string
}

type rocktecImportLayout struct{}

func newRocktecImportLayout() importLayout {
	return rocktecImportLayout{}
}

func (rocktecImportLayout) Key() string {
	return rocktecImportLayoutKey
}

func (rocktecImportLayout) Name() string {
	return "Rocktec"
}

func (rocktecImportLayout) SourceCompany() string {
	return "Rocktec"
}

func (rocktecImportLayout) Description() string {
	return "Layout atual da Rocktec com a mesma estrutura resumida da Trox, usando a aba Rocktec e cabecalho na linha 1."
}

func (rocktecImportLayout) SheetName() string {
	return rocktecImportSheetName
}

func (rocktecImportLayout) HeaderRowNumber() int {
	return rocktecImportHeaderRowNumber
}

func (rocktecImportLayout) PreviewWarnings() []dto.BudgetImportPreviewMessage {
	return []dto.BudgetImportPreviewMessage{
		{
			Code:    "ROCKTEC_STATUS_AS_FOLLOW_UP",
			Message: "A coluna Status da Rocktec passa a alimentar o follow-up atual, enquanto o status principal da importacao inicia como Em Negociacao.",
		},
		{
			Code:    "ROCKTEC_PRODUCT_LINE_AND_CUSTOMER_MAPPED",
			Message: "Linha de produtos passa a alimentar catalogo auxiliar e Nome Cliente passa a preencher Construtora na nova estrutura da Rocktec.",
		},
	}
}

func (rocktecImportLayout) FieldGroups() []dto.BudgetImportPreviewFieldGroup {
	return []dto.BudgetImportPreviewFieldGroup{
		{
			Key:         "domain",
			Title:       "Campos do dominio principal",
			Description: "Entram no cadastro principal apos normalizacao do novo layout da Rocktec.",
			Fields: []string{
				"Orcamento",
				"Revisao",
				"Data de Emissao",
				"Linha de produtos",
				"Construtora",
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
				"Codigo Cliente",
				"Fator Medio",
			},
		},
		{
			Key:         "business_notes",
			Title:       "Regras aplicadas no parser",
			Description: "Transformacoes e convencoes especificas usadas para a nova Rocktec.",
			Fields: []string{
				"Prefixo DECK - removido do vendedor",
				"Status principal definido como Em Negociacao",
				"Status da Rocktec enviado para follow-up atual",
				"Tipo e Linha de produtos nao viram tipo de obra",
			},
		},
	}
}

func (rocktecImportLayout) Governance() dto.BudgetImportPreviewGovernance {
	return dto.BudgetImportPreviewGovernance{
		DuplicateScope:      "source_company + budget_number + year_budget",
		DuplicatePolicy:     "A Rocktec concilia duplicidade pela origem Rocktec, numero do orcamento e ano. Registros legados sem origem definida ainda podem ser conciliados para evitar duplicacao na migracao.",
		MissingValuePolicy:  "Campos sem aderencia ao dominio principal usam Nao informado ou seguem apenas para rastreabilidade, conforme o mapeamento aprovado da nova Rocktec.",
		DefaultCatalogs:     []string{"Status", "Prioridade", "Instalador", "Linha de produto", "Obra", "Tipo de obra", "Vendedor", "Contato", "Motivo de perda"},
		LegacyMatchingScope: "Registros sem source_company continuam elegiveis como correspondencia legado durante a transicao.",
	}
}

func (rocktecImportLayout) HasExpectedHeader(header []string) bool {
	return rocktecHasExpectedHeader(header)
}

func (rocktecImportLayout) IsRowEmpty(rowValues []string) bool {
	return rocktecIsRowEmpty(rowValues)
}

func (rocktecImportLayout) ParseNormalizedRow(rowNumber int, rowValues []string) (normalizedBudgetImportRow, error) {
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

	row.statusName = "Em Negociacao"
	row.priorityName = notInformedName
	row.installerName = fallbackName(getCell(rowValues, 11))
	row.productLineName = normalizeDisplayText(getCell(rowValues, 6))
	row.projectName = fallbackName(getCell(rowValues, 9))
	row.projectTypeName = notInformedName
	row.salespersonName = fallbackName(normalizeTroxSalespersonName(getCell(rowValues, 10)))
	row.contactName = fallbackName(getCell(rowValues, 5))
	row.lossReasonName = notInformedName
	row.constructionCompany = fallbackName(getCell(rowValues, 8))
	row.competitorName = notInformedName
	row.projetistaName = notInformedName
	row.specification = notInformedName
	row.currentFollowUp = fallbackName(getCell(rowValues, 4))

	return row, nil
}

func rocktecHasExpectedHeader(header []string) bool {
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

func rocktecIsRowEmpty(values []string) bool {
	for column := 0; column < 14; column++ {
		if !isMissingValue(getCell(values, column)) {
			return false
		}
	}

	return true
}
