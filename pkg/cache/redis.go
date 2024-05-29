package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func Connect() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	rdb.Do(ctx, "flushall")

	return rdb
}

// GetCachedKey retrieves the value associated with the given key from the Redis cache.
// It takes a Redis client, a key string, and a pointer to a value of any type.
// The retrieved value is unmarshaled into the provided value pointer.
// If an error occurs during the retrieval or unmarshaling process, it is returned.
func GetCachedKey(rdb *redis.Client, key string, value any) error {
	res, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(res), value)
	return nil
}

func SetCachedKey(rdb *redis.Client, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, valueBytes, 5*time.Minute).Err()
}
