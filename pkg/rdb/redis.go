package rdb

import (
	"context"
	"crypto/tls" 
	"log"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Rdb *redis.Client
}

func Connect(Addr string, User string, Pass string) (*Redis, error) {
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:      Addr,
		Username:  User,
		Password:  Pass,
		TLSConfig: &tls.Config{},
	})

	if  _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal("Failed to connect to the redis server.")
		return nil, err
	}
	slog.Info("Connected to the redis server.")

	err := rdb.FlushAll(context.Background()).Err()
	if err != nil {
		log.Fatal("Failed to flush all keys in the redis server.")
		return nil, err
	}
	slog.Info("Flushed all keys in the redis server.")

	return &Redis{Rdb: rdb}, nil
}