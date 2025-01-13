package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sahilrush/src/models"
)

var Users = make(map[string]models.UserBalance)

func CreateUser(c *gin.Context) {

	var payload struct {
		UserId string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid request payload",
			Data:    nil,
		})
		return
	}

	fmt.Println("Current Users:", Users)

	if _, exists := Users[payload.UserId]; exists {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: false,
			Message: "User already exists",
			Data:    nil,
		})
		return
	}

	Users[payload.UserId] = models.UserBalance{
		Balance: 0,
		Locked:  0,
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "User created successfully",
		Data:    Users[payload.UserId],
	})
}

// on ramping the money
func OnrampUser(c *gin.Context) {
	var payload models.OnrampUser

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Invalid request data",
			Data:    nil,
		})
		return
	}

	if payload.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "Amount must be greater than zero",
			Data:    nil,
		})
		return
	}

	if userBalance, exists := Users[payload.UserId]; exists {
		userBalance.Balance += payload.Amount
		Users[payload.UserId] = userBalance

		c.JSON(http.StatusOK, models.UserResponse{
			Success: true,
			Message: "User balance updated",
			Data:    userBalance,
		})
		return
	}

	c.JSON(http.StatusBadRequest, models.UserResponse{
		Success: false,
		Message: "User does not exist",
		Data:    nil,
	})
}

func GetBalances(c *gin.Context) {

	if len(Users) == 0 {
		c.JSON(http.StatusBadRequest, models.UserResponse{
			Success: false,
			Message: "didont have any money",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		Success: true,
		Message: "The balance is ",
		Data:    Users,
	})

}

// can apply if else
func GetUserBalance(c *gin.Context) {
	userId := c.Param("userId")

	if userBalance, exists := Users[userId]; exists {
		c.JSON(http.StatusOK, models.UserResponse{
			Success: true,
			Message: "User balance is ",
			Data:    userBalance,
		})

	}

}
