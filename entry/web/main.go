package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/workflows"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve index.html at root path
	r.GET("/", func(c *gin.Context) {
		// Check if github_token cookie exists
		token, err := c.Cookie("github_token")
		if err == nil && token != "" {
			// User is authenticated, redirect to calculate page
			c.Redirect(http.StatusFound, "/calculate")
			return
		}

		// User is not authenticated, show login page
		c.File("index.html")
	})

	r.GET("/calculate", func(c *gin.Context) {
		// Check if github_token cookie exists
		token, err := c.Cookie("github_token")
		if err != nil || token == "" {
			// User is not authenticated, redirect to login page
			c.Redirect(http.StatusFound, "/")
			return
		}

		// User is authenticated, show calculate page
		c.File("calculate.html")
	})

	// Handle GitHub OAuth callback
	r.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")

		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No authorization code provided",
			})
			return
		}

		// Exchange code for access token
		clientID := os.Getenv("DUPCOST_GITHUB_CLIENT_ID")
		clientSecret := os.Getenv("DUPCOST_GITHUB_CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "GitHub OAuth credentials not configured",
			})
			return
		}

		// Prepare request to exchange code for token
		requestBody, _ := json.Marshal(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"code":          code,
		})

		req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create request",
			})
			return
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to exchange code for token",
			})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var tokenResponse struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			Scope       string `json:"scope"`
			Error       string `json:"error"`
			ErrorDesc   string `json:"error_description"`
		}

		if err := json.Unmarshal(body, &tokenResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse token response",
			})
			return
		}

		if tokenResponse.Error != "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       tokenResponse.Error,
				"description": tokenResponse.ErrorDesc,
			})
			return
		}

		if tokenResponse.AccessToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "No access token received",
			})
			return
		}

		// Set access token in HTTP-only cookie
		c.SetCookie(
			"github_token",            // name
			tokenResponse.AccessToken, // value
			3600,                      // max age (1 hour)
			"/",                       // path
			"",                        // domain (empty = current domain)
			false,                     // secure (set to true in production with HTTPS)
			true,                      // httpOnly
		)

		// Redirect to calculate page
		c.Redirect(http.StatusFound, "/calculate")
	})

	r.POST("/cost", func(c *gin.Context) {
		// Extract access token from cookie
		accessToken, err := c.Cookie("github_token")
		if err != nil || accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no access token found",
			})
			return
		}

		// Parse request body
		var requestBody struct {
			Repositories []string `json:"repositories"`
		}

		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		client := workflows.GetClient(accessToken)

		report := workflows.GenerateReport(client, requestBody.Repositories)

		c.JSON(http.StatusOK, report)
	})

	// Start server on port 8080
	r.Run(":8080")
}
