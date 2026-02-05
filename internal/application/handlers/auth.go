package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/client"
	"github.com/gin-gonic/gin"
)

// RequireAuth checks if the user is authenticated and redirects to login if not.
// Returns true if the user is authenticated, false if they were redirected.
func RequireAuth(c *gin.Context) bool {
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
			return false
		}
	}

	return true
}
