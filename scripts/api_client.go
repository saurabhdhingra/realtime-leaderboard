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

type User struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type Score struct {
	GameID string  `json:"game_id"`
	Score  float64 `json:"score"`
}

type ApiClient struct {
	client *http.Client
}

func NewApiClient() *ApiClient {
	return &ApiClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ApiClient) RegisterUser(user *User) error {
	reqBody, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/auth/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register user: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response struct {
		Message string `json:"message"`
		Token   string `json:"token"`
		User    *User  `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	user.ID = response.User.ID
	user.Token = response.Token

	return nil
}

func (c *ApiClient) LoginUser(email, password string) (*User, error) {
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to login: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response struct {
		Message string `json:"message"`
		Token   string `json:"token"`
		User    *User  `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	user := response.User
	user.Token = response.Token

	return user, nil
}

func (c *ApiClient) SubmitScore(token string, score *Score) error {
	reqBody, err := json.Marshal(score)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/leaderboard/score", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to submit score: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

func (c *ApiClient) GetLeaderboard(gameID string, start, count int) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/leaderboard/game/%s?start=%d&count=%d", apiBaseURL, gameID, start, count), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get leaderboard: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response struct {
		Leaderboard []map[string]interface{} `json:"leaderboard"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Leaderboard, nil
}

func (c *ApiClient) GetUserRank(token, gameID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiBaseURL+"/user/rank/"+gameID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user rank: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response struct {
		Ranking map[string]interface{} `json:"ranking"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Ranking, nil
}

func main() {
	client := NewApiClient()

	rand.Seed(time.Now().UnixNano())

	fmt.Println("Testing user registration and login...")

	timestamp := time.Now().UnixNano()
	user := &User{
		Username: fmt.Sprintf("user%d", timestamp),
		Email:    fmt.Sprintf("user%d@example.com", timestamp),
		Password: "password123",
	}

	fmt.Printf("Registering user: %s (%s)...\n", user.Username, user.Email)
	err := client.RegisterUser(user)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}
	fmt.Printf("User registered successfully with ID: %s\n", user.ID)

	fmt.Printf("Logging in user: %s...\n", user.Email)
	loggedInUser, err := client.LoginUser(user.Email, user.Password)
	if err != nil {
		log.Fatalf("Failed to login user: %v", err)
	}
	fmt.Printf("User logged in successfully with ID: %s\n", loggedInUser.ID)

	fmt.Println("\nTesting score submission...")
	gameID := "game1"
	
	for i := 0; i < 5; i++ {
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
		
		time.Sleep(100 * time.Millisecond)
	}

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