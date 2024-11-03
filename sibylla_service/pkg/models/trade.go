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
//
//	{
//	  "e": "trade",       // Event type
//	  "E": 1672515782136, // Event time
//	  "s": "BNBBTC",      // Symbol
//	  "t": 12345,         // Trade ID
//	  "p": "0.001",       // Price
//	  "q": "100",         // Quantity
//	  "T": 1672515782136, // Trade time
//	  "m": true,          // Is the buyer the market maker?
//	  "M": true           // Ignore
//	}
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

type BinanceMessageMultistream struct {
	Stream string       `json:"stream"`
	Data   BinanceTrade `json:"data"`
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

// Coinbase incoming trade data
type CoinbaseTrade struct {
	TradeID   string `json:"trade_id"`
	ProductID string `json:"product_id"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Side      string `json:"side"`
	Time      string `json:"time"`
}

type CoinbaseTradeEvent struct {
	Type   string          `json:"type"`
	Trades []CoinbaseTrade `json:"trades"`
}

type CoinbaseTradeMessage struct {
	Channel     string               `json:"channel"`
	ClientID    string               `json:"client_id"`
	Timestamp   string               `json:"timestamp"`
	SequenceNum int64                `json:"sequence_num"`
	Events      []CoinbaseTradeEvent `json:"events"`
}
