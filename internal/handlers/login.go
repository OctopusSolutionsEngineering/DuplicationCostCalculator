package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/config"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	if config.UsePrivateKeyAuth() {
		// Redirect to calculate page as there is no need to log in
		c.Redirect(http.StatusFound, "/calculate")
	}

	// Check if github_token cookie exists
	token, err := c.Cookie("github_token")
	if err == nil && token != "" {
		// User is authenticated, redirect to calculate page
		c.Redirect(http.StatusFound, "/calculate")
		return
	}

	// User is not authenticated, show login page
	c.File("index.html")
}
