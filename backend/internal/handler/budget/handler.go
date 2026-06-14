package budget

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budget"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   budgetservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service budgetservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/budgets")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)
	protectedRoutes.GET("/:budget_id", h.GetByID)

	adminRoutes := h.router.Group("/budgets")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:budget_id", h.Update)
	adminRoutes.DELETE("/:budget_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateBudgetRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	id, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) List(c *gin.Context) {
	filters, err := parseListFilters(c)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.service.List(
		c.Request.Context(),
		filters,
		middleware.Role(c),
		middleware.Username(c),
	)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(
		c.Request.Context(),
		budgetID,
		middleware.Role(c),
		middleware.Username(c),
	)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	var req dto.UpdateBudgetRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), budgetID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), budgetID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseListFilters(c *gin.Context) (*dto.ListBudgetsFilters, error) {
	yearBudget, err := parseOptionalInt(c.Query("year_budget"))
	if err != nil {
		return nil, err
	}
	statusID, err := parseOptionalInt64(c.Query("status_id"))
	if err != nil {
		return nil, err
	}
	salespersonID, err := parseOptionalInt64(c.Query("salesperson_id"))
	if err != nil {
		return nil, err
	}
	installerID, err := parseOptionalInt64(c.Query("installer_id"))
	if err != nil {
		return nil, err
	}
	priorityID, err := parseOptionalInt64(c.Query("priority_id"))
	if err != nil {
		return nil, err
	}
	projectTypeID, err := parseOptionalInt64(c.Query("project_type_id"))
	if err != nil {
		return nil, err
	}
	sentAtFrom, err := parseOptionalTime(c.Query("sent_at_from"))
	if err != nil {
		return nil, err
	}
	sentAtTo, err := parseOptionalTime(c.Query("sent_at_to"))
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
	page, err := parseOptionalIntValue(c.Query("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := parseOptionalIntValue(c.Query("page_size"), 20)
	if err != nil {
		return nil, err
	}

	return &dto.ListBudgetsFilters{
		BudgetNumber:   c.Query("budget_number"),
		YearBudget:     yearBudget,
		StatusID:       statusID,
		SalespersonID:  salespersonID,
		InstallerID:    installerID,
		PriorityID:     priorityID,
		ProjectTypeID:  projectTypeID,
		DesignerName:   c.Query("designer_name"),
		CompetitorName: c.Query("competitor_name"),
		SentAtFrom:     sentAtFrom,
		SentAtTo:       sentAtTo,
		GrossValueMin:  grossValueMin,
		GrossValueMax:  grossValueMax,
		Page:           page,
		PageSize:       pageSize,
		SortBy:         c.Query("sort_by"),
		SortOrder:      c.Query("sort_order"),
	}, nil
}

func parseBudgetID(c *gin.Context) (int64, bool) {
	budgetID, err := strconv.ParseInt(c.Param("budget_id"), 10, 64)
	if err != nil || budgetID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "invalid budget_id")
		return 0, false
	}

	return budgetID, true
}

func parseOptionalInt(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return nil, httpError("invalid integer value")
	}

	return &parsedValue, nil
}

func parseOptionalIntValue(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, httpError("invalid integer value")
	}

	return parsedValue, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, httpError("invalid integer value")
	}

	return &parsedValue, nil
}

func parseOptionalFloat64(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, httpError("invalid numeric value")
	}

	return &parsedValue, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	layouts := []string{time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		parsedValue, err := time.Parse(layout, value)
		if err == nil {
			return &parsedValue, nil
		}
	}

	return nil, httpError("invalid date value")
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
