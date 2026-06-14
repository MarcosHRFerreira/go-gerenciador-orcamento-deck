package auth

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	authservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router      gin.IRouter
	validate    *validator.Validate
	authService authservice.Service
	secretKey   string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, authService authservice.Service, secretKey string) *Handler {
	return &Handler{
		router:      router,
		validate:    validate,
		authService: authService,
		secretKey:   secretKey,
	}
}

func (h *Handler) RouteList() {
	authRoutes := h.router.Group("/auth")
	authRoutes.POST("/register", h.Register)
	authRoutes.POST("/login", h.Login)

	refreshRoutes := h.router.Group("/auth")
	refreshRoutes.Use(middleware.AuthRefresh(h.secretKey))
	refreshRoutes.POST("/refresh", h.Refresh)
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID: userID,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	token, refreshToken, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	token, refreshToken, err := h.authService.Refresh(c.Request.Context(), &req, middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
