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

func ConnectKrakenWebSocket(config exchangeconfig.Config, pairs []string) {
	for {
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

		log.Printf("pairs: %s", pairs)

		// Subscribe to the trade channel for the provided pairs
		subscribeMessage := map[string]interface{}{
			"method": "subscribe",
			"params": map[string]interface{}{
				"channel":  "trade",
				"symbol":   pairs,
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

				// Unpack the trade message into the KrakenTrade struct
				var krakenTrade trade.KrakenTradeMessage
				log.Printf("message: %s", message)
				err = json.Unmarshal(message, &krakenTrade)
				if err != nil {
					log.Printf("Could not unmarshal trade message: %v", err)
					continue
				}
				for _, tradeData := range krakenTrade.Data {
					pair, err := ConvertPairReverse(tradeData.Symbol, "kraken")

					// Map KrakenTradeMessage data to the Trade struct
					trade := trade.Trade{
						Exchange:     "kraken",
						Pair:         pair, // Map back to our language for pairs
						Price:        tradeData.Price,
						Quantity:     tradeData.Quantity,
						Timestamp:    func() int64 { t, _ := strconv.ParseInt(tradeData.Timestamp, 10, 64); return t }(),
						IsBuyerMaker: tradeData.Side == "sell",
					}

					// Push the trade struct into redis
					redisKey := "trades:kraken:" + tradeData.Symbol
					err = config.RedisClient.PushToList(redisKey, trade, 100)
					if err != nil {
						log.Printf("Could not push trade to Redis: %v", err)
					} else {
						log.Printf("Kraken trade pushed to redis for pair %s", tradeData.Symbol)
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
			case <-time.After(time.Hour): // Reset the connection every hour
				log.Println("Resetting WebSocket connection for Kraken")
				c.Close()
				goto reconnect
			}
		}
	reconnect:
		log.Println("Reconnecting to Kraken WebSocket")
	}
}
