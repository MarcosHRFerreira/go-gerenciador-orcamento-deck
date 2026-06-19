package productline

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	productlineservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/productline"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	router    gin.IRouter
	service   productlineservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, service productlineservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/product-lines")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}
