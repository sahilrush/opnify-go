package models

type UserBalance struct {
	Balance int `json:"balance"`
	Locked  int `json:"locked" `
}

var UserWithBalance = make(map[string]UserBalance)

type OnRampedUser struct {
	userId string `json:"userid" binding:"required"`
	amount int    `json:"amount" binding:"required"`
}

var INR_BALANCES = UserWithBalance
