package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const defaultServiceName = "go-gerenciador-orcamento-backend"

type contextKey string

const loggerContextKey contextKey = "request-logger"

func New(environment string, level string) (*slog.Logger, error) {
	logLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler).With(
		slog.String("service", defaultServiceName),
		slog.String("environment", normalizeEnvironment(environment)),
	), nil
}

func NewContext(ctx context.Context, log *slog.Logger) context.Context {
	if log == nil {
		return ctx
	}

	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	log, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok || log == nil {
		return slog.Default()
	}

	return log
}

func parseLevel(level string) (slog.Level, error) {
	normalizedLevel := strings.ToLower(strings.TrimSpace(level))
	switch normalizedLevel {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("nivel de log invalido: %s", level)
	}
}

func normalizeEnvironment(environment string) string {
	normalizedEnvironment := strings.TrimSpace(environment)
	if normalizedEnvironment == "" {
		return "development"
	}

	return normalizedEnvironment
}
