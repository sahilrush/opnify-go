package models

type Orders struct {
	Quantity int    `json:"quantity"`
	Type     string `json:"type"`
}

type OrderType struct {
	Total  int               `json:"total"`
	Orders map[string]Orders `json:"orders"`
}

type Pricing struct {
	Yes map[int]OrderType `json:"yes"`
	No  map[int]OrderType `json:"no"`
}

type Orderbook map[string]Pricing

var Orderbooks = Orderbook{}
