package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	apiBaseURL = "http://localhost:8080/api"
	concurrentUsers = 10
	requestsPerUser = 50
	requestDelay = 50
)

type User struct {
	ID       string
	Username string
	Email    string
	Password string
	Token    string
}

type Score struct {
	GameID string  `json:"game_id"`
	Score  float64 `json:"score"`
}

type Stats struct {
	TotalRequests     int
	SuccessfulRequests int
	FailedRequests    int
	TotalTime         time.Duration
	AverageTime       time.Duration
	MinTime           time.Duration
	MaxTime           time.Duration
	mu                sync.Mutex
}

func (s *Stats) AddRequest(success bool, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	
	if success {
		s.SuccessfulRequests++
	} else {
		s.FailedRequests++
	}
	
	s.TotalTime += duration
	
	if s.MinTime == 0 || duration < s.MinTime {
		s.MinTime = duration
	}
	
	if duration > s.MaxTime {
		s.MaxTime = duration
	}
}

func (s *Stats) CalculateAverage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.TotalRequests > 0 {
		s.AverageTime = s.TotalTime / time.Duration(s.TotalRequests)
	}
}

func registerUser(username, email, password string) (*User, error) {
	user := &User{
		Username: username,
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/auth/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to register user: status %d", resp.StatusCode)
	}

	var response struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	user.ID = response.User.ID
	user.Token = response.Token

	return user, nil
}

func submitScore(token string, gameID string, score float64, stats *Stats) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"game_id": gameID,
		"score":   score,
	})
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/leaderboard/score", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)
	
	success := err == nil && resp.StatusCode == http.StatusOK
	stats.AddRequest(success, duration)
	
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response: %d", resp.StatusCode)
	}
}

func getLeaderboard(gameID string, stats *Stats) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/leaderboard/game/%s", apiBaseURL, gameID), nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)
	
	success := err == nil && resp.StatusCode == http.StatusOK
	stats.AddRequest(success, duration)
	
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response: %d", resp.StatusCode)
	}
}

func runUserWorkload(user *User, gameID string, wg *sync.WaitGroup, stats *Stats) {
	defer wg.Done()
	
	for i := 0; i < requestsPerUser; i++ {
		score := float64(rand.Intn(1000) + 1)

		if rand.Float32() < 0.7 {
			submitScore(user.Token, gameID, score, stats)
		} else {
			getLeaderboard(gameID, stats)
		}

		time.Sleep(time.Duration(requestDelay) * time.Millisecond)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	stats := &Stats{
		MinTime: time.Hour,
	}
	
	fmt.Println("Realtime Leaderboard Load Test")
	fmt.Println("==============================")
	fmt.Printf("Concurrent users: %d\n", concurrentUsers)
	fmt.Printf("Requests per user: %d\n", requestsPerUser)
	fmt.Printf("Total requests: %d\n", concurrentUsers*requestsPerUser)
	fmt.Println("==============================")

	users := make([]*User, concurrentUsers)
	for i := 0; i < concurrentUsers; i++ {
		username := fmt.Sprintf("loadtest_user_%d_%d", i, time.Now().UnixNano())
		email := fmt.Sprintf("%s@example.com", username)
		password := "password123"
		
		user, err := registerUser(username, email, password)
		if err != nil {
			log.Fatalf("Failed to register user %s: %v", username, err)
		}
		
		users[i] = user
	}
	
	fmt.Printf("Registered %d test users\n", len(users))
	fmt.Println("Starting load test...")
	
	gameID := "loadtest_game"
	
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	for i := 0; i < concurrentUsers; i++ {
		go runUserWorkload(users[i], gameID, &wg, stats)
	}

	wg.Wait()

	totalTime := time.Since(startTime)
	stats.CalculateAverage()

	fmt.Println("\nLoad Test Results")
	fmt.Println("================")
	fmt.Printf("Total test time: %v\n", totalTime)
	fmt.Printf("Total requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful requests: %d (%.2f%%)\n", 
		stats.SuccessfulRequests, 
		float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100)
	fmt.Printf("Failed requests: %d (%.2f%%)\n", 
		stats.FailedRequests,
		float64(stats.FailedRequests) / float64(stats.TotalRequests) * 100)
	fmt.Printf("Requests per second: %.2f\n", float64(stats.TotalRequests) / totalTime.Seconds())
	fmt.Printf("Average response time: %v\n", stats.AverageTime)
	fmt.Printf("Min response time: %v\n", stats.MinTime)
	fmt.Printf("Max response time: %v\n", stats.MaxTime)
} 