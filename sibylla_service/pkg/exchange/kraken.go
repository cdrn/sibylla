package exchange

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	exchangeconfig "sibylla_service/pkg/config"
	trade "sibylla_service/pkg/models"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func ConnectKrakenWebSocket(config exchangeconfig.Config) {
	// Create a channel to receive OS signals
	interrupt := make(chan os.Signal, 1)
	// Notify the interrupt channel on receiving an interrupt signal
	signal.Notify(interrupt, os.Interrupt)

	// Define the WebSocket URL for Kraken
	u := url.URL{Scheme: "wss", Host: "ws.kraken.com", Path: "/v2"}
	log.Printf("connecting to %s", u.String())

	// Connect to the WebSocket server
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close() // Ensure the connection is closed when the function exits

	// Create a channel to signal when the connection is done
	done := make(chan struct{})

	// Subscribe to the trade channel for MATIC/USD
	subscribeMessage := map[string]interface{}{
		"method": "subscribe",
		"params": map[string]interface{}{
			"channel":  "trade",
			"symbol":   []string{"BTC/USD"},
			"snapshot": false,
		},
	}
	subscribeMessageJSON, err := json.Marshal(subscribeMessage)
	if err != nil {
		log.Fatal("subscribe message marshal:", err)
	}

	err = c.WriteMessage(websocket.TextMessage, subscribeMessageJSON)
	if err != nil {
		log.Fatal("subscribe message send:", err)
	}

	// Start a goroutine to read messages from the WebSocket
	go func() {
		// Defer executes after function completion. close socket.
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// Log every trade for now.
			log.Printf("kraken recv: %s", message)

			// Unpack the trade message into the KrakenTrade struct
			var krakenTrade trade.KrakenTradeMessage
			err = json.Unmarshal(message, &krakenTrade)
			if err != nil {
				log.Printf("Could not unmarshal trade message: %v", err)
				continue
			}
			for _, tradeData := range krakenTrade.Data {
				// Map KrakenTradeMessage data to the Trade struct
				trade := trade.Trade{
					Exchange:     "kraken",
					Pair:         tradeData.Symbol,
					Price:        tradeData.Price,
					Quantity:     tradeData.Quantity,
					Timestamp:    func() int64 { t, _ := strconv.ParseInt(tradeData.Timestamp, 10, 64); return t }(),
					IsBuyerMaker: tradeData.Side == "sell",
				}

				// Push the trade struct into redis
				err = config.RedisClient.PushToList("trades:kraken:BTC/USD", trade, 100)
				if err != nil {
					log.Printf("Could not push trade to Redis: %v", err)
				} else {
					log.Printf("Kraken trade added to redis")
				}
			}
		}
	}()

	for {
		select {
		case <-done: // Exit the loop on done sig
			return
		case <-interrupt: // If an interrupt signal is received
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done: // Wait for the done channel to be closed
			case <-time.After(time.Second): // Or timeout after 1 second
			}
			return
		}
	}
}
