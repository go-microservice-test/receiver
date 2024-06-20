package dbutils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func ConnectRedis(DBAddr, DBPassword string, DBName int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:        DBAddr,
		Password:    DBPassword,
		DB:          DBName,
		DialTimeout: 10 * time.Second, // allow 10 seconds to connect
	})

	// check connection with empty context
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v\n", err)
	}
	fmt.Println("Connected to Redis")
	return rdb
}
