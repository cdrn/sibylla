package exchangeconfig

import "sibylla_service/pkg/redisclient"

type Config struct {
	ConnectionString string
	RedisClient      *redisclient.RedisClient
}
