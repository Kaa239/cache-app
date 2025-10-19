package api

import (
	"cache-app/internal/api/handler"
	"cache-app/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	orderHandler *handler.OrderHandler
}

func NewRouter(orderService service.OrderService) *Router {
	return &Router{
		orderHandler: handler.NewOrderHandler(orderService),
	}
}

func (r *Router) SetupRoutes(router *gin.Engine) {
	// Swagger документация
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", r.orderHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/orders/:id", r.orderHandler.GetOrderByUID)
	}
}
