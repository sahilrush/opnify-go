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
	r.POST("/symbol/create", controllers.CreateSymbol)
	r.GET("/orderbook/:symbol", controllers.ViewOrderbook)
	r.GET("/orderbook/getorder", controllers.GetOrderBooks)
	r.GET("/getUserStock/:userId", controllers.GetUserStock)
	r.GET("/getStocks", controllers.GetStocks)
	r.POST("/sellyes", controllers.SellYes)
	r.POST("/sellno", controllers.SellNo)
	r.POST("/buyyes", controllers.BuyYes)
	r.POST("/buyno", controllers.BuyNo)
	//viewOrderbook
	r.Run(":8080")
}



//redis pubsub queue and subscribe should be http.Error
//(w, err.Error(), http.StatusInternalServerError)
//it will push to redis queue and it will subscribe to channel
