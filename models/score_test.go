package models

import (
	"testing"
	"time"
)

func TestLeaderboardEntry(t *testing.T) {
	entry := LeaderboardEntry{
		Rank:     1,
		UserID:   "user123",
		Username: "testuser",
		Score:    1000,
	}

	if entry.Rank != 1 {
		t.Errorf("Expected rank 1, got %d", entry.Rank)
	}

	if entry.UserID != "user123" {
		t.Errorf("Expected userID 'user123', got %s", entry.UserID)
	}

	if entry.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", entry.Username)
	}

	if entry.Score != 1000 {
		t.Errorf("Expected score 1000, got %f", entry.Score)
	}
}

func TestScoreCreation(t *testing.T) {
	now := time.Now()
	score := Score{
		UserID:    "user123",
		GameID:    "game456",
		Score:     500,
		Timestamp: now,
	}

	if score.UserID != "user123" {
		t.Errorf("Expected userID 'user123', got %s", score.UserID)
	}

	if score.GameID != "game456" {
		t.Errorf("Expected gameID 'game456', got %s", score.GameID)
	}

	if score.Score != 500 {
		t.Errorf("Expected score 500, got %f", score.Score)
	}

	if !score.Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, score.Timestamp)
	}
} 