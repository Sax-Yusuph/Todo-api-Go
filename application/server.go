package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func (a *App) StartServer(ctx context.Context) error {
	if err := a.rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	if err := os.Setenv("JWT_SECRET_KEY", "some-random-value"); err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}

		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeoutCtx)
	}
}
