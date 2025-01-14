package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

func SellYes(c *gin.Context) {
	// Define a local struct for the payload
	type YesPayload struct {
		UserId   string `json:"userId" binding:"required"`
		Stock    string `json:"stock" binding:"required"`
		Price    int    `json:"price" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	}

	var payload YesPayload

	// Bind the incoming JSON payload to the YesPayload struct
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid payload",
			Data:    err.Error(),
		})
		return
	}

	// Check if the user has stock balance
	userStock, ok := models.Stock_Balances[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available",
			Data:    nil,
		})
		return
	}

	// Check if the user has stock for the given symbol
	stockSymbol, ok := userStock[payload.Stock]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this movement",
		})
		return
	}

	// Check if the user has "yes" stock
	outcome, ok := stockSymbol["yes"]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available",
		})
		return
	}

	// Verify if the user has enough stock to sell
	if outcome.Quantity < payload.Quantity {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "User doesn't have enough stock",
			Data:    nil,
		})
		return
	}

	// Update user's stock balance
	outcome.Locked += payload.Quantity
	outcome.Quantity -= payload.Quantity
	stockSymbol["yes"] = outcome

	// Get or initialize the order book for the given stock symbol
	orderbooks, ok := models.Orderbooks[payload.Stock]
	if !ok {
		orderbooks = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
		models.Orderbooks[payload.Stock] = orderbooks
	}

	// Get or initialize the order type for the given price
	yesOrders, ok := orderbooks.Yes[payload.Price]
	if !ok {
		yesOrders = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
		orderbooks.Yes[payload.Price] = yesOrders
	}

	// Get or initialize the user's specific order
	userOrder, ok := yesOrders.Orders[payload.UserId]
	if !ok {
		userOrder = models.Orders{
			Quantity: 0,
			Type:     "normal",
		}
		yesOrders.Orders[payload.UserId] = userOrder
	}

	// Update the order book
	yesOrders.Total += payload.Quantity
	userOrder.Quantity += payload.Quantity
	yesOrders.Orders[payload.UserId] = userOrder
	orderbooks.Yes[payload.Price] = yesOrders
	models.Orderbooks[payload.Stock] = orderbooks

	// Return the success response
	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Stock sold successfully",
		Data:    models.Orderbooks[payload.Stock],
	})
}
