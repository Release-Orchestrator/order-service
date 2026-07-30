package handler

import (
	"net/http"

	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/Release-Orchestrator/order-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	svc service.OrderServiceInterface
}

func NewOrderHandler(svc service.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.Create)
		orders.GET("", h.List)
		orders.GET("/:id", h.Get)
		orders.DELETE("/:id", h.Cancel)
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	order, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": order})
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_ID", "message": "invalid order ID"}})
		return
	}

	order, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

func (h *OrderHandler) List(c *gin.Context) {
	userID := c.Query("user_id")

	orders, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_ID", "message": "invalid order ID"}})
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OrderHandler) handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrOrderNotFound:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "ORDER_NOT_FOUND", "message": "order not found"}})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "USER_NOT_FOUND", "message": "user not found"}})
	case service.ErrPaymentFailed:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "PAYMENT_FAILED", "message": "payment failed"}})
	case service.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": "invalid input"}})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": "internal server error"}})
	}
}
