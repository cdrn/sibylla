package main

import (
	"log"
	"net/http"
	"os"

	exchangeconfig "sibylla_service/pkg/config"
	"sibylla_service/pkg/exchange"
	"sibylla_service/pkg/redisclient"

	"github.com/joho/godotenv"
)

func main() {

	// load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	} else {
		log.Println(".env vars loaded")
	}

	// REDIS CLIENT //
	redisClient := redisclient.NewRedisClient(
		"localhost:6379", // Redis address
		"",               // no password by default
		0,                // use default Redis database (DB 0)
	)

	// load environment variables or configs
	port := getEnv("PORT", "8080")

	// initialize routes and handlers
	http.HandleFunc("/", homeHandler)

	// Initialize exchange listeners
	binanceConfig := exchangeconfig.Config{
		ConnectionString: getEnv("BINANCE_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	krakenConfig := exchangeconfig.Config{
		ConnectionString: getEnv("KRAKEN_WEBSOCKET_URL", ""),
		RedisClient:      redisClient,
	}

	go exchange.ConnectBinanceWebSocket(binanceConfig)
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

// helper function to load env variables with a default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
