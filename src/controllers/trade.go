package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

func SellYes(payload models.YesPayload) models.UserResponse {
	models.StockBalancesMutex.Lock()
	defer models.StockBalancesMutex.Unlock()

	userStock, ok := models.Stock_Balances[payload.UserId]
	if !ok {
		return models.UserResponse{
			Success: false,
			Message: "",
			Data:    userStock,
		}
	}

}

func SellNo(c *gin.Context) {

}

func BuyYes(c *gin.Context) {

}

func BuyNo(c *gin.Context) {

}
