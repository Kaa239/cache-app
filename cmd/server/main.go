package main

import (
	_ "cache-app/docs"
	"cache-app/internal/api"
	"cache-app/internal/config"
	"cache-app/internal/kafka"
	"cache-app/internal/repository"
	"cache-app/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// @title Orders API
// @version 1.0
// @description API for managing orders
// @host localhost:8080
func main() {
	cfg := config.Load()

	// Подключаемся к БД
	ctx := context.Background()
	db, err := repository.NewPostgresDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Инициализация репозитория и сервиса
	orderRepository := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepository)

	// Инициализация Kafka consumer
	kafkaConsumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, orderService)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()
	go kafkaConsumer.Start()

	// Настройка HTTP сервера
	router := gin.Default()
	apiRouter := api.NewRouter(orderService)
	apiRouter.SetupRoutes(router)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Запуск HTTP сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.HTTPPort)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", cfg.HTTPPort)

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")

}
