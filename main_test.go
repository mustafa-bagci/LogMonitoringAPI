package main

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestDB initializes a connection to the test PostgreSQL database.
func setupTestDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:bagci2001@localhost:5432/log_monitoring?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

// TestLogin tests the login endpoint.
func TestLogin(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid credentials",
			payload:        `{"username": "admin", "password": "password"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `"token":`,
		},
		{
			name:           "Invalid credentials",
			payload:        `{"username": "admin", "password": "wrong"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"error":"Invalid credentials"`,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Gin router
			r := gin.Default()
			r.POST("/login", login)

			// Create HTTP request
			req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			// Record HTTP response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestAddLog tests the addLog endpoint.
func TestAddLog(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		payload        string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid log entry",
			payload:        `{"message": "This is a test log"}`,
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzQwNjAxNTY2fQ.W7zuSwRnk1LaRRImtDUlkHNOx0q6CDPt_GwezQjTeT4",
			expectedStatus: http.StatusCreated,
			expectedBody:   `"message":"Log added successfully"`,
		},
		{
			name:           "Invalid token",
			payload:        `{"message": "This is a test log"}`,
			token:          "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `"error":"Invalid token"`,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Gin router
			r := gin.Default()
			r.POST("/logs", authMiddleware(), addLog)

			// Create HTTP request
			req, _ := http.NewRequest("POST", "/logs", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			// Record HTTP response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestGetLogsHandler tests the getLogsHandler endpoint.
func TestGetLogsHandler(t *testing.T) {
	// Connect to the test database
	db := setupTestDB()
	defer db.Close()

	// Insert test data
	_, err := db.Exec("INSERT INTO logs (level, message, service, created_at) VALUES ($1, $2, $3, $4)",
		"info", "Test log message", "test", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create Gin router
	r := gin.Default()
	r.GET("/logs", authMiddleware(), getLogsHandler)

	// Create HTTP request
	req, _ := http.NewRequest("GET", "/logs", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzQwNjAxNTY2fQ.W7zuSwRnk1LaRRImtDUlkHNOx0q6CDPt_GwezQjTeT4")

	// Record HTTP response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test log message")
}