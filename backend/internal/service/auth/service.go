package auth

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	jwtutil "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/jwt"
	refreshtoken "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/refreshtoken"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type Service interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (int64, error)
	Login(ctx context.Context, req *dto.LoginRequest) (string, string, error)
	Refresh(ctx context.Context, req *dto.RefreshTokenRequest, userID int64) (string, string, error)
}

type service struct {
	userRepo userrepository.Repository
	cfg      *config.Config
}

func NewService(userRepo userrepository.Repository, cfg *config.Config) Service {
	return &service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (int64, error) {
	totalUsers, err := s.userRepo.CountUsers(ctx)
	if err != nil {
		return 0, apperror.Internal("failed to count users", err)
	}

	if totalUsers > 0 {
		return 0, apperror.Forbidden("public registration is no longer available")
	}

	userID, err := s.createUser(ctx, req.Name, req.Email, req.Username, req.Password, model.RoleAdmin)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}
	if user == nil || !user.Active {
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", "", apperror.Unauthorized("wrong email or password")
	}

	return s.issueTokens(ctx, user)
}

func (s *service) Refresh(ctx context.Context, req *dto.RefreshTokenRequest, userID int64) (string, string, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal("failed to load user", err)
	}
	if user == nil || !user.Active {
		return "", "", apperror.Unauthorized("user is not active")
	}

	storedToken, err := s.userRepo.GetActiveRefreshTokenByUserID(ctx, userID, time.Now())
	if err != nil {
		return "", "", apperror.Internal("failed to load refresh token", err)
	}
	if storedToken == nil {
		return "", "", apperror.Unauthorized("refresh token expired")
	}
	if storedToken.RefreshToken != req.RefreshToken {
		return "", "", apperror.Unauthorized("refresh token not found")
	}

	return s.issueTokens(ctx, user)
}

func (s *service) createUser(ctx context.Context, name string, email string, username string, password string, role model.UserRole) (int64, error) {
	existingUser, err := s.userRepo.GetUserByEmailOrUsername(ctx, email, username)
	if err != nil {
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if existingUser != nil {
		return 0, apperror.Conflict("user already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	userID, err := s.userRepo.CreateUser(ctx, &model.UserModel{
		Name:         name,
		Email:        email,
		Username:     username,
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

func (s *service) issueTokens(ctx context.Context, user *model.UserModel) (string, string, error) {
	token, err := jwtutil.CreateToken(user.ID, user.Username, string(user.Role), s.cfg.SecretJWT)
	if err != nil {
		return "", "", apperror.Internal("failed to create token", err)
	}

	refreshToken, err := refreshtoken.Generate()
	if err != nil {
		return "", "", apperror.Internal("failed to generate refresh token", err)
	}

	now := time.Now()
	if err := s.userRepo.DeleteRefreshTokensByUserID(ctx, user.ID); err != nil {
		return "", "", apperror.Internal("failed to delete previous refresh token", err)
	}

	if err := s.userRepo.StoreRefreshToken(ctx, &model.RefreshTokenModel{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiredAt:    now.Add(refreshTokenTTL),
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		return "", "", apperror.Internal("failed to store refresh token", err)
	}

	return token, refreshToken, nil
}
