package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/realtime-leaderboard/models"
)

// SubmitScore handles score submission
func SubmitScore(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var submission models.ScoreSubmission
	// Bind request body to submission struct
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create score entry
	score := &models.Score{
		UserID: userID.(string),
		GameID: submission.GameID,
		Score:  submission.Score,
	}

	// Save score to leaderboard
	if err := models.SaveScore(score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save score"})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Score submitted successfully",
		"score":   score,
	})
}

// GetLeaderboard retrieves the leaderboard for a specific game
func GetLeaderboard(c *gin.Context) {
	// Get game ID from URL parameter
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	// Get pagination parameters (default: start=0, count=10)
	start, _ := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 64)
	
	// Calculate end index
	end := start + count - 1
	if end < 0 {
		end = 0
	}

	// Get leaderboard entries
	entries, err := models.GetLeaderboard(gameID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}

	// Return leaderboard
	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"game_id":     gameID,
		"start":       start,
		"count":       count,
	})
}

// GetGlobalLeaderboard retrieves the global leaderboard
func GetGlobalLeaderboard(c *gin.Context) {
	// Get pagination parameters (default: start=0, count=10)
	start, _ := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 64)
	
	// Calculate end index
	end := start + count - 1
	if end < 0 {
		end = 0
	}

	// Get global leaderboard entries
	entries, err := models.GetGlobalLeaderboard(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve global leaderboard"})
		return
	}

	// Return global leaderboard
	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"start":       start,
		"count":       count,
	})
}

// GetUserRanking retrieves a user's ranking in a specific game's leaderboard
func GetUserRanking(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get game ID from URL parameter
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	// Get user's rank
	entry, err := models.GetUserRank(userID.(string), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in leaderboard"})
		return
	}

	// Return user's rank
	c.JSON(http.StatusOK, gin.H{
		"ranking": entry,
		"game_id": gameID,
	})
}

// GetUserGlobalRanking retrieves a user's ranking in the global leaderboard
func GetUserGlobalRanking(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get user's global rank
	entry, err := models.GetUserGlobalRank(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in global leaderboard"})
		return
	}

	// Return user's global rank
	c.JSON(http.StatusOK, gin.H{
		"ranking": entry,
	})
}

// GetUserScoreHistory retrieves a user's score history for a specific game
func GetUserScoreHistory(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get game ID from URL parameter
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	// Get limit parameter (default: 10)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	// Get user's score history
	scores, err := models.GetUserScoreHistory(userID.(string), gameID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve score history"})
		return
	}

	// Return user's score history
	c.JSON(http.StatusOK, gin.H{
		"history": scores,
		"game_id": gameID,
		"user_id": userID,
		"limit":   limit,
	})
}

// GetTopPlayersByPeriod retrieves the top players for a specific time period
func GetTopPlayersByPeriod(c *gin.Context) {
	// Get game ID from URL parameter
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	// Get period parameters
	periodStr := c.DefaultQuery("period", "day")
	
	// Define time periods
	var startTime, endTime time.Time
	endTime = time.Now()
	
	switch periodStr {
	case "day":
		startTime = endTime.AddDate(0, 0, -1)
	case "week":
		startTime = endTime.AddDate(0, 0, -7)
	case "month":
		startTime = endTime.AddDate(0, -1, 0)
	case "year":
		startTime = endTime.AddDate(-1, 0, 0)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Valid values: day, week, month, year"})
		return
	}

	// Get limit parameter (default: 10)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	// Get top players for the period
	entries, err := models.GetTopPlayersByPeriod(gameID, startTime, endTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve top players"})
		return
	}

	// Return top players
	c.JSON(http.StatusOK, gin.H{
		"top_players": entries,
		"game_id":     gameID,
		"period":      periodStr,
		"start_time":  startTime.Format(time.RFC3339),
		"end_time":    endTime.Format(time.RFC3339),
		"limit":       limit,
	})
} 