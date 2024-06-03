package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

// Status godoc
// @Summary Responds with 200 if service is running
// @Description health check for service
// @Produce  json
// @Success 200 {string} Working!
// @Router /health/health [get]
func (h HealthController) Status(c *gin.Context) {
	ginMode := os.Getenv("GIN_MODE")
	host := os.Getenv("DB_HOST")
	c.JSON(http.StatusOK, gin.H{
		"message": "Working!",
		"version": "1.0.0",
		"ginMode": ginMode,
		"db-host": host,
	})
}

// Status godoc
// @Summary Responds with 200 if service is running
// @Description health check for service
// @Produce  json
// @Success 200 {string} pong
// @Router /health/ping [get]

func (h HealthController) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
