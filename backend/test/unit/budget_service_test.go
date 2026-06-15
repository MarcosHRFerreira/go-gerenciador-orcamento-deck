package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budget"
	"github.com/jackc/pgx/v5/pgconn"
)

type budgetRepositoryStub struct {
	createID                   int64
	createErr                  error
	changeStatusID             int64
	changeStatusErr            error
	listItems                  []model.BudgetModel
	listTotal                  int64
	listErr                    error
	existsByNumberAndYear      bool
	existsByNumberAndYearErr   error
	getByIDItem                *model.BudgetModel
	getByIDErr                 error
	updateErr                  error
	deleteErr                  error
	capturedCreateItem         *model.BudgetModel
	capturedChangeStatusParams *budgetrepository.ChangeStatusParams
	capturedUpdateItem         *model.BudgetModel
	capturedListFilters        *dto.ListBudgetsFilters
	capturedCurrentFollowUp    string
	capturedStatusID           int64
	existsByNumberAndYearCalls int
	updateCurrentFollowUpErr   error
	updateStatusErr            error
}

type salespersonRepositoryStub struct {
	getByUsernameItem *model.SalespersonModel
	getByUsernameErr  error
}

func (s *salespersonRepositoryStub) Create(_ context.Context, _ *model.SalespersonModel) (int64, error) {
	return 0, nil
}

func (s *salespersonRepositoryStub) List(_ context.Context) ([]model.SalespersonModel, error) {
	return nil, nil
}

func (s *salespersonRepositoryStub) GetByEmail(_ context.Context, _ string) (*model.SalespersonModel, error) {
	return nil, nil
}

func (s *salespersonRepositoryStub) GetByUsername(_ context.Context, _ string) (*model.SalespersonModel, error) {
	return s.getByUsernameItem, s.getByUsernameErr
}

func (s *salespersonRepositoryStub) GetByID(_ context.Context, _ int64) (*model.SalespersonModel, error) {
	return nil, nil
}

func (s *salespersonRepositoryStub) Update(_ context.Context, _ *model.SalespersonModel) error {
	return nil
}

func (s *salespersonRepositoryStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func (s *budgetRepositoryStub) Create(_ context.Context, item *model.BudgetModel) (int64, error) {
	s.capturedCreateItem = item
	return s.createID, s.createErr
}

func (s *budgetRepositoryStub) List(_ context.Context, filters *dto.ListBudgetsFilters) ([]model.BudgetModel, int64, error) {
	s.capturedListFilters = filters
	return s.listItems, s.listTotal, s.listErr
}

func (s *budgetRepositoryStub) ExistsByNumberAndYear(_ context.Context, _ string, _ int) (bool, error) {
	s.existsByNumberAndYearCalls++
	return s.existsByNumberAndYear, s.existsByNumberAndYearErr
}

func (s *budgetRepositoryStub) GetByNumberAndYear(_ context.Context, _ string, _ int) (*model.BudgetModel, error) {
	return nil, nil
}

func (s *budgetRepositoryStub) GetByID(_ context.Context, _ int64) (*model.BudgetModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *budgetRepositoryStub) GetByIDScoped(_ context.Context, _ int64, restrictedSalespersonID *int64) (*model.BudgetModel, error) {
	if s.capturedListFilters == nil {
		s.capturedListFilters = &dto.ListBudgetsFilters{}
	}
	s.capturedListFilters.RestrictedSalespersonID = restrictedSalespersonID
	return s.getByIDItem, s.getByIDErr
}

func (s *budgetRepositoryStub) Update(_ context.Context, item *model.BudgetModel) error {
	s.capturedUpdateItem = item
	return s.updateErr
}

func (s *budgetRepositoryStub) Delete(_ context.Context, _ int64) error {
	return s.deleteErr
}

func (s *budgetRepositoryStub) UpdateCurrentFollowUp(_ context.Context, _ int64, currentFollowUp string, _ time.Time) error {
	s.capturedCurrentFollowUp = currentFollowUp
	return s.updateCurrentFollowUpErr
}

func (s *budgetRepositoryStub) UpdateStatus(_ context.Context, _ int64, statusID int64, _ time.Time) error {
	s.capturedStatusID = statusID
	return s.updateStatusErr
}

func (s *budgetRepositoryStub) ChangeStatus(_ context.Context, params *budgetrepository.ChangeStatusParams) (int64, error) {
	s.capturedChangeStatusParams = params
	return s.changeStatusID, s.changeStatusErr
}

func TestBudgetServiceCreateShouldReturnBadRequestWhenBudgetNumberIsMissing(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &salespersonRepositoryStub{})

	_, err := service.Create(context.Background(), &dto.CreateBudgetRequest{})

	assertAppError(t, err, 400, "budget_number e obrigatorio")
}

func TestBudgetServiceCreateShouldReturnConflictWhenBudgetAlreadyExists(t *testing.T) {
	repo := &budgetRepositoryStub{
		existsByNumberAndYear: true,
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})

	_, err := service.Create(context.Background(), validCreateBudgetRequest())

	assertAppError(t, err, 409, "Ja existe um orcamento para o budget_number e year_budget informados")
}

func TestBudgetServiceCreateShouldTrimFieldsAndMapNullableValues(t *testing.T) {
	priorityID := int64(11)
	projectID := int64(22)
	competitorPrice := 890.5
	repo := &budgetRepositoryStub{
		createID: 77,
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})

	req := validCreateBudgetRequest()
	req.BudgetNumber = "  ORC-100  "
	req.PriorityID = &priorityID
	req.ProjectID = &projectID
	req.CompetitorPrice = &competitorPrice
	req.CompetitorName = "  Concorrente X  "
	req.DesignerName = "  Projetista Y  "
	req.SpecificationDetails = "  detalhe tecnico  "
	req.CurrentFollowUp = "  retorno agendado  "

	id, err := service.Create(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 77 {
		t.Fatalf("expected id 77, got %d", id)
	}
	if repo.capturedCreateItem == nil {
		t.Fatal("expected create item to be captured")
	}
	if repo.capturedCreateItem.BudgetNumber != "ORC-100" {
		t.Fatalf("expected trimmed budget number, got %s", repo.capturedCreateItem.BudgetNumber)
	}
	if repo.capturedCreateItem.CompetitorName != "Concorrente X" {
		t.Fatalf("expected trimmed competitor name, got %s", repo.capturedCreateItem.CompetitorName)
	}
	if repo.capturedCreateItem.DesignerName != "Projetista Y" {
		t.Fatalf("expected trimmed designer name, got %s", repo.capturedCreateItem.DesignerName)
	}
	if repo.capturedCreateItem.SpecificationDetails != "detalhe tecnico" {
		t.Fatalf("expected trimmed specification details, got %s", repo.capturedCreateItem.SpecificationDetails)
	}
	if repo.capturedCreateItem.CurrentFollowUp != "retorno agendado" {
		t.Fatalf("expected trimmed current follow up, got %s", repo.capturedCreateItem.CurrentFollowUp)
	}
	if !repo.capturedCreateItem.PriorityID.Valid || repo.capturedCreateItem.PriorityID.Int64 != priorityID {
		t.Fatalf("expected priority id %d, got %+v", priorityID, repo.capturedCreateItem.PriorityID)
	}
	if !repo.capturedCreateItem.ProjectID.Valid || repo.capturedCreateItem.ProjectID.Int64 != projectID {
		t.Fatalf("expected project id %d, got %+v", projectID, repo.capturedCreateItem.ProjectID)
	}
	if !repo.capturedCreateItem.CompetitorPrice.Valid || repo.capturedCreateItem.CompetitorPrice.Float64 != competitorPrice {
		t.Fatalf("expected competitor price %.2f, got %+v", competitorPrice, repo.capturedCreateItem.CompetitorPrice)
	}
}

func TestBudgetServiceListShouldNormalizeFiltersAndReturnPaginatedResponse(t *testing.T) {
	repo := &budgetRepositoryStub{
		listItems: []model.BudgetModel{
			{
				ID:           9,
				BudgetNumber: "ORC-9",
				YearBudget:   2026,
				SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				GrossValue:   1500,
				StatusID:     1,
				ProjectName: sql.NullString{
					String: "Projeto Centro",
					Valid:  true,
				},
				SalespersonName: sql.NullString{
					String: "Guilherme",
					Valid:  true,
				},
				ContactName: sql.NullString{
					String: "Contato Centro",
					Valid:  true,
				},
			},
		},
		listTotal: 1,
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})
	yearBudget := 2026

	response, err := service.List(context.Background(), &dto.ListBudgetsFilters{
		BudgetNumber: "  ORC  ",
		ProjectName:  "  Centro  ",
		YearBudget:   &yearBudget,
	}, model.RoleAdmin, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Total != 1 {
		t.Fatalf("expected total 1, got %d", response.Total)
	}
	if response.Page != 1 {
		t.Fatalf("expected page 1, got %d", response.Page)
	}
	if response.PageSize != 20 {
		t.Fatalf("expected page size 20, got %d", response.PageSize)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Items[0].ProjectName == nil || *response.Items[0].ProjectName != "Projeto Centro" {
		t.Fatalf("expected project name Projeto Centro, got %v", response.Items[0].ProjectName)
	}
	if response.Items[0].SalespersonName == nil || *response.Items[0].SalespersonName != "Guilherme" {
		t.Fatalf("expected salesperson name Guilherme, got %v", response.Items[0].SalespersonName)
	}
	if response.Items[0].ContactName == nil || *response.Items[0].ContactName != "Contato Centro" {
		t.Fatalf("expected contact name Contato Centro, got %v", response.Items[0].ContactName)
	}
	if repo.capturedListFilters == nil {
		t.Fatal("expected filters to be captured")
	}
	if repo.capturedListFilters.BudgetNumber != "ORC" {
		t.Fatalf("expected trimmed budget number filter, got %s", repo.capturedListFilters.BudgetNumber)
	}
	if repo.capturedListFilters.ProjectName != "Centro" {
		t.Fatalf("expected trimmed project name filter, got %s", repo.capturedListFilters.ProjectName)
	}
	if repo.capturedListFilters.SortBy != "sent_at" {
		t.Fatalf("expected default sort by sent_at, got %s", repo.capturedListFilters.SortBy)
	}
	if repo.capturedListFilters.SortOrder != "desc" {
		t.Fatalf("expected default sort order desc, got %s", repo.capturedListFilters.SortOrder)
	}
}

func TestBudgetServiceListShouldReturnBadRequestWhenPageSizeIsTooLarge(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &salespersonRepositoryStub{})

	_, err := service.List(context.Background(), &dto.ListBudgetsFilters{
		PageSize: 101,
	}, model.RoleAdmin, "")

	assertAppError(t, err, 400, "page_size nao pode ser maior que 100")
}

func TestBudgetServiceListShouldRestrictUserBySalespersonResolvedFromUsername(t *testing.T) {
	repo := &budgetRepositoryStub{}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 15, Active: true},
	})

	_, err := service.List(context.Background(), &dto.ListBudgetsFilters{}, model.RoleUser, "sales.alpha")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedListFilters == nil || repo.capturedListFilters.RestrictedSalespersonID == nil {
		t.Fatal("expected restricted salesperson id to be applied")
	}
	if *repo.capturedListFilters.RestrictedSalespersonID != 15 {
		t.Fatalf("expected restricted salesperson id 15, got %d", *repo.capturedListFilters.RestrictedSalespersonID)
	}
}

func TestBudgetServiceGetByIDShouldReturnNotFoundWhenBudgetDoesNotExist(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &salespersonRepositoryStub{})

	_, err := service.GetByID(context.Background(), 10, model.RoleAdmin, "")

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetServiceGetByIDShouldRestrictUserBySalespersonResolvedFromUsername(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{ID: 20, BudgetNumber: "ORC-020", YearBudget: 2026, SentAt: time.Now(), GrossValue: 1000, StatusID: 1},
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 18, Active: true},
	})

	_, err := service.GetByID(context.Background(), 20, model.RoleUser, "sales.beta")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedListFilters == nil || repo.capturedListFilters.RestrictedSalespersonID == nil {
		t.Fatal("expected restricted salesperson id to be applied")
	}
	if *repo.capturedListFilters.RestrictedSalespersonID != 18 {
		t.Fatalf("expected restricted salesperson id 18, got %d", *repo.capturedListFilters.RestrictedSalespersonID)
	}
}

func TestBudgetServiceUpdateShouldReturnConflictWhenNumberAndYearAlreadyExist(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:           5,
			BudgetNumber: "ORC-OLD",
			YearBudget:   2025,
			StatusID:     1,
		},
		existsByNumberAndYear: true,
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})

	err := service.Update(context.Background(), 5, &dto.UpdateBudgetRequest{
		BudgetNumber: "ORC-NEW",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	})

	assertAppError(t, err, 409, "Ja existe um orcamento para o budget_number e year_budget informados")
}

func TestBudgetServiceUpdateShouldSkipUniquenessCheckWhenNumberAndYearDoNotChange(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:           5,
			BudgetNumber: "ORC-100",
			YearBudget:   2026,
			StatusID:     1,
		},
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})

	err := service.Update(context.Background(), 5, &dto.UpdateBudgetRequest{
		BudgetNumber: "  ORC-100  ",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.existsByNumberAndYearCalls != 0 {
		t.Fatalf("expected uniqueness check to be skipped, got %d calls", repo.existsByNumberAndYearCalls)
	}
	if repo.capturedUpdateItem == nil {
		t.Fatal("expected update item to be captured")
	}
	if repo.capturedUpdateItem.BudgetNumber != "ORC-100" {
		t.Fatalf("expected trimmed budget number, got %s", repo.capturedUpdateItem.BudgetNumber)
	}
}

func TestBudgetServiceDeleteShouldReturnNotFoundWhenBudgetDoesNotExist(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &salespersonRepositoryStub{})

	err := service.Delete(context.Background(), 8)

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetServiceCreateShouldMapPersistenceErrorFromForeignKey(t *testing.T) {
	repo := &budgetRepositoryStub{
		createErr: &pgconn.PgError{ConstraintName: "fk_budgets_status_id"},
	}
	service := budgetservice.NewService(repo, &salespersonRepositoryStub{})

	_, err := service.Create(context.Background(), validCreateBudgetRequest())

	assertAppError(t, err, 400, "Status de orcamento nao encontrado")
}

func validCreateBudgetRequest() *dto.CreateBudgetRequest {
	return &dto.CreateBudgetRequest{
		BudgetNumber: "ORC-001",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	}
}

func assertAppError(t *testing.T, err error, expectedStatusCode int, expectedMessage string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apperror.StatusCode(err) != expectedStatusCode {
		t.Fatalf("expected status code %d, got %d", expectedStatusCode, apperror.StatusCode(err))
	}
	if err.Error() != expectedMessage {
		t.Fatalf("expected message %s, got %s", expectedMessage, err.Error())
	}
}
