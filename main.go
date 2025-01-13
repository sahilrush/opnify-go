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
	r.GET("/balance/inr", controllers.GetBalances)
	r.GET("/balance/inr/:userId", controllers.GetUserBalance)
	r.GET("/symbol/create", controllers.CreateSymbol)
	r.GET("/orderbook/getorder", controllers.GetOrderBooks)
	r.GET("/getUserStock/:userId", controllers.GetUserStock)
	r.GET("/getStocks", controllers.GetStocks)
	r.Run(":8080")
}
