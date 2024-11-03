// sibylla_service/pkg/exchange/pair_mapping.go
package exchange

import "fmt"

// Define a map for each exchange's pair format
var pairMappings = map[string]map[string]string{
	"binance": {
		"BTCUSD":   "btcusd",
		"BTCUSDT":  "btcusdt",
		"ETHUSD":   "ethusd",
		"ETHUSDT":  "ethusdt",
		"BNBBTC":   "bnbbtc",
		"WBTCUSDT": "wbtcusdt",
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

// Reverse mappings will be generated from pairMappings
var reversePairMappings = generateReverseMappings(pairMappings)

// Function to generate reverse mappings
func generateReverseMappings(mappings map[string]map[string]string) map[string]map[string]string {
	reverseMappings := make(map[string]map[string]string)
	for exchange, pairs := range mappings {
		reverseMappings[exchange] = make(map[string]string)
		for bespoke, exchangeSpecific := range pairs {
			reverseMappings[exchange][exchangeSpecific] = bespoke
		}
	}
	return reverseMappings
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

// ConvertPairReverse converts an exchange-specific pair format to the bespoke format
func ConvertPairReverse(exchangePair, exchange string) (string, error) {
	if exchangePairs, ok := reversePairMappings[exchange]; ok {
		if bespokePair, ok := exchangePairs[exchangePair]; ok {
			return bespokePair, nil
		}
		return "", nil
	}
	return "", fmt.Errorf("exchange %s not supported", exchange)
}

// ConvertPairsReverse converts a list of exchange-specific pair formats to the bespoke formats
func ConvertPairsReverse(exchangePairs []string, exchange string) ([]string, error) {
	var bespokePairs []string
	for _, pair := range exchangePairs {
		bespokePair, err := ConvertPairReverse(pair, exchange)
		if err != nil {
			return nil, err
		}
		// If missing a pair, log it and continue anyway
		if bespokePair == "" {
			fmt.Printf("Missing reverse pair mapping for pair: %s, exchange: %s\n", pair, exchange)
			continue
		}
		bespokePairs = append(bespokePairs, bespokePair)
	}
	return bespokePairs, nil
}
