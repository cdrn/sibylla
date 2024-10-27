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

// Binance incoming trade data
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

// Kraken incoming trade data
type KrakenTradeMessage struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Price     float64 `json:"price"`
		Quantity  float64 `json:"qty"`
		OrderType string  `json:"ord_type"`
		TradeID   int64   `json:"trade_id"`
		Timestamp string  `json:"timestamp"`
	} `json:"data"`
}
