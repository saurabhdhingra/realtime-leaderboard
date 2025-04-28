package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/user/realtime-leaderboard/config"
)

// Score represents a user's score entry
type Score struct {
	UserID    string    `json:"user_id"`
	GameID    string    `json:"game_id"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// ScoreSubmission represents score data submitted by a user
type ScoreSubmission struct {
	GameID string  `json:"game_id" binding:"required"`
	Score  float64 `json:"score" binding:"required"`
}

// LeaderboardEntry represents an entry in the leaderboard
type LeaderboardEntry struct {
	Rank     int64   `json:"rank"`
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
}

// SaveScore adds or updates a user's score in the leaderboard
func SaveScore(score *Score) error {
	// Set the timestamp
	score.Timestamp = time.Now()

	// Get the leaderboard key for the game
	leaderboardKey := fmt.Sprintf("leaderboard:%s", score.GameID)
	
	// Get the historical scores key for this user and game
	historyKey := fmt.Sprintf("history:%s:%s", score.UserID, score.GameID)
	
	// Add or update the score in the leaderboard (sorted set)
	err := config.RedisClient.ZAdd(config.Ctx, leaderboardKey, &redis.Z{
		Score:  score.Score,
		Member: score.UserID,
	}).Err()
	if err != nil {
		return err
	}
	
	// Serialize the score for historical record
	scoreJSON, err := json.Marshal(score)
	if err != nil {
		return err
	}
	
	// Add the score to the user's history (sorted set with timestamp as score)
	err = config.RedisClient.ZAdd(config.Ctx, historyKey, &redis.Z{
		Score:  float64(score.Timestamp.Unix()),
		Member: string(scoreJSON),
	}).Err()
	if err != nil {
		return err
	}
	
	// Add to global leaderboard
	globalKey := "leaderboard:global"
	return config.RedisClient.ZIncrBy(config.Ctx, globalKey, score.Score, score.UserID).Err()
}

// GetLeaderboard retrieves the leaderboard for a specific game
func GetLeaderboard(gameID string, start, end int64) ([]LeaderboardEntry, error) {
	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)
	
	// Get top scores with rank
	leaderboardData, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, leaderboardKey, start, end).Result()
	if err != nil {
		return nil, err
	}
	
	// Create leaderboard entries from data
	var entries []LeaderboardEntry
	for i, data := range leaderboardData {
		userID := data.Member.(string)
		
		// Get user for additional info
		user, err := GetUserByID(userID)
		if err != nil {
			continue
		}
		
		// Get rank (Redis ranks are 0-based, so add 1)
		rank, err := config.RedisClient.ZRevRank(config.Ctx, leaderboardKey, userID).Result()
		if err != nil {
			rank = int64(i) // Fallback to index if rank retrieval fails
		}
		
		entries = append(entries, LeaderboardEntry{
			Rank:     rank + 1, // Convert to 1-based ranking
			UserID:   userID,
			Username: user.Username,
			Score:    data.Score,
		})
	}
	
	return entries, nil
}

// GetGlobalLeaderboard retrieves the global leaderboard across all games
func GetGlobalLeaderboard(start, end int64) ([]LeaderboardEntry, error) {
	globalKey := "leaderboard:global"
	
	// Get top scores with rank
	leaderboardData, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, globalKey, start, end).Result()
	if err != nil {
		return nil, err
	}
	
	// Create leaderboard entries from data
	var entries []LeaderboardEntry
	for i, data := range leaderboardData {
		userID := data.Member.(string)
		
		// Get user for additional info
		user, err := GetUserByID(userID)
		if err != nil {
			continue
		}
		
		// Get rank (Redis ranks are 0-based, so add 1)
		rank, err := config.RedisClient.ZRevRank(config.Ctx, globalKey, userID).Result()
		if err != nil {
			rank = int64(i) // Fallback to index if rank retrieval fails
		}
		
		entries = append(entries, LeaderboardEntry{
			Rank:     rank + 1, // Convert to 1-based ranking
			UserID:   userID,
			Username: user.Username,
			Score:    data.Score,
		})
	}
	
	return entries, nil
}

// GetUserRank gets a user's rank in a specific game's leaderboard
func GetUserRank(userID, gameID string) (*LeaderboardEntry, error) {
	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)
	
	// Get user's score
	score, err := config.RedisClient.ZScore(config.Ctx, leaderboardKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	// Get user's rank
	rank, err := config.RedisClient.ZRevRank(config.Ctx, leaderboardKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	// Get user for additional info
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	return &LeaderboardEntry{
		Rank:     rank + 1, // Convert to 1-based ranking
		UserID:   userID,
		Username: user.Username,
		Score:    score,
	}, nil
}

// GetUserGlobalRank gets a user's rank in the global leaderboard
func GetUserGlobalRank(userID string) (*LeaderboardEntry, error) {
	globalKey := "leaderboard:global"
	
	// Get user's score
	score, err := config.RedisClient.ZScore(config.Ctx, globalKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	// Get user's rank
	rank, err := config.RedisClient.ZRevRank(config.Ctx, globalKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	// Get user for additional info
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	return &LeaderboardEntry{
		Rank:     rank + 1, // Convert to 1-based ranking
		UserID:   userID,
		Username: user.Username,
		Score:    score,
	}, nil
}

// GetUserScoreHistory gets a user's score history for a specific game
func GetUserScoreHistory(userID, gameID string, limit int64) ([]Score, error) {
	historyKey := fmt.Sprintf("history:%s:%s", userID, gameID)
	
	// Get score history sorted by time (newest first)
	results, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, historyKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	
	var scores []Score
	for _, result := range results {
		scoreJSON := result.Member.(string)
		var score Score
		if err := json.Unmarshal([]byte(scoreJSON), &score); err != nil {
			continue
		}
		scores = append(scores, score)
	}
	
	return scores, nil
}

// GetTopPlayersByPeriod gets the top players for a specific time period
func GetTopPlayersByPeriod(gameID string, startTime, endTime time.Time, limit int64) ([]LeaderboardEntry, error) {
	// Generate a temporary key for this period
	periodKey := fmt.Sprintf("leaderboard:%s:period:%d-%d", 
		gameID, startTime.Unix(), endTime.Unix())
	
	// Get the history keys for all users and aggregate scores within the period
	// This uses a Redis pipeline for efficiency
	pipe := config.RedisClient.Pipeline()
	
	// First, get all user scores for this game
	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)
	userScores, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, leaderboardKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	
	// Create a temporary sorted set for the period
	for _, userScore := range userScores {
		userID := userScore.Member.(string)
		historyKey := fmt.Sprintf("history:%s:%s", userID, gameID)
		
		// Get scores within the time period
		startScore := float64(startTime.Unix())
		endScore := float64(endTime.Unix())
		
		// Get score history for this period
		scores, err := config.RedisClient.ZRangeByScore(config.Ctx, historyKey, &redis.ZRangeBy{
			Min: strconv.FormatFloat(startScore, 'f', 0, 64),
			Max: strconv.FormatFloat(endScore, 'f', 0, 64),
		}).Result()
		
		if err != nil || len(scores) == 0 {
			continue
		}
		
		// Get the highest score in the period
		var highestScore float64
		for _, scoreData := range scores {
			var score Score
			if err := json.Unmarshal([]byte(scoreData), &score); err != nil {
				continue
			}
			if score.Score > highestScore {
				highestScore = score.Score
			}
		}
		
		// Add to the period leaderboard
		pipe.ZAdd(config.Ctx, periodKey, &redis.Z{
			Score:  highestScore,
			Member: userID,
		})
	}
	
	// Execute the pipeline
	_, err = pipe.Exec(config.Ctx)
	if err != nil {
		return nil, err
	}
	
	// Get the top players for this period
	periodScores, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, periodKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	
	// Create leaderboard entries
	var entries []LeaderboardEntry
	for i, data := range periodScores {
		userID := data.Member.(string)
		
		// Get user for additional info
		user, err := GetUserByID(userID)
		if err != nil {
			continue
		}
		
		entries = append(entries, LeaderboardEntry{
			Rank:     int64(i + 1),
			UserID:   userID,
			Username: user.Username,
			Score:    data.Score,
		})
	}
	
	// Clean up the temporary key
	config.RedisClient.Del(config.Ctx, periodKey)
	
	return entries, nil
} 