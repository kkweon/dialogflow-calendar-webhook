package main

import (
	"net/http"
	"os"

	dialogflow_calendar_webhook "github.com/kkweon/dialogflow-calendar-webhook"
)

func main() {
	http.HandleFunc("/", dialogflow_calendar_webhook.MainHTTP)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
