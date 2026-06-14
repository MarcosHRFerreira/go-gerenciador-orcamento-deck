package budgetimport

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetimportservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetimport"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   budgetimportservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service budgetimportservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	adminRoutes := h.router.Group("/budget-imports")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("/preview", h.Preview)
	adminRoutes.POST("", h.ExecuteImport)
}

func (h *Handler) Preview(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, "Arquivo obrigatorio")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, "failed to open uploaded file")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, "failed to read uploaded file")
		return
	}

	options, err := parsePreviewOptions(c)
	if err != nil {
		httpresponse.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Preview(
		c.Request.Context(),
		fileHeader.Filename,
		fileData,
		options,
	)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ExecuteImport(c *gin.Context) {
	var req dto.ExecuteBudgetImportRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	response, err := h.service.ExecuteImport(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func parsePreviewOptions(c *gin.Context) (dto.PreviewBudgetImportOptions, error) {
	createMissingCatalogs, err := parseBoolDefaultTrue(c.PostForm("create_missing_catalogs"))
	if err != nil {
		return dto.PreviewBudgetImportOptions{}, httpError("create_missing_catalogs deve ser booleano")
	}

	useDefaultNotInformed, err := parseBoolDefaultTrue(c.PostForm("use_default_not_informed"))
	if err != nil {
		return dto.PreviewBudgetImportOptions{}, httpError("use_default_not_informed deve ser booleano")
	}

	return dto.PreviewBudgetImportOptions{
		DuplicateStrategy:     strings.TrimSpace(c.PostForm("duplicate_strategy")),
		CreateMissingCatalogs: createMissingCatalogs,
		UseDefaultNotInformed: useDefaultNotInformed,
	}, nil
}

func parseBoolDefaultTrue(value string) (bool, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return true, nil
	}

	parsedValue, err := strconv.ParseBool(normalizedValue)
	if err != nil {
		return false, err
	}

	return parsedValue, nil
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
