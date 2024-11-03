package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	exchangeconfig "sibylla_service/pkg/config"
	"sibylla_service/pkg/exchange"
	handlers "sibylla_service/pkg/handlers"
	"sibylla_service/pkg/redisclient"

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
	http.HandleFunc("/api/trades", handlers.TradesHandler(redisClient))

	// Initialize exchange listeners
	binanceConfig := exchangeconfig.Config{
		ConnectionString: getEnv("BINANCE_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	krakenConfig := exchangeconfig.Config{
		ConnectionString: getEnv("KRAKEN_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	// coinbaseConfig := exchangeconfig.Config{
	// 	ConnectionString: getEnv("COINBASE_WEBSOCKET_URL", ""),
	// 	RedisClient:      redisClient,
	// }

	// Base pairs that we are watching
	basePairs := []string{"BTCUSDT", "ETHUSDT", "BTCUSD", "ETHUSD", "WBTCUSDT"}

	binancePairs, err := exchange.ConvertPairs(basePairs, "binance")
	krakenPairs, err := exchange.ConvertPairs(basePairs, "kraken")
	if err != nil {
		log.Fatalf("Failed to convert pairs for Kraken: %v", err)
	}
	// coinbasePairs, err := exchange.ConvertPairs(basePairs, "coinbase")
	// if err != nil {
	// 	log.Fatalf("Failed to convert pairs for Coinbase: %v", err)
	// }

	go exchange.ConnectBinanceWebSocket(binanceConfig, binancePairs)
	go exchange.ConnectKrakenWebSocket(krakenConfig, krakenPairs)
	// go exchange.ConnectCoinbaseWebSocket(coinbaseConfig, coinbasePairs)

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

// helper function to load env variables with a default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
