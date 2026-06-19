package project

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	projectRoutes := h.router.Group("/projects")
	projectRoutes.Use(middleware.Auth(h.secretKey))
	projectRoutes.Use(middleware.RequireRoles(model.RoleAdmin, model.RoleUser))
	projectRoutes.GET("/next-code", h.GetNextCode)
	projectRoutes.GET("", h.List)
	projectRoutes.GET("/:project_id", h.GetByID)
	projectRoutes.POST("", h.Create)
	projectRoutes.PUT("/:project_id", h.Update)
	projectRoutes.DELETE("/:project_id", h.Delete)
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
	// #region debug-point A:projects-list-entry
	reportProjectDebugEvent(
		"A",
		"project/handler.go:List:entry",
		"[DEBUG] GET /projects entry",
		map[string]interface{}{
			"path":     c.FullPath(),
			"role":     string(middleware.Role(c)),
			"username": middleware.Username(c),
			"userId":   middleware.UserID(c),
		},
	)
	// #endregion
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		// #region debug-point B:projects-list-error
		reportProjectDebugEvent(
			"B",
			"project/handler.go:List:error",
			"[DEBUG] GET /projects failed",
			map[string]interface{}{
				"path":     c.FullPath(),
				"role":     string(middleware.Role(c)),
				"username": middleware.Username(c),
				"userId":   middleware.UserID(c),
				"error":    err.Error(),
			},
		)
		// #endregion
		httpresponse.JSONAppError(c, err)
		return
	}

	// #region debug-point C:projects-list-success
	reportProjectDebugEvent(
		"C",
		"project/handler.go:List:success",
		"[DEBUG] GET /projects success",
		map[string]interface{}{
			"path":      c.FullPath(),
			"role":      string(middleware.Role(c)),
			"username":  middleware.Username(c),
			"userId":    middleware.UserID(c),
			"itemCount": len(items),
		},
	)
	// #endregion
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

func reportProjectDebugEvent(hypothesisID string, location string, message string, data map[string]interface{}) {
	debugServerURL := "http://127.0.0.1:7777/event"
	debugSessionID := "projects-user-load"

	if envContent, err := os.ReadFile(".dbg/projects-user-load.env"); err == nil {
		for _, line := range strings.Split(string(envContent), "\n") {
			switch {
			case strings.HasPrefix(line, "DEBUG_SERVER_URL="):
				debugServerURL = strings.TrimSpace(strings.TrimPrefix(line, "DEBUG_SERVER_URL="))
			case strings.HasPrefix(line, "DEBUG_SESSION_ID="):
				debugSessionID = strings.TrimSpace(strings.TrimPrefix(line, "DEBUG_SESSION_ID="))
			}
		}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"sessionId":    debugSessionID,
		"runId":        "pre-fix",
		"hypothesisId": hypothesisID,
		"location":     location,
		"msg":          message,
		"data":         data,
		"ts":           time.Now().UnixMilli(),
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, debugServerURL, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 500 * time.Millisecond}
	response, err := client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
}
