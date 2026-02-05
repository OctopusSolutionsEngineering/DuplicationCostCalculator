package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/client"
	"github.com/gin-gonic/gin"
)

// IsAuthenticated checks if the user is authenticated and redirects to login if not.
// Returns true if the user is authenticated, false if they were redirected.
func IsAuthenticated(c *gin.Context) bool {
	if client.UsePrivateKeyAuth() {
		return true
	}

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
		return false
	}

	return true
}
