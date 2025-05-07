package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/user/realtime-leaderboard/config"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRegistration struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SaveUser(user *User) error {
	userKey := fmt.Sprintf("user:%s", user.ID)
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	

	user.CreatedAt = time.Now()
	
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = config.RedisClient.Set(config.Ctx, userKey, userJSON, 0).Err()
	if err != nil {
		return err
	}

	err = config.RedisClient.Set(config.Ctx, fmt.Sprintf("username:%s", user.Username), user.ID, 0).Err()
	if err != nil {
		return err
	}
	
	return config.RedisClient.Set(config.Ctx, fmt.Sprintf("email:%s", user.Email), user.ID, 0).Err()
}

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

func GetUserByEmail(email string) (*User, error) {
	userID, err := config.RedisClient.Get(config.Ctx, fmt.Sprintf("email:%s", email)).Result()
	if err != nil {
		return nil, err
	}
	
	return GetUserByID(userID)
}

func GetUserByUsername(username string) (*User, error) {

	userID, err := config.RedisClient.Get(config.Ctx, fmt.Sprintf("username:%s", username)).Result()
	if err != nil {
		return nil, err
	}
	
	return GetUserByID(userID)
}

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