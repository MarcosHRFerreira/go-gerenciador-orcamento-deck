package middleware

import (
	"net/http"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
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
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}

		c.Set(contextUserIDKey, userID)
		c.Set(contextUsernameKey, username)
		c.Set(contextRoleKey, role)
		c.Set(contextMustChangePasswordKey, mustChangePassword)

		if mustChangePassword && !isPasswordChangeAllowedPath(c.FullPath()) {
			httpresponse.AbortJSONError(c, http.StatusForbidden, "password change required before accessing the system")
			return
		}

		c.Next()
	}
}

func AuthRefresh(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, username, role, mustChangePassword, err := jwt.ValidateToken(c.GetHeader("Authorization"), secretKey, false)
		if err != nil {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
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
			httpresponse.AbortJSONError(c, http.StatusForbidden, "missing authenticated role")
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			httpresponse.AbortJSONError(c, http.StatusForbidden, "invalid authenticated role")
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == string(allowedRole) {
				c.Next()
				return
			}
		}

		httpresponse.AbortJSONError(c, http.StatusForbidden, "insufficient permissions")
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
