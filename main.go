package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/controllers"
)

func main() {

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Probo-Backend"})
	})

	r.POST("/user/create", controllers.CreateUser)
	r.POST("/onramp/inr", controllers.OnrampUser)
	r.POST("/balance/inr", controllers.GetBalances)
	r.POST("/balance/inr/:userId", controllers.GetUserBalance)
	r.Run(":8080")
}
