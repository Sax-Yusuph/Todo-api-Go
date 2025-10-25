package application

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	Router http.Handler
	rdb    *redis.Client
	Config Config
}

func New(config Config) *App {
	app := &App{
		rdb:    redis.NewClient(config.RedisOptions),
		Config: config,
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()
	if err := app.rdb.Ping(timeoutCtx).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	app.loadRoutes()

	return app
}
