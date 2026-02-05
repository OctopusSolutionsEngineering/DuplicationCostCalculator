package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/configuration"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/oauth"
	"github.com/gin-gonic/gin"
)

func CallbackHandler(c *gin.Context) {
	CallbackHandlerWrapped(c, oauth.ExchangeCodeForToken, configuration.GetClientId, configuration.GetClientSecret)
}

func CallbackHandlerWrapped(
	c *gin.Context,
	oauthTokenExchange func(string, string, string, string) (oauth.TokenResponse, error),
	clientIdSetting func() string,
	clientSecretSetting func() string) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No authorization code provided",
		})
		return
	}

	// Exchange code for access token
	clientID := clientIdSetting()
	clientSecret := clientSecretSetting()

	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "GitHub OAuth credentials not configured",
		})
		return
	}

	tokenResponse, err := oauthTokenExchange(clientID, clientSecret, code, "https://github.com/login/oauth/access_token")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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

	c.SetSameSite(http.SameSiteStrictMode)

	// Set access token in HTTP-only cookie
	c.SetCookie(
		"github_token",            // name
		tokenResponse.AccessToken, // value
		3600,                      // max age (1 hour)
		"/",                       // path
		"",                        // domain (empty = current domain)
		true,                      // secure (set to true in production with HTTPS)
		true,                      // httpOnly
	)

	// Redirect to repos page with repos query param if state was provided
	if state != "" {
		c.Redirect(http.StatusFound, "/repos?repos="+state)
	} else {
		c.Redirect(http.StatusFound, "/repos")
	}
}
