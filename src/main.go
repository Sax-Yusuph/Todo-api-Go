package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdaurl"
	"github.com/sax-yusuph/todo/application"
)

func main() {
	app := application.New()

	handler := lambdaurl.Wrap(app.Router)
	lambda.Start(handler)

}

// Handler local server
func Handler() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := application.New()
	log.Fatal(app.StartServer(ctx))
}
