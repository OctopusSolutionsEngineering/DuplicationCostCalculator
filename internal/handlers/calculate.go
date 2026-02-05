package handlers

import (
	"github.com/gin-gonic/gin"
)

func Calculate(c *gin.Context) {
	if !RequireAuth(c) {
		return
	}

	// User is authenticated, show calculate page
	c.File("calculate.html")
}
