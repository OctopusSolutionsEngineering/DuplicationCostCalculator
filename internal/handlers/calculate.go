package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/client"
	"github.com/gin-gonic/gin"
)

func Calculate(c *gin.Context) {
	// We can only define one redirection URL for a github app. To test the app locally,
	// we can use a private key for authentication instead of OAuth.
	if !client.UsePrivateKeyAuth() {
		// Check if github_token cookie exists
		token, err := c.Cookie("github_token")
		if err != nil || token == "" {
			// User is not authenticated, redirect to login page with repos query param if present
			reposParam := c.Query("repos")
			if reposParam != "" {
				c.Redirect(http.StatusFound, "/?repos="+reposParam)
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			return
		}
	}

	// User is authenticated, show calculate page
	c.File("calculate.html")
}
