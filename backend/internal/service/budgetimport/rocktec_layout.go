package budgetimport

import (
	"fmt"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

const (
	rocktecImportLayoutKey       = "rocktec"
	rocktecImportSheetName       = "ORCAMENTOS"
	rocktecImportHeaderRowNumber = 10
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
	return "Layout legado da Rocktec com aba ORCAMENTOS e cabecalho na linha 10."
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
			Code:    "COMMISSION_INTERPRETATION_ASSUMED",
			Message: "A coluna COMISSAO foi tratada como valor numerico simples.",
		},
		{
			Code:    "COLUMN_MAPPING_ASSUMED",
			Message: "A coluna PRIORIDADE foi tratada como status catalogado e a coluna STATUS como follow-up atual.",
		},
	}
}

func (rocktecImportLayout) FieldGroups() []dto.BudgetImportPreviewFieldGroup {
	return []dto.BudgetImportPreviewFieldGroup{
		{
			Key:         "domain",
			Title:       "Campos do dominio principal",
			Description: "Entram diretamente no cadastro principal de orcamentos e catalogos relacionados.",
			Fields: []string{
				"Numero",
				"Ano",
				"Revisao",
				"Data",
				"Valor bruto",
				"Comissao",
				"Area m2",
				"Status",
				"Prioridade",
				"Instalador",
				"Obra",
				"Tipo de obra",
				"Vendedor",
				"Contato",
				"Motivo de perda",
				"Concorrente",
				"Preco concorrente",
				"Projetista",
				"Especificacao",
				"Follow-up atual",
			},
		},
		{
			Key:         "tracking",
			Title:       "Campos preservados para rastreabilidade",
			Description: "O lote e as linhas processadas ficam auditados no banco para consulta posterior.",
			Fields: []string{
				"Arquivo importado",
				"Aba",
				"Linha original",
				"Linha normalizada",
				"Mensagens do preview",
				"Origem Rocktec",
			},
		},
	}
}

func (rocktecImportLayout) Governance() dto.BudgetImportPreviewGovernance {
	return dto.BudgetImportPreviewGovernance{
		DuplicateScope:      "source_company + budget_number + year_budget",
		DuplicatePolicy:     "A Rocktec concilia duplicidade pela origem Rocktec, numero do orcamento e ano. Registros legados sem origem definida ainda podem ser conciliados para evitar duplicacao na migracao.",
		MissingValuePolicy:  "Campos ausentes podem usar o item padrao Nao informado quando a opcao correspondente estiver ativa no preview.",
		DefaultCatalogs:     []string{"Status", "Prioridade", "Instalador", "Obra", "Tipo de obra", "Vendedor", "Contato", "Motivo de perda"},
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
		rowNumber:    rowNumber,
		budgetNumber: normalizeCellText(getCell(rowValues, 1)),
		warnings:     []string{},
	}

	if row.budgetNumber == "" {
		return row, fmt.Errorf("Numero do orcamento nao informado.")
	}

	sentAt, err := parseExcelDate(getCell(rowValues, 0))
	if err != nil {
		return row, fmt.Errorf("Data do orcamento invalida.")
	}
	row.sentAt = sentAt
	row.yearBudget = sentAt.Year()

	row.grossValue, err = parseOptionalNumber(getCell(rowValues, 8), true)
	if err != nil {
		return row, fmt.Errorf("Valor bruto invalido.")
	}

	row.commissionValue, err = parseOptionalNumber(getCell(rowValues, 9), false)
	if err != nil {
		return row, fmt.Errorf("Comissao invalida.")
	}

	row.areaM2, err = parseOptionalNumber(getCell(rowValues, 10), false)
	if err != nil {
		return row, fmt.Errorf("M2 invalido.")
	}

	if !isMissingValue(getCell(rowValues, 15)) {
		value, valueErr := parseOptionalNumber(getCell(rowValues, 15), false)
		if valueErr != nil {
			return row, fmt.Errorf("Valor concorrente invalido.")
		}
		row.competitorPrice = &value
	}

	rawRevision := getCell(rowValues, 2)
	row.revision = extractRevision(rawRevision)
	if row.revision == 0 && !isMissingValue(rawRevision) {
		row.warnings = append(row.warnings, "Revisao nao reconhecida, sera usada revisao 0.")
	}

	if isMissingValue(getCell(rowValues, 12)) {
		row.warnings = append(row.warnings, "Follow-up atual nao informado, sera usado Nao informado.")
	}
	if isMissingValue(getCell(rowValues, 13)) {
		row.warnings = append(row.warnings, "Concorrente nao informado, sera usado Nao informado.")
	}
	if isMissingValue(getCell(rowValues, 16)) {
		row.warnings = append(row.warnings, "Projetista nao informado, sera usado Nao informado.")
	}
	if isMissingValue(getCell(rowValues, 17)) {
		row.warnings = append(row.warnings, "Especificacoes nao informadas, sera usado Nao informado.")
	}

	row.statusName = fallbackName(getCell(rowValues, 11))
	row.priorityName = notInformedName
	row.installerName = fallbackName(getCell(rowValues, 3))
	row.projectName = fallbackName(getCell(rowValues, 4))
	row.projectTypeName = fallbackName(getCell(rowValues, 5))
	row.salespersonName = fallbackName(getCell(rowValues, 6))
	row.contactName = fallbackName(getCell(rowValues, 7))
	row.lossReasonName = fallbackName(getCell(rowValues, 14))
	row.competitorName = fallbackName(getCell(rowValues, 13))
	row.projetistaName = fallbackName(getCell(rowValues, 16))
	row.specification = fallbackName(getCell(rowValues, 17))
	row.currentFollowUp = fallbackName(getCell(rowValues, 12))

	return row, nil
}

func rocktecHasExpectedHeader(header []string) bool {
	if len(header) < 18 {
		return false
	}

	first := normalizeLookupKey(getCell(header, 0))
	fourth := normalizeLookupKey(getCell(header, 3))
	fifth := normalizeLookupKey(getCell(header, 4))
	seventh := normalizeLookupKey(getCell(header, 6))
	eighth := normalizeLookupKey(getCell(header, 7))

	return first == "data" &&
		fourth != "" &&
		fifth != "" &&
		seventh != "" &&
		eighth != ""
}

func rocktecIsRowEmpty(values []string) bool {
	for column := 0; column < 18; column++ {
		if !isMissingValue(getCell(values, column)) {
			return false
		}
	}

	return true
}
