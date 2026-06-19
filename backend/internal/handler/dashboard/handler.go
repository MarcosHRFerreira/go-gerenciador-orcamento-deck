package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/logger"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	dashboardservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/dashboard"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	router    gin.IRouter
	service   dashboardservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, service dashboardservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	adminRoutes := h.router.Group("/dashboard")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.GET("/salespeople", h.GetSalespeopleDashboard)
}

func (h *Handler) GetSalespeopleDashboard(c *gin.Context) {
	startedAt := time.Now()
	filters, err := parseSalespeopleDashboardFilters(c)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	requestLogger := logger.FromContext(c.Request.Context()).With(
		slog.String("dashboard_action", "salespeople"),
		slog.String("username", middleware.Username(c)),
		slog.String("role", string(middleware.Role(c))),
		slog.String("source_company", filters.SourceCompany),
		slog.Bool("has_salesperson_filter", filters.SalespersonID != nil),
		slog.Bool("has_year_filter", filters.Year != nil),
		slog.Bool("has_month_filter", filters.Month != nil),
	)

	response, err := h.service.GetSalespeopleDashboard(
		c.Request.Context(),
		filters,
		middleware.Role(c),
		middleware.Username(c),
	)
	if err != nil {
		requestLogger.ErrorContext(
			c.Request.Context(),
			"falha ao carregar dashboard de vendedores",
			slog.Int("duration_ms", int(time.Since(startedAt).Milliseconds())),
			slog.Int("status_code", httpresponse.StatusCodeFromError(err)),
		)
		httpresponse.JSONAppError(c, err)
		return
	}

	requestLogger.InfoContext(
		c.Request.Context(),
		"dashboard de vendedores carregado com sucesso",
		slog.Int("duration_ms", int(time.Since(startedAt).Milliseconds())),
		slog.Int("active_salespeople", response.Summary.ActiveSalespeople),
		slog.Int("total_budgets", response.Summary.TotalBudgets),
	)
	c.JSON(http.StatusOK, response)
}

func parseSalespeopleDashboardFilters(c *gin.Context) (*dto.DashboardSalespeopleFilters, error) {
	salespersonID, err := parseOptionalInt64(c.Query("salesperson_id"))
	if err != nil {
		return nil, err
	}
	year, err := parseOptionalInt(c.Query("year"))
	if err != nil {
		return nil, err
	}
	month, err := parseOptionalInt(c.Query("month"))
	if err != nil {
		return nil, err
	}

	return &dto.DashboardSalespeopleFilters{
		Month:         month,
		SalespersonID: salespersonID,
		SourceCompany: c.Query("source_company"),
		Year:          year,
	}, nil
}

func parseOptionalInt(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return nil, httpError("Valor inteiro invalido")
	}

	return &parsedValue, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, httpError("Valor inteiro invalido")
	}

	return &parsedValue, nil
}

func httpError(message string) error {
	return &requestError{message: message}
}

type requestError struct {
	message string
}

func (e *requestError) Error() string {
	return e.message
}
