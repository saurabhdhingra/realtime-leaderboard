package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/user/realtime-leaderboard/config"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRegistration represents the data needed for user registration
type UserRegistration struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserLogin represents the data needed for user login
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SaveUser saves a user to Redis
func SaveUser(user *User) error {
	// Generate user key
	userKey := fmt.Sprintf("user:%s", user.ID)
	
	// Hash the password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	
	// Set creation time
	user.CreatedAt = time.Now()
	
	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	
	// Save user in Redis
	err = config.RedisClient.Set(config.Ctx, userKey, userJSON, 0).Err()
	if err != nil {
		return err
	}
	
	// Save username to ID mapping for username lookup
	err = config.RedisClient.Set(config.Ctx, fmt.Sprintf("username:%s", user.Username), user.ID, 0).Err()
	if err != nil {
		return err
	}
	
	// Save email to ID mapping for email lookup
	return config.RedisClient.Set(config.Ctx, fmt.Sprintf("email:%s", user.Email), user.ID, 0).Err()
}

// GetUserByID retrieves a user by ID
func GetUserByID(id string) (*User, error) {
	userKey := fmt.Sprintf("user:%s", id)
	userJSON, err := config.RedisClient.Get(config.Ctx, userKey).Result()
	if err != nil {
		return nil, err
	}
	
	var user User
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	// Get user ID from email
	userID, err := config.RedisClient.Get(config.Ctx, fmt.Sprintf("email:%s", email)).Result()
	if err != nil {
		return nil, err
	}
	
	return GetUserByID(userID)
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*User, error) {
	// Get user ID from username
	userID, err := config.RedisClient.Get(config.Ctx, fmt.Sprintf("username:%s", username)).Result()
	if err != nil {
		return nil, err
	}
	
	return GetUserByID(userID)
}

// ValidateCredentials validates user credentials
func ValidateCredentials(email, password string) (*User, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	return user, nil
} 