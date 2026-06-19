package unit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	estimatorservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/estimator"
)

type estimatorRepositoryStub struct {
	createID        int64
	createErr       error
	nextCode        string
	nextCodeErr     error
	listItems       []model.EstimatorModel
	listErr         error
	getByIDItem     *model.EstimatorModel
	getByIDErr      error
	getByCodeItem   *model.EstimatorModel
	getByCodeErr    error
	getByUserIDItem *model.EstimatorModel
	getByUserIDErr  error
	updateErr       error
	deleteErr       error

	capturedCreateItem *model.EstimatorModel
	capturedUpdateItem *model.EstimatorModel
	deletedEstimatorID int64
}

func (s *estimatorRepositoryStub) Create(_ context.Context, item *model.EstimatorModel) (int64, error) {
	s.capturedCreateItem = item
	return s.createID, s.createErr
}

func (s *estimatorRepositoryStub) GetNextCode(_ context.Context) (string, error) {
	return s.nextCode, s.nextCodeErr
}

func (s *estimatorRepositoryStub) List(_ context.Context) ([]model.EstimatorModel, error) {
	return s.listItems, s.listErr
}

func (s *estimatorRepositoryStub) GetByID(_ context.Context, _ int64) (*model.EstimatorModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *estimatorRepositoryStub) GetByCode(_ context.Context, _ string) (*model.EstimatorModel, error) {
	return s.getByCodeItem, s.getByCodeErr
}

func (s *estimatorRepositoryStub) GetByUserID(_ context.Context, _ int64) (*model.EstimatorModel, error) {
	return s.getByUserIDItem, s.getByUserIDErr
}

func (s *estimatorRepositoryStub) Update(_ context.Context, item *model.EstimatorModel) error {
	s.capturedUpdateItem = item
	return s.updateErr
}

func (s *estimatorRepositoryStub) Delete(_ context.Context, estimatorID int64) error {
	s.deletedEstimatorID = estimatorID
	return s.deleteErr
}

func TestEstimatorServiceCreateShouldGenerateCodeAndPersistLinkedUser(t *testing.T) {
	repo := &estimatorRepositoryStub{
		createID:      21,
		nextCode:      "EST-000021",
		getByCodeItem: nil,
	}
	userRepo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       8,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}
	service := estimatorservice.NewService(repo, userRepo)
	userID := int64(8)

	id, err := service.Create(context.Background(), &dto.CreateEstimatorRequest{
		Name:   "Orcamentista 1",
		Email:  "orc1@example.com",
		Phone:  "11999999999",
		Notes:  "principal",
		UserID: &userID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 21 {
		t.Fatalf("expected estimator id 21, got %d", id)
	}
	if repo.capturedCreateItem == nil {
		t.Fatal("expected created estimator to be captured")
	}
	if repo.capturedCreateItem.Code != "EST-000021" {
		t.Fatalf("expected generated code EST-000021, got %s", repo.capturedCreateItem.Code)
	}
	if !repo.capturedCreateItem.UserID.Valid || repo.capturedCreateItem.UserID.Int64 != 8 {
		t.Fatalf("expected linked user id 8, got %+v", repo.capturedCreateItem.UserID)
	}
	if !repo.capturedCreateItem.Active {
		t.Fatal("expected new estimator to be active")
	}
}

func TestEstimatorServiceCreateShouldRejectLinkedUserWithoutEstimatorKind(t *testing.T) {
	repo := &estimatorRepositoryStub{
		nextCode: "EST-000001",
	}
	userRepo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       8,
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}
	service := estimatorservice.NewService(repo, userRepo)
	userID := int64(8)

	_, err := service.Create(context.Background(), &dto.CreateEstimatorRequest{
		Name:   "Orcamentista 1",
		UserID: &userID,
	})

	assertAppError(t, err, 400, "Usuario informado precisa ter user_kind estimator")
}

func TestEstimatorServiceListShouldMapResponse(t *testing.T) {
	now := time.Date(2026, time.June, 18, 12, 0, 0, 0, time.UTC)
	repo := &estimatorRepositoryStub{
		listItems: []model.EstimatorModel{
			{
				ID:        4,
				Code:      "EST-000004",
				Name:      "Orcamentista 4",
				Email:     "orc4@example.com",
				Phone:     "1133334444",
				Active:    true,
				Notes:     "observacao",
				UserID:    sql.NullInt64{Int64: 12, Valid: true},
				UserName:  sql.NullString{String: "user.est4", Valid: true},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	service := estimatorservice.NewService(repo, &userRepositoryStub{})

	items, err := service.List(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 estimator, got %d", len(items))
	}
	if items[0].UserID == nil || *items[0].UserID != 12 {
		t.Fatalf("expected linked user id 12, got %+v", items[0].UserID)
	}
	if items[0].UserName == nil || *items[0].UserName != "user.est4" {
		t.Fatalf("expected linked user name user.est4, got %+v", items[0].UserName)
	}
}

func TestEstimatorServiceUpdateShouldPersistChanges(t *testing.T) {
	repo := &estimatorRepositoryStub{
		getByIDItem: &model.EstimatorModel{
			ID:     9,
			Code:   "EST-000009",
			Name:   "Atual",
			Email:  "atual@example.com",
			Phone:  "1100000000",
			Active: true,
			Notes:  "antes",
		},
	}
	userRepo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       15,
			Role:     model.RoleUser,
			UserKind: model.UserKindEstimator,
			Active:   true,
		},
	}
	service := estimatorservice.NewService(repo, userRepo)
	userID := int64(15)

	err := service.Update(context.Background(), 9, &dto.UpdateEstimatorRequest{
		Code:   "EST-000009",
		Name:   "Atualizado",
		Email:  "novo@example.com",
		Phone:  "11911112222",
		Active: false,
		Notes:  "depois",
		UserID: &userID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedUpdateItem == nil {
		t.Fatal("expected updated estimator to be captured")
	}
	if repo.capturedUpdateItem.Name != "Atualizado" {
		t.Fatalf("expected updated name, got %s", repo.capturedUpdateItem.Name)
	}
	if repo.capturedUpdateItem.UserID.Int64 != 15 || !repo.capturedUpdateItem.UserID.Valid {
		t.Fatalf("expected linked user id 15, got %+v", repo.capturedUpdateItem.UserID)
	}
	if repo.capturedUpdateItem.Active {
		t.Fatal("expected updated estimator to be inactive")
	}
}

func TestEstimatorServiceDeleteShouldReturnNotFoundWhenEstimatorDoesNotExist(t *testing.T) {
	service := estimatorservice.NewService(&estimatorRepositoryStub{}, &userRepositoryStub{})

	err := service.Delete(context.Background(), 3)

	assertAppError(t, err, 404, "Orcamentista nao encontrado")
}
