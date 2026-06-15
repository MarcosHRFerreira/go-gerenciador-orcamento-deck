package budgetimport

import (
	"fmt"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
)

type importLayout interface {
	Key() string
	Name() string
	SourceCompany() string
	Description() string
	SheetName() string
	HeaderRowNumber() int
	PreviewWarnings() []dto.BudgetImportPreviewMessage
	FieldGroups() []dto.BudgetImportPreviewFieldGroup
	Governance() dto.BudgetImportPreviewGovernance
	HasExpectedHeader(header []string) bool
	IsRowEmpty(rowValues []string) bool
	ParseNormalizedRow(rowNumber int, rowValues []string) (normalizedBudgetImportRow, error)
}

var registeredImportLayouts = []importLayout{
	newRocktecImportLayout(),
	newTroxImportLayout(),
}

func resolveImportLayout(fileData []byte) (importLayout, *workbookData, error) {
	if len(registeredImportLayouts) == 1 {
		layout := registeredImportLayouts[0]
		workbook, err := loadWorkbookForLayout(fileData, layout)
		if err != nil {
			return nil, nil, err
		}

		return layout, workbook, nil
	}

	for _, layout := range registeredImportLayouts {
		workbook, err := loadWorkbookForLayout(fileData, layout)
		if err != nil {
			continue
		}

		return layout, workbook, nil
	}

	return nil, nil, apperror.BadRequest("Nenhum layout de importacao compativel foi identificado no arquivo xlsx")
}

func loadWorkbookForLayout(fileData []byte, layout importLayout) (*workbookData, error) {
	workbook, err := parseWorkbook(fileData, layout.SheetName())
	if err != nil {
		return nil, err
	}

	if !layout.HasExpectedHeader(workbook.rows[layout.HeaderRowNumber()]) {
		return nil, apperror.BadRequest(fmt.Sprintf("Cabecalho invalido na aba %s", layout.SheetName()))
	}

	return workbook, nil
}

func findImportLayoutByKey(layoutKey string) (importLayout, bool) {
	for _, layout := range registeredImportLayouts {
		if layout.Key() == layoutKey {
			return layout, true
		}
	}

	return nil, false
}

func clonePreviewWarnings(messages []dto.BudgetImportPreviewMessage) []dto.BudgetImportPreviewMessage {
	return append([]dto.BudgetImportPreviewMessage{}, messages...)
}

func clonePreviewFieldGroups(
	fieldGroups []dto.BudgetImportPreviewFieldGroup,
) []dto.BudgetImportPreviewFieldGroup {
	cloned := make([]dto.BudgetImportPreviewFieldGroup, 0, len(fieldGroups))
	for _, fieldGroup := range fieldGroups {
		cloned = append(cloned, dto.BudgetImportPreviewFieldGroup{
			Key:         fieldGroup.Key,
			Title:       fieldGroup.Title,
			Description: fieldGroup.Description,
			Fields:      append([]string{}, fieldGroup.Fields...),
		})
	}

	return cloned
}

func clonePreviewGovernance(
	governance dto.BudgetImportPreviewGovernance,
) dto.BudgetImportPreviewGovernance {
	return dto.BudgetImportPreviewGovernance{
		DuplicateScope:      governance.DuplicateScope,
		DuplicatePolicy:     governance.DuplicatePolicy,
		MissingValuePolicy:  governance.MissingValuePolicy,
		DefaultCatalogs:     append([]string{}, governance.DefaultCatalogs...),
		LegacyMatchingScope: governance.LegacyMatchingScope,
	}
}
