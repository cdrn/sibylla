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
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func ConnectBinanceWebSocket(config exchangeconfig.Config, pairs []string) {
	for {
		// Create a channel to receive OS signals
		interrupt := make(chan os.Signal, 1)
		// Notify the interrupt channel on receiving an interrupt signal
		signal.Notify(interrupt, os.Interrupt)

		// Define the WebSocket URL for Binance with multiple pairs
		streams := strings.Join(pairs, "@trade/")
		u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/stream", RawQuery: "streams=" + streams + "@trade"}

		log.Printf("connecting to %s", u.String())

		// Connect to the WebSocket server
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}

		// Create a channel to signal when the connection is done
		done := make(chan struct{})

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

				// Unpack the trade message into the BinanceTrade struct
				var binanceMessage trade.BinanceMessageMultistream
				err = json.Unmarshal(message, &binanceMessage)
				if err != nil {
					log.Printf("Could not unmarshal trade message: %v", err)
					continue
				}
				// log.Printf("btrade %s", message)

				// Map BinanceTrade to the Trade struct
				tradeData := trade.Trade{
					Exchange:     "binance",
					Pair:         binanceMessage.Data.Symbol,
					Price:        func() float64 { p, _ := strconv.ParseFloat(binanceMessage.Data.Price, 64); return p }(),
					Quantity:     func() float64 { q, _ := strconv.ParseFloat(binanceMessage.Data.Quantity, 64); return q }(),
					Timestamp:    binanceMessage.Data.TradeTime,
					IsBuyerMaker: binanceMessage.Data.IsBuyerMaker,
				}

				// Push the trade struct into redis
				err = config.RedisClient.PushToList("trades:binance:"+binanceMessage.Data.Symbol, tradeData, 100)
				if err != nil {
					log.Printf("Could not push trade to Redis: %v", err)
				} else {
					// log.Printf("Binance trade pushed to Redis for pair %s", tradeData.Pair)
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
				log.Println("Resetting WebSocket connection for Binance")
				c.Close()
				goto reconnect
			}
		}
	reconnect:
		log.Println("Reconnecting to Binance WebSocket")
	}
}
