package unit

import (
	"context"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	authservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/auth"
	userservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/user"
	jwtutil "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/jwt"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	countUsersResult                 int64
	countUsersErr                    error
	createUserID                     int64
	createUserErr                    error
	getUserByEmailItem               *model.UserModel
	getUserByEmailErr                error
	getUserByEmailOrUsernameItem     *model.UserModel
	getUserByEmailOrUsernameErr      error
	getUserByIDItem                  *model.UserModel
	getUserByIDErr                   error
	listUsersItems                   []model.UserModel
	listUsersErr                     error
	getActiveRefreshTokenByUserID    *model.RefreshTokenModel
	getActiveRefreshTokenByUserIDErr error
	storeRefreshTokenErr             error
	deleteRefreshTokensByUserIDErr   error
	capturedCreateUser               *model.UserModel
	capturedStoreRefreshToken        *model.RefreshTokenModel
	deletedRefreshTokensUserID       int64
}

func (s *userRepositoryStub) CountUsers(_ context.Context) (int64, error) {
	return s.countUsersResult, s.countUsersErr
}

func (s *userRepositoryStub) CreateUser(_ context.Context, user *model.UserModel) (int64, error) {
	s.capturedCreateUser = user
	return s.createUserID, s.createUserErr
}

func (s *userRepositoryStub) GetUserByEmail(_ context.Context, _ string) (*model.UserModel, error) {
	return s.getUserByEmailItem, s.getUserByEmailErr
}

func (s *userRepositoryStub) GetUserByUsername(_ context.Context, _ string) (*model.UserModel, error) {
	return nil, nil
}

func (s *userRepositoryStub) GetUserByEmailOrUsername(_ context.Context, _, _ string) (*model.UserModel, error) {
	return s.getUserByEmailOrUsernameItem, s.getUserByEmailOrUsernameErr
}

func (s *userRepositoryStub) GetUserByID(_ context.Context, _ int64) (*model.UserModel, error) {
	return s.getUserByIDItem, s.getUserByIDErr
}

func (s *userRepositoryStub) ListUsers(_ context.Context) ([]model.UserModel, error) {
	return s.listUsersItems, s.listUsersErr
}

func (s *userRepositoryStub) GetActiveRefreshTokenByUserID(_ context.Context, _ int64, _ time.Time) (*model.RefreshTokenModel, error) {
	return s.getActiveRefreshTokenByUserID, s.getActiveRefreshTokenByUserIDErr
}

func (s *userRepositoryStub) StoreRefreshToken(_ context.Context, token *model.RefreshTokenModel) error {
	s.capturedStoreRefreshToken = token
	return s.storeRefreshTokenErr
}

func (s *userRepositoryStub) DeleteRefreshTokensByUserID(_ context.Context, userID int64) error {
	s.deletedRefreshTokensUserID = userID
	return s.deleteRefreshTokensByUserIDErr
}

func TestAuthServiceRegisterShouldCreateFirstUserAsAdmin(t *testing.T) {
	repo := &userRepositoryStub{
		createUserID: 11,
	}
	service := authservice.NewService(repo, &config.Config{SecretJWT: "local-secret"})

	userID, err := service.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Administrador",
		Email:           "admin@example.com",
		Username:        "admin",
		Password:        "123456",
		PasswordConfirm: "123456",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 11 {
		t.Fatalf("expected user id 11, got %d", userID)
	}
	if repo.capturedCreateUser == nil {
		t.Fatal("expected created user to be captured")
	}
	if repo.capturedCreateUser.Role != model.RoleAdmin {
		t.Fatalf("expected admin role, got %s", repo.capturedCreateUser.Role)
	}
	if repo.capturedCreateUser.PasswordHash == "123456" {
		t.Fatal("expected password hash to be generated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.capturedCreateUser.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("expected valid password hash, got %v", err)
	}
}

func TestAuthServiceRegisterShouldReturnForbiddenWhenUsersAlreadyExist(t *testing.T) {
	service := authservice.NewService(&userRepositoryStub{
		countUsersResult: 1,
	}, &config.Config{SecretJWT: "local-secret"})

	_, err := service.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Administrador",
		Email:           "admin@example.com",
		Username:        "admin",
		Password:        "123456",
		PasswordConfirm: "123456",
	})

	assertAppError(t, err, 403, "public registration is no longer available")
}

func TestAuthServiceLoginShouldReturnUnauthorizedWhenPasswordIsWrong(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	service := authservice.NewService(&userRepositoryStub{
		getUserByEmailItem: &model.UserModel{
			ID:           7,
			Username:     "user",
			PasswordHash: string(passwordHash),
			Role:         model.RoleUser,
			Active:       true,
		},
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, loginErr := service.Login(context.Background(), &dto.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})

	assertAppError(t, loginErr, 401, "wrong email or password")
}

func TestAuthServiceLoginShouldIssueTokenAndStoreRefreshToken(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	repo := &userRepositoryStub{
		getUserByEmailItem: &model.UserModel{
			ID:           7,
			Name:         "User",
			Email:        "user@example.com",
			Username:     "user",
			PasswordHash: string(passwordHash),
			Role:         model.RoleUser,
			Active:       true,
		},
	}
	service := authservice.NewService(repo, &config.Config{SecretJWT: "local-secret"})

	token, refreshToken, loginErr := service.Login(context.Background(), &dto.LoginRequest{
		Email:    "user@example.com",
		Password: "123456",
	})

	if loginErr != nil {
		t.Fatalf("expected no error, got %v", loginErr)
	}
	if token == "" {
		t.Fatal("expected token to be generated")
	}
	if refreshToken == "" {
		t.Fatal("expected refresh token to be generated")
	}
	if repo.deletedRefreshTokensUserID != 7 {
		t.Fatalf("expected refresh tokens to be deleted for user 7, got %d", repo.deletedRefreshTokensUserID)
	}
	if repo.capturedStoreRefreshToken == nil {
		t.Fatal("expected refresh token to be stored")
	}
	if repo.capturedStoreRefreshToken.UserID != 7 {
		t.Fatalf("expected stored refresh token user id 7, got %d", repo.capturedStoreRefreshToken.UserID)
	}
	if repo.capturedStoreRefreshToken.RefreshToken != refreshToken {
		t.Fatalf("expected stored refresh token %s, got %s", refreshToken, repo.capturedStoreRefreshToken.RefreshToken)
	}
	userID, username, role, validateErr := jwtutil.ValidateToken(token, "local-secret", true)
	if validateErr != nil {
		t.Fatalf("expected valid token, got %v", validateErr)
	}
	if userID != 7 || username != "user" || role != "user" {
		t.Fatalf("expected token claims 7/user/user, got %d/%s/%s", userID, username, role)
	}
}

func TestAuthServiceRefreshShouldReturnUnauthorizedWhenRefreshTokenDoesNotMatch(t *testing.T) {
	service := authservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       7,
			Username: "user",
			Role:     model.RoleUser,
			Active:   true,
		},
		getActiveRefreshTokenByUserID: &model.RefreshTokenModel{
			UserID:       7,
			RefreshToken: "stored-token",
		},
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, err := service.Refresh(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "another-token",
	}, 7)

	assertAppError(t, err, 401, "refresh token not found")
}

func TestAuthServiceRefreshShouldIssueNewTokens(t *testing.T) {
	repo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       7,
			Username: "user",
			Role:     model.RoleUser,
			Active:   true,
		},
		getActiveRefreshTokenByUserID: &model.RefreshTokenModel{
			UserID:       7,
			RefreshToken: "stored-token",
		},
	}
	service := authservice.NewService(repo, &config.Config{SecretJWT: "local-secret"})

	token, refreshToken, err := service.Refresh(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "stored-token",
	}, 7)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" || refreshToken == "" {
		t.Fatal("expected new token pair")
	}
	if repo.deletedRefreshTokensUserID != 7 {
		t.Fatalf("expected deletion of previous refresh token for user 7, got %d", repo.deletedRefreshTokensUserID)
	}
}

func TestUserServiceCreateShouldReturnBadRequestWhenRoleIsInvalid(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{})

	_, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        "123456",
		PasswordConfirm: "123456",
		Role:            "manager",
	})

	assertAppError(t, err, 400, "invalid role")
}

func TestUserServiceCreateShouldReturnConflictWhenUserAlreadyExists(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByEmailOrUsernameItem: &model.UserModel{ID: 1},
	})

	_, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        "123456",
		PasswordConfirm: "123456",
		Role:            "user",
	})

	assertAppError(t, err, 409, "user already exists")
}

func TestUserServiceCreateShouldHashPasswordAndCreateUser(t *testing.T) {
	repo := &userRepositoryStub{
		createUserID: 19,
	}
	service := userservice.NewService(repo)

	userID, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        "123456",
		PasswordConfirm: "123456",
		Role:            "user",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 19 {
		t.Fatalf("expected user id 19, got %d", userID)
	}
	if repo.capturedCreateUser == nil {
		t.Fatal("expected created user to be captured")
	}
	if repo.capturedCreateUser.Role != model.RoleUser {
		t.Fatalf("expected role user, got %s", repo.capturedCreateUser.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.capturedCreateUser.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("expected valid password hash, got %v", err)
	}
}

func TestUserServiceListShouldMapResponse(t *testing.T) {
	now := time.Date(2026, time.June, 13, 10, 0, 0, 0, time.UTC)
	service := userservice.NewService(&userRepositoryStub{
		listUsersItems: []model.UserModel{
			{
				ID:        1,
				Name:      "Admin",
				Email:     "admin@example.com",
				Username:  "admin",
				Role:      model.RoleAdmin,
				Active:    true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	})

	users, err := service.List(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Role != "admin" {
		t.Fatalf("expected role admin, got %s", users[0].Role)
	}
}

func TestUserServiceGetMeShouldReturnNotFoundWhenUserDoesNotExist(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{})

	_, err := service.GetMe(context.Background(), 9)

	assertAppError(t, err, 404, "user not found")
}
