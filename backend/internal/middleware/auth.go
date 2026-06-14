package middleware

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql/jwt"
	"github.com/gin-gonic/gin"
)

const (
	contextUserIDKey   = "userID"
	contextUsernameKey = "username"
	contextRoleKey     = "role"
)

func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, username, role, err := jwt.ValidateToken(c.GetHeader("Authorization"), secretKey, true)
		if err != nil {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}

		c.Set(contextUserIDKey, userID)
		c.Set(contextUsernameKey, username)
		c.Set(contextRoleKey, role)
		c.Next()
	}
}

func AuthRefresh(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, username, role, err := jwt.ValidateToken(c.GetHeader("Authorization"), secretKey, false)
		if err != nil {
			httpresponse.AbortJSONError(c, http.StatusUnauthorized, "invalid authorization token")
			return
		}

		c.Set(contextUserIDKey, userID)
		c.Set(contextUsernameKey, username)
		c.Set(contextRoleKey, role)
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
