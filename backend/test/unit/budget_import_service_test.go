package unit

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetimportservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetimport"
)

type budgetImportBudgetRepositoryStub struct {
	exists             bool
	err                error
	getItem            *model.BudgetModel
	createID           int64
	createErr          error
	updateErr          error
	capturedCreateItem *model.BudgetModel
	capturedUpdateItem *model.BudgetModel
}

func (s *budgetImportBudgetRepositoryStub) Create(_ context.Context, item *model.BudgetModel) (int64, error) {
	s.capturedCreateItem = item
	if s.createID == 0 {
		s.createID = 1
	}
	return s.createID, s.createErr
}

func (s *budgetImportBudgetRepositoryStub) ExistsByNumberAndYear(_ context.Context, _ string, _ int) (bool, error) {
	return s.exists, s.err
}

func (s *budgetImportBudgetRepositoryStub) GetByNumberAndYear(_ context.Context, _ string, _ int) (*model.BudgetModel, error) {
	return s.getItem, s.err
}

func (s *budgetImportBudgetRepositoryStub) Update(_ context.Context, item *model.BudgetModel) error {
	s.capturedUpdateItem = item
	return s.updateErr
}

type budgetImportStatusRepositoryStub struct {
	items []model.BudgetStatusModel
}

func (s *budgetImportStatusRepositoryStub) Create(_ context.Context, status *model.BudgetStatusModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.BudgetStatusModel{
		ID:   id,
		Code: status.Code,
		Name: status.Name,
	})
	return id, nil
}

func (s *budgetImportStatusRepositoryStub) GetByCodeOrName(_ context.Context, code string, name string) (*model.BudgetStatusModel, error) {
	for _, item := range s.items {
		if item.Code == code || item.Name == name {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *budgetImportStatusRepositoryStub) List(_ context.Context) ([]model.BudgetStatusModel, error) {
	return s.items, nil
}

type budgetImportPriorityRepositoryStub struct {
	items []model.PriorityModel
}

func (s *budgetImportPriorityRepositoryStub) Create(_ context.Context, priority *model.PriorityModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.PriorityModel{
		ID:   id,
		Code: priority.Code,
		Name: priority.Name,
	})
	return id, nil
}

func (s *budgetImportPriorityRepositoryStub) GetByCodeOrName(_ context.Context, code string, name string) (*model.PriorityModel, error) {
	for _, item := range s.items {
		if item.Code == code || item.Name == name {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *budgetImportPriorityRepositoryStub) List(_ context.Context) ([]model.PriorityModel, error) {
	return s.items, nil
}

type budgetImportInstallerRepositoryStub struct {
	items []model.InstallerModel
}

func (s *budgetImportInstallerRepositoryStub) Create(_ context.Context, installer *model.InstallerModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.InstallerModel{
		ID:   id,
		Name: installer.Name,
	})
	return id, nil
}

func (s *budgetImportInstallerRepositoryStub) List(_ context.Context) ([]model.InstallerModel, error) {
	return s.items, nil
}

type budgetImportProjectRepositoryStub struct {
	items []model.ProjectModel
}

func (s *budgetImportProjectRepositoryStub) Create(_ context.Context, item *model.ProjectModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.ProjectModel{
		ID:   id,
		Name: item.Name,
	})
	return id, nil
}

func (s *budgetImportProjectRepositoryStub) List(_ context.Context) ([]model.ProjectModel, error) {
	return s.items, nil
}

type budgetImportProjectTypeRepositoryStub struct {
	items []model.ProjectTypeModel
}

func (s *budgetImportProjectTypeRepositoryStub) Create(_ context.Context, item *model.ProjectTypeModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.ProjectTypeModel{
		ID:   id,
		Code: item.Code,
		Name: item.Name,
	})
	return id, nil
}

func (s *budgetImportProjectTypeRepositoryStub) GetByCodeOrName(_ context.Context, code string, name string) (*model.ProjectTypeModel, error) {
	for _, item := range s.items {
		if item.Code == code || item.Name == name {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *budgetImportProjectTypeRepositoryStub) List(_ context.Context) ([]model.ProjectTypeModel, error) {
	return s.items, nil
}

type budgetImportSalespersonRepositoryStub struct {
	items []model.SalespersonModel
}

func (s *budgetImportSalespersonRepositoryStub) Create(_ context.Context, salesperson *model.SalespersonModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.SalespersonModel{
		ID:   id,
		Name: salesperson.Name,
	})
	return id, nil
}

func (s *budgetImportSalespersonRepositoryStub) List(_ context.Context) ([]model.SalespersonModel, error) {
	return s.items, nil
}

type budgetImportContactRepositoryStub struct {
	items []model.ContactModel
}

func (s *budgetImportContactRepositoryStub) Create(_ context.Context, contact *model.ContactModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.ContactModel{
		ID:          id,
		InstallerID: contact.InstallerID,
		Name:        contact.Name,
	})
	return id, nil
}

func (s *budgetImportContactRepositoryStub) List(_ context.Context, _ *int64) ([]model.ContactModel, error) {
	return s.items, nil
}

type budgetImportLossReasonRepositoryStub struct {
	items []model.LossReasonModel
}

func (s *budgetImportLossReasonRepositoryStub) Create(_ context.Context, reason *model.LossReasonModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.LossReasonModel{
		ID:   id,
		Code: reason.Code,
		Name: reason.Name,
	})
	return id, nil
}

func (s *budgetImportLossReasonRepositoryStub) GetByCodeOrName(_ context.Context, code string, name string) (*model.LossReasonModel, error) {
	for _, item := range s.items {
		if item.Code == code || item.Name == name {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *budgetImportLossReasonRepositoryStub) List(_ context.Context) ([]model.LossReasonModel, error) {
	return s.items, nil
}

func TestBudgetImportPreviewShouldParseWorkbookAndSummarizeCatalogActions(t *testing.T) {
	service := budgetimportservice.NewService(
		&budgetImportBudgetRepositoryStub{},
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{Name: "FECHADO"},
			},
		},
		&budgetImportPriorityRepositoryStub{},
		&budgetImportInstallerRepositoryStub{},
		&budgetImportProjectRepositoryStub{},
		&budgetImportProjectTypeRepositoryStub{},
		&budgetImportSalespersonRepositoryStub{},
		&budgetImportContactRepositoryStub{},
		&budgetImportLossReasonRepositoryStub{},
	)

	response, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45660",
				"B": "1001",
				"C": "R2",
				"D": "Instalador A",
				"E": "Obra XPTO",
				"F": "Industrial",
				"G": "Vendedor A",
				"H": "Contato A",
				"I": "1234.56",
				"J": "0.05",
				"K": "10",
				"L": "FECHADO",
				"M": "Negociando",
				"N": "-",
				"O": "-",
				"P": "1000",
				"Q": "-",
				"R": "-",
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.SheetName != "ORCAMENTOS" {
		t.Fatalf("expected sheet ORCAMENTOS, got %s", response.SheetName)
	}
	if response.Summary.RowsRead != 1 {
		t.Fatalf("expected rows_read 1, got %d", response.Summary.RowsRead)
	}
	if response.Summary.RowsValid != 1 {
		t.Fatalf("expected rows_valid 1, got %d", response.Summary.RowsValid)
	}
	if response.Summary.NewBudgets != 1 {
		t.Fatalf("expected new budgets 1, got %d", response.Summary.NewBudgets)
	}
	if response.CatalogActions.InstallersToCreate < 1 {
		t.Fatalf("expected at least one installer to create, got %d", response.CatalogActions.InstallersToCreate)
	}
	if response.CatalogActions.ProjectsToCreate < 1 {
		t.Fatalf("expected at least one project to create, got %d", response.CatalogActions.ProjectsToCreate)
	}
	if response.CatalogActions.ProjectTypesToCreate < 1 {
		t.Fatalf("expected at least one project type to create, got %d", response.CatalogActions.ProjectTypesToCreate)
	}
	if response.CatalogActions.SalespeopleToCreate < 1 {
		t.Fatalf("expected at least one salesperson to create, got %d", response.CatalogActions.SalespeopleToCreate)
	}
	if response.CatalogActions.LossReasonsToCreate < 1 {
		t.Fatalf("expected at least one loss reason to create, got %d", response.CatalogActions.LossReasonsToCreate)
	}
	if response.CatalogActions.PrioritiesToCreate < 1 {
		t.Fatalf("expected at least one priority to create, got %d", response.CatalogActions.PrioritiesToCreate)
	}
	if len(response.SampleRows) != 1 {
		t.Fatalf("expected one sample row, got %d", len(response.SampleRows))
	}
	if response.SampleRows[0].Action != "create" {
		t.Fatalf("expected sample row action create, got %s", response.SampleRows[0].Action)
	}
	if response.SampleRows[0].Status != "warning" {
		t.Fatalf("expected sample row status warning, got %s", response.SampleRows[0].Status)
	}
}

func TestBudgetImportPreviewShouldMarkExistingBudgetAsIgnore(t *testing.T) {
	service := budgetimportservice.NewService(
		&budgetImportBudgetRepositoryStub{exists: true},
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{Name: "FECHADO"},
				{Name: "Nao informado"},
			},
		},
		&budgetImportPriorityRepositoryStub{
			items: []model.PriorityModel{
				{Name: "Nao informado"},
			},
		},
		&budgetImportInstallerRepositoryStub{
			items: []model.InstallerModel{
				{ID: 1, Name: "Instalador A"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportProjectRepositoryStub{
			items: []model.ProjectModel{
				{Name: "Obra XPTO"},
				{Name: "Nao informado"},
			},
		},
		&budgetImportProjectTypeRepositoryStub{
			items: []model.ProjectTypeModel{
				{Name: "Industrial"},
				{Name: "Nao informado"},
			},
		},
		&budgetImportSalespersonRepositoryStub{
			items: []model.SalespersonModel{
				{Name: "Vendedor A"},
				{Name: "Nao informado"},
			},
		},
		&budgetImportContactRepositoryStub{
			items: []model.ContactModel{
				{InstallerID: 1, Name: "Contato A"},
				{InstallerID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportLossReasonRepositoryStub{
			items: []model.LossReasonModel{
				{Name: "Nao informado"},
			},
		},
	)

	response, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45660",
				"B": "1001",
				"C": "R1",
				"D": "Instalador A",
				"E": "Obra XPTO",
				"F": "Industrial",
				"G": "Vendedor A",
				"H": "Contato A",
				"I": "1234.56",
				"J": "0.05",
				"K": "10",
				"L": "FECHADO",
				"M": "Negociando",
				"N": "Concorrente X",
				"O": "-",
				"P": "1000",
				"Q": "Projetista Y",
				"R": "Detalhe tecnico",
			},
		}),
		dto.PreviewBudgetImportOptions{
			DuplicateStrategy:     "ignore",
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Summary.ExistingBudgets != 1 {
		t.Fatalf("expected one existing budget, got %d", response.Summary.ExistingBudgets)
	}
	if response.SampleRows[0].Action != "ignore" {
		t.Fatalf("expected ignore action, got %s", response.SampleRows[0].Action)
	}
}

func TestBudgetImportPreviewShouldExposeInconsistencyRows(t *testing.T) {
	service := budgetimportservice.NewService(
		&budgetImportBudgetRepositoryStub{},
		&budgetImportStatusRepositoryStub{},
		&budgetImportPriorityRepositoryStub{},
		&budgetImportInstallerRepositoryStub{},
		&budgetImportProjectRepositoryStub{},
		&budgetImportProjectTypeRepositoryStub{},
		&budgetImportSalespersonRepositoryStub{},
		&budgetImportContactRepositoryStub{},
		&budgetImportLossReasonRepositoryStub{},
	)

	response, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45660",
				"B": "",
				"C": "R2",
				"D": "Instalador A",
				"E": "Obra XPTO",
				"F": "Industrial",
				"G": "Vendedor A",
				"H": "Contato A",
				"I": "1234.56",
				"J": "0.05",
				"K": "10",
				"L": "FECHADO",
				"M": "Negociando",
				"N": "-",
				"O": "-",
				"P": "1000",
				"Q": "-",
				"R": "-",
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Summary.RowsWithError != 1 {
		t.Fatalf("expected one row with error, got %d", response.Summary.RowsWithError)
	}
	if len(response.InconsistencyRows) != 1 {
		t.Fatalf("expected one inconsistency row, got %d", len(response.InconsistencyRows))
	}
	if response.InconsistencyRows[0].Status != "error" {
		t.Fatalf("expected inconsistency row status error, got %s", response.InconsistencyRows[0].Status)
	}
}

func TestBudgetImportExecuteShouldCreateBudgetFromStoredPreview(t *testing.T) {
	budgetRepo := &budgetImportBudgetRepositoryStub{}
	service := budgetimportservice.NewService(
		budgetRepo,
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{ID: 1, Name: "FECHADO"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportPriorityRepositoryStub{
			items: []model.PriorityModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportInstallerRepositoryStub{
			items: []model.InstallerModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportProjectRepositoryStub{
			items: []model.ProjectModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportProjectTypeRepositoryStub{
			items: []model.ProjectTypeModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportSalespersonRepositoryStub{
			items: []model.SalespersonModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportContactRepositoryStub{
			items: []model.ContactModel{
				{ID: 1, InstallerID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportLossReasonRepositoryStub{
			items: []model.LossReasonModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
	)

	previewResponse, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45660",
				"B": "1001",
				"C": "R1",
				"D": "-",
				"E": "-",
				"F": "-",
				"G": "-",
				"H": "-",
				"I": "1234.56",
				"J": "0.05",
				"K": "10",
				"L": "FECHADO",
				"M": "-",
				"N": "-",
				"O": "-",
				"P": "1000",
				"Q": "-",
				"R": "-",
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)
	if err != nil {
		t.Fatalf("expected preview without error, got %v", err)
	}

	importResponse, err := service.ExecuteImport(context.Background(), &dto.ExecuteBudgetImportRequest{
		PreviewID: previewResponse.PreviewID,
	})
	if err != nil {
		t.Fatalf("expected import without error, got %v", err)
	}
	if importResponse.Summary.BudgetsCreated != 1 {
		t.Fatalf("expected one created budget, got %d", importResponse.Summary.BudgetsCreated)
	}
	if budgetRepo.capturedCreateItem == nil {
		t.Fatal("expected captured create budget item")
	}
	if budgetRepo.capturedCreateItem.BudgetNumber != "1001" {
		t.Fatalf("expected budget number 1001, got %s", budgetRepo.capturedCreateItem.BudgetNumber)
	}
}

func buildImportWorkbook(t *testing.T, rowValues []map[string]string) []byte {
	t.Helper()

	headers := []string{
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
	}

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	writeZipFile(t, zipWriter, "[Content_Types].xml", contentTypesXML())
	writeZipFile(t, zipWriter, "_rels/.rels", rootRelationshipsXML())
	writeZipFile(t, zipWriter, "xl/workbook.xml", workbookXMLContent())
	writeZipFile(t, zipWriter, "xl/_rels/workbook.xml.rels", workbookRelationshipsXMLContent())
	writeZipFile(t, zipWriter, "xl/worksheets/sheet1.xml", buildSheetXML(headers, rowValues))

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buffer.Bytes()
}

func writeZipFile(t *testing.T, zipWriter *zip.Writer, name string, content string) {
	t.Helper()

	writer, err := zipWriter.Create(name)
	if err != nil {
		t.Fatalf("failed to create zip entry %s: %v", name, err)
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write zip entry %s: %v", name, err)
	}
}

func buildSheetXML(headers []string, rows []map[string]string) string {
	builder := strings.Builder{}
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	builder.WriteString(fmt.Sprintf(`<row r="%d">`, 10))
	for index, header := range headers {
		builder.WriteString(buildInlineStringCell(columnName(index+1)+"10", header))
	}
	builder.WriteString(`</row>`)

	for index, row := range rows {
		rowNumber := 11 + index
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowNumber))
		for columnIndex := 1; columnIndex <= len(headers); columnIndex++ {
			column := columnName(columnIndex)
			value, exists := row[column]
			if !exists {
				continue
			}
			if isNumericCell(value) {
				builder.WriteString(fmt.Sprintf(`<c r="%s%d"><v>%s</v></c>`, column, rowNumber, value))
				continue
			}
			builder.WriteString(buildInlineStringCell(fmt.Sprintf("%s%d", column, rowNumber), value))
		}
		builder.WriteString(`</row>`)
	}

	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func buildInlineStringCell(reference string, value string) string {
	escaped := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(value)
	return fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, reference, escaped)
}

func isNumericCell(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && char != '.' && char != '-' {
			return false
		}
	}
	return true
}

func columnName(index int) string {
	result := ""
	for index > 0 {
		index--
		result = string(rune('A'+(index%26))) + result
		index /= 26
	}
	return result
}

func contentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`
}

func rootRelationshipsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`
}

func workbookXMLContent() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="ORCAMENTOS" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`
}

func workbookRelationshipsXMLContent() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`
}
