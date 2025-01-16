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

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON format",
			"error":   err.Error(),
		})
		return
	}

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

	user := Users[payload.UserId]
	if user.Balance < payload.Price*payload.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Insufficient balance of user",
			"data":    map[string]interface{}{},
		})
		return
	}

	// Initialize orderbook if it doesn't exist
	orderbook, exists := models.Orderbooks[payload.Stock]
	if !exists {
		orderbook = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
		models.Orderbooks[payload.Stock] = orderbook
	}

	// Get the current orderbook state
	currentPricing := models.Orderbooks[payload.Stock]

	// Initialize Yes and No if they're nil
	if currentPricing.Yes == nil {
		currentPricing.Yes = make(map[int]models.OrderType)
	}
	if currentPricing.No == nil {
		currentPricing.No = make(map[int]models.OrderType)
	}

	// Check for existing orders at this price
	_, priceExists := currentPricing.Yes[payload.Price]

	// If no matching YES orders exist, create inverse NO order
	if !priceExists {
		newPrice := 1000 - payload.Price

		// Create new order
		orderType := models.OrderType{
			Total:  payload.Quantity,
			Orders: make(map[string]models.Orders),
		}
		orderType.Orders[payload.UserId] = models.Orders{
			Quantity: payload.Quantity,
			Type:     "inverse",
		}

		// Update the No orderbook
		currentPricing.No[newPrice] = orderType

		// Save the updated pricing back to the main orderbook
		models.Orderbooks[payload.Stock] = currentPricing

		// Update user balance
		user.Locked += payload.Price * payload.Quantity
		user.Balance -= payload.Price * payload.Quantity
		Users[payload.UserId] = user

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Inverse order created",
			"data":    models.Orderbooks[payload.Stock],
		})
		return
	}

	// Handle matching with existing YES orders
	if orders, ok := currentPricing.Yes[payload.Price]; ok {
		totalAmount := payload.Quantity
		modifiedOrders := orders

		for userId, order := range orders.Orders {
			if totalAmount <= 0 {
				break
			}

			currentQuantity := order.Quantity
			matchQuantity := min(totalAmount, currentQuantity)

			// Update seller's balance
			seller := Users[userId]
			seller.Balance += payload.Price * matchQuantity
			seller.Locked -= payload.Price * matchQuantity
			Users[userId] = seller

			// Update order
			if currentQuantity > matchQuantity {
				modifiedOrders.Orders[userId] = models.Orders{
					Quantity: currentQuantity - matchQuantity,
					Type:     order.Type,
				}
			} else {
				delete(modifiedOrders.Orders, userId)
			}

			totalAmount -= matchQuantity
			modifiedOrders.Total -= matchQuantity

			// Initialize buyer's stock balance if needed
			stockBalances, exists := models.Stock_Balances[payload.Stock]
			if !exists {
				stockBalances = make(map[string]models.Stocksymbol)
				models.Stock_Balances[payload.Stock] = stockBalances
			}

			userStocks, exists := stockBalances[payload.UserId]
			if !exists {
				userStocks = make(map[string]models.OutCome)
				stockBalances[payload.UserId] = userStocks
			}

			// Update buyer's stock balance
			currentOutcome := userStocks["yes"]
			currentOutcome.Quantity += matchQuantity
			userStocks["yes"] = currentOutcome
			stockBalances[payload.UserId] = userStocks
			models.Stock_Balances[payload.Stock] = stockBalances
		}

		// Update the orderbook
		if modifiedOrders.Total > 0 {
			currentPricing.Yes[payload.Price] = modifiedOrders
		} else {
			delete(currentPricing.Yes, payload.Price)
		}
		models.Orderbooks[payload.Stock] = currentPricing

		// Update buyer's balance
		user.Balance -= payload.Price * payload.Quantity
		Users[payload.UserId] = user

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Orders matched successfully",
			"data":    models.Orderbooks[payload.Stock],
		})
		return
	}

	// If no matches, create a new NO inverse order
	newPrice := 1000 - payload.Price
	orderType := models.OrderType{
		Total:  payload.Quantity,
		Orders: make(map[string]models.Orders),
	}
	orderType.Orders[payload.UserId] = models.Orders{
		Quantity: payload.Quantity,
		Type:     "inverse",
	}

	currentPricing.No[newPrice] = orderType
	models.Orderbooks[payload.Stock] = currentPricing

	user.Balance -= payload.Price * payload.Quantity
	user.Locked += payload.Price * payload.Quantity
	Users[payload.UserId] = user

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inverse order created",
		"data":    models.Orderbooks[payload.Stock],
	})
}

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
