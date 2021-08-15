package dialogflow_calendar_webhook

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"net/http"
	"sync"
	"time"
)

const calendarID = "l9c4qhf35d3102u4mmsmlggeqo@group.calendar.google.com"

var calendarService *calendar.Service

func init() {
	var once sync.Once

	once.Do(func() {
		var err error
		calendarService, err = calendar.NewService(context.Background())
		if err != nil {
			log.WithError(err).Fatal("failed to initialize calendar service")
			return
		}
	})
}

type CreateEventParams struct {
	startTime time.Time
	endTime   time.Time
	title     string
	location  string
}

func MainHTTP(w http.ResponseWriter, r *http.Request) {
	webhookRequest := dialogflowpb.WebhookRequest{}
	if err := jsonpb.Unmarshal(r.Body, &webhookRequest); err != nil {
		log.WithError(err).Warn("failed to unmarshal dialogflowpb.WebhookRequest")
		return
	}

	log.WithField("webhookRequest", webhookRequest.String()).Info("received the request")

	params := extractCreateEventParams(&webhookRequest)

	event, err := CreateEvent(params)
	if err != nil {
		log.WithError(err).Warn("failed to create a calendar event")
		return
	}

	_, _ = fmt.Fprint(w,
		"event was created at https://www.google.com/calendar/event?eid="+base64.StdEncoding.EncodeToString([]byte(event.Id+" "+calendarID)))
}

func extractCreateEventParams(webhookRequest *dialogflowpb.WebhookRequest) CreateEventParams {
	fieldsMap := webhookRequest.GetQueryResult().GetParameters().GetFields()

	startTimeText := fieldsMap["date-time"].GetStructValue().GetFields()["startTime"].GetStringValue()
	if startTimeText == "" {
		startTimeText = fieldsMap["date-time"].GetStringValue()
	}
	endTimeText := fieldsMap["date-time"].GetStructValue().GetFields()["endTime"].GetStringValue()

	startTime_, err := time.Parse(time.RFC3339, startTimeText)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"startTimeText": startTimeText,
		}).Warn("failed to parse startTimeText with RFC3339 format")
	}

	var endTime_ time.Time

	if endTimeText == "" {
		endTime_ = startTime_.Add(time.Hour)
	} else {
		endTime_, err = time.Parse(time.RFC3339, endTimeText)
	}

	location := fieldsMap["location"].GetStringValue()
	title := fieldsMap["title"].GetStringValue()

	params := CreateEventParams{
		startTime: startTime_,
		endTime:   endTime_,
		title:     title,
		location:  location,
	}
	return params
}

func CreateEvent(params CreateEventParams) (*calendar.Event, error) {
	event, err := calendarService.Events.Insert(calendarID,
		&calendar.Event{
			End: &calendar.EventDateTime{
				DateTime: params.endTime.Format(time.RFC3339),
			},
			EndTimeUnspecified: false,
			GuestsCanModify:    true,
			Location:           params.location,
			Start: &calendar.EventDateTime{
				DateTime: params.startTime.Format(time.RFC3339),
			},
			Summary: params.title,
		}).Do()
	if err != nil {
		return nil, err
	}

	return event, nil
}
