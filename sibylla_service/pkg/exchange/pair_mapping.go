// sibylla_service/pkg/exchange/pair_mapping.go
package exchange

import (
	"fmt"
)

// Define a map for each exchange's pair format
var pairMappings = map[string]map[string]string{
	"binance": {
		"BTCUSD":  "BTCUSD",
		"BTCUSDT": "BTCUSDT",
		"ETHUSD":  "ETHUSD",
		"ETHUSDT": "ETHUSDT",
		"BNBBTC":  "BNBBTC",
	},
	"kraken": {
		"BTCUSD": "BTC/USD",
		"ETHUSD": "ETH/USD",
	},
	// Add more exchanges as needed
}

// ConvertPair converts a bespoke pair format to the exchange-specific format
func ConvertPair(bespokePair, exchange string) (string, error) {
	if exchangePairs, ok := pairMappings[exchange]; ok {
		if exchangePair, ok := exchangePairs[bespokePair]; ok {
			return exchangePair, nil
		}
		return "", fmt.Errorf("pair %s not found for exchange %s", bespokePair, exchange)
	}
	return "", fmt.Errorf("exchange %s not supported", exchange)
}
