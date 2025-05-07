package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/realtime-leaderboard/models"
)

func SubmitScore(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var submission models.ScoreSubmission

	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score := &models.Score{
		UserID: userID.(string),
		GameID: submission.GameID,
		Score:  submission.Score,
	}

	if err := models.SaveScore(score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Score submitted successfully",
		"score":   score,
	})
}

func GetLeaderboard(c *gin.Context) {
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	start, _ := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 64)
	
	end := start + count - 1
	if end < 0 {
		end = 0
	}

	entries, err := models.GetLeaderboard(gameID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"game_id":     gameID,
		"start":       start,
		"count":       count,
	})
}

func GetGlobalLeaderboard(c *gin.Context) {
	start, _ := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 64)
	
	end := start + count - 1
	if end < 0 {
		end = 0
	}

	entries, err := models.GetGlobalLeaderboard(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve global leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"start":       start,
		"count":       count,
	})
}

func GetUserRanking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	entry, err := models.GetUserRank(userID.(string), gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ranking": entry,
		"game_id": gameID,
	})
}

func GetUserGlobalRanking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	entry, err := models.GetUserGlobalRank(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in global leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ranking": entry,
	})
}

func GetUserScoreHistory(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	scores, err := models.GetUserScoreHistory(userID.(string), gameID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve score history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": scores,
		"game_id": gameID,
		"user_id": userID,
		"limit":   limit,
	})
}

func GetTopPlayersByPeriod(c *gin.Context) {
	gameID := c.Param("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	periodStr := c.DefaultQuery("period", "day")
	
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

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	entries, err := models.GetTopPlayersByPeriod(gameID, startTime, endTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve top players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_players": entries,
		"game_id":     gameID,
		"period":      periodStr,
		"start_time":  startTime.Format(time.RFC3339),
		"end_time":    endTime.Format(time.RFC3339),
		"limit":       limit,
	})
} 