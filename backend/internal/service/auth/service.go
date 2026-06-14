package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/security"
	jwtutil "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/jwt"
	refreshtoken "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/refreshtoken"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type Service interface {
	Register(ctx context.Context, req *dto.RegisterRequest, setupToken string) (int64, error)
	Login(ctx context.Context, req *dto.LoginRequest) (string, string, error)
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
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

func (s *service) Register(ctx context.Context, req *dto.RegisterRequest, setupToken string) (int64, error) {
	totalUsers, err := s.userRepo.CountUsers(ctx)
	if err != nil {
		logAuthError(ctx, "Falha ao contar usuarios no bootstrap inicial", err)
		return 0, apperror.Internal("failed to count users", err)
	}

	if totalUsers > 0 {
		logAuthWarn(ctx, "Tentativa de bootstrap inicial bloqueada", slog.String("auth_action", "register_initial"), slog.String("reason", "users_already_exist"))
		return 0, apperror.Forbidden("Cadastro publico nao esta mais disponivel")
	}
	if strings.TrimSpace(s.cfg.InitialAdminSetupToken) == "" {
		logAuthWarn(ctx, "Bootstrap inicial bloqueado por configuracao ausente", slog.String("auth_action", "register_initial"), slog.String("reason", "setup_token_not_configured"))
		return 0, apperror.Forbidden("Token de configuracao inicial do admin nao foi definido")
	}
	if strings.TrimSpace(setupToken) != s.cfg.InitialAdminSetupToken {
		logAuthWarn(ctx, "Bootstrap inicial bloqueado por token invalido", slog.String("auth_action", "register_initial"), slog.String("reason", "invalid_setup_token"))
		return 0, apperror.Forbidden("Token de configuracao invalido")
	}

	userID, err := s.createUser(ctx, req.Name, req.Email, req.Username, req.Password, model.RoleAdmin)
	if err != nil {
		return 0, err
	}

	logAuthInfo(ctx, "Bootstrap inicial concluido", slog.String("auth_action", "register_initial"), slog.Int64("user_id", userID), slog.String("username", strings.TrimSpace(req.Username)), slog.String("role", string(model.RoleAdmin)))
	return userID, nil
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logAuthError(ctx, "Falha ao carregar usuario no login", err, slog.String("auth_action", "login"))
		return "", "", apperror.Internal("failed to load user", err)
	}
	if user == nil {
		logAuthWarn(ctx, "Tentativa de login com credenciais invalidas", slog.String("auth_action", "login"), slog.String("reason", "user_not_found"))
		return "", "", apperror.Unauthorized("E-mail ou senha invalidos")
	}
	if !user.Active {
		logAuthWarn(ctx, "Tentativa de login com usuario desativado", slog.String("auth_action", "login"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("reason", "inactive_user"))
		return "", "", apperror.Unauthorized("Usuario desativado")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logAuthWarn(ctx, "Tentativa de login com credenciais invalidas", slog.String("auth_action", "login"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("reason", "invalid_password"))
		return "", "", apperror.Unauthorized("E-mail ou senha invalidos")
	}

	token, refreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		logAuthError(ctx, "Falha ao emitir tokens no login", err, slog.String("auth_action", "login"), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
		return "", "", err
	}

	logAuthInfo(ctx, "Login concluido", slog.String("auth_action", "login"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("role", string(user.Role)))
	return token, refreshToken, nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	refreshTokenHash := hashRefreshToken(refreshToken)
	if refreshTokenHash == "" {
		logAuthWarn(ctx, "Refresh bloqueado por token ausente", slog.String("auth_action", "refresh"), slog.String("reason", "missing_refresh_token"))
		return "", "", apperror.Unauthorized("Refresh token nao encontrado")
	}

	storedToken, err := s.userRepo.GetActiveRefreshTokenByHash(ctx, refreshTokenHash, time.Now())
	if err != nil {
		logAuthError(ctx, "Falha ao carregar refresh token", err, slog.String("auth_action", "refresh"))
		return "", "", apperror.Internal("failed to load refresh token", err)
	}
	if storedToken == nil {
		logAuthWarn(ctx, "Refresh bloqueado por token expirado ou invalido", slog.String("auth_action", "refresh"), slog.String("reason", "refresh_token_not_found"))
		return "", "", apperror.Unauthorized("Refresh token expirado")
	}

	user, err := s.userRepo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		logAuthError(ctx, "Falha ao carregar usuario no refresh", err, slog.String("auth_action", "refresh"), slog.Int64("user_id", storedToken.UserID))
		return "", "", apperror.Internal("failed to load user", err)
	}
	if user == nil || !user.Active {
		logAuthWarn(ctx, "Refresh bloqueado por usuario desativado", slog.String("auth_action", "refresh"), slog.Int64("user_id", storedToken.UserID), slog.String("reason", "inactive_user"))
		return "", "", apperror.Unauthorized("Usuario desativado")
	}

	token, nextRefreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		logAuthError(ctx, "Falha ao emitir tokens no refresh", err, slog.String("auth_action", "refresh"), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
		return "", "", err
	}

	logAuthInfo(ctx, "Refresh concluido", slog.String("auth_action", "refresh"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("role", string(user.Role)))
	return token, nextRefreshToken, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenHash := hashRefreshToken(refreshToken)
	if refreshTokenHash == "" {
		logAuthInfo(ctx, "Logout concluido sem refresh token", slog.String("auth_action", "logout"), slog.String("reason", "missing_refresh_token"))
		return nil
	}

	if err := s.userRepo.DeleteRefreshTokenByHash(ctx, refreshTokenHash); err != nil {
		logAuthError(ctx, "Falha ao revogar refresh token no logout", err, slog.String("auth_action", "logout"))
		return apperror.Internal("failed to revoke refresh token", err)
	}

	logAuthInfo(ctx, "Logout concluido", slog.String("auth_action", "logout"))
	return nil
}

func (s *service) createUser(ctx context.Context, name string, email string, username string, password string, role model.UserRole) (int64, error) {
	existingUser, err := s.userRepo.GetUserByEmailOrUsername(ctx, email, username)
	if err != nil {
		logAuthError(ctx, "Falha ao verificar usuario existente", err, slog.String("auth_action", "create_user"))
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if existingUser != nil {
		logAuthWarn(ctx, "Criacao de usuario bloqueada por duplicidade", slog.String("auth_action", "create_user"), slog.String("username", strings.TrimSpace(username)), slog.String("reason", "user_already_exists"))
		return 0, apperror.Conflict("Usuario ja existe")
	}

	if err := security.ValidateStrongPassword(password); err != nil {
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logAuthError(ctx, "Falha ao gerar hash de senha", err, slog.String("auth_action", "create_user"))
		return 0, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	userID, err := s.userRepo.CreateUser(ctx, &model.UserModel{
		Name:               name,
		Email:              email,
		Username:           username,
		PasswordHash:       string(passwordHash),
		Role:               role,
		Active:             true,
		MustChangePassword: false,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		logAuthError(ctx, "Falha ao criar usuario", err, slog.String("auth_action", "create_user"), slog.String("username", strings.TrimSpace(username)), slog.String("role", string(role)))
		return 0, apperror.Internal("failed to create user", err)
	}

	return userID, nil
}

func (s *service) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) (string, string, error) {
	if userID <= 0 {
		logAuthWarn(ctx, "Troca de senha bloqueada por usuario autenticado invalido", slog.String("auth_action", "change_password"), slog.String("reason", "invalid_authenticated_user"))
		return "", "", apperror.Forbidden("Usuario autenticado invalido")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logAuthError(ctx, "Falha ao carregar usuario na troca de senha", err, slog.String("auth_action", "change_password"), slog.Int64("user_id", userID))
		return "", "", apperror.Internal("failed to load user", err)
	}
	if user == nil || !user.Active {
		logAuthWarn(ctx, "Troca de senha bloqueada por usuario desativado", slog.String("auth_action", "change_password"), slog.Int64("user_id", userID), slog.String("reason", "inactive_user"))
		return "", "", apperror.Unauthorized("Usuario desativado")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		logAuthWarn(ctx, "Troca de senha bloqueada por senha atual invalida", slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("reason", "invalid_current_password"))
		return "", "", apperror.Unauthorized("Senha atual invalida")
	}
	if req.CurrentPassword == req.NewPassword {
		logAuthWarn(ctx, "Troca de senha bloqueada por reutilizacao da senha atual", slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("reason", "same_password"))
		return "", "", apperror.BadRequest("A nova senha deve ser diferente da senha atual")
	}
	if err := security.ValidateStrongPassword(req.NewPassword); err != nil {
		logAuthWarn(ctx, "Troca de senha bloqueada por senha fraca", slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("reason", "weak_password"))
		return "", "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logAuthError(ctx, "Falha ao gerar hash na troca de senha", err, slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID))
		return "", "", apperror.Internal("failed to hash password", err)
	}

	if err := s.userRepo.UpdateUserPassword(ctx, user.ID, string(passwordHash), false, time.Now()); err != nil {
		logAuthError(ctx, "Falha ao atualizar senha do usuario", err, slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
		return "", "", apperror.Internal("failed to update user password", err)
	}

	user.PasswordHash = string(passwordHash)
	user.MustChangePassword = false

	token, refreshToken, err := s.issueTokens(ctx, user)
	if err != nil {
		logAuthError(ctx, "Falha ao emitir tokens apos troca de senha", err, slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
		return "", "", err
	}

	logAuthInfo(ctx, "Troca de senha concluida", slog.String("auth_action", "change_password"), slog.Int64("user_id", user.ID), slog.String("username", user.Username), slog.String("role", string(user.Role)))
	return token, refreshToken, nil
}

func (s *service) issueTokens(ctx context.Context, user *model.UserModel) (string, string, error) {
	token, err := jwtutil.CreateToken(user.ID, user.Username, string(user.Role), user.MustChangePassword, s.cfg.SecretJWT)
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
		RefreshToken: hashRefreshToken(refreshToken),
		ExpiredAt:    now.Add(refreshTokenTTL),
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		return "", "", apperror.Internal("failed to store refresh token", err)
	}

	return token, refreshToken, nil
}

func hashRefreshToken(refreshToken string) string {
	normalizedToken := strings.TrimSpace(refreshToken)
	if normalizedToken == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(normalizedToken))
	return hex.EncodeToString(sum[:])
}

func logAuthInfo(ctx context.Context, message string, attrs ...slog.Attr) {
	logWithLevel(ctx, slog.LevelInfo, message, attrs...)
}

func logAuthWarn(ctx context.Context, message string, attrs ...slog.Attr) {
	logWithLevel(ctx, slog.LevelWarn, message, attrs...)
}

func logAuthError(ctx context.Context, message string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	logWithLevel(ctx, slog.LevelError, message, attrs...)
}

func logWithLevel(ctx context.Context, level slog.Level, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, level, message, attrs...)
}
