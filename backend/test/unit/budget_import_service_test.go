package unit

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetimportservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetimport"
)

type budgetImportBudgetRepositoryStub struct {
	exists             bool
	existsBySource     bool
	err                error
	getItem            *model.BudgetModel
	getItemBySource    *model.BudgetModel
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

func (s *budgetImportBudgetRepositoryStub) ExistsBySourceAndNumberAndYear(_ context.Context, _ string, _ string, _ int) (bool, error) {
	return s.existsBySource, s.err
}

func (s *budgetImportBudgetRepositoryStub) GetByNumberAndYear(_ context.Context, _ string, _ int) (*model.BudgetModel, error) {
	return s.getItem, s.err
}

func (s *budgetImportBudgetRepositoryStub) GetBySourceAndNumberAndYear(_ context.Context, _ string, _ string, _ int) (*model.BudgetModel, error) {
	return s.getItemBySource, s.err
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

type budgetImportProductLineRepositoryStub struct {
	items []model.ProductLineModel
}

func (s *budgetImportProductLineRepositoryStub) Create(_ context.Context, item *model.ProductLineModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.ProductLineModel{
		ID:   id,
		Code: item.Code,
		Name: item.Name,
	})
	return id, nil
}

func (s *budgetImportProductLineRepositoryStub) GetByCodeOrName(_ context.Context, code string, name string) (*model.ProductLineModel, error) {
	for _, item := range s.items {
		if item.Code == code || item.Name == name {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (s *budgetImportProductLineRepositoryStub) List(_ context.Context) ([]model.ProductLineModel, error) {
	return s.items, nil
}

type budgetImportProjectRepositoryStub struct {
	items []model.ProjectModel
}

func (s *budgetImportProjectRepositoryStub) Create(_ context.Context, item *model.ProjectModel) (int64, error) {
	id := int64(len(s.items) + 1)
	s.items = append(s.items, model.ProjectModel{
		ID:   id,
		Code: item.Code,
		Name: item.Name,
	})
	return id, nil
}

func (s *budgetImportProjectRepositoryStub) GetNextCode(_ context.Context) (string, error) {
	return fmt.Sprintf("OBR-%06d", len(s.items)+1), nil
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

type budgetImportAuditRepositoryStub struct {
	createBatchID       int64
	capturedBatchCreate *model.BudgetImportBatchModel
	capturedBatchUpdate *model.BudgetImportBatchModel
	capturedRows        []*model.BudgetImportRowRawModel
	batchesByID         map[int64]*model.BudgetImportBatchModel
}

func (s *budgetImportAuditRepositoryStub) CreateBatch(_ context.Context, item *model.BudgetImportBatchModel) (int64, error) {
	s.capturedBatchCreate = item
	if s.createBatchID == 0 {
		s.createBatchID = 1
	}
	if s.batchesByID == nil {
		s.batchesByID = make(map[int64]*model.BudgetImportBatchModel)
	}
	batchCopy := *item
	batchCopy.ID = s.createBatchID
	s.batchesByID[s.createBatchID] = &batchCopy
	return s.createBatchID, nil
}

func (s *budgetImportAuditRepositoryStub) UpdateBatch(_ context.Context, item *model.BudgetImportBatchModel) error {
	s.capturedBatchUpdate = item
	if s.batchesByID == nil {
		s.batchesByID = make(map[int64]*model.BudgetImportBatchModel)
	}
	batchCopy := *item
	s.batchesByID[item.ID] = &batchCopy
	return nil
}

func (s *budgetImportAuditRepositoryStub) GetBatchByID(_ context.Context, batchID int64) (*model.BudgetImportBatchModel, error) {
	if s.batchesByID == nil {
		return nil, nil
	}

	item, exists := s.batchesByID[batchID]
	if !exists {
		return nil, nil
	}

	batchCopy := *item
	return &batchCopy, nil
}

func (s *budgetImportAuditRepositoryStub) CreateRowRaw(_ context.Context, item *model.BudgetImportRowRawModel) (int64, error) {
	s.capturedRows = append(s.capturedRows, item)
	return int64(len(s.capturedRows)), nil
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
		&budgetImportAuditRepositoryStub{},
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
	if response.SheetName != "Rocktec" {
		t.Fatalf("expected sheet Rocktec, got %s", response.SheetName)
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
	if response.Governance.DuplicateScope != "source_company + budget_number + year_budget" {
		t.Fatalf("expected governance duplicate scope, got %s", response.Governance.DuplicateScope)
	}
	if len(response.Governance.DefaultCatalogs) == 0 {
		t.Fatal("expected governance default catalogs")
	}
	if response.CatalogActions.InstallersToCreate < 1 {
		t.Fatalf("expected at least one installer to create, got %d", response.CatalogActions.InstallersToCreate)
	}
	if response.CatalogActions.ProjectsToCreate < 1 {
		t.Fatalf("expected at least one project to create, got %d", response.CatalogActions.ProjectsToCreate)
	}
	if response.CatalogActions.ProductLinesToCreate < 1 {
		t.Fatalf("expected at least one product line to create, got %d", response.CatalogActions.ProductLinesToCreate)
	}
	if response.CatalogActions.SalespeopleToCreate < 1 {
		t.Fatalf("expected at least one salesperson to create, got %d", response.CatalogActions.SalespeopleToCreate)
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
		&budgetImportBudgetRepositoryStub{
			exists:         true,
			existsBySource: true,
		},
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
		&budgetImportAuditRepositoryStub{},
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

func TestBudgetImportPreviewShouldReuseSalespersonWhenFirstNameAlreadyExists(t *testing.T) {
	service := budgetimportservice.NewService(
		&budgetImportBudgetRepositoryStub{},
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{Name: "Fechado"},
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
				{ID: 10, Name: "Marcos"},
				{ID: 12, Name: "Marcos Ferreira"},
				{ID: 11, Name: "Nao informado"},
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
		&budgetImportAuditRepositoryStub{},
	)

	response, err := service.Preview(
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
				"G": "MARCOS FERREIRA",
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

	if response.CatalogActions.SalespeopleToCreate != 0 {
		t.Fatalf("expected no salesperson creation when first name already exists, got %d", response.CatalogActions.SalespeopleToCreate)
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
		&budgetImportAuditRepositoryStub{},
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
		&budgetImportAuditRepositoryStub{},
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
	if budgetRepo.capturedCreateItem.ProjectID.Valid {
		t.Fatalf("expected budget project_id to remain null when obra is '-', got %d", budgetRepo.capturedCreateItem.ProjectID.Int64)
	}
}

func TestBudgetImportExecuteShouldNotAssociateProjectWhenObraIsNaoInformado(t *testing.T) {
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
		&budgetImportAuditRepositoryStub{},
	)

	previewResponse, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45661",
				"B": "1002",
				"C": "R1",
				"D": "-",
				"E": "Nao informado",
				"F": "Nao informado",
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

	_, err = service.ExecuteImport(context.Background(), &dto.ExecuteBudgetImportRequest{
		PreviewID: previewResponse.PreviewID,
	})
	if err != nil {
		t.Fatalf("expected import without error, got %v", err)
	}
	if budgetRepo.capturedCreateItem == nil {
		t.Fatal("expected captured create budget item")
	}
	if budgetRepo.capturedCreateItem.ProjectID.Valid {
		t.Fatalf("expected budget project_id to remain null when obra is 'Nao informado', got %d", budgetRepo.capturedCreateItem.ProjectID.Int64)
	}
}

func TestBudgetImportStartImportShouldProcessInBackgroundAndExposeStatus(t *testing.T) {
	budgetRepo := &budgetImportBudgetRepositoryStub{}
	auditRepo := &budgetImportAuditRepositoryStub{}
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
		auditRepo,
	)

	previewResponse, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45661",
				"B": "1002",
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

	startResponse, err := service.StartImport(context.Background(), &dto.ExecuteBudgetImportRequest{
		PreviewID: previewResponse.PreviewID,
	})
	if err != nil {
		t.Fatalf("expected async start without error, got %v", err)
	}
	if startResponse.Status != "processing" {
		t.Fatalf("expected processing status, got %s", startResponse.Status)
	}
	if startResponse.Summary.RowsExpected != 1 {
		t.Fatalf("expected one expected row, got %d", startResponse.Summary.RowsExpected)
	}

	var statusResponse *dto.ExecuteBudgetImportResponse
	for attempt := 0; attempt < 20; attempt++ {
		statusResponse, err = service.GetImportStatus(context.Background(), 1)
		if err != nil {
			t.Fatalf("expected import status without error, got %v", err)
		}
		if statusResponse != nil && statusResponse.Status != "processing" {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if statusResponse == nil {
		t.Fatal("expected import status response")
	}
	if statusResponse.Status != "completed" {
		t.Fatalf("expected completed status, got %s", statusResponse.Status)
	}
	if statusResponse.Summary.BudgetsCreated != 1 {
		t.Fatalf("expected one created budget, got %d", statusResponse.Summary.BudgetsCreated)
	}
	if budgetRepo.capturedCreateItem == nil {
		t.Fatal("expected background import to create budget")
	}
}

func TestBudgetImportExecuteShouldNormalizeCatalogNamesAndReuseSalespersonByFirstName(t *testing.T) {
	budgetRepo := &budgetImportBudgetRepositoryStub{}
	salespersonRepo := &budgetImportSalespersonRepositoryStub{
		items: []model.SalespersonModel{
			{ID: 7, Name: "Marcos"},
			{ID: 9, Name: "Marcos Ferreira"},
			{ID: 8, Name: "Nao informado"},
		},
	}
	projectRepo := &budgetImportProjectRepositoryStub{
		items: []model.ProjectModel{
			{ID: 1, Name: "Nao informado"},
		},
	}
	installerRepo := &budgetImportInstallerRepositoryStub{
		items: []model.InstallerModel{
			{ID: 1, Name: "Nao informado"},
		},
	}

	service := budgetimportservice.NewService(
		budgetRepo,
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{ID: 1, Name: "Fechado"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportPriorityRepositoryStub{
			items: []model.PriorityModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		installerRepo,
		projectRepo,
		&budgetImportProjectTypeRepositoryStub{
			items: []model.ProjectTypeModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		salespersonRepo,
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
		&budgetImportAuditRepositoryStub{},
	)

	previewResponse, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbook(t, []map[string]string{
			{
				"A": "45662",
				"B": "1003",
				"C": "R1",
				"D": "INSTALADOR ALFA",
				"E": "OBRA CENTRAL NORTE",
				"F": "-",
				"G": "MARCOS FERREIRA",
				"H": "-",
				"I": "1234.56",
				"J": "0.05",
				"K": "10",
				"L": "FECHADO",
				"M": "EM NEGOCIACAO",
				"N": "CONCORRENTE XPTO",
				"O": "-",
				"P": "1000",
				"Q": "PROJETISTA SUL",
				"R": "DETALHE TECNICO FINAL",
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

	_, err = service.ExecuteImport(context.Background(), &dto.ExecuteBudgetImportRequest{
		PreviewID: previewResponse.PreviewID,
	})
	if err != nil {
		t.Fatalf("expected import without error, got %v", err)
	}

	if budgetRepo.capturedCreateItem == nil {
		t.Fatal("expected created budget to be captured")
	}
	if !budgetRepo.capturedCreateItem.SalespersonID.Valid || budgetRepo.capturedCreateItem.SalespersonID.Int64 != 7 {
		t.Fatalf("expected salesperson id 7 to be reused by first name, got %+v", budgetRepo.capturedCreateItem.SalespersonID)
	}
	if budgetRepo.capturedCreateItem.ConstructionCompany != "Concorrente Xpto" {
		t.Fatalf("expected construction company Concorrente Xpto, got %s", budgetRepo.capturedCreateItem.ConstructionCompany)
	}
	if budgetRepo.capturedCreateItem.ProductLineID.Valid {
		t.Fatalf("expected product line id to remain null in create item, got %+v", budgetRepo.capturedCreateItem.ProductLineID)
	}
	if budgetRepo.capturedCreateItem.ProjetistaName != "Nao informado" {
		t.Fatalf("expected projetista name Nao informado, got %s", budgetRepo.capturedCreateItem.ProjetistaName)
	}
	if budgetRepo.capturedCreateItem.SpecificationDetails != "Nao informado" {
		t.Fatalf("expected specification Nao informado, got %s", budgetRepo.capturedCreateItem.SpecificationDetails)
	}
	if budgetRepo.capturedCreateItem.CurrentFollowUp != "Em Negociacao" {
		t.Fatalf("expected normalized current follow-up, got %s", budgetRepo.capturedCreateItem.CurrentFollowUp)
	}
	if len(salespersonRepo.items) != 3 {
		t.Fatalf("expected salesperson list to remain unchanged, got %d items", len(salespersonRepo.items))
	}
	if len(projectRepo.items) != 2 || projectRepo.items[1].Name != "Obra Central Norte" {
		t.Fatalf("expected normalized project creation, got %+v", projectRepo.items)
	}
	if projectRepo.items[1].Code != "OBR-000002" {
		t.Fatalf("expected imported project code OBR-000002, got %s", projectRepo.items[1].Code)
	}
	if len(installerRepo.items) != 2 || installerRepo.items[1].Name != "Instalador Alfa" {
		t.Fatalf("expected normalized installer creation, got %+v", installerRepo.items)
	}
}

func TestBudgetImportPreviewShouldRejectWorkbookWithoutRocktecSheet(t *testing.T) {
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
		&budgetImportAuditRepositoryStub{},
	)

	_, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbookWithOptions(t, importWorkbookFixtureOptions{
			sheetName: "Resumo",
			rowValues: []map[string]string{
				{
					"A": "45660",
					"B": "1001",
				},
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)

	if err == nil {
		t.Fatal("expected error when workbook does not contain Rocktec sheet")
	}
	if !strings.Contains(err.Error(), "Nenhum layout de importacao compativel foi identificado") {
		t.Fatalf("expected no compatible layout error, got %v", err)
	}
}

func TestBudgetImportPreviewShouldRejectWorkbookWithInvalidRocktecHeader(t *testing.T) {
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
		&budgetImportAuditRepositoryStub{},
	)

	_, err := service.Preview(
		context.Background(),
		"orcamentos.xlsx",
		buildImportWorkbookWithOptions(t, importWorkbookFixtureOptions{
			headers: []string{
				"Data",
				"Numero",
				"Revisao",
			},
			rowValues: []map[string]string{
				{
					"A": "45660",
					"B": "1001",
				},
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)

	if err == nil {
		t.Fatal("expected error when workbook header is invalid for Rocktec")
	}
	if !strings.Contains(err.Error(), "Nenhum layout de importacao compativel foi identificado") {
		t.Fatalf("expected no compatible layout error, got %v", err)
	}
}

func TestBudgetImportPreviewShouldDetectTroxLayout(t *testing.T) {
	productLineRepo := &budgetImportProductLineRepositoryStub{}
	service := budgetimportservice.NewService(
		&budgetImportBudgetRepositoryStub{
			exists:         true,
			existsBySource: false,
		},
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
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
				{Name: "Nao informado"},
			},
		},
		&budgetImportProjectRepositoryStub{
			items: []model.ProjectModel{
				{Name: "Nao informado"},
			},
		},
		&budgetImportProjectTypeRepositoryStub{
			items: []model.ProjectTypeModel{
				{Name: "Nao informado"},
			},
		},
		&budgetImportSalespersonRepositoryStub{
			items: []model.SalespersonModel{
				{Name: "EMANUEL FERRI"},
				{Name: "Nao informado"},
			},
		},
		&budgetImportContactRepositoryStub{
			items: []model.ContactModel{
				{Name: "Nao informado"},
			},
		},
		&budgetImportLossReasonRepositoryStub{
			items: []model.LossReasonModel{
				{Name: "Nao informado"},
			},
		},
		&budgetImportAuditRepositoryStub{},
		productLineRepo,
	)

	response, err := service.Preview(
		context.Background(),
		"trox.xlsx",
		buildTroxWorkbook(t, []map[string]string{
			{
				"A": "477139",
				"B": "0",
				"C": "09/06/2026",
				"D": "Consulta de preco",
				"E": "Informado",
				"F": "ELDER J. BONETTI",
				"G": "FILTROS",
				"H": "BR1007854",
				"I": "ABECON ENGENHARIA E CLIMATIZACAO LT",
				"J": "DIVERSOS DE JUNHO",
				"K": "DECK - EMANUEL FERRI",
				"L": "ABECON ENGENHARIA E CLIMATIZACAO LT",
				"M": "65515.83",
				"N": "0.8",
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)
	if err != nil {
		t.Fatalf("expected Trox preview without error, got %v", err)
	}
	if response.SheetName != "Capa" {
		t.Fatalf("expected sheet Capa, got %s", response.SheetName)
	}
	if response.HeaderRow != 1 {
		t.Fatalf("expected header row 1, got %d", response.HeaderRow)
	}
	if response.Layout.Key != "trox" {
		t.Fatalf("expected layout key trox, got %s", response.Layout.Key)
	}
	if response.Layout.SourceCompany != "Trox" {
		t.Fatalf("expected source company Trox, got %s", response.Layout.SourceCompany)
	}
	if len(response.FieldGroups) == 0 {
		t.Fatal("expected preview field groups for Trox layout")
	}
	if response.Governance.DuplicateScope != "source_company + budget_number + year_budget" {
		t.Fatalf("expected governance duplicate scope for Trox, got %s", response.Governance.DuplicateScope)
	}
	if response.Summary.RowsValid != 1 {
		t.Fatalf("expected one valid row, got %d", response.Summary.RowsValid)
	}
	if response.Summary.NewBudgets != 1 {
		t.Fatalf("expected one new budget because duplicate is source-aware, got %d", response.Summary.NewBudgets)
	}
	if response.Summary.ExistingBudgets != 0 {
		t.Fatalf("expected zero existing budgets because duplicate is source-aware, got %d", response.Summary.ExistingBudgets)
	}
	if response.CatalogActions.ProductLinesToCreate != 1 {
		t.Fatalf("expected one product line to create, got %d", response.CatalogActions.ProductLinesToCreate)
	}
	if response.SampleRows[0].Status != "warning" {
		t.Fatalf("expected warning row because Trox uses defaults, got %s", response.SampleRows[0].Status)
	}
}

func TestBudgetImportExecuteShouldCreateBudgetFromTroxPreview(t *testing.T) {
	productLineRepo := &budgetImportProductLineRepositoryStub{}
	budgetRepo := &budgetImportBudgetRepositoryStub{
		getItem: &model.BudgetModel{
			ID:            99,
			BudgetNumber:  "477139",
			YearBudget:    2026,
			SourceCompany: "Rocktec",
		},
	}
	auditRepo := &budgetImportAuditRepositoryStub{}
	service := budgetimportservice.NewService(
		budgetRepo,
		&budgetImportStatusRepositoryStub{
			items: []model.BudgetStatusModel{
				{ID: 1, Name: "Em Negociacao"},
			},
		},
		&budgetImportPriorityRepositoryStub{
			items: []model.PriorityModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportInstallerRepositoryStub{
			items: []model.InstallerModel{
				{ID: 1, Name: "ABECON ENGENHARIA E CLIMATIZACAO LT"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportProjectRepositoryStub{
			items: []model.ProjectModel{
				{ID: 1, Name: "DIVERSOS DE JUNHO"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportProjectTypeRepositoryStub{
			items: []model.ProjectTypeModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		&budgetImportSalespersonRepositoryStub{
			items: []model.SalespersonModel{
				{ID: 1, Name: "EMANUEL FERRI"},
				{ID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportContactRepositoryStub{
			items: []model.ContactModel{
				{ID: 1, InstallerID: 1, Name: "ELDER J. BONETTI"},
				{ID: 2, InstallerID: 2, Name: "Nao informado"},
			},
		},
		&budgetImportLossReasonRepositoryStub{
			items: []model.LossReasonModel{
				{ID: 1, Name: "Nao informado"},
			},
		},
		auditRepo,
		productLineRepo,
	)

	previewResponse, err := service.Preview(
		context.Background(),
		"trox.xlsx",
		buildTroxWorkbook(t, []map[string]string{
			{
				"A": "477139",
				"B": "0",
				"C": "09/06/2026",
				"D": "Consulta de preco",
				"E": "Informado",
				"F": "ELDER J. BONETTI",
				"G": "FILTROS",
				"H": "BR1007854",
				"I": "ABECON ENGENHARIA E CLIMATIZACAO LT",
				"J": "DIVERSOS DE JUNHO",
				"K": "DECK - EMANUEL FERRI",
				"L": "ABECON ENGENHARIA E CLIMATIZACAO LT",
				"M": "65515.83",
				"N": "0.8",
			},
		}),
		dto.PreviewBudgetImportOptions{
			CreateMissingCatalogs: true,
			UseDefaultNotInformed: true,
		},
	)
	if err != nil {
		t.Fatalf("expected Trox preview without error, got %v", err)
	}

	importResponse, err := service.ExecuteImport(context.Background(), &dto.ExecuteBudgetImportRequest{
		PreviewID: previewResponse.PreviewID,
	})
	if err != nil {
		t.Fatalf("expected Trox import without error, got %v", err)
	}
	if importResponse.Summary.BudgetsCreated != 1 {
		t.Fatalf("expected one created budget, got %d", importResponse.Summary.BudgetsCreated)
	}
	if budgetRepo.capturedCreateItem == nil {
		t.Fatal("expected captured create item for Trox import")
	}
	if budgetRepo.capturedCreateItem.BudgetNumber != "477139" {
		t.Fatalf("expected Trox budget number 477139, got %s", budgetRepo.capturedCreateItem.BudgetNumber)
	}
	if budgetRepo.capturedCreateItem.YearBudget != 2026 {
		t.Fatalf("expected Trox year budget 2026, got %d", budgetRepo.capturedCreateItem.YearBudget)
	}
	if budgetRepo.capturedCreateItem.CurrentFollowUp != "Informado" {
		t.Fatalf("expected current follow-up Informado, got %s", budgetRepo.capturedCreateItem.CurrentFollowUp)
	}
	if budgetRepo.capturedCreateItem.StatusID != 1 {
		t.Fatalf("expected status id 1 for Em Negociacao, got %d", budgetRepo.capturedCreateItem.StatusID)
	}
	if !budgetRepo.capturedCreateItem.ProjectID.Valid || budgetRepo.capturedCreateItem.ProjectID.Int64 != 1 {
		t.Fatalf("expected project id 1, got %+v", budgetRepo.capturedCreateItem.ProjectID)
	}
	if !budgetRepo.capturedCreateItem.ProductLineID.Valid || budgetRepo.capturedCreateItem.ProductLineID.Int64 != 1 {
		t.Fatalf("expected product line id 1, got %+v", budgetRepo.capturedCreateItem.ProductLineID)
	}
	if budgetRepo.capturedCreateItem.ConstructionCompany != "Abecon Engenharia E Climatizacao Lt" {
		t.Fatalf("expected construction company mapped from Nome Cliente, got %s", budgetRepo.capturedCreateItem.ConstructionCompany)
	}
	if budgetRepo.capturedCreateItem.SourceCompany != "Trox" {
		t.Fatalf("expected source company Trox, got %s", budgetRepo.capturedCreateItem.SourceCompany)
	}
	if budgetRepo.capturedCreateItem.SourceLayout != "trox" {
		t.Fatalf("expected source layout trox, got %s", budgetRepo.capturedCreateItem.SourceLayout)
	}
	if !budgetRepo.capturedCreateItem.ImportBatchID.Valid || budgetRepo.capturedCreateItem.ImportBatchID.Int64 != 1 {
		t.Fatalf("expected import batch id 1, got %+v", budgetRepo.capturedCreateItem.ImportBatchID)
	}
	if auditRepo.capturedBatchCreate == nil || auditRepo.capturedBatchCreate.SourceLayout != "trox" {
		t.Fatalf("expected captured Trox import batch, got %+v", auditRepo.capturedBatchCreate)
	}
	if auditRepo.capturedBatchUpdate == nil || auditRepo.capturedBatchUpdate.BudgetsCreated != 1 {
		t.Fatalf("expected batch update with one created budget, got %+v", auditRepo.capturedBatchUpdate)
	}
	if len(auditRepo.capturedRows) != 1 {
		t.Fatalf("expected one persisted raw import row, got %d", len(auditRepo.capturedRows))
	}
	if auditRepo.capturedRows[0].Action != "create" {
		t.Fatalf("expected raw import row action create, got %s", auditRepo.capturedRows[0].Action)
	}
	if len(productLineRepo.items) != 1 || productLineRepo.items[0].Name != "Filtros" {
		t.Fatalf("expected product line catalog to be created from Trox import, got %+v", productLineRepo.items)
	}
}

func buildImportWorkbook(t *testing.T, rowValues []map[string]string) []byte {
	t.Helper()

	normalizedRows := make([]map[string]string, 0, len(rowValues))
	for _, rowValue := range rowValues {
		normalizedRows = append(normalizedRows, normalizeRocktecFixtureRow(rowValue))
	}

	return buildImportWorkbookWithOptions(t, importWorkbookFixtureOptions{
		headers:   defaultStructuredImportWorkbookHeaders(),
		headerRow: 1,
		rowValues: normalizedRows,
		sheetName: "Rocktec",
	})
}

func normalizeRocktecFixtureRow(rowValue map[string]string) map[string]string {
	if len(rowValue) == 0 {
		return rowValue
	}

	if _, hasLegacyBudgetNumber := rowValue["B"]; !hasLegacyBudgetNumber {
		return rowValue
	}

	statusValue := rowValue["M"]
	if strings.TrimSpace(statusValue) == "" || statusValue == "-" {
		statusValue = rowValue["L"]
	}

	customerNameValue := rowValue["N"]
	if strings.TrimSpace(customerNameValue) == "" || customerNameValue == "-" {
		customerNameValue = rowValue["D"]
	}

	return map[string]string{
		"A": rowValue["B"],
		"B": rowValue["C"],
		"C": formatFixtureDateBR(rowValue["A"]),
		"D": fallbackFixtureValue(rowValue["F"], "Consulta de preco"),
		"E": fallbackFixtureValue(statusValue, "Nao informado"),
		"F": fallbackFixtureValue(rowValue["H"], "-"),
		"G": fallbackFixtureValue(rowValue["F"], "Nao informado"),
		"H": "BR-TESTE-001",
		"I": fallbackFixtureValue(customerNameValue, "Cliente teste"),
		"J": fallbackFixtureValue(rowValue["E"], "-"),
		"K": fallbackFixtureValue(rowValue["G"], "-"),
		"L": fallbackFixtureValue(rowValue["D"], "-"),
		"M": fallbackFixtureValue(rowValue["I"], "0"),
		"N": fallbackFixtureValue(rowValue["J"], "1"),
	}
}

func formatFixtureDateBR(raw string) string {
	if strings.Contains(raw, "/") {
		return raw
	}

	serialNumber, err := strconv.Atoi(strings.TrimSpace(raw))
	if err == nil {
		baseDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
		return baseDate.AddDate(0, 0, serialNumber).Format("02/01/2006")
	}

	return raw
}

func fallbackFixtureValue(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}

	return trimmedValue
}

func buildTroxWorkbook(t *testing.T, rowValues []map[string]string) []byte {
	t.Helper()

	return buildImportWorkbookWithOptions(t, importWorkbookFixtureOptions{
		headers: []string{
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
		},
		headerRow: 1,
		rowValues: rowValues,
		sheetName: "Capa",
	})
}

type importWorkbookFixtureOptions struct {
	headers   []string
	headerRow int
	rowValues []map[string]string
	sheetName string
}

func buildImportWorkbookWithOptions(t *testing.T, options importWorkbookFixtureOptions) []byte {
	t.Helper()

	headers := options.headers
	if len(headers) == 0 {
		headers = defaultImportWorkbookHeaders()
	}

	headerRow := options.headerRow
	if headerRow == 0 {
		headerRow = 1
	}

	sheetName := options.sheetName
	if sheetName == "" {
		sheetName = "Rocktec"
	}

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	writeZipFile(t, zipWriter, "[Content_Types].xml", contentTypesXML())
	writeZipFile(t, zipWriter, "_rels/.rels", rootRelationshipsXML())
	writeZipFile(t, zipWriter, "xl/workbook.xml", workbookXMLContent(sheetName))
	writeZipFile(t, zipWriter, "xl/_rels/workbook.xml.rels", workbookRelationshipsXMLContent())
	writeZipFile(t, zipWriter, "xl/worksheets/sheet1.xml", buildSheetXML(headers, headerRow, options.rowValues))

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

func defaultImportWorkbookHeaders() []string {
	return defaultStructuredImportWorkbookHeaders()
}

func defaultStructuredImportWorkbookHeaders() []string {
	return []string{
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
	}
}

func buildSheetXML(headers []string, headerRow int, rows []map[string]string) string {
	builder := strings.Builder{}
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	builder.WriteString(fmt.Sprintf(`<row r="%d">`, headerRow))
	for index, header := range headers {
		builder.WriteString(buildInlineStringCell(fmt.Sprintf("%s%d", columnName(index+1), headerRow), header))
	}
	builder.WriteString(`</row>`)

	for index, row := range rows {
		rowNumber := headerRow + 1 + index
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

func workbookXMLContent(sheetName string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="%s" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`, sheetName)
}

func workbookRelationshipsXMLContent() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`
}
