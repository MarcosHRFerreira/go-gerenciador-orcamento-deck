package budgetstatushistory

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	budgetstatushistoryservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatushistory"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   budgetstatushistoryservice.Service
	secretKey string
}

func NewHandler(
	router gin.IRouter,
	validate *validator.Validate,
	service budgetstatushistoryservice.Service,
	secretKey string,
) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	statusRoutes := h.router.Group("/budgets/:budget_id/status")
	statusRoutes.Use(middleware.Auth(h.secretKey))
	statusRoutes.PATCH("", h.ChangeStatus)

	historyRoutes := h.router.Group("/budgets/:budget_id/status-history")
	historyRoutes.Use(middleware.Auth(h.secretKey))
	historyRoutes.GET("", h.List)
}

func (h *Handler) ChangeStatus(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	var req dto.ChangeBudgetStatusRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	id, err := h.service.ChangeStatus(c.Request.Context(), budgetID, middleware.UserID(c), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) List(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	items, err := h.service.ListByBudgetID(c.Request.Context(), budgetID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func parseBudgetID(c *gin.Context) (int64, bool) {
	budgetID, err := strconv.ParseInt(c.Param("budget_id"), 10, 64)
	if err != nil || budgetID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "invalid budget_id")
		return 0, false
	}

	return budgetID, true
}
