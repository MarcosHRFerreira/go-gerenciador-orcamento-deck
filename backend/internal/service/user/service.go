package user

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (int64, error)
	List(ctx context.Context) ([]dto.UserResponse, error)
	GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error)
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
	role := model.UserRole(req.Role)
	if role != model.RoleAdmin && role != model.RoleUser {
		return 0, apperror.BadRequest("invalid role")
	}

	existingUser, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if existingUser != nil {
		return 0, apperror.Conflict("user already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	userID, err := s.userRepo.CreateUser(ctx, &model.UserModel{
		Name:         req.Name,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Role:         role,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
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

func toResponse(user model.UserModel) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		Role:      string(user.Role),
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
