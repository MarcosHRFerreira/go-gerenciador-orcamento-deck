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
	createID                            int64
	createErr                           error
	changeStatusID                      int64
	changeStatusErr                     error
	electProjectWinnerErr               error
	updateAndChangeStatusErr            error
	listItems                           []model.BudgetModel
	listTotal                           int64
	listErr                             error
	existsByNumberAndYear               bool
	existsByNumberAndYearErr            error
	getByIDItem                         *model.BudgetModel
	getByIDErr                          error
	updateErr                           error
	deleteErr                           error
	deletedBudgetID                     int64
	capturedCreateItem                  *model.BudgetModel
	capturedChangeStatusParams          *budgetrepository.ChangeStatusParams
	capturedElectProjectWinnerParams    *budgetrepository.ElectProjectWinnerParams
	capturedUpdateItem                  *model.BudgetModel
	capturedUpdateAndChangeStatusItem   *model.BudgetModel
	capturedUpdateAndChangeStatusParams *budgetrepository.ChangeStatusParams
	capturedListFilters                 *dto.ListBudgetsFilters
	capturedCurrentFollowUp             string
	capturedStatusID                    int64
	existsByNumberAndYearCalls          int
	updateCurrentFollowUpErr            error
	updateStatusErr                     error
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

func (s *budgetRepositoryStub) GetGrossValueRange(_ context.Context, _ *dto.ListBudgetsFilters) (*dto.BudgetGrossValueRangeResponse, error) {
	return &dto.BudgetGrossValueRangeResponse{}, nil
}

func (s *budgetRepositoryStub) ListDeliveryMonitor(_ context.Context, _ *dto.ListBudgetDeliveryMonitorFilters) ([]model.BudgetDeliveryMonitorModel, int64, *dto.BudgetDeliveryMonitorSummaryResponse, error) {
	return nil, 0, nil, nil
}

func (s *budgetRepositoryStub) ExistsByNumberAndYear(_ context.Context, _ string, _ int) (bool, error) {
	s.existsByNumberAndYearCalls++
	return s.existsByNumberAndYear, s.existsByNumberAndYearErr
}

func (s *budgetRepositoryStub) ExistsBySourceAndNumberAndYear(_ context.Context, _ string, _ string, _ int) (bool, error) {
	return false, nil
}

func (s *budgetRepositoryStub) GetByNumberAndYear(_ context.Context, _ string, _ int) (*model.BudgetModel, error) {
	return nil, nil
}

func (s *budgetRepositoryStub) GetBySourceAndNumberAndYear(_ context.Context, _ string, _ string, _ int) (*model.BudgetModel, error) {
	return nil, nil
}

func (s *budgetRepositoryStub) GetByID(_ context.Context, _ int64) (*model.BudgetModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *budgetRepositoryStub) GetByIDScoped(_ context.Context, _ int64, restrictedSalespersonID *int64, restrictedEstimatorID *int64) (*model.BudgetModel, error) {
	if s.capturedListFilters == nil {
		s.capturedListFilters = &dto.ListBudgetsFilters{}
	}
	s.capturedListFilters.RestrictedSalespersonID = restrictedSalespersonID
	s.capturedListFilters.RestrictedEstimatorID = restrictedEstimatorID
	return s.getByIDItem, s.getByIDErr
}

func (s *budgetRepositoryStub) Update(_ context.Context, item *model.BudgetModel) error {
	s.capturedUpdateItem = item
	return s.updateErr
}

func (s *budgetRepositoryStub) UpdateAndChangeStatus(_ context.Context, item *model.BudgetModel, params *budgetrepository.ChangeStatusParams) error {
	s.capturedUpdateAndChangeStatusItem = item
	s.capturedUpdateAndChangeStatusParams = params
	return s.updateAndChangeStatusErr
}

func (s *budgetRepositoryStub) Delete(_ context.Context, budgetID int64) error {
	s.deletedBudgetID = budgetID
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

func (s *budgetRepositoryStub) ElectProjectWinner(_ context.Context, params *budgetrepository.ElectProjectWinnerParams) error {
	s.capturedElectProjectWinnerParams = params
	return s.electProjectWinnerErr
}

func TestBudgetServiceCreateShouldReturnBadRequestWhenBudgetNumberIsMissing(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	_, err := service.Create(context.Background(), model.RoleAdmin, "", &dto.CreateBudgetRequest{})

	assertAppError(t, err, 400, "budget_number e obrigatorio")
}

func TestBudgetServiceCreateShouldReturnConflictWhenBudgetAlreadyExists(t *testing.T) {
	repo := &budgetRepositoryStub{
		existsByNumberAndYear: true,
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	_, err := service.Create(context.Background(), model.RoleAdmin, "", validCreateBudgetRequest())

	assertAppError(t, err, 409, "Ja existe um orcamento para o budget_number e year_budget informados")
}

func TestBudgetServiceCreateShouldTrimFieldsAndMapNullableValues(t *testing.T) {
	productLineID := int64(14)
	projectID := int64(22)
	competitorPrice := 890.5
	repo := &budgetRepositoryStub{
		createID: 77,
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	req := validCreateBudgetRequest()
	req.BudgetNumber = "  ORC-100  "
	req.ProductLineID = &productLineID
	req.ProjectID = &projectID
	req.ConstructionCompany = "  Construtora XPTO  "
	req.CompetitorPrice = &competitorPrice
	req.CompetitorName = "  Concorrente X  "
	req.ProjetistaName = "  Projetista Y  "
	req.SpecificationDetails = "  detalhe tecnico  "
	req.CurrentFollowUp = "  retorno agendado  "

	id, err := service.Create(context.Background(), model.RoleAdmin, "", req)

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
	if repo.capturedCreateItem.ProjetistaName != "Projetista Y" {
		t.Fatalf("expected trimmed projetista name, got %s", repo.capturedCreateItem.ProjetistaName)
	}
	if repo.capturedCreateItem.SpecificationDetails != "detalhe tecnico" {
		t.Fatalf("expected trimmed specification details, got %s", repo.capturedCreateItem.SpecificationDetails)
	}
	if repo.capturedCreateItem.CurrentFollowUp != "retorno agendado" {
		t.Fatalf("expected trimmed current follow up, got %s", repo.capturedCreateItem.CurrentFollowUp)
	}
	if !repo.capturedCreateItem.PriorityID.Valid || repo.capturedCreateItem.PriorityID.Int64 != 1 {
		t.Fatalf("expected derived priority id 1, got %+v", repo.capturedCreateItem.PriorityID)
	}
	if !repo.capturedCreateItem.ProductLineID.Valid || repo.capturedCreateItem.ProductLineID.Int64 != productLineID {
		t.Fatalf("expected product line id %d, got %+v", productLineID, repo.capturedCreateItem.ProductLineID)
	}
	if !repo.capturedCreateItem.ProjectID.Valid || repo.capturedCreateItem.ProjectID.Int64 != projectID {
		t.Fatalf("expected project id %d, got %+v", projectID, repo.capturedCreateItem.ProjectID)
	}
	if !repo.capturedCreateItem.CompetitorPrice.Valid || repo.capturedCreateItem.CompetitorPrice.Float64 != competitorPrice {
		t.Fatalf("expected competitor price %.2f, got %+v", competitorPrice, repo.capturedCreateItem.CompetitorPrice)
	}
	if repo.capturedCreateItem.ConstructionCompany != "Construtora XPTO" {
		t.Fatalf("expected trimmed construction company, got %s", repo.capturedCreateItem.ConstructionCompany)
	}
}

func TestBudgetServiceCreateShouldAllowEstimatorUserToManageAssignmentsFreely(t *testing.T) {
	repo := &budgetRepositoryStub{
		createID: 55,
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       81,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{
		getByUserIDItem: &model.EstimatorModel{
			ID:     19,
			Active: true,
		},
	})

	req := validCreateBudgetRequest()
	salespersonID := int64(9)
	estimatorID := int64(21)
	req.SalespersonID = &salespersonID
	req.EstimatorID = &estimatorID

	id, err := service.Create(context.Background(), model.RoleUser, "estimator.user", req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 55 {
		t.Fatalf("expected id 55, got %d", id)
	}
	if repo.capturedCreateItem == nil {
		t.Fatal("expected create item to be captured")
	}
	if !repo.capturedCreateItem.EstimatorID.Valid || repo.capturedCreateItem.EstimatorID.Int64 != estimatorID {
		t.Fatalf("expected estimator id %d to be preserved, got %+v", estimatorID, repo.capturedCreateItem.EstimatorID)
	}
	if !repo.capturedCreateItem.SalespersonID.Valid || repo.capturedCreateItem.SalespersonID.Int64 != salespersonID {
		t.Fatalf("expected salesperson id %d to be preserved, got %+v", salespersonID, repo.capturedCreateItem.SalespersonID)
	}
}

func TestBudgetServiceCreateShouldRejectSalespersonUser(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       82,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{
			ID:     22,
			Active: true,
		},
	}, &estimatorRepositoryStub{})

	_, err := service.Create(context.Background(), model.RoleUser, "sales.user", validCreateBudgetRequest())

	assertAppError(t, err, 403, "Perfil comercial nao pode criar orcamentos")
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
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})
	yearBudget := 2026

	response, err := service.List(context.Background(), &dto.ListBudgetsFilters{
		BudgetNumber: "  ORC  ",
		ProjectCode:  "  OBR-000090  ",
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
	if response.PageSize != 50 {
		t.Fatalf("expected page size 50, got %d", response.PageSize)
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
	if response.Items[0].EstimatorName != nil {
		t.Fatalf("expected estimator name nil, got %v", response.Items[0].EstimatorName)
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
	if repo.capturedListFilters.ProjectCode != "OBR-000090" {
		t.Fatalf("expected trimmed project code filter, got %s", repo.capturedListFilters.ProjectCode)
	}
	if repo.capturedListFilters.SortBy != "sent_at" {
		t.Fatalf("expected default sort by sent_at, got %s", repo.capturedListFilters.SortBy)
	}
	if repo.capturedListFilters.SortOrder != "desc" {
		t.Fatalf("expected default sort order desc, got %s", repo.capturedListFilters.SortOrder)
	}
}

func TestBudgetServiceListShouldReturnBadRequestWhenPageSizeIsTooLarge(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	_, err := service.List(context.Background(), &dto.ListBudgetsFilters{
		PageSize: 101,
	}, model.RoleAdmin, "")

	assertAppError(t, err, 400, "page_size nao pode ser maior que 100")
}

func TestBudgetServiceListShouldRestrictUserBySalespersonResolvedFromUsername(t *testing.T) {
	repo := &budgetRepositoryStub{}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       91,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 15, Active: true},
	}, &estimatorRepositoryStub{})

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
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	_, err := service.GetByID(context.Background(), 10, model.RoleAdmin, "")

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetServiceGetByIDShouldRestrictUserBySalespersonResolvedFromUsername(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{ID: 20, BudgetNumber: "ORC-020", YearBudget: 2026, SentAt: time.Now(), GrossValue: 1000, StatusID: 1},
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       92,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 18, Active: true},
	}, &estimatorRepositoryStub{})

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
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	err := service.Update(context.Background(), 5, model.RoleAdmin, "", &dto.UpdateBudgetRequest{
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
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	err := service.Update(context.Background(), 5, model.RoleAdmin, "", &dto.UpdateBudgetRequest{
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

func TestBudgetServiceUpdateShouldUseStatusWorkflowWithoutProjectWinnerRule(t *testing.T) {
	projectID := int64(77)
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:        5,
			StatusID:  1,
			ProjectID: sql.NullInt64{Int64: projectID, Valid: true},
		},
	}
	service := budgetservice.NewService(
		repo,
		&budgetStatusRepositoryStub{
			getByIDItem: &model.BudgetStatusModel{ID: 2, Code: "PEDIDO", Name: "Pedido"},
		},
		&budgetImportPriorityRepositoryStub{},
		&userRepositoryStub{
			getUserByUsernameItem: &model.UserModel{ID: 40, Active: true},
		},
		&salespersonRepositoryStub{},
		&estimatorRepositoryStub{},
	)

	err := service.Update(context.Background(), 5, model.RoleAdmin, "admin.master", &dto.UpdateBudgetRequest{
		BudgetNumber: "ORC-100",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     2,
		ProjectID:    &projectID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedUpdateItem != nil {
		t.Fatal("expected plain update not to be used when status changes")
	}
	if repo.capturedUpdateAndChangeStatusItem == nil {
		t.Fatal("expected update and change status item to be captured")
	}
	if repo.capturedUpdateAndChangeStatusParams == nil {
		t.Fatal("expected update and change status params to be captured")
	}
	if repo.capturedUpdateAndChangeStatusParams.EnforceProjectWinnerRule {
		t.Fatal("expected project winner rule to stay disabled in common update flow")
	}
	if repo.capturedUpdateAndChangeStatusParams.UserID != 40 {
		t.Fatalf("expected user id 40, got %d", repo.capturedUpdateAndChangeStatusParams.UserID)
	}
}

func TestBudgetServiceElectProjectWinnerShouldEnsureStatusesAndCallRepository(t *testing.T) {
	projectID := int64(77)
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:        5,
			StatusID:  1,
			ProjectID: sql.NullInt64{Int64: projectID, Valid: true},
		},
	}
	statusRepo := &budgetStatusRepositoryStub{
		createID:            91,
		getByCodeOrNameItem: nil,
	}
	service := budgetservice.NewService(
		repo,
		statusRepo,
		&budgetImportPriorityRepositoryStub{},
		&userRepositoryStub{},
		&salespersonRepositoryStub{},
		&estimatorRepositoryStub{},
	)

	err := service.ElectProjectWinner(context.Background(), 5, 40, model.RoleAdmin, "admin.master", &dto.ElectBudgetWinnerRequest{
		Notes: "  Orcamento escolhido como vencedor  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedElectProjectWinnerParams == nil {
		t.Fatal("expected elect project winner params to be captured")
	}
	if repo.capturedElectProjectWinnerParams.BudgetID != 5 {
		t.Fatalf("expected budget id 5, got %d", repo.capturedElectProjectWinnerParams.BudgetID)
	}
	if repo.capturedElectProjectWinnerParams.UserID != 40 {
		t.Fatalf("expected user id 40, got %d", repo.capturedElectProjectWinnerParams.UserID)
	}
	if repo.capturedElectProjectWinnerParams.Notes != "Orcamento escolhido como vencedor" {
		t.Fatalf("expected trimmed notes, got %s", repo.capturedElectProjectWinnerParams.Notes)
	}
	if repo.capturedElectProjectWinnerParams.PedidoStatusID == 0 {
		t.Fatal("expected pedido status id to be resolved")
	}
	if repo.capturedElectProjectWinnerParams.CancelledStatusID == 0 {
		t.Fatal("expected cancelled status id to be resolved")
	}
	if statusRepo.createCalls != 2 {
		t.Fatalf("expected 2 required statuses to be created, got %d", statusRepo.createCalls)
	}
}

func TestBudgetServiceElectProjectWinnerShouldReturnConflictWhenProjectAlreadyHasPedido(t *testing.T) {
	projectID := int64(77)
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:        5,
			StatusID:  1,
			ProjectID: sql.NullInt64{Int64: projectID, Valid: true},
		},
		electProjectWinnerErr: budgetrepository.ErrProjectAlreadyHasPedido,
	}
	service := budgetservice.NewService(
		repo,
		&budgetStatusRepositoryStub{
			getByCodeOrNameItem: &model.BudgetStatusModel{ID: 3, Code: "PEDIDO", Name: "Pedido"},
		},
		&budgetImportPriorityRepositoryStub{},
		&userRepositoryStub{},
		&salespersonRepositoryStub{},
		&estimatorRepositoryStub{},
	)

	err := service.ElectProjectWinner(context.Background(), 5, 40, model.RoleAdmin, "admin.master", &dto.ElectBudgetWinnerRequest{})

	assertAppError(t, err, 409, "Ja existe outro orcamento da obra marcado como Fechado")
}

func TestBudgetServiceUpdateShouldRestrictUserBySalespersonResolvedFromUsername(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:           5,
			BudgetNumber: "ORC-100",
			YearBudget:   2026,
			StatusID:     1,
		},
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       93,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 27, Active: true},
	}, &estimatorRepositoryStub{})

	err := service.Update(context.Background(), 5, model.RoleUser, "sales.gamma", &dto.UpdateBudgetRequest{
		BudgetNumber: "ORC-100",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedListFilters == nil || repo.capturedListFilters.RestrictedSalespersonID == nil {
		t.Fatal("expected restricted salesperson id to be applied on update")
	}
	if *repo.capturedListFilters.RestrictedSalespersonID != 27 {
		t.Fatalf("expected restricted salesperson id 27, got %d", *repo.capturedListFilters.RestrictedSalespersonID)
	}
}

func TestBudgetServiceUpdateShouldReturnNotFoundWhenUserIsOutsideScope(t *testing.T) {
	repo := &budgetRepositoryStub{}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       94,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 19, Active: true},
	}, &estimatorRepositoryStub{})

	err := service.Update(context.Background(), 5, model.RoleUser, "sales.delta", &dto.UpdateBudgetRequest{
		BudgetNumber: "ORC-100",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	})

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetServiceDeleteShouldReturnNotFoundWhenBudgetDoesNotExist(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	err := service.Delete(context.Background(), 8, model.RoleAdmin, "admin")

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetServiceDeleteShouldAllowEstimatorUserOutsideOwnScope(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:           9,
			BudgetNumber: "ORC-200",
			YearBudget:   2026,
			StatusID:     1,
		},
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       77,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	err := service.Delete(context.Background(), 9, model.RoleUser, "estimator.gamma")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedBudgetID != 9 {
		t.Fatalf("expected deleted budget id 9, got %d", repo.deletedBudgetID)
	}
	if repo.capturedListFilters == nil {
		t.Fatal("expected scoped lookup filters to be captured on delete")
	}
	if repo.capturedListFilters.RestrictedEstimatorID != nil {
		t.Fatalf("expected estimator restriction to be nil on delete, got %v", *repo.capturedListFilters.RestrictedEstimatorID)
	}
}

func TestBudgetServiceDeleteShouldBlockSalespersonUser(t *testing.T) {
	service := budgetservice.NewService(&budgetRepositoryStub{}, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       88,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}, &salespersonRepositoryStub{
		getByUsernameItem: &model.SalespersonModel{ID: 31, Active: true},
	}, &estimatorRepositoryStub{})

	err := service.Delete(context.Background(), 9, model.RoleUser, "sales.delta")

	assertAppError(t, err, 403, "Perfil comercial nao pode excluir orcamentos")
}

func TestBudgetServiceCreateShouldMapPersistenceErrorFromForeignKey(t *testing.T) {
	repo := &budgetRepositoryStub{
		createErr: &pgconn.PgError{ConstraintName: "fk_budgets_status_id"},
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{})

	_, err := service.Create(context.Background(), model.RoleAdmin, "", validCreateBudgetRequest())

	assertAppError(t, err, 400, "Status de orcamento nao encontrado")
}

func TestBudgetServiceListShouldAllowEstimatorUserToViewFullBudgetScreen(t *testing.T) {
	repo := &budgetRepositoryStub{}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       55,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{
		getByUserIDItem: &model.EstimatorModel{
			ID:     44,
			Active: true,
		},
	})

	_, err := service.List(context.Background(), &dto.ListBudgetsFilters{}, model.RoleUser, "estimator.alpha")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedListFilters == nil {
		t.Fatal("expected filters to be captured")
	}
	if repo.capturedListFilters.RestrictedEstimatorID != nil {
		t.Fatalf("expected estimator restriction to be nil, got %v", *repo.capturedListFilters.RestrictedEstimatorID)
	}
	if repo.capturedListFilters.RestrictedSalespersonID != nil {
		t.Fatalf("expected salesperson restriction to be nil, got %v", *repo.capturedListFilters.RestrictedSalespersonID)
	}
}

func TestBudgetServiceUpdateShouldAllowEstimatorUserToEditOutsideOwnScope(t *testing.T) {
	repo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:           5,
			BudgetNumber: "ORC-100",
			YearBudget:   2026,
			StatusID:     1,
		},
	}
	service := budgetservice.NewService(repo, &budgetStatusRepositoryStub{}, &budgetImportPriorityRepositoryStub{}, &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       56,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}, &salespersonRepositoryStub{}, &estimatorRepositoryStub{
		getByUserIDItem: &model.EstimatorModel{
			ID:     45,
			Active: true,
		},
	})

	err := service.Update(context.Background(), 5, model.RoleUser, "estimator.beta", &dto.UpdateBudgetRequest{
		BudgetNumber: "ORC-100",
		YearBudget:   2026,
		SentAt:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		GrossValue:   1000,
		StatusID:     1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedListFilters == nil {
		t.Fatal("expected filters to be captured on update")
	}
	if repo.capturedListFilters.RestrictedEstimatorID != nil {
		t.Fatalf("expected estimator restriction to be nil on update, got %v", *repo.capturedListFilters.RestrictedEstimatorID)
	}
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
