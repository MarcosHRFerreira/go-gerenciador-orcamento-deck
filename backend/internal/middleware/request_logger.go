package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"
const contextRequestIDKey = "requestID"

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		requestID := resolveRequestID(c.GetHeader(requestIDHeader))
		c.Set(contextRequestIDKey, requestID)
		c.Header(requestIDHeader, requestID)

		requestLogger := slog.Default().With(
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		requestContext := logger.NewContext(c.Request.Context(), requestLogger)
		c.Request = c.Request.WithContext(requestContext)

		c.Next()

		routePath := strings.TrimSpace(c.FullPath())
		if routePath == "" {
			routePath = c.Request.URL.Path
		}

		logArgs := []any{
			slog.String("route", routePath),
			slog.Int("status_code", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(startedAt).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		}

		userID := UserID(c)
		if userID > 0 {
			logArgs = append(logArgs, slog.Int64("user_id", userID))
		}

		username := Username(c)
		if username != "" {
			logArgs = append(logArgs, slog.String("username", username))
		}

		role := strings.TrimSpace(string(Role(c)))
		if role != "" {
			logArgs = append(logArgs, slog.String("role", role))
		}

		logLevel := slog.LevelInfo
		switch {
		case c.Writer.Status() >= http.StatusInternalServerError:
			logLevel = slog.LevelError
		case c.Writer.Status() >= http.StatusBadRequest:
			logLevel = slog.LevelWarn
		}

		requestLogger.Log(c.Request.Context(), logLevel, "request completed", logArgs...)
	}
}

func RequestID(c *gin.Context) string {
	value, exists := c.Get(contextRequestIDKey)
	if !exists {
		return ""
	}

	requestID, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(requestID)
}

func resolveRequestID(currentRequestID string) string {
	normalizedRequestID := strings.TrimSpace(currentRequestID)
	if normalizedRequestID != "" {
		return normalizedRequestID
	}

	return generateRequestID()
}

func generateRequestID() string {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmtFallbackRequestID()
	}

	return "req_" + hex.EncodeToString(randomBytes)
}

func fmtFallbackRequestID() string {
	return "req_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
