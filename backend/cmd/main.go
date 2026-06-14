package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/server"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/pkg/internalsql"
	"github.com/go-playground/validator/v10"
)

const (
	serverReadHeaderTimeout = 2 * time.Second
	serverReadTimeout       = 10 * time.Second
	serverWriteTimeout      = 15 * time.Second
	serverIdleTimeout       = 60 * time.Second
	shutdownTimeout         = 10 * time.Second
)

func main() {
	validate := validator.New()

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao carregar configuracao: %v\n", err)
		os.Exit(1)
	}

	appLogger, err := logger.New(cfg.AppEnv, cfg.LogLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "falha ao inicializar logger: %v\n", err)
		os.Exit(1)
	}
	slog.SetDefault(appLogger)

	db, err := internalsql.ConnectPostgres(cfg)
	if err != nil {
		appLogger.Error("falha ao conectar ao banco de dados", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Warn("falha ao fechar conexao com banco de dados", slog.Any("error", err))
		}
	}()

	router := server.NewRouter(validate, server.Dependencies{
		DB:            db,
		HealthChecker: db,
		Config:        cfg,
	})

	httpServer := &http.Server{
		Addr:              cfg.ServerAddress(),
		Handler:           router,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info("iniciando servidor http", slog.String("address", cfg.ServerAddress()))
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		close(serverErrors)
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			appLogger.Error("servidor http finalizado de forma inesperada", slog.Any("error", err))
			os.Exit(1)
		}
	case <-shutdownSignal.Done():
		appLogger.Info("sinal de encerramento recebido")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("falha ao encerrar servidor http", slog.Any("error", err))
		os.Exit(1)
	}

	appLogger.Info("servidor encerrado com sucesso")
}
