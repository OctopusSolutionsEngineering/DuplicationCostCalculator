package handlers

import (
	"github.com/gin-gonic/gin"
)

func ReposHandler(c *gin.Context) {
	if !IsAuthenticated(c) {
		return
	}

	// User is authenticated, show repos page
	c.File("html/repos.html")
}
