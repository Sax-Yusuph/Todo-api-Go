package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/sax-yusuph/todo/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := application.New()
	log.Fatal(app.StartServer(ctx))
}
