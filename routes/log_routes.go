package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mustafa-bagci/LogMonitoringAPI/controllers"
)

// SetupLogRoutes configures log-related routes
func SetupLogRoutes(r *gin.Engine) {
	r.GET("/logs", controllers.GetLogsHandler)
}