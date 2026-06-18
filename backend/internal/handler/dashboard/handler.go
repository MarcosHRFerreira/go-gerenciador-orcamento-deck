package dashboard

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
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
	protectedRoutes := h.router.Group("/dashboard")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("/salespeople", h.GetSalespeopleDashboard)
}

func (h *Handler) GetSalespeopleDashboard(c *gin.Context) {
	filters, err := parseSalespeopleDashboardFilters(c)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.GetSalespeopleDashboard(
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
