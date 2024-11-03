package utils

import (
	"encoding/json"
	"log"
	"sibylla_service/pkg/models"
)

// Helper function to parse price from Trade data
func ParseTradePrice(trade string) float64 {
	var tradeData models.Trade
	err := json.Unmarshal([]byte(trade), &tradeData)
	if err != nil {
		log.Printf("Error parsing trade data: %v", err)
		return 0.0
	}
	return tradeData.Price
}
