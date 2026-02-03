package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/client"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/workflows"
	"github.com/gin-gonic/gin"
)

func CostHandler(c *gin.Context) {
	accessToken := ""

	if !client.UsePrivateKeyAuth() {
		// Extract access token from cookie
		accessToken, err := c.Cookie("github_token")
		if err != nil || accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no access token found",
			})
			return
		}
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

	githubClient := client.GetClient(accessToken)

	report := workflows.GenerateReport(githubClient, requestBody.Repositories)

	c.JSON(http.StatusOK, report)
}
