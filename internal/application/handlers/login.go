package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/client"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	reposParam := c.Query("repos")

	if client.UsePrivateKeyAuth() {
		// Redirect to repos page as there is no need to log in
		if reposParam != "" {
			c.Redirect(http.StatusFound, "/repos?repos="+reposParam)
		} else {
			c.Redirect(http.StatusFound, "/repos")
		}
		return
	}

	// Check if github_token cookie exists
	token, err := c.Cookie("github_token")
	if err == nil && token != "" {
		// User is authenticated, redirect to repos page
		if reposParam != "" {
			c.Redirect(http.StatusFound, "/repos?repos="+reposParam)
		} else {
			c.Redirect(http.StatusFound, "/repos")
		}
		return
	}

	// User is not authenticated, show login page
	c.File("index.html")
}
