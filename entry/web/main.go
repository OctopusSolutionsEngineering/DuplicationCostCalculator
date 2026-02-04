package main

import (
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve index.html at root path
	r.GET("/", handlers.Login)

	r.GET("/repos", handlers.ReposHandler)

	r.GET("/calculate", handlers.Calculate)

	// Handle GitHub OAuth callback
	r.GET("/callback", handlers.CallbackHandler)

	r.POST("/cost", handlers.CostHandler)

	// Default handler for unmatched routes - redirect to login page
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	// Start server on port 8080
	r.Run(":8080")
}
