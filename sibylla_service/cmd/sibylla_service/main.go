package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	exchangeconfig "sibylla_service/pkg/config"
	"sibylla_service/pkg/exchange"
	"sibylla_service/pkg/redisclient"

	"encoding/json"

	"sibylla_service/pkg/models"

	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	// load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	} else {
		log.Println(".env vars loaded")
	}

	// REDIS CLIENT //
	redisClient := redisclient.NewRedisClient(
		fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", ""), getEnv("REDIS_PORT", "")),
		getEnv("REDIS_PASSWORD", ""), // no password by default
		0,                            // use default Redis database (DB 0)
	)

	// ENVS //
	port := getEnv("PORT", "8080")

	// ROUTES //
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/trades", tradesHandler(redisClient))

	// Websocket
	http.HandleFunc("/ws", websocketHandler(redisClient))

	// Initialize exchange listeners
	binanceConfig := exchangeconfig.Config{
		ConnectionString: getEnv("BINANCE_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	krakenConfig := exchangeconfig.Config{
		ConnectionString: getEnv("KRAKEN_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	binancePairs := []string{"btcusdt", "ethusdt"}
	go exchange.ConnectBinanceWebSocket(binanceConfig, binancePairs)
	go exchange.ConnectKrakenWebSocket(krakenConfig)

	// start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	} else {
		log.Printf("Server started! 📈")
	}
}

// simple handler function
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Sibylla online"))
}

func tradesHandler(redisClient *redisclient.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tradesBinance, err := redisClient.GetList("trades:binance:BTCUSDT", 1)
		if err != nil || len(tradesBinance) == 0 {
			log.Printf("No binance trades found")
			tradesBinance = []string{"{\"Price\":0}"}
		}

		tradesKraken, err := redisClient.GetList("trades:kraken:BTC/USD", 1)
		if err != nil || len(tradesKraken) == 0 {
			log.Printf("No kraken trades found")
			tradesKraken = []string{"{\"Price\":0}"}
		}

		binancePrice := parseTradePrice(tradesBinance[0])
		krakenPrice := parseTradePrice(tradesKraken[0])

		// Calculate arbitrage opportunity
		arbOpportunity := binancePrice - krakenPrice

		response := map[string]interface{}{
			"binance": tradesBinance,
			"kraken":  tradesKraken,
			"delta":   arbOpportunity,
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

func websocketHandler(redisClient *redisclient.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		for {
			// Fetch the latest trade data
			tradesBinance, err := redisClient.GetList("trades:binance:BTC/USDT", 1)
			if err != nil || len(tradesBinance) == 0 {
				log.Printf("No binance trades found")
				tradesBinance = []string{"{\"Price\":0}"}
			}

			tradesKraken, err := redisClient.GetList("trades:kraken:BTC/USD", 1)
			if err != nil || len(tradesKraken) == 0 {
				log.Printf("No kraken trades found")
				tradesKraken = []string{"{\"Price\":0}"}
			}

			binancePrice := parseTradePrice(tradesBinance[0])
			krakenPrice := parseTradePrice(tradesKraken[0])

			arbOpportunity := krakenPrice - binancePrice

			response := map[string]interface{}{
				"binance": tradesBinance,
				"kraken":  tradesKraken,
				"spread":  arbOpportunity,
			}

			// Send the data to the client
			err = conn.WriteJSON(response)
			if err != nil {
				log.Printf("Error writing JSON to WebSocket: %v", err)
				break
			}

			// Wait for a short period before sending the next update
			time.Sleep(1 * time.Second)
		}
	}
}

// Helper function to parse price from Trade data
func parseTradePrice(trade string) float64 {
	var tradeData models.Trade
	err := json.Unmarshal([]byte(trade), &tradeData)
	if err != nil {
		log.Printf("Error parsing trade data: %v", err)
		return 0.0
	}
	return tradeData.Price
}

// helper function to load env variables with a default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
