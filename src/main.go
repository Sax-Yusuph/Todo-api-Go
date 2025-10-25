package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/sax-yusuph/todo/application"
)

func main() {
	config, err := application.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(config)

	lambdaurl.Start(app.Router)

}

// Handler local server
func LocalServerHandler() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := application.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(config)
	log.Fatal(app.StartServer(ctx))
}
