package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/mustafa-bagci/LogMonitoringAPI/docs"
)

var (
	db     *sql.DB
	jwtKey []byte
)

// Metrics
var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

// JWT Authentication Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

// Login Route
// @Summary Login to the application
// @Description Authenticate user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
// @Example
// {
//   "username": "admin",
//   "password": "password"
// }
func login(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Hardcoded credentials for demonstration
	if creds.Username != "admin" || creds.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// LogEntry struct for the addLog endpoint
type LogEntry struct {
	Message string `json:"message"`
}

type LogResponse struct {
	ID        int64  `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	CreatedAt string `json:"created_at"`
}

// Add Log Route
// @Summary Add a new log
// @Description Add a new log entry
// @Tags logs
// @Accept json
// @Produce json
// @Param log body LogEntry true "Log message"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /logs [post]
func addLog(c *gin.Context) {
	var logEntry LogEntry

	if err := c.BindJSON(&logEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if logEntry.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	_, err := db.Exec("INSERT INTO logs (message, created_at) VALUES ($1, NOW())", logEntry.Message)
	if err != nil {
		log.Printf("Error inserting log: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Log added successfully"})
}

// Get Logs Route
// @Summary Get all logs
// @Description Retrieve all log entries
// @Tags logs
// @Produce json
// @Success 200 {array} LogResponse
// @Failure 401 {object} map[string]string
// @Router /logs [get]
func getLogsHandler(c *gin.Context) {
	// Fetch logs from database
	rows, err := db.Query("SELECT id, level, message, service, created_at FROM logs")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving logs"})
		return
	}
	defer rows.Close()

	var logs []LogResponse
	for rows.Next() {
		var (
			id        sql.NullInt64
			level     sql.NullString
			message   sql.NullString
			service   sql.NullString
			createdAt sql.NullTime
		)

		// Scan the row into variables
		if err := rows.Scan(&id, &level, &message, &service, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning logs"})
			return
		}

		// Create a LogResponse struct
		logEntry := LogResponse{
			ID:        id.Int64,
			Level:     level.String,
			Message:   message.String,
			Service:   service.String,
			CreatedAt: createdAt.Time.Format(time.RFC3339),
		}

		// Append the log entry to the logs slice
		logs = append(logs, logEntry)
	}

	// Return the logs as JSON
	c.JSON(http.StatusOK, logs)
}

// Delete Log Route
// @Summary Delete a log by ID
// @Description Delete a log entry by its ID
// @Tags logs
// @Produce json
// @Param id path int true "Log ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /logs/{id} [delete]
func deleteLogHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	_, err := db.Exec("DELETE FROM logs WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting log: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}

// Update Log Route
// @Summary Update a log by ID
// @Description Update a log entry completely by its ID
// @Tags logs
// @Accept json
// @Produce json
// @Param id path int true "Log ID"
// @Param log body LogEntry true "Log message"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /logs/{id} [put]
func updateLogHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var logEntry LogEntry
	if err := c.BindJSON(&logEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if logEntry.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	_, err := db.Exec("UPDATE logs SET message = $1 WHERE id = $2", logEntry.Message, id)
	if err != nil {
		log.Printf("Error updating log: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log updated successfully"})
}

// Patch Log Route
// @Summary Partially update a log by ID
// @Description Partially update a log entry by its ID
// @Tags logs
// @Accept json
// @Produce json
// @Param id path int true "Log ID"
// @Param log body LogEntry true "Log message"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /logs/{id} [patch]
func patchLogHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var logEntry LogEntry
	if err := c.BindJSON(&logEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if logEntry.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	_, err := db.Exec("UPDATE logs SET message = $1 WHERE id = $2", logEntry.Message, id)
	if err != nil {
		log.Printf("Error patching log: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log patched successfully"})
}

// Middleware to measure request metrics
func measureRequest(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		handler(c)
		duration := time.Since(start).Seconds()
		httpRequests.WithLabelValues(c.Request.Method, c.FullPath(), fmt.Sprint(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// @title Log Monitoring API
// @version 1.0
// @description This is a sample API for log monitoring.
// @host localhost:8080
// @BasePath /
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtKey = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtKey) == 0 {
		log.Fatal("JWT_SECRET is empty! Check your .env file.")
	}

	// Connect to PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal("Unable to ping database:", err)
	}
	log.Println("✅ Connected to PostgreSQL successfully!")

	// Initialize Gin Router
	r := gin.Default()

	// Add Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Security headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public Routes
	r.POST("/login", measureRequest(login))

	// Protected Routes
	authRoutes := r.Group("/")
	authRoutes.Use(authMiddleware())
	{
		authRoutes.GET("/logs", measureRequest(getLogsHandler))
		authRoutes.POST("/logs", measureRequest(addLog))
		authRoutes.DELETE("/logs/:id", measureRequest(deleteLogHandler))
		authRoutes.PUT("/logs/:id", measureRequest(updateLogHandler))
		authRoutes.PATCH("/logs/:id", measureRequest(patchLogHandler))
	}

	// Graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}