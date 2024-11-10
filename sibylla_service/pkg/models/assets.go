package models

type Asset struct {
	Symbol        string   `json:"symbol"`
	RelatedAssets []string `json:"related_assets"`
}

var assetsMap = map[string]Asset{
	"USDT": {Symbol: "USDT", RelatedAssets: []string{"USD", "BTC", "ETH"}},
	"BTC":  {Symbol: "BTC", RelatedAssets: []string{"WBTC", "USDT"}},
	"WBTC": {Symbol: "WBTC", RelatedAssets: []string{"BTC"}},
}
