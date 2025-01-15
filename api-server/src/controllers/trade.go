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

func BuyYes(c *gin.Context) {

	type BuyYes struct {
		UserId    string `json:"userid"`
		Stock     string `json:"stock"`
		Price     int    `json:"price"`
		Quantity  int    `json:"quantity"`
		StockType string `json:"stocktype"`
	}

	var payload BuyYes

	// Bind the JSON request body to the payload struct
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON format",
			"error":   err.Error(),
		})
		return
	}

	// Validate the payload
	if payload.Stock == "" || payload.Price <= 0 ||
		payload.UserId == "" || payload.Quantity <= 0 ||
		payload.StockType == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: All fields must be provided with valid values",
			"data":    nil,
		})
		return
	}

	user := models.INR_BALANCES[payload.UserId]
	if user.Balance < payload.Price*payload.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Insufficient balance of user",
			"data":    map[string]interface{}{},
		})
		return
	}

	_, exists := models.Orderbooks[payload.Stock]
	if !exists {
		models.Orderbooks[payload.Stock] = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
	}

	if models.Orderbooks[payload.Stock].Yes == nil {
		models.Orderbooks[payload.Stock] = models.Pricing{
			Yes: make(map[int]models.OrderType),
		}
	}

	_, priceExists := models.Orderbooks[payload.Stock].Yes[payload.Price]
	newPrice := 1000 - payload.Price
	if !priceExists {
		if _, noPriceExists := models.Orderbooks[payload.Stock].No[payload.Price]; !noPriceExists {
			models.Orderbooks[payload.Stock].No[newPrice] = models.OrderType{
				Total:  0,
				Orders: make(map[string]models.Orders),
			}
		}

		newOrder, orderExists := models.Orderbooks[payload.Stock].No[newPrice].Orders[payload.UserId]
		if orderExists {
			newOrder.Quantity += payload.Quantity
			newOrder.Type = "inverse"

			existingOrder := models.Orderbooks[payload.Stock].No[newPrice]
			existingOrder.Total += payload.Quantity
			user.Locked += payload.Price * payload.Quantity
			user.Balance -= payload.Price * payload.Quantity

			models.INR_BALANCES[payload.UserId] = user

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Orderbook updated",
				"data":    newOrder,
			})
			return
		} else {
			newBook := models.Orderbooks[payload.Stock].No[newPrice].Orders[payload.UserId]
			newBook.Quantity += payload.Quantity
			newBook.Type = "inverse"

			copiedBook := models.Orderbooks[payload.Stock].No[newPrice]
			copiedBook.Total += payload.Quantity
			user.Locked += payload.Price * payload.Quantity
			user.Balance -= payload.Price * payload.Quantity

			models.INR_BALANCES[payload.UserId] = user

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Orderbook updated",
				"data":    models.Orderbooks[payload.Stock],
			})
			return
		}
	}

	// Handling the orders in 'Yes'
	if orders, ok := models.Orderbooks[payload.Stock].Yes[payload.Price]; ok {
		totalAmount := payload.Quantity
		for userId, order := range orders.Orders {
			if totalAmount <= 0 {
				break
			}
			currentQuantity := order.Quantity
			subtraction := min(totalAmount, currentQuantity)
			existingUser := models.INR_BALANCES[userId]
			existingUser.Balance += payload.Price * subtraction

			models.INR_BALANCES[userId] = existingUser

			newUser := models.Orderbooks[payload.Stock].Yes[payload.Price].Orders[userId]
			newUser.Quantity -= subtraction
			totalAmount -= subtraction

			if _, exists := models.Stock_Balances[payload.UserId]; !exists {
				models.Stock_Balances[payload.UserId] = make(map[string]models.Stocksymbol)
			}
			if stock, exists := models.Stock_Balances[payload.UserId][payload.Stock]; !exists {
				stock[payload.Stock] = models.OutCome{
					Quantity: 0,
					Locked:   0,
				}
			}
			existingStock := models.Stock_Balances[payload.UserId][payload.Stock]["yes"]
			existingStock.Quantity -= subtraction
		}

		orderbook := models.Orderbooks[payload.Stock].Yes[payload.Price]
		orderbook.Total -= payload.Quantity - totalAmount
		if orderbook.Total == 0 {
			delete(models.Orderbooks[payload.Stock].Yes, payload.Price)
		}
		user.Balance -= payload.Price * payload.Quantity
		models.INR_BALANCES[payload.UserId] = user

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Orderbook updated",
			"data":    models.Orderbooks[payload.Stock],
		})
		return
	}

	// If no matching order, deduct balance and lock funds
	user.Balance -= payload.Price * payload.Quantity
	user.Locked += payload.Price * payload.Quantity
	models.INR_BALANCES[payload.UserId] = user

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Orderbook updated",
		"data":    models.Orderbooks[payload.Stock],
	})
}

// Helper function to find minimum between two numbers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func BuyNo(c *gin.Context) {

	type BuyNo struct {
		UserId    string `json:"userid"`
		Stock     string `json:"stock"`
		Price     int    `json:"price"`
		Quantity  int    `json:"quantity"`
		StockType string `json:"stocktype"`
	}

	var payload BuyNo
	// Bind the JSON request body to the payload struct
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON format",
			"error":   err.Error(),
		})
		return
	}

	// Validate required fields
	if payload.Stock == "" || payload.Price <= 0 ||
		payload.UserId == "" || payload.Quantity <= 0 ||
		payload.StockType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: All fields must be provided with valid values",
		})
		return
	}

	// Log the user balance and requested order details
	user := models.INR_BALANCES[payload.UserId]
	totalPrice := payload.Price * payload.Quantity
	fmt.Printf("User balance: %d, Total price: %d\n", user.Balance, totalPrice)

	// Check if the user has sufficient balance
	if user.Balance < totalPrice {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Insufficient balance",
			"data":    map[string]interface{}{},
		})
		return
	}

	// Check if the stock exists in the order book
	pricing, exists := models.Orderbooks[payload.Stock]
	if !exists {
		pricing = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
	}

	// Check if the price already exists for the stock
	ordertype, priceExists := pricing.No[payload.Price]
	if !priceExists {
		ordertype = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
	}

	// Update or create the order
	newOrder, orderExists := ordertype.Orders[payload.UserId]
	if orderExists {
		newOrder.Quantity += payload.Quantity
	} else {
		newOrder = models.Orders{
			Quantity: payload.Quantity,
			Type:     "inverse",
		}
	}

	// Update the orderbook with the new order
	ordertype.Orders[payload.UserId] = newOrder
	ordertype.Total += payload.Quantity
	pricing.No[payload.Price] = ordertype
	models.Orderbooks[payload.Stock] = pricing

	// Update the user's balance and locked funds
	user.Locked += totalPrice
	user.Balance -= totalPrice
	models.INR_BALANCES[payload.UserId] = user

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Orderbook updated",
		"data":    models.Orderbooks[payload.Stock],
	})
}
