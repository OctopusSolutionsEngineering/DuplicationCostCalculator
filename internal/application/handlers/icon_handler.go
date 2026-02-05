package handlers

import "github.com/gin-gonic/gin"

func IconHandler(c *gin.Context) {
	c.File("images/octopus.png")
}
