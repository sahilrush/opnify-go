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

	// Check if the stock already exists
	if _, exists := models.Orderbooks[payload.Stock]; exists {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Stock already exists",
			Data:    models.Orderbooks[payload.Stock],
		})
		return
	}

	// Initialize the orderbook for the stock
	models.Orderbooks[payload.Stock] = models.Pricing{
		Yes: make(map[int]models.OrderType),
		No:  make(map[int]models.OrderType),
	}

	// Initialize the stock balance for the user and stock
	if _, exists := models.Stock_Balances[payload.Stock]; !exists {
		models.Stock_Balances[payload.Stock] = map[string]models.Stocksymbol{}
	}

	// Initialize the user's stock symbol if not present
	if _, exists := models.Stock_Balances[payload.Stock][payload.UserId]; !exists {
		models.Stock_Balances[payload.Stock][payload.UserId] = map[string]models.OutCome{
			"yes": {Quantity: 100, Locked: 0}, // Initialize with 100 units
			"no":  {Quantity: 100, Locked: 0}, // Initialize with 100 units
		}
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Symbol created successfully",
		Data:    models.Orderbooks[payload.Stock],
	})
}

func GetOrderBooks(c *gin.Context) {
	if len(ORDERBOOKS) == 0 {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "no orderbook available",
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Following orderbooks are available",
		Data:    ORDERBOOKS,
	})

}
func ViewOrderbook(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Symbol is required",
			Data:    nil,
		})
		return
	}
	orderbook, exists := ORDERBOOKS[symbol]
	if !exists {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: false,
			Message: "no orderbook found for given symbol",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Here is the Orderbook " + symbol,
		Data:    orderbook,
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
