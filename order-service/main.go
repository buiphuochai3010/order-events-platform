package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"order-service/internal/config"
	"order-service/internal/db"
	"order-service/internal/handlers"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.InitSchema(ctx, pool); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	orderHandler := handlers.NewOrderHandler(pool)

	router := gin.Default()
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.GET("/orders", orderHandler.ListOrders)
	}

	log.Printf("order-service listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
