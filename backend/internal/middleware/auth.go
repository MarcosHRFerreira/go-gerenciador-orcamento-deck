package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/jwt"
	"github.com/gin-gonic/gin"
)

const (
	contextUserIDKey             = "userID"
	contextUsernameKey           = "username"
	contextRoleKey               = "role"
	contextMustChangePasswordKey = "mustChangePassword"
)

func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, username, role, mustChangePassword, err := jwt.ValidateToken(c.GetHeader("Authorization"), secretKey, true)
		if err != nil {
			logSecurityWarn(c, "Requisicao bloqueada por token de autorizacao invalido", slog.String("security_action", "auth"), slog.String("reason", "invalid_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "Token de autorizacao invalido")
			return
		}

		c.Set(contextUserIDKey, userID)
		c.Set(contextUsernameKey, username)
		c.Set(contextRoleKey, role)
		c.Set(contextMustChangePasswordKey, mustChangePassword)

		if mustChangePassword && !isPasswordChangeAllowedPath(c.FullPath()) {
			logSecurityWarn(c, "Requisicao bloqueada por troca de senha obrigatoria", slog.String("security_action", "auth"), slog.Int64("user_id", userID), slog.String("username", username), slog.String("role", role), slog.String("reason", "must_change_password"))
			httpresponse.AbortJSONError(c, http.StatusForbidden, "Troca de senha obrigatoria antes de acessar o sistema")
			return
		}

		c.Next()
	}
}

func AuthRefresh(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, username, role, mustChangePassword, err := jwt.ValidateToken(c.GetHeader("Authorization"), secretKey, false)
		if err != nil {
			logSecurityWarn(c, "Requisicao de refresh bloqueada por token invalido", slog.String("security_action", "auth_refresh"), slog.String("reason", "invalid_authorization_token"))
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "Token de autorizacao invalido")
			return
		}

		c.Set(contextUserIDKey, userID)
		c.Set(contextUsernameKey, username)
		c.Set(contextRoleKey, role)
		c.Set(contextMustChangePasswordKey, mustChangePassword)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(contextRoleKey)
		if !exists {
			logSecurityWarn(c, "Requisicao bloqueada por papel autenticado ausente", slog.String("security_action", "require_roles"), slog.String("reason", "missing_authenticated_role"))
			httpresponse.AbortJSONError(c, http.StatusForbidden, "Papel do usuario autenticado nao encontrado")
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			logSecurityWarn(c, "Requisicao bloqueada por papel autenticado invalido", slog.String("security_action", "require_roles"), slog.String("reason", "invalid_authenticated_role"))
			httpresponse.AbortJSONError(c, http.StatusForbidden, "Papel do usuario autenticado invalido")
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == string(allowedRole) {
				c.Next()
				return
			}
		}

		logSecurityWarn(c, "Requisicao bloqueada por permissoes insuficientes", slog.String("security_action", "require_roles"), slog.Int64("user_id", UserID(c)), slog.String("username", Username(c)), slog.String("role", role), slog.String("reason", "insufficient_permissions"))
		httpresponse.AbortJSONError(c, http.StatusForbidden, "Permissoes insuficientes")
	}
}

func UserID(c *gin.Context) int64 {
	value, exists := c.Get(contextUserIDKey)
	if !exists {
		return 0
	}

	userID, ok := value.(int64)
	if !ok {
		return 0
	}

	return userID
}

func Username(c *gin.Context) string {
	value, exists := c.Get(contextUsernameKey)
	if !exists {
		return ""
	}

	username, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(username)
}

func Role(c *gin.Context) model.UserRole {
	value, exists := c.Get(contextRoleKey)
	if !exists {
		return ""
	}

	role, ok := value.(string)
	if !ok {
		return ""
	}

	return model.UserRole(strings.TrimSpace(role))
}

func isPasswordChangeAllowedPath(path string) bool {
	return path == "/users/me" || path == "/auth/change-password"
}

func logSecurityWarn(c *gin.Context, message string, attrs ...slog.Attr) {
	logger.FromContext(c.Request.Context()).LogAttrs(c.Request.Context(), slog.LevelWarn, message, attrs...)
}
