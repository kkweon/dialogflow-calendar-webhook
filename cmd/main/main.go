package main

import (
	"context"
	"os"

	webhook "github.com/kkweon/dialogflow-calendar-webhook"
	log "github.com/sirupsen/logrus"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("using the port %s", port)

	if err := funcframework.RegisterHTTPFunctionContext(context.Background(), "/", webhook.MainHTTP); err != nil {
		log.WithError(err).Fatal("failed to register a function")
	}

	if err := funcframework.Start(port); err != nil {
		log.WithError(err).Fatalf("funcframework.Start(port: %s) has failed", port)
	}
}
