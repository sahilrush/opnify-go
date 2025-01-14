package controllers

import (
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

	// Get user's stock balance
	userStock, ok := models.Stock_Balances[payload.Stock]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this symbol",
			Data:    nil,
		})
		return
	}

	stockSymbol, ok := userStock[payload.UserId]
	if !ok {
		c.JSON(http.StatusNotFound, models.UserResponse{
			Success: false,
			Message: "No stock available for this user",
			Data:    nil,
		})
		return
	}

	outcome, ok := stockSymbol["yes"]
	if !ok || outcome.Quantity < payload.Quantity {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Insufficient stock quantity",
			Data:    nil,
		})
		return
	}

	// Update user's stock balance
	outcome.Locked += payload.Quantity
	outcome.Quantity -= payload.Quantity
	stockSymbol["yes"] = outcome
	userStock[payload.UserId] = stockSymbol
	models.Stock_Balances[payload.Stock] = userStock

	// Initialize or get orderbook
	orderbook, exists := models.Orderbooks[payload.Stock]
	if !exists {
		orderbook = models.Pricing{
			Yes: make(map[int]models.OrderType),
			No:  make(map[int]models.OrderType),
		}
	}

	// Initialize or get the price level in YES orderbook
	priceLevel, exists := orderbook.Yes[payload.Price]
	if !exists {
		priceLevel = models.OrderType{
			Total:  0,
			Orders: make(map[string]models.Orders),
		}
	}

	// Update user's order
	userOrder, exists := priceLevel.Orders[payload.UserId]
	if !exists {
		userOrder = models.Orders{
			Quantity: 0,
			Type:     "sell",
		}
	}

	userOrder.Quantity += payload.Quantity
	priceLevel.Orders[payload.UserId] = userOrder
	priceLevel.Total += payload.Quantity

	// Update the orderbook
	orderbook.Yes[payload.Price] = priceLevel
	models.Orderbooks[payload.Stock] = orderbook

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "Stock sold successfully",
		Data: map[string]interface{}{
			"orderbook":         models.Orderbooks[payload.Stock],
			"remaining_balance": outcome,
		},
	})
}
