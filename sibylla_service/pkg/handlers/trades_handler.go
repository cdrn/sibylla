package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sibylla_service/pkg/redisclient"
	"sibylla_service/pkg/utils"
)

func TradesHandler(redisClient *redisclient.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all keys matching the pattern "trades:*"
		keys, err := redisClient.Keys("trades:*")
		if err != nil {
			http.Error(w, "Failed to retrieve trade keys", http.StatusInternalServerError)
			return
		}

		response := make(map[string]interface{})

		for _, key := range keys {
			trades, err := redisClient.GetList(key, 1)
			if err != nil || len(trades) == 0 {
				log.Printf("No trades found for key: %s", key)
				trades = []string{"{\"Price\":0}"}
			}

			var tradeData map[string]interface{}
			if err := json.Unmarshal([]byte(trades[0]), &tradeData); err != nil {
				log.Printf("Failed to unmarshal trade data for key: %s", key)
				continue
			}

			price := utils.ParseTradePrice(trades[0])
			response[key] = map[string]interface{}{
				"trades": tradeData,
				"price":  price,
			}
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}
}
