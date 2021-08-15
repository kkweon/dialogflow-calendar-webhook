package dialogflow_calendar_webhook

import "net/http"

func MainHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Dialogflow Calendar Webhook!"))
}
