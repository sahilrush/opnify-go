package models

type UserBalance struct {
	Balance int `json:"balance"`
	Locked  int `json:"locked" `
}

var UserWithBalance = make(map[string]UserBalance)

type OnrampUser struct {
	UserId string `json:"userId" binding:"required"`
	Amount int    `json:"amount" binding:"required"`
}

var INR_BALANCES = UserWithBalance
