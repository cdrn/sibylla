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

func ConnectCoinbaseWebSocket(config exchangeconfig.Config, pairs []string) {
	for {
		// Create a channel to receive OS signals
		interrupt := make(chan os.Signal, 1)
		// Notify the interrupt channel on receiving an interrupt signal
		signal.Notify(interrupt, os.Interrupt)

		// Define the WebSocket URL for Coinbase
		u := url.URL{Scheme: "wss", Host: "ws-feed.pro.coinbase.com", Path: ""}
		log.Printf("connecting to %s", u.String())

		// Connect to the WebSocket server
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close() // Ensure the connection is closed when the function exits

		// Create a channel to signal when the connection is done
		done := make(chan struct{})

		// Subscribe to the trade channel for the provided pairs
		subscribeMessage := map[string]interface{}{
			"type": "subscribe",
			"channels": []map[string]interface{}{
				{
					"name":        "matches",
					"product_ids": pairs,
				},
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

				// Unpack the trade message into the CoinbaseTrade struct
				var coinbaseTrade trade.CoinbaseTradeMessage
				err = json.Unmarshal(message, &coinbaseTrade)
				if err != nil {
					log.Printf("Could not unmarshal trade message: %v", err)
					continue
				}

				// Iterate over the events in the CoinbaseTradeMessage
				for _, event := range coinbaseTrade.Events {
					for _, tradeData := range event.Trades {
						// Map CoinbaseTrade data to the Trade struct
						trade := trade.Trade{
							Exchange:     "coinbase",
							Pair:         tradeData.ProductID,
							Price:        func() float64 { p, _ := strconv.ParseFloat(tradeData.Price, 64); return p }(),
							Quantity:     func() float64 { q, _ := strconv.ParseFloat(tradeData.Size, 64); return q }(),
							Timestamp:    func() int64 { t, _ := time.Parse(time.RFC3339, tradeData.Time); return t.Unix() }(),
							IsBuyerMaker: tradeData.Side == "sell",
						}

						// Push the trade struct into redis
						redisKey := "trades:coinbase:" + tradeData.ProductID
						err = config.RedisClient.PushToList(redisKey, trade, 100)
						if err != nil {
							log.Printf("Could not push trade to Redis: %v", err)
						} else {
							log.Printf("Coinbase trade added to redis")
						}
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
				log.Println("Resetting WebSocket connection for Coinbase")
				c.Close()
				goto reconnect
			}
		}
	reconnect:
		log.Println("Reconnecting to Coinbase WebSocket")
	}
}
