package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

func SellYes(c *gin.Context) {

	var payload models.YesPayload
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

	stock, ok := userStock[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this user",
			Data:    fmt.Sprintf("User stock state: %v", userStock),
		})
		return
	}

	// Debug: Print outcome state
	fmt.Printf("Stock symbol found. Available types: %v\n", getKeys(stock))

	outcome, ok := stock["yes"]
	if !ok {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "No YES tokens available",
			Data:    fmt.Sprintf("Available token types: %v", getKeys(stock)),
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
	stock["yes"] = outcome
	userStock[payload.UserId] = stock
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
	stock, ok := userStock[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stocks available for this user",
			Data:    nil,
		})
		return
	}

	// Check if "NO" tokens exist for the user
	outcome, ok := stock["no"]
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
	stock["no"] = outcome
	userStock[payload.UserId] = stock
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
	var payload models.BuyYes

	// Bind JSON payload to the BuyYes structure
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid JSON payload",
			Data:    nil,
		})
		return
	}

	// Validate required fields
	if payload.Stock == "" || payload.Price <= 0 ||
		payload.UserId == "" || payload.Quantity <= 0 ||
		payload.StockType == "" {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid request: All fields must be provided with valid values",
			Data:    nil,
		})
		return
	}

	// Check if the user exists in INR_BALANCES
	user, exists := models.INR_BALANCES[payload.UserId]
	if !exists {
		// Initialize the user if not found
		user = models.UserBalance{
			Balance: 10000, // Set a default balance
			Locked:  0,
		}
		models.INR_BALANCES[payload.UserId] = user
	}

	// Calculate total cost of the buy operation
	totalCost := float64(payload.Price * payload.Quantity)
	if user.Balance < int(totalCost) {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Insufficient balance",
			Data: map[string]interface{}{
				"required":  totalCost,
				"available": user.Balance,
			},
		})
		return
	}

	// Initialize or get the orderbook for the specified stock
	if models.Orderbooks == nil {
		models.Orderbooks = make(map[string]models.Pricing)
	}

	// Check if stock exists in the orderbook, else create a new entry
	if _, exists := models.Orderbooks[payload.Stock]; !exists {
		models.Orderbooks[payload.Stock] = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
	}

	pricing := models.Orderbooks[payload.Stock]

	// Calculate 'NO' price based on 'YES' price
	newPrice := 10 - payload.Price // Fixed difference (e.g., 10 - YES price)

	if _, exists := pricing.No[newPrice]; !exists {
		pricing.No[newPrice] = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
	}

	// Create or update the inverse order on the 'no' side of the book
	orderType := pricing.No[newPrice]
	order := models.Orders{
		Quantity: payload.Quantity,
		Type:     "inverse",
	}

	// Update or create order in the orderbook for user at this price level
	if existingOrder, exists := orderType.Orders[payload.UserId]; exists {
		// If user already has an order at this price, increase the quantity
		existingOrder.Quantity += payload.Quantity
		orderType.Orders[payload.UserId] = existingOrder
	} else {
		// Create new order for user if not exists
		orderType.Orders[payload.UserId] = order
	}

	// Update the total quantity of orders at this price level
	orderType.Total += payload.Quantity

	// Save the updated order in the orderbook
	pricing.No[newPrice] = orderType
	models.Orderbooks[payload.Stock] = pricing

	// Lock funds and update user balance
	user.Locked += int(totalCost)
	user.Balance -= int(totalCost)
	models.INR_BALANCES[payload.UserId] = user

	// Debug logs
	fmt.Printf("Orderbook after updating: %+v\n", models.Orderbooks[payload.Stock])

	// Return success response
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Orderbook updated successfully",
		Data:    models.Orderbooks[payload.Stock],
	})
}

func BuyNo(c *gin.Context) {

	var payload models.BuyNo
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
	user := Users[payload.UserId]
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
	Users[payload.UserId] = user

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Orderbook updated",
		"data":    models.Orderbooks[payload.Stock],
	})
}
