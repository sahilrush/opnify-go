package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

func SellYes(c *gin.Context) {
	type YesPayload struct {
		UserId   string `json:"userId" binding:"required"`
		Stock    string `json:"stock" binding:"required"`
		Price    int    `json:"price" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	}

	var payload YesPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid payload",
			Data:    err.Error(),
		})
		return
	}

	if payload.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Quantity must be greater than zero",
			Data:    nil,
		})
		return
	}

	// Debug: Print initial state
	fmt.Printf("Checking stock: %s for user: %s\n", payload.Stock, payload.UserId)

	userStock, ok := models.Stock_Balances[payload.Stock]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this symbol",
			Data:    fmt.Sprintf("Available stocks: %v", getKeys(models.Stock_Balances)),
		})
		return
	}

	// Debug: Print user stock state
	fmt.Printf("User stock found. Available users: %v\n", getKeys(userStock))

	stockSymbol, ok := userStock[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this user",
			Data:    fmt.Sprintf("User stock state: %v", userStock),
		})
		return
	}

	// Debug: Print outcome state
	fmt.Printf("Stock symbol found. Available types: %v\n", getKeys(stockSymbol))

	outcome, ok := stockSymbol["yes"]
	if !ok {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "No YES tokens available",
			Data:    fmt.Sprintf("Available token types: %v", getKeys(stockSymbol)),
		})
		return
	}

	// Debug: Print quantities
	fmt.Printf("Current quantities - Available: %d, Trying to sell: %d\n",
		outcome.Quantity, payload.Quantity)

	if outcome.Quantity < payload.Quantity {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Insufficient stock quantity",
			Data: map[string]interface{}{
				"available": outcome.Quantity,
				"requested": payload.Quantity,
				"locked":    outcome.Locked,
			},
		})
		return
	}

	// Rest of your existing code for updating balances...
	outcome.Locked += payload.Quantity
	outcome.Quantity -= payload.Quantity
	stockSymbol["yes"] = outcome
	userStock[payload.UserId] = stockSymbol
	models.Stock_Balances[payload.Stock] = userStock

	// Initialize orderbook price level if it doesn't exist
	orderbooks := models.Orderbooks[payload.Stock]
	if _, ok := orderbooks.Yes[payload.Price]; !ok {
		orderbooks.Yes[payload.Price] = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
	}

	yesOrders := orderbooks.Yes[payload.Price]
	userOrder := yesOrders.Orders[payload.UserId]

	yesOrders.Total += payload.Quantity
	userOrder.Quantity += payload.Quantity
	userOrder.Type = "sell" // Adding order type
	yesOrders.Orders[payload.UserId] = userOrder
	orderbooks.Yes[payload.Price] = yesOrders
	models.Orderbooks[payload.Stock] = orderbooks

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Stock sold successfully",
		Data: map[string]interface{}{
			"orderbook":         models.Orderbooks[payload.Stock],
			"remaining_balance": outcome,
		},
	})
}

// Helper function to get map keys
func getKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func SellNo(c *gin.Context) {
	type NoPayload struct {
		UserId   string `json:"userId" binding:"required"`
		Stock    string `json:"stock" binding:"required"`
		Price    int    `json:"price" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	}

	var payload NoPayload

	// Bind JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid payload",
			Data:    err.Error(),
		})
		return
	}

	// Check if stock exists for the symbol
	userStock, ok := models.Stock_Balances[payload.Stock]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock exists for this symbol",
			Data:    nil,
		})
		return
	}

	// Check if user has any stocks
	stockSymbol, ok := userStock[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stocks available for this user",
			Data:    nil,
		})
		return
	}

	// Check if "NO" tokens exist for the user
	outcome, ok := stockSymbol["no"]
	if !ok {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "No 'NO' tokens available",
			Data:    nil,
		})
		return
	}

	// Validate quantity
	if outcome.Quantity < payload.Quantity {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Insufficient stock quantity",
			Data: map[string]interface{}{
				"available": outcome.Quantity,
				"requested": payload.Quantity,
				"locked":    outcome.Locked,
			},
		})
		return
	}

	// Update user's stock balance
	outcome.Locked += payload.Quantity
	outcome.Quantity -= payload.Quantity
	stockSymbol["no"] = outcome
	userStock[payload.UserId] = stockSymbol
	models.Stock_Balances[payload.Stock] = userStock

	// Initialize or get orderbook for the stock symbol
	orderbook, ok := models.Orderbooks[payload.Stock]
	if !ok {
		orderbook = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
		models.Orderbooks[payload.Stock] = orderbook
	}

	// Initialize or get price level in "NO" orderbook
	priceLevel, ok := orderbook.No[payload.Price]
	if !ok {
		priceLevel = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
		orderbook.No[payload.Price] = priceLevel
	}

	// Update user's order at this price level
	userOrder := priceLevel.Orders[payload.UserId]
	userOrder.Quantity += payload.Quantity
	userOrder.Type = "sell" // Changed from "normal" to "sell" for clarity

	// Update the orderbook
	priceLevel.Total += payload.Quantity
	priceLevel.Orders[payload.UserId] = userOrder
	orderbook.No[payload.Price] = priceLevel
	models.Orderbooks[payload.Stock] = orderbook

	// Return successful response
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "NO tokens sold successfully",
		Data: map[string]interface{}{
			"orderbook":         models.Orderbooks[payload.Stock],
			"remaining_balance": outcome,
		},
	})

}
