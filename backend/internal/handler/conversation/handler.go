package conversation

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	conversationservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/conversation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router              gin.IRouter
	validate            *validator.Validate
	conversationService conversationservice.Service
	secretKey           string
}

func NewHandler(
	router gin.IRouter,
	validate *validator.Validate,
	conversationService conversationservice.Service,
	secretKey string,
) *Handler {
	return &Handler{
		router:              router,
		validate:            validate,
		conversationService: conversationService,
		secretKey:           secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/conversations")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.ListConversations)
	protectedRoutes.GET("/available-users", h.ListAvailableUsers)
	protectedRoutes.GET("/unread-count", h.GetUnreadCount)
	protectedRoutes.GET("/:conversation_id/messages", h.ListMessages)
	protectedRoutes.POST("", h.CreateConversation)
	protectedRoutes.POST("/:conversation_id/messages", h.SendMessage)
	protectedRoutes.PATCH("/:conversation_id/read", h.MarkAsRead)
}

func (h *Handler) CreateConversation(c *gin.Context) {
	var req dto.CreateConversationRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	conversationID, err := h.conversationService.Create(c.Request.Context(), middleware.UserID(c), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateConversationResponse{ID: conversationID})
}

func (h *Handler) ListConversations(c *gin.Context) {
	items, err := h.conversationService.ListByUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) ListAvailableUsers(c *gin.Context) {
	items, err := h.conversationService.ListAvailableUsers(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	count, err := h.conversationService.CountUnreadByUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ConversationUnreadCountResponse{Count: count})
}

func (h *Handler) ListMessages(c *gin.Context) {
	conversationID, ok := parseConversationID(c)
	if !ok {
		return
	}

	items, err := h.conversationService.ListMessagesByConversation(c.Request.Context(), middleware.UserID(c), conversationID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) SendMessage(c *gin.Context) {
	conversationID, ok := parseConversationID(c)
	if !ok {
		return
	}

	var req dto.SendConversationMessageRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	messageID, err := h.conversationService.SendMessage(c.Request.Context(), middleware.UserID(c), conversationID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.SendConversationMessageResponse{ID: messageID})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	conversationID, ok := parseConversationID(c)
	if !ok {
		return
	}

	if err := h.conversationService.MarkAsRead(c.Request.Context(), middleware.UserID(c), conversationID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseConversationID(c *gin.Context) (int64, bool) {
	rawID := strings.TrimSpace(c.Param("conversation_id"))
	conversationID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || conversationID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "conversation_id deve ser um inteiro valido")
		return 0, false
	}

	return conversationID, true
}
