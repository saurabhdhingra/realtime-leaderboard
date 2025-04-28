package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/user/realtime-leaderboard/config"
	"github.com/user/realtime-leaderboard/handlers"
	"github.com/user/realtime-leaderboard/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Redis connection
	if err := config.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add metrics middleware to track request metrics
	router.Use(middleware.MetricsMiddleware())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", middleware.MetricsHandler)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// User routes (auth required)
		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", handlers.GetProfile)
			user.GET("/rank/:gameID", handlers.GetUserRanking)
			user.GET("/global-rank", handlers.GetUserGlobalRanking)
			user.GET("/history/:gameID", handlers.GetUserScoreHistory)
		}

		// Leaderboard routes
		leaderboard := api.Group("/leaderboard")
		{
			// Public routes (no auth required)
			leaderboard.GET("/game/:gameID", handlers.GetLeaderboard)
			leaderboard.GET("/global", handlers.GetGlobalLeaderboard)
			leaderboard.GET("/top/:gameID", handlers.GetTopPlayersByPeriod)

			// Protected routes (auth required)
			leaderboard.Use(middleware.AuthMiddleware())
			leaderboard.POST("/score", handlers.SubmitScore)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start metrics reporting in background
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			middleware.PrintMetrics()
		}
	}()

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 