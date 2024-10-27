package exchange

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	exchangeconfig "sibylla_service/pkg/config"
	"time"

	"github.com/gorilla/websocket"
)

func ConnectBinanceWebSocket(config exchangeconfig.Config) {
	// Create a channel to receive OS signals
	interrupt := make(chan os.Signal, 1)
	// Notify the interrupt channel on receiving an interrupt signal
	signal.Notify(interrupt, os.Interrupt)

	// Define the WebSocket URL for Binance
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@trade"}
	log.Printf("connecting to %s", u.String())

	// Connect to the WebSocket server
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close() // Ensure the connection is closed when the function exits

	// Create a channel to signal when the connection is done
	done := make(chan struct{})

	// Start a goroutine to read messages from the WebSocket
	go func() {
		defer close(done) // Close the done channel when the goroutine exits
		for {
			_, message, err := c.ReadMessage() // Read a message from the WebSocket
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message) // Log the received message
		}
	}()

	for {
		select {
		case <-done: // If the done channel is closed, exit the loop
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
