package models

import (
	"encoding/json"
)

// Trade struct definition
type Trade struct {
	Exchange     string
	Pair         string
	Price        float64
	Quantity     float64
	Timestamp    int64
	IsBuyerMaker bool
}

// Implement the encoding.BinaryMarshaler interface
func (t Trade) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

// Define a struct to match the incoming JSON structure
type BinanceTrade struct {
	Event        string `json:"e"`
	EventTime    int64  `json:"E"`
	Symbol       string `json:"s"`
	TradeID      int64  `json:"t"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	TradeTime    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
	Ignore       bool   `json:"M"`
}
