package user

import (
	"net/http"

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
