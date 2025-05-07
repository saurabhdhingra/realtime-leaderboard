package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/user/realtime-leaderboard/config"
)

type Score struct {
	UserID    string    `json:"user_id"`
	GameID    string    `json:"game_id"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

type ScoreSubmission struct {
	GameID string  `json:"game_id" binding:"required"`
	Score  float64 `json:"score" binding:"required"`
}

type LeaderboardEntry struct {
	Rank     int64   `json:"rank"`
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
}

func SaveScore(score *Score) error {
	score.Timestamp = time.Now()

	leaderboardKey := fmt.Sprintf("leaderboard:%s", score.GameID)

	historyKey := fmt.Sprintf("history:%s:%s", score.UserID, score.GameID)

	err := config.RedisClient.ZAdd(config.Ctx, leaderboardKey, &redis.Z{
		Score:  score.Score,
		Member: score.UserID,
	}).Err()
	if err != nil {
		return err
	}
	
	scoreJSON, err := json.Marshal(score)
	if err != nil {
		return err
	}

	err = config.RedisClient.ZAdd(config.Ctx, historyKey, &redis.Z{
		Score:  float64(score.Timestamp.Unix()),
		Member: string(scoreJSON),
	}).Err()
	if err != nil {
		return err
	}

	globalKey := "leaderboard:global"
	return config.RedisClient.ZIncrBy(config.Ctx, globalKey, score.Score, score.UserID).Err()
}

func GetLeaderboard(gameID string, start, end int64) ([]LeaderboardEntry, error) {
	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)
	
	leaderboardData, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, leaderboardKey, start, end).Result()
	if err != nil {
		return nil, err
	}
	
	var entries []LeaderboardEntry
	for i, data := range leaderboardData {
		userID := data.Member.(string)
		
		user, err := GetUserByID(userID)
		if err != nil {
			continue
		}

		rank, err := config.RedisClient.ZRevRank(config.Ctx, leaderboardKey, userID).Result()
		if err != nil {
			rank = int64(i) 
		
		entries = append(entries, LeaderboardEntry{
			Rank:     rank + 1,
			UserID:   userID,
			Username: user.Username,
			Score:    data.Score,
		})
	}
	
	return entries, nil
}


func GetGlobalLeaderboard(start, end int64) ([]LeaderboardEntry, error) {
	globalKey := "leaderboard:global"
	
	leaderboardData, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, globalKey, start, end).Result()
	if err != nil {
		return nil, err
	}
	
	var entries []LeaderboardEntry
	for i, data := range leaderboardData {
		userID := data.Member.(string)
		
		user, err := GetUserByID(userID)
		if err != nil {
			continue
		}
		
		rank, err := config.RedisClient.ZRevRank(config.Ctx, globalKey, userID).Result()
		if err != nil {
			rank = int64(i)
		}
		
		entries = append(entries, LeaderboardEntry{
			Rank:     rank + 1,
			UserID:   userID,
			Username: user.Username,
			Score:    data.Score,
		})
	}
	
	return entries, nil
}

func GetUserRank(userID, gameID string) (*LeaderboardEntry, error) {
	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)

	score, err := config.RedisClient.ZScore(config.Ctx, leaderboardKey, userID).Result()
	if err != nil {
		return nil, err
	}

	rank, err := config.RedisClient.ZRevRank(config.Ctx, leaderboardKey, userID).Result()
	if err != nil {
		return nil, err
	}

	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	return &LeaderboardEntry{
		Rank:     rank + 1,
		UserID:   userID,
		Username: user.Username,
		Score:    score,
	}, nil
}

func GetUserGlobalRank(userID string) (*LeaderboardEntry, error) {
	globalKey := "leaderboard:global"
	
	score, err := config.RedisClient.ZScore(config.Ctx, globalKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	rank, err := config.RedisClient.ZRevRank(config.Ctx, globalKey, userID).Result()
	if err != nil {
		return nil, err
	}
	
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	return &LeaderboardEntry{
		Rank:     rank + 1,
		UserID:   userID,
		Username: user.Username,
		Score:    score,
	}, nil
}

func GetUserScoreHistory(userID, gameID string, limit int64) ([]Score, error) {
	historyKey := fmt.Sprintf("history:%s:%s", userID, gameID)

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

func GetTopPlayersByPeriod(gameID string, startTime, endTime time.Time, limit int64) ([]LeaderboardEntry, error) {
	periodKey := fmt.Sprintf("leaderboard:%s:period:%d-%d", 
		gameID, startTime.Unix(), endTime.Unix())
	
	pipe := config.RedisClient.Pipeline()

	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)
	userScores, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, leaderboardKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	for _, userScore := range userScores {
		userID := userScore.Member.(string)
		historyKey := fmt.Sprintf("history:%s:%s", userID, gameID)

		startScore := float64(startTime.Unix())
		endScore := float64(endTime.Unix())

		scores, err := config.RedisClient.ZRangeByScore(config.Ctx, historyKey, &redis.ZRangeBy{
			Min: strconv.FormatFloat(startScore, 'f', 0, 64),
			Max: strconv.FormatFloat(endScore, 'f', 0, 64),
		}).Result()
		
		if err != nil || len(scores) == 0 {
			continue
		}
		
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
		
		pipe.ZAdd(config.Ctx, periodKey, &redis.Z{
			Score:  highestScore,
			Member: userID,
		})
	}
	
	_, err = pipe.Exec(config.Ctx)
	if err != nil {
		return nil, err
	}

	periodScores, err := config.RedisClient.ZRevRangeWithScores(config.Ctx, periodKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	var entries []LeaderboardEntry
	for i, data := range periodScores {
		userID := data.Member.(string)
		
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
	
	config.RedisClient.Del(config.Ctx, periodKey)
	
	return entries, nil
} 