package redisclient

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return &RedisClient{client: rdb}
}

func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Could not set key %s: %v", key, err)
		return err
	}
	return nil
}

func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Key %s does not exist", key)
			return "", nil
		}
		log.Printf("Could not get key %s: %v", key, err)
		return "", err
	}
	return val, nil
}

func (r *RedisClient) Del(key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Could not delete key %s: %v", key, err)
		return err
	}
	return nil
}

// PushToList adds a value to the beginning of a Redis list and trims it to the specified length.
func (r *RedisClient) PushToList(key string, value interface{}, maxLength int64) error {
	err := r.client.LPush(ctx, key, value).Err()
	if err != nil {
		log.Printf("Could not push value to list %s: %v", key, err)
		return err
	}

	// Trim the list to the specified max length
	err = r.client.LTrim(ctx, key, 0, maxLength-1).Err()
	if err != nil {
		log.Printf("Could not trim list %s: %v", key, err)
		return err
	}
	return nil
}

// GetList retrieves the latest items from a Redis list up to the specified max length.
func (r *RedisClient) GetList(key string, maxLength int64) ([]string, error) {
	vals, err := r.client.LRange(ctx, key, 0, maxLength-1).Result()
	if err != nil {
		log.Printf("Could not get list %s: %v", key, err)
		return nil, err
	}
	return vals, nil
}
