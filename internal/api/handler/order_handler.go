package handler

import (
	"cache-app/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// GetOrderByUID godoc
// @Summary Получение заказа по UID
// @Description Возвращает заказ по его уникальному идентификатору (UID)
// @Tags orders
// @Accept  json
// @Produce  json
// @Param id path string true "UID заказа"
// @Success 200 {object} response.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderByUID(c *gin.Context) {
	print("GetOrderByUID")
	orderUID := c.Param("id")
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	order, err := h.orderService.GetOrderByUID(c.Request.Context(), orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// HealthCheck godoc
// @Summary Проверка состояния сервиса
// @Description Возвращает статус, что сервис работает
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Order service is running",
	})
}
