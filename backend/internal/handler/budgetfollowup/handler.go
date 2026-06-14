package budgetfollowup

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	budgetfollowupservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetfollowup"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   budgetfollowupservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service budgetfollowupservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	routes := h.router.Group("/budgets/:budget_id/follow-ups")
	routes.Use(middleware.Auth(h.secretKey))
	routes.GET("", h.List)
	routes.POST("", h.Create)
}

func (h *Handler) Create(c *gin.Context) {
	budgetID, ok := parseBudgetID(c)
	if !ok {
		return
	}

	var req dto.CreateBudgetFollowUpRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	id, err := h.service.Create(
		c.Request.Context(),
		budgetID,
		middleware.UserID(c),
		middleware.Role(c),
		middleware.Username(c),
		&req,
	)
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

	items, err := h.service.ListByBudgetID(
		c.Request.Context(),
		budgetID,
		middleware.Role(c),
		middleware.Username(c),
	)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func parseBudgetID(c *gin.Context) (int64, bool) {
	budgetID, err := strconv.ParseInt(c.Param("budget_id"), 10, 64)
	if err != nil || budgetID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "budget_id invalido")
		return 0, false
	}

	return budgetID, true
}
