package main

import (
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve index.html at root path
	r.GET("/", handlers.Login)

	r.GET("/calculate", handlers.Calculate)

	// Handle GitHub OAuth callback
	r.GET("/callback", handlers.CallbackHandler)

	r.POST("/cost", handlers.CostHandler)

	// Start server on port 8080
	r.Run(":8080")
}
