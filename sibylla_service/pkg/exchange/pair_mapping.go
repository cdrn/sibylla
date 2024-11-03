// sibylla_service/pkg/exchange/pair_mapping.go
package exchange

import "fmt"

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
	"coinbase": {
		"BTCUSD": "BTC-USD",
		"ETHUSD": "ETH-USD",
	},
	// Add more exchanges as needed
}

// ConvertPair converts a bespoke pair format to the exchange-specific format
func ConvertPair(bespokePair, exchange string) (string, error) {
	if exchangePairs, ok := pairMappings[exchange]; ok {
		if exchangePair, ok := exchangePairs[bespokePair]; ok {
			return exchangePair, nil
		}
		return "", nil
	}
	return "", fmt.Errorf("exchange %s not supported", exchange)
}

// ConvertPairs converts a list of bespoke pair formats to the exchange-specific formats
func ConvertPairs(bespokePairs []string, exchange string) ([]string, error) {
	var convertedPairs []string
	for _, pair := range bespokePairs {
		convertedPair, err := ConvertPair(pair, exchange)
		if err != nil {
			return nil, err
		}
		// If missing a pair, log it and continue anyway
		if convertedPair == "" {
			fmt.Printf("Missing pair mapping for pair: %s, exchange: %s\n", pair, exchange)
			continue
		}
		convertedPairs = append(convertedPairs, convertedPair)
	}
	return convertedPairs, nil
}
