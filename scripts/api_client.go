package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	apiBaseURL = "http://localhost:8080/api"
)

// User represents a basic user for the API client
type User struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

// Score represents a score submission
type Score struct {
	GameID string  `json:"game_id"`
	Score  float64 `json:"score"`
}

// ApiClient is a simple client for the leaderboard API
type ApiClient struct {
	client *http.Client
}

// NewApiClient creates a new API client
func NewApiClient() *ApiClient {
	return &ApiClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterUser registers a new user
func (c *ApiClient) RegisterUser(user *User) error {
	// Create request body
	reqBody, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", apiBaseURL+"/auth/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register user: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var response struct {
		Message string `json:"message"`
		Token   string `json:"token"`
		User    *User  `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	// Update user with ID and token
	user.ID = response.User.ID
	user.Token = response.Token

	return nil
}

// LoginUser logs in a user
func (c *ApiClient) LoginUser(email, password string) (*User, error) {
	// Create request body
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequest("POST", apiBaseURL+"/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to login: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var response struct {
		Message string `json:"message"`
		Token   string `json:"token"`
		User    *User  `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Create user with token
	user := response.User
	user.Token = response.Token

	return user, nil
}

// SubmitScore submits a score for a user
func (c *ApiClient) SubmitScore(token string, score *Score) error {
	// Create request body
	reqBody, err := json.Marshal(score)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", apiBaseURL+"/leaderboard/score", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit score: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// GetLeaderboard gets the leaderboard for a game
func (c *ApiClient) GetLeaderboard(gameID string, start, count int) ([]map[string]interface{}, error) {
	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/leaderboard/game/%s?start=%d&count=%d", apiBaseURL, gameID, start, count), nil)
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get leaderboard: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var response struct {
		Leaderboard []map[string]interface{} `json:"leaderboard"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Leaderboard, nil
}

// GetUserRank gets a user's rank in a game
func (c *ApiClient) GetUserRank(token, gameID string) (map[string]interface{}, error) {
	// Create request
	req, err := http.NewRequest("GET", apiBaseURL+"/user/rank/"+gameID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user rank: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var response struct {
		Ranking map[string]interface{} `json:"ranking"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Ranking, nil
}

func main() {
	// Initialize the client
	client := NewApiClient()

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Test user registration and login
	fmt.Println("Testing user registration and login...")

	// Generate a unique user
	timestamp := time.Now().UnixNano()
	user := &User{
		Username: fmt.Sprintf("user%d", timestamp),
		Email:    fmt.Sprintf("user%d@example.com", timestamp),
		Password: "password123",
	}

	// Register the user
	fmt.Printf("Registering user: %s (%s)...\n", user.Username, user.Email)
	err := client.RegisterUser(user)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("User registered successfully with ID: %s\n", user.ID)

	// Login the user
	fmt.Printf("Logging in user: %s...\n", user.Email)
	loggedInUser, err := client.LoginUser(user.Email, user.Password)
	if err != nil {
		log.Fatalf("Failed to login user: %v", err)
	}
	fmt.Printf("User logged in successfully with ID: %s\n", loggedInUser.ID)

	// Test score submission
	fmt.Println("\nTesting score submission...")
	gameID := "game1"
	
	// Submit 5 random scores
	for i := 0; i < 5; i++ {
		// Generate a random score between 1 and 1000
		score := &Score{
			GameID: gameID,
			Score:  float64(rand.Intn(1000) + 1),
		}

		fmt.Printf("Submitting score for game %s: %.0f...\n", score.GameID, score.Score)
		err = client.SubmitScore(user.Token, score)
		if err != nil {
			log.Fatalf("Failed to submit score: %v", err)
		}
		fmt.Println("Score submitted successfully")
		
		// Wait a moment to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// Get the leaderboard
	fmt.Println("\nGetting leaderboard...")
	leaderboard, err := client.GetLeaderboard(gameID, 0, 10)
	if err != nil {
		log.Fatalf("Failed to get leaderboard: %v", err)
	}

	fmt.Printf("Leaderboard for game %s:\n", gameID)
	for i, entry := range leaderboard {
		fmt.Printf("%d. %s (Score: %.0f, Rank: %v)\n", 
			i+1, 
			entry["username"], 
			entry["score"].(float64), 
			int64(entry["rank"].(float64)))
	}

	// Get the user's rank
	fmt.Println("\nGetting user rank...")
	rank, err := client.GetUserRank(user.Token, gameID)
	if err != nil {
		log.Fatalf("Failed to get user rank: %v", err)
	}

	fmt.Printf("User %s rank in game %s: %v (Score: %.0f)\n", 
		user.Username, 
		gameID, 
		int64(rank["rank"].(float64)), 
		rank["score"].(float64))

	fmt.Println("\nAPI test completed successfully!")
} 