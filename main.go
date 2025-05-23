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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	if err := config.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	router := gin.Default()

	router.Use(middleware.MetricsMiddleware())

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

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/metrics", middleware.MetricsHandler)


	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", handlers.GetProfile)
			user.GET("/rank/:gameID", handlers.GetUserRanking)
			user.GET("/global-rank", handlers.GetUserGlobalRanking)
			user.GET("/history/:gameID", handlers.GetUserScoreHistory)
		}

		leaderboard := api.Group("/leaderboard")
		{
			leaderboard.GET("/game/:gameID", handlers.GetLeaderboard)
			leaderboard.GET("/global", handlers.GetGlobalLeaderboard)
			leaderboard.GET("/top/:gameID", handlers.GetTopPlayersByPeriod)

			leaderboard.Use(middleware.AuthMiddleware())
			leaderboard.POST("/score", handlers.SubmitScore)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			middleware.PrintMetrics()
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 