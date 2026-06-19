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

const (
	testStrongPassword  = "Strong@123"
	testUpdatedPassword = "Updated@123"
	testWeakPassword    = "12345678"
	testSetupToken      = "unit-bootstrap-token"
)

type userRepositoryStub struct {
	countUsersResult               int64
	countUsersErr                  error
	countActiveAdminsResult        int64
	countActiveAdminsErr           error
	createUserID                   int64
	createUserErr                  error
	getUserByEmailItem             *model.UserModel
	getUserByEmailErr              error
	getUserByUsernameItem          *model.UserModel
	getUserByUsernameErr           error
	getUserByEmailOrUsernameItem   *model.UserModel
	getUserByEmailOrUsernameErr    error
	getUserByIDItem                *model.UserModel
	getUserByIDErr                 error
	listUsersItems                 []model.UserModel
	listUsersErr                   error
	getActiveRefreshTokenByHash    *model.RefreshTokenModel
	getActiveRefreshTokenByHashErr error
	storeRefreshTokenErr           error
	deleteRefreshTokensByUserIDErr error
	deleteRefreshTokenByHashErr    error
	capturedCreateUser             *model.UserModel
	capturedStoreRefreshToken      *model.RefreshTokenModel
	deletedRefreshTokensUserID     int64
	deletedRefreshTokenHash        string
	updatedUserEmail               string
	updatedUserName                string
	updatedUsername                string
	updatedUserUserID              int64
	updatedUserRole                model.UserRole
	updatedUserKind                model.UserKind
	updatedUserRoleUserID          int64
	updatedUserActive              bool
	updatedUserActiveUserID        int64
	updatedUserPasswordHash        string
	updatedUserPasswordUserID      int64
	updatedUserMustChangePassword  bool
}

func (s *userRepositoryStub) CountUsers(_ context.Context) (int64, error) {
	return s.countUsersResult, s.countUsersErr
}

func (s *userRepositoryStub) CountActiveAdmins(_ context.Context) (int64, error) {
	return s.countActiveAdminsResult, s.countActiveAdminsErr
}

func (s *userRepositoryStub) CreateUser(_ context.Context, user *model.UserModel) (int64, error) {
	s.capturedCreateUser = user
	return s.createUserID, s.createUserErr
}

func (s *userRepositoryStub) GetUserByEmail(_ context.Context, _ string) (*model.UserModel, error) {
	return s.getUserByEmailItem, s.getUserByEmailErr
}

func (s *userRepositoryStub) GetUserByUsername(_ context.Context, _ string) (*model.UserModel, error) {
	return s.getUserByUsernameItem, s.getUserByUsernameErr
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

func (s *userRepositoryStub) UpdateUser(_ context.Context, userID int64, name string, email string, username string, role model.UserRole, userKind model.UserKind, _ time.Time) error {
	s.updatedUserUserID = userID
	s.updatedUserName = name
	s.updatedUserEmail = email
	s.updatedUsername = username
	s.updatedUserRole = role
	s.updatedUserKind = userKind
	return nil
}

func (s *userRepositoryStub) UpdateUserRole(_ context.Context, userID int64, role model.UserRole, userKind model.UserKind, _ time.Time) error {
	s.updatedUserRoleUserID = userID
	s.updatedUserRole = role
	s.updatedUserKind = userKind
	return nil
}

func (s *userRepositoryStub) UpdateUserActive(_ context.Context, userID int64, active bool, _ time.Time) error {
	s.updatedUserActiveUserID = userID
	s.updatedUserActive = active
	return nil
}

func (s *userRepositoryStub) UpdateUserPassword(_ context.Context, userID int64, passwordHash string, mustChangePassword bool, _ time.Time) error {
	s.updatedUserPasswordUserID = userID
	s.updatedUserPasswordHash = passwordHash
	s.updatedUserMustChangePassword = mustChangePassword
	return nil
}

func (s *userRepositoryStub) GetActiveRefreshTokenByHash(_ context.Context, _ string, _ time.Time) (*model.RefreshTokenModel, error) {
	return s.getActiveRefreshTokenByHash, s.getActiveRefreshTokenByHashErr
}

func (s *userRepositoryStub) StoreRefreshToken(_ context.Context, token *model.RefreshTokenModel) error {
	s.capturedStoreRefreshToken = token
	return s.storeRefreshTokenErr
}

func (s *userRepositoryStub) DeleteRefreshTokensByUserID(_ context.Context, userID int64) error {
	s.deletedRefreshTokensUserID = userID
	return s.deleteRefreshTokensByUserIDErr
}

func (s *userRepositoryStub) DeleteRefreshTokenByHash(_ context.Context, refreshTokenHash string) error {
	s.deletedRefreshTokenHash = refreshTokenHash
	return s.deleteRefreshTokenByHashErr
}

func TestAuthServiceRegisterShouldCreateFirstUserAsAdmin(t *testing.T) {
	repo := &userRepositoryStub{
		createUserID: 11,
	}
	service := authservice.NewService(repo, &config.Config{
		SecretJWT:              "local-secret",
		InitialAdminSetupToken: testSetupToken,
	})

	userID, err := service.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Administrador",
		Email:           "admin@example.com",
		Username:        "admin",
		Password:        testStrongPassword,
		PasswordConfirm: testStrongPassword,
	}, testSetupToken)

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
	if repo.capturedCreateUser.MustChangePassword {
		t.Fatal("expected first registered admin to skip forced password change")
	}
	if repo.capturedCreateUser.PasswordHash == testStrongPassword {
		t.Fatal("expected password hash to be generated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.capturedCreateUser.PasswordHash), []byte(testStrongPassword)); err != nil {
		t.Fatalf("expected valid password hash, got %v", err)
	}
}

func TestAuthServiceRegisterShouldRejectWeakPassword(t *testing.T) {
	service := authservice.NewService(&userRepositoryStub{}, &config.Config{
		SecretJWT:              "local-secret",
		InitialAdminSetupToken: testSetupToken,
	})

	_, err := service.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Administrador",
		Email:           "admin@example.com",
		Username:        "admin",
		Password:        testWeakPassword,
		PasswordConfirm: testWeakPassword,
	}, testSetupToken)

	assertAppError(t, err, 400, "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial")
}

func TestAuthServiceRegisterShouldReturnForbiddenWhenUsersAlreadyExist(t *testing.T) {
	service := authservice.NewService(&userRepositoryStub{
		countUsersResult: 1,
	}, &config.Config{
		SecretJWT:              "local-secret",
		InitialAdminSetupToken: testSetupToken,
	})

	_, err := service.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Administrador",
		Email:           "admin@example.com",
		Username:        "admin",
		Password:        testStrongPassword,
		PasswordConfirm: testStrongPassword,
	}, testSetupToken)

	assertAppError(t, err, 403, "Cadastro publico nao esta mais disponivel")
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

	assertAppError(t, loginErr, 401, "E-mail ou senha invalidos")
}

func TestAuthServiceLoginShouldReturnUnauthorizedWhenUserIsInactive(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testStrongPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	service := authservice.NewService(&userRepositoryStub{
		getUserByEmailItem: &model.UserModel{
			ID:           7,
			Username:     "user",
			PasswordHash: string(passwordHash),
			Role:         model.RoleUser,
			Active:       false,
		},
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, loginErr := service.Login(context.Background(), &dto.LoginRequest{
		Email:    "user@example.com",
		Password: testStrongPassword,
	})

	assertAppError(t, loginErr, 401, "Usuario desativado")
}

func TestAuthServiceLoginShouldIssueTokenAndStoreRefreshToken(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testStrongPassword), bcrypt.DefaultCost)
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
		Password: testStrongPassword,
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
	if repo.capturedStoreRefreshToken.RefreshToken == refreshToken {
		t.Fatal("expected stored refresh token to be hashed")
	}
	if len(repo.capturedStoreRefreshToken.RefreshToken) != 64 {
		t.Fatalf("expected stored refresh token hash with 64 chars, got %d", len(repo.capturedStoreRefreshToken.RefreshToken))
	}
	userID, username, role, mustChangePassword, validateErr := jwtutil.ValidateToken(token, "local-secret", true)
	if validateErr != nil {
		t.Fatalf("expected valid token, got %v", validateErr)
	}
	if userID != 7 || username != "user" || role != "user" {
		t.Fatalf("expected token claims 7/user/user, got %d/%s/%s", userID, username, role)
	}
	if mustChangePassword {
		t.Fatal("expected login token without forced password change for regular auth flow")
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
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, err := service.Refresh(context.Background(), "another-token")

	assertAppError(t, err, 401, "Refresh token expirado")
}

func TestAuthServiceRefreshShouldIssueNewTokens(t *testing.T) {
	repo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       7,
			Username: "user",
			Role:     model.RoleUser,
			Active:   true,
		},
		getActiveRefreshTokenByHash: &model.RefreshTokenModel{
			UserID:       7,
			RefreshToken: "hashed-refresh-token",
		},
	}
	service := authservice.NewService(repo, &config.Config{SecretJWT: "local-secret"})

	token, refreshToken, err := service.Refresh(context.Background(), "stored-token")

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
		Password:        testWeakPassword,
		PasswordConfirm: testWeakPassword,
		Role:            "manager",
	})

	assertAppError(t, err, 400, "Perfil invalido")
}

func TestUserServiceCreateShouldReturnConflictWhenUserAlreadyExists(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByEmailOrUsernameItem: &model.UserModel{ID: 1},
	})

	_, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        testStrongPassword,
		PasswordConfirm: testStrongPassword,
		Role:            "user",
		UserKind:        stringPointer(string(model.UserKindSalesperson)),
	})

	assertAppError(t, err, 409, "Usuario ja existe")
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
		Password:        testStrongPassword,
		PasswordConfirm: testStrongPassword,
		Role:            "user",
		UserKind:        stringPointer(string(model.UserKindSalesperson)),
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
	if repo.capturedCreateUser.UserKind != model.UserKindSalesperson {
		t.Fatalf("expected default salesperson kind for user role, got %s", repo.capturedCreateUser.UserKind)
	}
	if !repo.capturedCreateUser.MustChangePassword {
		t.Fatal("expected admin-created user to require password change on first access")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.capturedCreateUser.PasswordHash), []byte(testStrongPassword)); err != nil {
		t.Fatalf("expected valid password hash, got %v", err)
	}
}

func TestUserServiceCreateShouldRejectWeakPassword(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{})

	_, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        testWeakPassword,
		PasswordConfirm: testWeakPassword,
		Role:            "user",
		UserKind:        stringPointer(string(model.UserKindSalesperson)),
	})

	assertAppError(t, err, 400, "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial")
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
	if users[0].UserKind != nil {
		t.Fatalf("expected admin user kind to be nil, got %v", *users[0].UserKind)
	}
}

func TestUserServiceGetMeShouldReturnNotFoundWhenUserDoesNotExist(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{})

	_, err := service.GetMe(context.Background(), 9)

	assertAppError(t, err, 404, "Usuario nao encontrado")
}

func TestAuthServiceChangePasswordShouldUpdatePasswordAndIssueNewTokens(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testStrongPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	repo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:                 17,
			Email:              "user@example.com",
			Username:           "user",
			PasswordHash:       string(passwordHash),
			Role:               model.RoleUser,
			Active:             true,
			MustChangePassword: true,
		},
	}
	service := authservice.NewService(repo, &config.Config{SecretJWT: "local-secret"})

	token, refreshToken, changeErr := service.ChangePassword(context.Background(), 17, &dto.ChangePasswordRequest{
		CurrentPassword:    testStrongPassword,
		NewPassword:        testUpdatedPassword,
		NewPasswordConfirm: testUpdatedPassword,
	})

	if changeErr != nil {
		t.Fatalf("expected no error, got %v", changeErr)
	}
	if token == "" || refreshToken == "" {
		t.Fatal("expected token pair after changing password")
	}
	if repo.updatedUserPasswordUserID != 17 {
		t.Fatalf("expected password update for user 17, got %d", repo.updatedUserPasswordUserID)
	}
	if repo.updatedUserMustChangePassword {
		t.Fatal("expected forced password change flag to be cleared")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.updatedUserPasswordHash), []byte(testUpdatedPassword)); err != nil {
		t.Fatalf("expected updated password hash, got %v", err)
	}

	_, _, _, mustChangePassword, validateErr := jwtutil.ValidateToken(token, "local-secret", true)
	if validateErr != nil {
		t.Fatalf("expected valid token, got %v", validateErr)
	}
	if mustChangePassword {
		t.Fatal("expected updated token without forced password change")
	}
}

func TestAuthServiceChangePasswordShouldRejectSamePassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testStrongPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	service := authservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:                 17,
			Username:           "user",
			PasswordHash:       string(passwordHash),
			Role:               model.RoleUser,
			Active:             true,
			MustChangePassword: true,
		},
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, changeErr := service.ChangePassword(context.Background(), 17, &dto.ChangePasswordRequest{
		CurrentPassword:    testStrongPassword,
		NewPassword:        testStrongPassword,
		NewPasswordConfirm: testStrongPassword,
	})

	assertAppError(t, changeErr, 400, "A nova senha deve ser diferente da senha atual")
}

func TestAuthServiceChangePasswordShouldRejectWeakNewPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testStrongPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	service := authservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:                 17,
			Username:           "user",
			PasswordHash:       string(passwordHash),
			Role:               model.RoleUser,
			Active:             true,
			MustChangePassword: true,
		},
	}, &config.Config{SecretJWT: "local-secret"})

	_, _, changeErr := service.ChangePassword(context.Background(), 17, &dto.ChangePasswordRequest{
		CurrentPassword:    testStrongPassword,
		NewPassword:        testWeakPassword,
		NewPasswordConfirm: testWeakPassword,
	})

	assertAppError(t, changeErr, 400, "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial")
}

func TestUserServiceUpdateRoleShouldPreventChangingOwnRole(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:     7,
			Role:   model.RoleAdmin,
			Active: true,
		},
	})

	err := service.UpdateRole(context.Background(), 7, 7, &dto.UpdateUserRoleRequest{
		Role: "user",
	})

	assertAppError(t, err, 403, "Nao e permitido alterar o proprio perfil")
}

func TestUserServiceUpdateShouldUpdateUserData(t *testing.T) {
	repo := &userRepositoryStub{
		getUserByEmailItem: nil,
		getUserByIDItem: &model.UserModel{
			ID:       9,
			Name:     "Usuario Antigo",
			Email:    "old.user@local.dev",
			Username: "old_user",
			Role:     model.RoleUser,
			Active:   true,
		},
	}
	service := userservice.NewService(repo)

	err := service.Update(context.Background(), 1, 9, &dto.UpdateUserRequest{
		Name:     "Usuario Atualizado",
		Email:    "new.user@local.dev",
		Username: "new_user",
		Role:     "admin",
		UserKind: nil,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedUserUserID != 9 {
		t.Fatalf("expected update for user 9, got %d", repo.updatedUserUserID)
	}
	if repo.updatedUserName != "Usuario Atualizado" {
		t.Fatalf("expected updated name, got %s", repo.updatedUserName)
	}
	if repo.updatedUserEmail != "new.user@local.dev" {
		t.Fatalf("expected updated email, got %s", repo.updatedUserEmail)
	}
	if repo.updatedUsername != "new_user" {
		t.Fatalf("expected updated username, got %s", repo.updatedUsername)
	}
	if repo.updatedUserRole != model.RoleAdmin {
		t.Fatalf("expected updated role admin, got %s", repo.updatedUserRole)
	}
	if repo.updatedUserKind != "" {
		t.Fatalf("expected updated user kind to be cleared for admin, got %s", repo.updatedUserKind)
	}
}

func TestUserServiceUpdateShouldRejectDuplicatedEmail(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByEmailItem: &model.UserModel{
			ID:    15,
			Email: "duplicated@local.dev",
		},
		getUserByIDItem: &model.UserModel{
			ID:       9,
			Name:     "Usuario Atual",
			Email:    "current@local.dev",
			Username: "current_user",
			Role:     model.RoleUser,
			Active:   true,
		},
	})

	err := service.Update(context.Background(), 1, 9, &dto.UpdateUserRequest{
		Name:     "Usuario Atual",
		Email:    "duplicated@local.dev",
		Username: "current_user",
		Role:     "user",
		UserKind: stringPointer(string(model.UserKindSalesperson)),
	})

	assertAppError(t, err, 409, "E-mail ja esta em uso")
}

func TestUserServiceUpdateRoleShouldPreventRemovingLastActiveAdmin(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		countActiveAdminsResult: 1,
		getUserByIDItem: &model.UserModel{
			ID:     9,
			Role:   model.RoleAdmin,
			Active: true,
		},
	})

	err := service.UpdateRole(context.Background(), 1, 9, &dto.UpdateUserRoleRequest{
		Role: "user",
	})

	assertAppError(t, err, 403, "Nao e permitido remover o perfil do ultimo administrador ativo")
}

func TestUserServiceUpdateRoleShouldUpdateRole(t *testing.T) {
	repo := &userRepositoryStub{
		countActiveAdminsResult: 2,
		getUserByIDItem: &model.UserModel{
			ID:     9,
			Role:   model.RoleUser,
			Active: true,
		},
	}
	service := userservice.NewService(repo)

	err := service.UpdateRole(context.Background(), 1, 9, &dto.UpdateUserRoleRequest{
		Role: "admin",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedUserRoleUserID != 9 {
		t.Fatalf("expected role update for user 9, got %d", repo.updatedUserRoleUserID)
	}
	if repo.updatedUserRole != model.RoleAdmin {
		t.Fatalf("expected role admin, got %s", repo.updatedUserRole)
	}
	if repo.updatedUserKind != "" {
		t.Fatalf("expected user kind to be cleared for admin role update, got %s", repo.updatedUserKind)
	}
}

func TestUserServiceCreateShouldRequireUserKindWhenRoleIsUser(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{})

	_, err := service.Create(context.Background(), &dto.CreateUserRequest{
		Name:            "User",
		Email:           "user@example.com",
		Username:        "user",
		Password:        testStrongPassword,
		PasswordConfirm: testStrongPassword,
		Role:            "user",
	})

	assertAppError(t, err, 400, "user_kind e obrigatorio para perfil user")
}

func TestUserServiceUpdateRoleShouldRequireUserKindWhenChangingToUser(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		countActiveAdminsResult: 2,
		getUserByIDItem: &model.UserModel{
			ID:     9,
			Role:   model.RoleAdmin,
			Active: true,
		},
	})

	err := service.UpdateRole(context.Background(), 1, 9, &dto.UpdateUserRoleRequest{
		Role: "user",
	})

	assertAppError(t, err, 400, "user_kind e obrigatorio para perfil user")
}

func TestUserServiceUpdateShouldPersistEstimatorUserKind(t *testing.T) {
	repo := &userRepositoryStub{
		getUserByEmailItem: nil,
		getUserByIDItem: &model.UserModel{
			ID:       9,
			Name:     "Usuario Atual",
			Email:    "current@local.dev",
			Username: "current_user",
			Role:     model.RoleUser,
			UserKind: model.UserKindSalesperson,
			Active:   true,
		},
	}
	service := userservice.NewService(repo)

	err := service.Update(context.Background(), 1, 9, &dto.UpdateUserRequest{
		Name:     "Usuario Atual",
		Email:    "current@local.dev",
		Username: "current_user",
		Role:     "user",
		UserKind: stringPointer(string(model.UserKindEstimator)),
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedUserKind != model.UserKindEstimator {
		t.Fatalf("expected updated user kind estimator, got %s", repo.updatedUserKind)
	}
}

func stringPointer(value string) *string {
	return &value
}

func TestUserServiceUpdateActiveShouldPreventSelfDeactivation(t *testing.T) {
	active := false
	service := userservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:     7,
			Role:   model.RoleAdmin,
			Active: true,
		},
	})

	err := service.UpdateActive(context.Background(), 7, 7, &dto.UpdateUserActiveRequest{
		Active: &active,
	})

	assertAppError(t, err, 403, "Nao e permitido desativar o proprio usuario")
}

func TestUserServiceUpdateActiveShouldPreventDeactivatingLastActiveAdmin(t *testing.T) {
	active := false
	service := userservice.NewService(&userRepositoryStub{
		countActiveAdminsResult: 1,
		getUserByIDItem: &model.UserModel{
			ID:     9,
			Role:   model.RoleAdmin,
			Active: true,
		},
	})

	err := service.UpdateActive(context.Background(), 1, 9, &dto.UpdateUserActiveRequest{
		Active: &active,
	})

	assertAppError(t, err, 403, "Nao e permitido desativar o ultimo administrador ativo")
}

func TestUserServiceUpdateActiveShouldUpdateStatus(t *testing.T) {
	active := false
	repo := &userRepositoryStub{
		countActiveAdminsResult: 2,
		getUserByIDItem: &model.UserModel{
			ID:     9,
			Role:   model.RoleUser,
			Active: true,
		},
	}
	service := userservice.NewService(repo)

	err := service.UpdateActive(context.Background(), 1, 9, &dto.UpdateUserActiveRequest{
		Active: &active,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedUserActiveUserID != 9 {
		t.Fatalf("expected active update for user 9, got %d", repo.updatedUserActiveUserID)
	}
	if repo.updatedUserActive {
		t.Fatal("expected user to be deactivated")
	}
}

func TestUserServiceResetPasswordShouldPreventResettingOwnPassword(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       7,
			Username: "admin",
			Role:     model.RoleAdmin,
			Active:   true,
		},
	})

	err := service.ResetPassword(context.Background(), 7, 7, &dto.ResetUserPasswordRequest{
		Password:        testUpdatedPassword,
		PasswordConfirm: testUpdatedPassword,
	})

	assertAppError(t, err, 403, "Nao e permitido resetar a propria senha")
}

func TestUserServiceResetPasswordShouldUpdatePasswordAndRequireChange(t *testing.T) {
	repo := &userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       9,
			Username: "user",
			Role:     model.RoleUser,
			Active:   true,
		},
	}
	service := userservice.NewService(repo)

	err := service.ResetPassword(context.Background(), 1, 9, &dto.ResetUserPasswordRequest{
		Password:        testUpdatedPassword,
		PasswordConfirm: testUpdatedPassword,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedUserPasswordUserID != 9 {
		t.Fatalf("expected password reset for user 9, got %d", repo.updatedUserPasswordUserID)
	}
	if !repo.updatedUserMustChangePassword {
		t.Fatal("expected forced password change after admin reset")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.updatedUserPasswordHash), []byte(testUpdatedPassword)); err != nil {
		t.Fatalf("expected updated password hash, got %v", err)
	}
}

func TestUserServiceResetPasswordShouldRejectWeakPassword(t *testing.T) {
	service := userservice.NewService(&userRepositoryStub{
		getUserByIDItem: &model.UserModel{
			ID:       9,
			Username: "user",
			Role:     model.RoleUser,
			Active:   true,
		},
	})

	err := service.ResetPassword(context.Background(), 1, 9, &dto.ResetUserPasswordRequest{
		Password:        testWeakPassword,
		PasswordConfirm: testWeakPassword,
	})

	assertAppError(t, err, 400, "A senha deve conter pelo menos 8 caracteres, letra maiuscula, letra minuscula, numero e caractere especial")
}
