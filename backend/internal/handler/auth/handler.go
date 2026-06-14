package auth

import (
	"net/http"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
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
	cfg         *config.Config
}

func NewHandler(router gin.IRouter, validate *validator.Validate, authService authservice.Service, cfg *config.Config) *Handler {
	return &Handler{
		router:      router,
		validate:    validate,
		authService: authService,
		secretKey:   cfg.SecretJWT,
		cfg:         cfg,
	}
}

func (h *Handler) RouteList() {
	authRoutes := h.router.Group("/auth")
	authRoutes.Use(middleware.RateLimit(10, time.Minute, "auth-public"))
	authRoutes.POST("/register", h.Register)
	authRoutes.POST("/login", h.Login)
	authRoutes.POST("/refresh", h.Refresh)
	authRoutes.POST("/logout", h.Logout)

	protectedRoutes := h.router.Group("/auth")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.Use(middleware.RateLimit(20, time.Minute, "auth-protected"))
	protectedRoutes.PATCH("/change-password", h.ChangePassword)
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), &req, c.GetHeader("X-Setup-Token"))
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

	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, dto.LoginResponse{
		Token: token,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	token, refreshToken, err := h.authService.Refresh(c.Request.Context(), h.readRefreshCookie(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		Token: token,
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	token, refreshToken, err := h.authService.ChangePassword(c.Request.Context(), middleware.UserID(c), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, dto.ChangePasswordResponse{
		Token: token,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.authService.Logout(c.Request.Context(), h.readRefreshCookie(c)); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) setRefreshCookie(c *gin.Context, refreshToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.cfg.RefreshCookieName,
		refreshToken,
		int((7 * 24 * time.Hour).Seconds()),
		"/",
		h.cfg.RefreshCookieDomain,
		h.cfg.RefreshCookieSecure,
		true,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.cfg.RefreshCookieName,
		"",
		-1,
		"/",
		h.cfg.RefreshCookieDomain,
		h.cfg.RefreshCookieSecure,
		true,
	)
}

func (h *Handler) readRefreshCookie(c *gin.Context) string {
	refreshToken, err := c.Cookie(h.cfg.RefreshCookieName)
	if err != nil {
		return ""
	}

	return refreshToken
}
