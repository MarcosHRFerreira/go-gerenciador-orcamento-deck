package user

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	userservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router      gin.IRouter
	validate    *validator.Validate
	userService userservice.Service
	secretKey   string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, userService userservice.Service, secretKey string) *Handler {
	return &Handler{
		router:      router,
		validate:    validate,
		userService: userService,
		secretKey:   secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/users")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("/me", h.GetMe)

	adminRoutes := h.router.Group("/users")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("", h.CreateUser)
	adminRoutes.GET("", h.ListUsers)
	adminRoutes.PUT("/:user_id", h.UpdateUser)
	adminRoutes.PATCH("/:user_id/role", h.UpdateRole)
	adminRoutes.PATCH("/:user_id/active", h.UpdateActive)
	adminRoutes.PATCH("/:user_id/reset-password", h.ResetPassword)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateUserResponse{
		ID: userID,
	})
}

func (h *Handler) GetMe(c *gin.Context) {
	user, err := h.userService.GetMe(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userService.List(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.userService.Update(c.Request.Context(), middleware.UserID(c), userID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserRoleRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.userService.UpdateRole(c.Request.Context(), middleware.UserID(c), userID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateActive(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserActiveRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.userService.UpdateActive(c.Request.Context(), middleware.UserID(c), userID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.ResetUserPasswordRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), middleware.UserID(c), userID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUserID(c *gin.Context) (int64, bool) {
	rawID := c.Param("user_id")
	userID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || userID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "user_id deve ser um inteiro valido")
		return 0, false
	}

	return userID, true
}
