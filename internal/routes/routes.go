package routes

import (
	"net/http"

	"github.com/Sarthak-Nagaria/ticket-system/internal/config"
	"github.com/Sarthak-Nagaria/ticket-system/internal/handlers"
	"github.com/Sarthak-Nagaria/ticket-system/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup wires up all application routes onto the provided Gin engine.
func Setup(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	authHandler := handlers.NewAuthHandler(db, cfg)
	ticketHandler := handlers.NewTicketHandler(db)

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Ticket System API is running",
		})
	})

	// Public health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Public auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected ticket routes
	tickets := router.Group("/tickets")
	tickets.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		tickets.POST("", ticketHandler.CreateTicket)
		tickets.GET("", ticketHandler.ListTickets)
		tickets.GET("/:id", ticketHandler.GetTicket)
		tickets.PATCH("/:id/status", ticketHandler.UpdateStatus)
	}
}
