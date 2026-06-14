package user

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (int64, error)
	List(ctx context.Context) ([]dto.UserResponse, error)
	GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error)
	UpdateRole(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserRoleRequest) error
	UpdateActive(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserActiveRequest) error
	ResetPassword(ctx context.Context, actorUserID int64, userID int64, req *dto.ResetUserPasswordRequest) error
}

type service struct {
	userRepo userrepository.Repository
}

func NewService(userRepo userrepository.Repository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Create(ctx context.Context, req *dto.CreateUserRequest) (int64, error) {
	role, err := normalizeRole(req.Role)
	if err != nil {
		return 0, err
	}

	existingUser, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if existingUser != nil {
		return 0, apperror.Conflict("user already exists")
	}

	if err := security.ValidateStrongPassword(req.Password); err != nil {
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	userID, err := s.userRepo.CreateUser(ctx, &model.UserModel{
		Name:               req.Name,
		Email:              req.Email,
		Username:           req.Username,
		PasswordHash:       string(passwordHash),
		Role:               role,
		Active:             true,
		MustChangePassword: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return 0, apperror.Internal("failed to create user", err)
	}

	return userID, nil
}

func (s *service) List(ctx context.Context) ([]dto.UserResponse, error) {
	users, err := s.userRepo.ListUsers(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list users", err)
	}

	response := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toResponse(user))
	}

	return response, nil
}

func (s *service) GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to load user", err)
	}
	if user == nil {
		return nil, apperror.NotFound("user not found")
	}

	response := toResponse(*user)

	return &response, nil
}

func (s *service) UpdateRole(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserRoleRequest) error {
	if userID <= 0 {
		return apperror.BadRequest("user_id is required")
	}
	if actorUserID <= 0 {
		return apperror.Forbidden("invalid authenticated user")
	}

	nextRole, err := normalizeRole(req.Role)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		return apperror.NotFound("user not found")
	}

	if actorUserID == userID {
		return apperror.Forbidden("cannot change your own role")
	}

	if user.Role == nextRole {
		return nil
	}

	if user.Role == model.RoleAdmin && nextRole != model.RoleAdmin {
		activeAdminsCount, err := s.userRepo.CountActiveAdmins(ctx)
		if err != nil {
			return apperror.Internal("failed to count active admins", err)
		}
		if user.Active && activeAdminsCount <= 1 {
			return apperror.Forbidden("cannot remove role from last active admin")
		}
	}

	if err := s.userRepo.UpdateUserRole(ctx, userID, nextRole, time.Now()); err != nil {
		return apperror.Internal("failed to update user role", err)
	}

	return nil
}

func (s *service) UpdateActive(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserActiveRequest) error {
	if userID <= 0 {
		return apperror.BadRequest("user_id is required")
	}
	if actorUserID <= 0 {
		return apperror.Forbidden("invalid authenticated user")
	}
	if req.Active == nil {
		return apperror.BadRequest("active is required")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		return apperror.NotFound("user not found")
	}

	if actorUserID == userID && !*req.Active {
		return apperror.Forbidden("cannot deactivate your own user")
	}

	if user.Active == *req.Active {
		return nil
	}

	if user.Role == model.RoleAdmin && !*req.Active {
		activeAdminsCount, err := s.userRepo.CountActiveAdmins(ctx)
		if err != nil {
			return apperror.Internal("failed to count active admins", err)
		}
		if activeAdminsCount <= 1 {
			return apperror.Forbidden("cannot deactivate last active admin")
		}
	}

	if err := s.userRepo.UpdateUserActive(ctx, userID, *req.Active, time.Now()); err != nil {
		return apperror.Internal("failed to update user active status", err)
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, actorUserID int64, userID int64, req *dto.ResetUserPasswordRequest) error {
	if userID <= 0 {
		return apperror.BadRequest("user_id is required")
	}
	if actorUserID <= 0 {
		return apperror.Forbidden("invalid authenticated user")
	}
	if req.Password == "" {
		return apperror.BadRequest("password is required")
	}
	if actorUserID == userID {
		return apperror.Forbidden("cannot reset your own password")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		return apperror.NotFound("user not found")
	}

	if err := security.ValidateStrongPassword(req.Password); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("failed to hash password", err)
	}

	if err := s.userRepo.UpdateUserPassword(ctx, userID, string(passwordHash), true, time.Now()); err != nil {
		return apperror.Internal("failed to reset user password", err)
	}

	return nil
}

func normalizeRole(role string) (model.UserRole, error) {
	userRole := model.UserRole(role)
	if userRole != model.RoleAdmin && userRole != model.RoleUser {
		return "", apperror.BadRequest("invalid role")
	}

	return userRole, nil
}

func toResponse(user model.UserModel) dto.UserResponse {
	return dto.UserResponse{
		ID:                 user.ID,
		Name:               user.Name,
		Email:              user.Email,
		Username:           user.Username,
		Role:               string(user.Role),
		Active:             user.Active,
		MustChangePassword: user.MustChangePassword,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}
