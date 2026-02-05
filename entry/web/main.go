package main

import (
	handlers2 "github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/application/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve index.html at root path
	r.GET("/", handlers2.Login)

	r.GET("/favicon.ico", handlers2.IconHandler)

	r.GET("/repos", handlers2.ReposHandler)

	r.GET("/calculate", handlers2.Calculate)

	// Handle GitHub OAuth callback
	r.GET("/callback", handlers2.CallbackHandler)

	r.POST("/cost", handlers2.CostHandler)

	// Default handler for unmatched routes - redirect to login page
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	// Start server on port 8080
	r.Run(":8080")
}
