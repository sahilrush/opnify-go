package models

type OutCome struct {
	Quantity int `json:"quantity"`
	Locked   int `json:"locked"`
}

type Stocksymbol map[string]OutCome
type User map[string]Stocksymbol

type Stock map[string]User

var Stock_Balances = Stock{}
