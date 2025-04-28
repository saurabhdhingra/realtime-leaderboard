package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/realtime-leaderboard/models"
	"github.com/user/realtime-leaderboard/utils"
)

// Register handles user registration
func Register(c *gin.Context) {
	var registration models.UserRegistration

	// Bind request body to registration struct
	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	_, err := models.GetUserByEmail(registration.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Check if username already exists
	_, err = models.GetUserByUsername(registration.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	// Create new user
	user := &models.User{
		ID:       uuid.New().String(),
		Username: registration.Username,
		Email:    registration.Email,
		Password: registration.Password,
	}

	// Save user to database
	if err := models.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return success with token
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var login models.UserLogin

	// Bind request body to login struct
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate credentials
	user, err := models.ValidateCredentials(login.Email, login.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return success with token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// GetProfile retrieves the current user's profile
func GetProfile(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Return user profile
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.(*models.User).ID,
			"username":  user.(*models.User).Username,
			"email":     user.(*models.User).Email,
			"createdAt": user.(*models.User).CreatedAt.Format(time.RFC3339),
		},
	})
} 