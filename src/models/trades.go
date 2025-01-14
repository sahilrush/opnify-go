package models

type YesPayload struct {
	UserId      string `json:"userId"`
	Stocksymbol string `json:"stocksymbol"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
}

type NoPayload struct {
	UserId      string `json:"userId"`
	Stocksymbol string `json:"stocksymbol"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
}

type BuyYes struct {
	UserId      string `json:"userid"`
	Stocksymbol string `json:"stocksymbol"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	StockType   string `json:"stocktype"`
}

type BuyNo struct {
	UserId      string `json:"userid"`
	Stocksymbol string `json:"stocksymbol"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	StockType   string `json:"stocktype"`
}
