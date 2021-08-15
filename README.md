# Dialogflow + Go + Google Calendar + Cloud Function integration

## How to deploy

```bash
gcloud functions deploy dialogflow-calendar-integration-webhook --entry-point=MainHTTP --runtime=go113 --trigger-http --allow-unauthenticated
```
