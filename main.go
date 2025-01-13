package main

import (
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Probo-Backend"})
	})

	r.POST("/user/create")
	r.POST("/onramp/inr")
	r.POST("/balance/inr")
	r.POST("/balance/inr/:userId")
}
