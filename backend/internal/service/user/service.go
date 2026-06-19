package user

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (int64, error)
	List(ctx context.Context) ([]dto.UserResponse, error)
	GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error)
	Update(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserRequest) error
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
		logUserWarn(ctx, "Criacao de usuario bloqueada por perfil invalido", slog.String("user_action", "create_user"), slog.String("reason", "invalid_role"))
		return 0, err
	}
	userKind, err := normalizeUserKind(role, req.UserKind)
	if err != nil {
		logUserWarn(ctx, "Criacao de usuario bloqueada por tipo funcional invalido", slog.String("user_action", "create_user"), slog.String("reason", "invalid_user_kind"))
		return 0, err
	}

	existingUser, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		logUserError(ctx, "Falha ao verificar usuario existente", err, slog.String("user_action", "create_user"))
		return 0, apperror.Internal("failed to check existing user", err)
	}
	if existingUser != nil {
		logUserWarn(ctx, "Criacao de usuario bloqueada por duplicidade", slog.String("user_action", "create_user"), slog.String("username", req.Username), slog.String("reason", "user_already_exists"))
		return 0, apperror.Conflict("Usuario ja existe")
	}

	if err := security.ValidateStrongPassword(req.Password); err != nil {
		logUserWarn(ctx, "Criacao de usuario bloqueada por senha fraca", slog.String("user_action", "create_user"), slog.String("username", req.Username), slog.String("reason", "weak_password"))
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logUserError(ctx, "Falha ao gerar hash de senha na criacao de usuario", err, slog.String("user_action", "create_user"))
		return 0, apperror.Internal("failed to hash password", err)
	}

	now := time.Now()
	userID, err := s.userRepo.CreateUser(ctx, &model.UserModel{
		Name:               req.Name,
		Email:              req.Email,
		Username:           req.Username,
		PasswordHash:       string(passwordHash),
		Role:               role,
		UserKind:           userKind,
		Active:             true,
		MustChangePassword: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		logUserError(ctx, "Falha ao criar usuario", err, slog.String("user_action", "create_user"), slog.String("username", req.Username), slog.String("role", string(role)))
		return 0, apperror.Internal("failed to create user", err)
	}

	logUserInfo(ctx, "Usuario criado com sucesso", slog.String("user_action", "create_user"), slog.Int64("target_user_id", userID), slog.String("username", req.Username), slog.String("role", string(role)))
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
		return nil, apperror.NotFound("Usuario nao encontrado")
	}

	response := toResponse(*user)

	return &response, nil
}

func (s *service) Update(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserRequest) error {
	if userID <= 0 {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por user_id invalido", slog.String("user_action", "update_user"), slog.String("reason", "invalid_target_user"))
		return apperror.BadRequest("user_id e obrigatorio")
	}
	if actorUserID <= 0 {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por usuario autenticado invalido", slog.String("user_action", "update_user"), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_actor"))
		return apperror.Forbidden("Usuario autenticado invalido")
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	username := strings.TrimSpace(req.Username)
	nextRole, err := normalizeRole(req.Role)
	if err != nil {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por perfil invalido", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_role"))
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logUserError(ctx, "Falha ao carregar usuario na atualizacao", err, slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por usuario inexistente", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "user_not_found"))
		return apperror.NotFound("Usuario nao encontrado")
	}

	if actorUserID == userID && user.Role != nextRole {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por autoalteracao de perfil", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "self_role_change"))
		return apperror.Forbidden("Nao e permitido alterar o proprio perfil")
	}

	existingByEmail, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		logUserError(ctx, "Falha ao verificar duplicidade de e-mail", err, slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check email duplication", err)
	}
	if existingByEmail != nil && existingByEmail.ID != userID {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por e-mail duplicado", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("email", email), slog.String("reason", "duplicated_email"))
		return apperror.Conflict("E-mail ja esta em uso")
	}

	existingByUsername, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		logUserError(ctx, "Falha ao verificar duplicidade de username", err, slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check username duplication", err)
	}
	if existingByUsername != nil && existingByUsername.ID != userID {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por username duplicado", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("username", username), slog.String("reason", "duplicated_username"))
		return apperror.Conflict("Username ja esta em uso")
	}

	if user.Role == model.RoleAdmin && nextRole != model.RoleAdmin {
		activeAdminsCount, err := s.userRepo.CountActiveAdmins(ctx)
		if err != nil {
			logUserError(ctx, "Falha ao contar administradores ativos", err, slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
			return apperror.Internal("failed to count active admins", err)
		}
		if user.Active && activeAdminsCount <= 1 {
			logUserWarn(ctx, "Atualizacao de usuario bloqueada para preservar ultimo admin ativo", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "last_active_admin"))
			return apperror.Forbidden("Nao e permitido remover o perfil do ultimo administrador ativo")
		}
	}

	nextUserKind, err := normalizeUserKind(nextRole, req.UserKind)
	if err != nil {
		logUserWarn(ctx, "Atualizacao de usuario bloqueada por tipo funcional invalido", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_user_kind"))
		return err
	}

	if user.Name == name && user.Email == email && user.Username == username && user.Role == nextRole && user.UserKind == nextUserKind {
		logUserInfo(ctx, "Atualizacao de usuario ignorada por nao haver mudanca", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return nil
	}

	if err := s.userRepo.UpdateUser(ctx, userID, name, email, username, nextRole, nextUserKind, time.Now()); err != nil {
		logUserError(ctx, "Falha ao atualizar usuario", err, slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("username", username), slog.String("role", string(nextRole)))
		return apperror.Internal("failed to update user", err)
	}

	logUserInfo(ctx, "Usuario atualizado com sucesso", slog.String("user_action", "update_user"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("username", username), slog.String("role", string(nextRole)))
	return nil
}

func (s *service) UpdateRole(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserRoleRequest) error {
	if userID <= 0 {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por user_id invalido", slog.String("user_action", "update_role"), slog.String("reason", "invalid_target_user"))
		return apperror.BadRequest("user_id e obrigatorio")
	}
	if actorUserID <= 0 {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por usuario autenticado invalido", slog.String("user_action", "update_role"), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_actor"))
		return apperror.Forbidden("Usuario autenticado invalido")
	}

	nextRole, err := normalizeRole(req.Role)
	if err != nil {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por perfil invalido", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_role"))
		return err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logUserError(ctx, "Falha ao carregar usuario na alteracao de perfil", err, slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por usuario inexistente", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "user_not_found"))
		return apperror.NotFound("Usuario nao encontrado")
	}

	if actorUserID == userID {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por autoalteracao", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "self_role_change"))
		return apperror.Forbidden("Nao e permitido alterar o proprio perfil")
	}

	if user.Role == model.RoleAdmin && nextRole != model.RoleAdmin {
		activeAdminsCount, err := s.userRepo.CountActiveAdmins(ctx)
		if err != nil {
			logUserError(ctx, "Falha ao contar administradores ativos", err, slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
			return apperror.Internal("failed to count active admins", err)
		}
		if user.Active && activeAdminsCount <= 1 {
			logUserWarn(ctx, "Alteracao de perfil bloqueada para preservar ultimo admin ativo", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "last_active_admin"))
			return apperror.Forbidden("Nao e permitido remover o perfil do ultimo administrador ativo")
		}
	}

	nextUserKind, err := normalizeUserKind(nextRole, req.UserKind)
	if err != nil {
		logUserWarn(ctx, "Alteracao de perfil bloqueada por tipo funcional invalido", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_user_kind"))
		return err
	}

	if user.Role == nextRole && user.UserKind == nextUserKind {
		logUserInfo(ctx, "Alteracao de perfil ignorada por nao haver mudanca", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("role", string(nextRole)))
		return nil
	}

	if err := s.userRepo.UpdateUserRole(ctx, userID, nextRole, nextUserKind, time.Now()); err != nil {
		logUserError(ctx, "Falha ao atualizar perfil do usuario", err, slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("role", string(nextRole)))
		return apperror.Internal("failed to update user role", err)
	}

	logUserInfo(ctx, "Perfil do usuario atualizado", slog.String("user_action", "update_role"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("role", string(nextRole)))
	return nil
}

func (s *service) UpdateActive(ctx context.Context, actorUserID int64, userID int64, req *dto.UpdateUserActiveRequest) error {
	if userID <= 0 {
		logUserWarn(ctx, "Alteracao de status do usuario bloqueada por user_id invalido", slog.String("user_action", "update_active"), slog.String("reason", "invalid_target_user"))
		return apperror.BadRequest("user_id e obrigatorio")
	}
	if actorUserID <= 0 {
		logUserWarn(ctx, "Alteracao de status do usuario bloqueada por usuario autenticado invalido", slog.String("user_action", "update_active"), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_actor"))
		return apperror.Forbidden("Usuario autenticado invalido")
	}
	if req.Active == nil {
		logUserWarn(ctx, "Alteracao de status do usuario bloqueada por payload invalido", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "missing_active_flag"))
		return apperror.BadRequest("active e obrigatorio")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logUserError(ctx, "Falha ao carregar usuario na alteracao de status", err, slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		logUserWarn(ctx, "Alteracao de status do usuario bloqueada por usuario inexistente", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "user_not_found"))
		return apperror.NotFound("Usuario nao encontrado")
	}

	if actorUserID == userID && !*req.Active {
		logUserWarn(ctx, "Alteracao de status do usuario bloqueada por autodesativacao", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "self_deactivation"))
		return apperror.Forbidden("Nao e permitido desativar o proprio usuario")
	}

	if user.Active == *req.Active {
		logUserInfo(ctx, "Alteracao de status do usuario ignorada por nao haver mudanca", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.Bool("active", *req.Active))
		return nil
	}

	if user.Role == model.RoleAdmin && !*req.Active {
		activeAdminsCount, err := s.userRepo.CountActiveAdmins(ctx)
		if err != nil {
			logUserError(ctx, "Falha ao contar administradores ativos", err, slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
			return apperror.Internal("failed to count active admins", err)
		}
		if activeAdminsCount <= 1 {
			logUserWarn(ctx, "Alteracao de status bloqueada para preservar ultimo admin ativo", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "last_active_admin"))
			return apperror.Forbidden("Nao e permitido desativar o ultimo administrador ativo")
		}
	}

	if err := s.userRepo.UpdateUserActive(ctx, userID, *req.Active, time.Now()); err != nil {
		logUserError(ctx, "Falha ao atualizar status do usuario", err, slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.Bool("active", *req.Active))
		return apperror.Internal("failed to update user active status", err)
	}

	logUserInfo(ctx, "Status do usuario atualizado", slog.String("user_action", "update_active"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.Bool("active", *req.Active))
	return nil
}

func (s *service) ResetPassword(ctx context.Context, actorUserID int64, userID int64, req *dto.ResetUserPasswordRequest) error {
	if userID <= 0 {
		logUserWarn(ctx, "Reset de senha bloqueado por user_id invalido", slog.String("user_action", "reset_password"), slog.String("reason", "invalid_target_user"))
		return apperror.BadRequest("user_id e obrigatorio")
	}
	if actorUserID <= 0 {
		logUserWarn(ctx, "Reset de senha bloqueado por usuario autenticado invalido", slog.String("user_action", "reset_password"), slog.Int64("target_user_id", userID), slog.String("reason", "invalid_actor"))
		return apperror.Forbidden("Usuario autenticado invalido")
	}
	if req.Password == "" {
		logUserWarn(ctx, "Reset de senha bloqueado por senha ausente", slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "missing_password"))
		return apperror.BadRequest("Senha obrigatoria")
	}
	if actorUserID == userID {
		logUserWarn(ctx, "Reset de senha bloqueado por autoredefinicao", slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "self_reset"))
		return apperror.Forbidden("Nao e permitido resetar a propria senha")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logUserError(ctx, "Falha ao carregar usuario no reset de senha", err, slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to check user", err)
	}
	if user == nil {
		logUserWarn(ctx, "Reset de senha bloqueado por usuario inexistente", slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "user_not_found"))
		return apperror.NotFound("Usuario nao encontrado")
	}

	if err := security.ValidateStrongPassword(req.Password); err != nil {
		logUserWarn(ctx, "Reset de senha bloqueado por senha fraca", slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID), slog.String("reason", "weak_password"))
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logUserError(ctx, "Falha ao gerar hash no reset de senha", err, slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to hash password", err)
	}

	if err := s.userRepo.UpdateUserPassword(ctx, userID, string(passwordHash), true, time.Now()); err != nil {
		logUserError(ctx, "Falha ao resetar senha do usuario", err, slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
		return apperror.Internal("failed to reset user password", err)
	}

	logUserInfo(ctx, "Reset de senha concluido", slog.String("user_action", "reset_password"), slog.Int64("actor_user_id", actorUserID), slog.Int64("target_user_id", userID))
	return nil
}

func normalizeRole(role string) (model.UserRole, error) {
	userRole := model.UserRole(role)
	if userRole != model.RoleAdmin && userRole != model.RoleUser {
		return "", apperror.BadRequest("Perfil invalido")
	}

	return userRole, nil
}

func toResponse(user model.UserModel) dto.UserResponse {
	var userKind *string
	if user.UserKind != "" {
		value := string(user.UserKind)
		userKind = &value
	}

	return dto.UserResponse{
		ID:                 user.ID,
		Name:               user.Name,
		Email:              user.Email,
		Username:           user.Username,
		Role:               string(user.Role),
		UserKind:           userKind,
		Active:             user.Active,
		MustChangePassword: user.MustChangePassword,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}

func normalizeUserKind(role model.UserRole, userKind *string) (model.UserKind, error) {
	if role == model.RoleAdmin {
		if userKind != nil && strings.TrimSpace(*userKind) != "" {
			return "", apperror.BadRequest("user_kind deve ser informado apenas para perfil user")
		}

		return "", nil
	}

	if userKind == nil {
		return "", apperror.BadRequest("user_kind e obrigatorio para perfil user")
	}

	normalizedUserKind := model.UserKind(strings.TrimSpace(*userKind))
	if normalizedUserKind != model.UserKindSalesperson && normalizedUserKind != model.UserKindEstimator {
		return "", apperror.BadRequest("Tipo funcional invalido")
	}

	return normalizedUserKind, nil
}

func logUserInfo(ctx context.Context, message string, attrs ...slog.Attr) {
	logUserWithLevel(ctx, slog.LevelInfo, message, attrs...)
}

func logUserWarn(ctx context.Context, message string, attrs ...slog.Attr) {
	logUserWithLevel(ctx, slog.LevelWarn, message, attrs...)
}

func logUserError(ctx context.Context, message string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	logUserWithLevel(ctx, slog.LevelError, message, attrs...)
}

func logUserWithLevel(ctx context.Context, level slog.Level, message string, attrs ...slog.Attr) {
	logger.FromContext(ctx).LogAttrs(ctx, level, message, attrs...)
}
