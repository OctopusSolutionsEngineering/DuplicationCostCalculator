package handlers

import (
	"net/http"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/workflows"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/client"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v57/github"
)

func CostHandler(c *gin.Context) {
	CostHandlerWrapped(c, client.GetClient, workflows.GenerateReport)
}

func CostHandlerWrapped(c *gin.Context, getClient func(string) *github.Client, generateReport func(*github.Client, []string) models.Report) {
	accessToken := ""

	if !client.UsePrivateKeyAuth() {
		// Extract access token from cookie
		token, err := c.Cookie("github_token")
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no access token found",
			})
			return
		}

		accessToken = token
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

	githubClient := getClient(accessToken)

	report := generateReport(githubClient, requestBody.Repositories)

	c.JSON(http.StatusOK, report)
}
