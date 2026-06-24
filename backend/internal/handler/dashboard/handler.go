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
	adminRoutes.GET("/salespeople/gross-value-range", h.GetSalespeopleDashboardGrossValueRange)
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
		slog.Bool("has_installer_filter", filters.InstallerID != nil),
		slog.Bool("has_salesperson_filter", filters.SalespersonID != nil),
		slog.Bool("has_status_filter", filters.StatusID != nil),
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

func (h *Handler) GetSalespeopleDashboardGrossValueRange(c *gin.Context) {
	filters, err := parseSalespeopleDashboardFilters(c)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.GetGrossValueRange(
		c.Request.Context(),
		filters,
		middleware.Role(c),
		middleware.Username(c),
	)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseSalespeopleDashboardFilters(c *gin.Context) (*dto.DashboardSalespeopleFilters, error) {
	installerID, err := parseOptionalInt64(c.Query("installer_id"))
	if err != nil {
		return nil, err
	}
	salespersonID, err := parseOptionalInt64(c.Query("salesperson_id"))
	if err != nil {
		return nil, err
	}
	statusID, err := parseOptionalInt64(c.Query("status_id"))
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
	grossValueMin, err := parseOptionalFloat64(c.Query("gross_value_min"))
	if err != nil {
		return nil, err
	}
	grossValueMax, err := parseOptionalFloat64(c.Query("gross_value_max"))
	if err != nil {
		return nil, err
	}

	return &dto.DashboardSalespeopleFilters{
		InstallerID:   installerID,
		Month:         month,
		SalespersonID: salespersonID,
		SourceCompany: c.Query("source_company"),
		StatusID:      statusID,
		Year:          year,
		GrossValueMin: grossValueMin,
		GrossValueMax: grossValueMax,
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

func parseOptionalFloat64(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, httpError("Valor numerico invalido")
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
