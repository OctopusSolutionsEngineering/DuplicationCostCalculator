package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CallbackHandler(c *gin.Context) {
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
}
