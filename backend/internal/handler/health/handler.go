package health

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

type Checker interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	router  gin.IRouter
	checker Checker
	timeout time.Duration
}

func NewHandler(router gin.IRouter, checker Checker, timeout time.Duration) *Handler {
	return &Handler{
		router:  router,
		checker: checker,
		timeout: timeout,
	}
}

func (h *Handler) RouteList() {
	h.router.GET("/check-health", h.CheckHealth)
}

func (h *Handler) CheckHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.checker.PingContext(ctx); err != nil {
		httpresponse.JSONError(c, 503, "database unavailable")
		return
	}

	httpresponse.JSON(c, 200, gin.H{
		"message": "service is healthy",
	})
}
