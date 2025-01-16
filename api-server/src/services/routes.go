package services

// import (
// 	"github.com/gin-gonic/gin"
// 	"github.com/sahilrush/src/controllers"
// )

// func SetupRoutes(router *gin.Engine) {
// 	api := router.Group("/")
// 	{
// 		api.POST("/user/create/:userId", controllers.ForwardReq("/user/create/:userId"))
// 		api.POST("/symbol/create/:stockSymbol", controllers.ForwardReq("/symbol/create/:stockSymbol"))
// 		api.POST("/onramp/inr", controllers.ForwardReq("/onramp/inr"))
// 		api.POST("/trade/mint", controllers.ForwardReq("/trade/mint"))
// 		api.POST("/reset", controllers.ForwardReq("/reset"))

// 		api.GET("/balances/inr", controllers.ForwardReq("/balances/inr"))
// 		api.GET("/balances/inr/:userId", controllers.ForwardReq("/balances/inr/:userId"))
// 		api.GET("/balances/stock", controllers.ForwardReq("/balances/stock"))
// 		api.GET("/balances/stock/:stockSymbol", controllers.ForwardReq("/balances/stock/:stockSymbol"))

// 		api.GET("/orderbook", controllers.ForwardReq("/orderbook"))
// 		api.GET("/orderbook/:stockSymbol", controllers.ForwardReq("/orderbook/:stockSymbol"))

// 		api.POST("/order/buy", controllers.ForwardReq("/order/buy"))
// 		api.POST("/order/sell", controllers.ForwardReq("/order/sell"))
// 		api.POST("/order/cancel", controllers.ForwardReq("/order/cancel"))
// 	}
// }
