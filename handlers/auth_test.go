package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/user/realtime-leaderboard/models"
)

func TestLoginValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/login", Login)

	tests := []struct {
		name       string
		loginBody  models.UserLogin
		statusCode int
	}{
		{
			name: "missing_email",
			loginBody: models.UserLogin{
				Email:    "",
				Password: "password123",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "missing_password",
			loginBody: models.UserLogin{
				Email:    "test@example.com",
				Password: "",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid_email_format",
			loginBody: models.UserLogin{
				Email:    "invalid-email",
				Password: "password123",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.loginBody)

			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			
			r.ServeHTTP(w, req)
			
			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
} 