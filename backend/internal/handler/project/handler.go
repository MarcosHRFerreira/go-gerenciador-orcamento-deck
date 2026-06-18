package project

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	projectservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/project"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   projectservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service projectservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	adminRoutes := h.router.Group("/projects")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.GET("/next-code", h.GetNextCode)
	adminRoutes.GET("", h.List)
	adminRoutes.GET("/:project_id", h.GetByID)
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:project_id", h.Update)
	adminRoutes.DELETE("/:project_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateProjectRequest
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

func (h *Handler) GetNextCode(c *gin.Context) {
	code, err := h.service.GetNextCode(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code})
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), projectID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req dto.UpdateProjectRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), projectID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), projectID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseProjectID(c *gin.Context) (int64, bool) {
	rawID := c.Param("project_id")
	projectID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || projectID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "project_id deve ser um inteiro valido")
		return 0, false
	}

	return projectID, true
}
