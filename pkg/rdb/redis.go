package rdb

import (
	"context"
	"crypto/tls"
	"fmt"
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

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis server at %s: %w", Addr, err)
	}
	slog.Info("Connected to the redis server")

	// WARNING: FlushAll should NEVER be called in production!
	// This is only for development/testing. Remove this in production.
	// Uncomment only if you need to clear Redis during development:
	// err := rdb.FlushAll(context.Background()).Err()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to flush redis keys: %w", err)
	// }
	// slog.Info("Flushed all keys in the redis server.")

	return &Redis{Rdb: rdb}, nil
}
