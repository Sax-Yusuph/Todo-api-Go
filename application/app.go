package application

import (
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	if err := os.Setenv("JWT_SECRET_KEY", "some-random-value"); err != nil {
		panic(err)
	}

	app := &App{
		rdb: redis.NewClient(&redis.Options{}),
	}

	app.loadRoutes()

	return app
}
