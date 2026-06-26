package main

import (
	"log"

	"github.com/Sarthak-Nagaria/ticket-system/internal/config"
	"github.com/Sarthak-Nagaria/ticket-system/internal/database"
	"github.com/Sarthak-Nagaria/ticket-system/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.GinMode)

	db := database.Init(cfg.DatabasePath)

	router := gin.Default()
	routes.Setup(router, db, cfg)

	log.Printf("Starting ticket system server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
