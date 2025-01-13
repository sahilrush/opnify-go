package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

var STOCK_BALANCES = models.Stock_Balances
var ORDERBOOKS = models.Orderbooks

func CreateSymbol(c *gin.Context) {
	var payload struct {
		UserId string `json:"userId"  binding:"required"`
		Stock  string `json:"stock"  binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid payload",
			Data:    nil,
		})
		return
	}

	if _, exists := ORDERBOOKS[payload.Stock]; exists {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "stock already exists",
			Data:    ORDERBOOKS[payload.Stock],
		})
		return
	}

	ORDERBOOKS[payload.Stock] = models.Pricing{
		Yes: make(map[int]models.OrderType),
		No:  make(map[int]models.OrderType),
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Symbol created succefully",
		Data:    ORDERBOOKS[payload.Stock],
	})

}

func GetOrderBooks(c *gin.Context) {
	if len(ORDERBOOKS) == 0 {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: false,
			Message: "no orderbook available",
			Data:    nil,
		})
	}
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Following orderbooks are available",
		Data:    ORDERBOOKS,
	})

}

func GetStocks(c *gin.Context) {
	if len(STOCK_BALANCES) == 0 {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: false,
			Message: "No stocks are avaliable",
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "these are the stocks",
		Data:    STOCK_BALANCES,
	})
}

func GetUserStock(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "User Id is required",
			Data:    nil,
		})
		return
	}

	userStocks, exists := STOCK_BALANCES[userId]
	if !exists {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: false,
			Message: "no stocks found for the given user",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "here are the stocks for the user",
		Data:    userStocks,
	})
}
