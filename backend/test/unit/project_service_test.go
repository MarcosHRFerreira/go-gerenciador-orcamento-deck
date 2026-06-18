package unit

import (
	"context"
	"database/sql"
	"testing"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	projectservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/project"
)

type projectRepositoryStub struct {
	createID             int64
	createErr            error
	getByIDItem          *model.ProjectModel
	getByIDErr           error
	updateErr            error
	deleteErr            error
	nextCode             string
	nextCodeErr          error
	capturedCreateItem   *model.ProjectModel
	capturedUpdateItem   *model.ProjectModel
	getNextCodeCallCount int
}

type projectTypeRepositoryStub struct {
	getByIDItem *model.ProjectTypeModel
	getByIDErr  error
}

func (s *projectRepositoryStub) Create(_ context.Context, item *model.ProjectModel) (int64, error) {
	s.capturedCreateItem = item
	return s.createID, s.createErr
}

func (s *projectRepositoryStub) GetNextCode(_ context.Context) (string, error) {
	s.getNextCodeCallCount++
	return s.nextCode, s.nextCodeErr
}

func (s *projectRepositoryStub) List(_ context.Context) ([]model.ProjectModel, error) {
	return nil, nil
}

func (s *projectRepositoryStub) GetByID(_ context.Context, _ int64) (*model.ProjectModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *projectRepositoryStub) Update(_ context.Context, item *model.ProjectModel) error {
	s.capturedUpdateItem = item
	return s.updateErr
}

func (s *projectRepositoryStub) Delete(_ context.Context, _ int64) error {
	return s.deleteErr
}

func (s *projectTypeRepositoryStub) Create(_ context.Context, _ *model.ProjectTypeModel) (int64, error) {
	return 0, nil
}

func (s *projectTypeRepositoryStub) List(_ context.Context) ([]model.ProjectTypeModel, error) {
	return nil, nil
}

func (s *projectTypeRepositoryStub) GetByCodeOrName(_ context.Context, _ string, _ string) (*model.ProjectTypeModel, error) {
	return nil, nil
}

func (s *projectTypeRepositoryStub) GetByID(_ context.Context, _ int64) (*model.ProjectTypeModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *projectTypeRepositoryStub) Update(_ context.Context, _ *model.ProjectTypeModel) error {
	return nil
}

func (s *projectTypeRepositoryStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func TestProjectServiceCreateShouldGenerateCodeWhenCodeIsMissing(t *testing.T) {
	repo := &projectRepositoryStub{
		createID: 51,
		nextCode: "OBR-000123",
	}
	service := projectservice.NewService(repo, &projectTypeRepositoryStub{})

	projectID, err := service.Create(context.Background(), &dto.CreateProjectRequest{
		Code:  "   ",
		Name:  "  Obra teste  ",
		City:  "  Campinas  ",
		State: "  SP  ",
		Notes: "  observacao  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if projectID != 51 {
		t.Fatalf("expected project id 51, got %d", projectID)
	}
	if repo.getNextCodeCallCount != 1 {
		t.Fatalf("expected get next code to be called once, got %d", repo.getNextCodeCallCount)
	}
	if repo.capturedCreateItem == nil {
		t.Fatal("expected create item to be captured")
	}
	if repo.capturedCreateItem.Code != "OBR-000123" {
		t.Fatalf("expected generated code OBR-000123, got %s", repo.capturedCreateItem.Code)
	}
	if repo.capturedCreateItem.Name != "Obra teste" {
		t.Fatalf("expected trimmed name, got %s", repo.capturedCreateItem.Name)
	}
	if repo.capturedCreateItem.City != "Campinas" {
		t.Fatalf("expected trimmed city, got %s", repo.capturedCreateItem.City)
	}
	if repo.capturedCreateItem.State != "SP" {
		t.Fatalf("expected trimmed state, got %s", repo.capturedCreateItem.State)
	}
	if repo.capturedCreateItem.Notes != "observacao" {
		t.Fatalf("expected trimmed notes, got %s", repo.capturedCreateItem.Notes)
	}
}

func TestProjectServiceCreateShouldReturnInternalErrorWhenCodeGenerationFails(t *testing.T) {
	repo := &projectRepositoryStub{
		nextCodeErr: assertInternalError(),
	}
	service := projectservice.NewService(repo, &projectTypeRepositoryStub{})

	_, err := service.Create(context.Background(), &dto.CreateProjectRequest{
		Name: "Obra teste",
	})

	assertAppError(t, err, 500, "failed to generate project code")
}

func TestProjectServiceGetNextCodeShouldReturnGeneratedCode(t *testing.T) {
	repo := &projectRepositoryStub{
		nextCode: "OBR-000321",
	}
	service := projectservice.NewService(repo, &projectTypeRepositoryStub{})

	code, err := service.GetNextCode(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != "OBR-000321" {
		t.Fatalf("expected generated code OBR-000321, got %s", code)
	}
	if repo.getNextCodeCallCount != 1 {
		t.Fatalf("expected get next code to be called once, got %d", repo.getNextCodeCallCount)
	}
}

func TestProjectServiceUpdateShouldPreserveCurrentCodeWhenPayloadCodeIsBlank(t *testing.T) {
	repo := &projectRepositoryStub{
		getByIDItem: &model.ProjectModel{
			ID:   7,
			Code: "OBR-000007",
			Name: "Obra atual",
		},
	}
	service := projectservice.NewService(repo, &projectTypeRepositoryStub{})

	err := service.Update(context.Background(), 7, &dto.UpdateProjectRequest{
		Code:  "   ",
		Name:  "  Obra atualizada  ",
		City:  "  Sao Paulo  ",
		State: "  SP  ",
		Notes: "  notas  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedUpdateItem == nil {
		t.Fatal("expected update item to be captured")
	}
	if repo.capturedUpdateItem.Code != "OBR-000007" {
		t.Fatalf("expected current code to be preserved, got %s", repo.capturedUpdateItem.Code)
	}
	if repo.capturedUpdateItem.Name != "Obra atualizada" {
		t.Fatalf("expected trimmed name, got %s", repo.capturedUpdateItem.Name)
	}
	if repo.capturedUpdateItem.City != "Sao Paulo" {
		t.Fatalf("expected trimmed city, got %s", repo.capturedUpdateItem.City)
	}
	if repo.capturedUpdateItem.State != "SP" {
		t.Fatalf("expected trimmed state, got %s", repo.capturedUpdateItem.State)
	}
	if repo.capturedUpdateItem.Notes != "notas" {
		t.Fatalf("expected trimmed notes, got %s", repo.capturedUpdateItem.Notes)
	}
}

func assertInternalError() error {
	return apperror.Internal("db error", sql.ErrConnDone)
}
