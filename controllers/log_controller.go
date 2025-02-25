package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetLogsHandler handles the GET /logs request
func GetLogsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logs retrieved successfully"})
}