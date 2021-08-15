# Dialogflow + Go + Google Calendar + Cloud Function integration

Also, I had to add Cloud Function's service account to my calendar so that the service account can create an event on my calendar.

## How to deploy

```bash
gcloud functions deploy dialogflow-calendar-integration-webhook --entry-point=MainHTTP --runtime=go113 --trigger-http --allow-unauthenticated
```
