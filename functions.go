package dialogflow_calendar_webhook

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"net/http"
	"os"
	"sync"
	"time"
)

const calendarID = "l9c4qhf35d3102u4mmsmlggeqo@group.calendar.google.com"

var startTimeKeys = []string{
	"startDate", "startTime", "startDateTime", "date_time",
}

var endTimeKeys = []string{
	"endDate", "endTime", "endDateTime",
}

var calendarService *calendar.Service
var isTestMode bool

func init() {
	var once sync.Once

	once.Do(func() {
		var err error

		apiKey := os.Getenv("GCP_API_KEY")
		if apiKey != "" {
			isTestMode = true
			calendarService, err = calendar.NewService(context.Background(), option.WithAPIKey(apiKey))
		} else {
			calendarService, err = calendar.NewService(context.Background())
		}

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

// MainHTTP is the main entry function.
func MainHTTP(w http.ResponseWriter, r *http.Request) {
	webhookRequest := dialogflowpb.WebhookRequest{}
	if err := jsonpb.Unmarshal(r.Body, &webhookRequest); err != nil {
		log.WithError(err).Warn("failed to unmarshal dialogflowpb.WebhookRequest")
		return
	}

	log.WithField("webhookRequest", webhookRequest.String()).Info("received the request")

	params, err := extractCreateEventParams(&webhookRequest)
	if err != nil {
		log.WithError(err).Warn("failed to parse event parameters")
		return
	}

	event, err := CreateEvent(params)
	if err != nil {
		log.WithError(err).Warn("failed to create a calendar event")
		return
	}

	resp := dialogflowpb.WebhookResponse{
		FulfillmentMessages: []*dialogflowpb.Intent_Message{
			{
				Message: &dialogflowpb.Intent_Message_Text_{
					Text: &dialogflowpb.Intent_Message_Text{
						Text: []string{fmt.Sprintf("Event (%s) was created on %s. You can find your event at %s", params.title, params.startTime.Format(time.UnixDate), getCalendarLink(event.Id, calendarID))}}},
			},
		},
	}

	marshaller := jsonpb.Marshaler{}
	_ = marshaller.Marshal(w, &resp)
}

func getCalendarLink(eventID, calendarID string) string {
	return "https://www.google.com/calendar/event?eid=" + base64.StdEncoding.EncodeToString([]byte(eventID+" "+calendarID))
}

// extractCreateEventParams extracts necessary parameters from the request for creating an event.
func extractCreateEventParams(webhookRequest *dialogflowpb.WebhookRequest) (*CreateEventParams, error) {
	fieldsMap := webhookRequest.GetQueryResult().GetParameters().GetFields()

	dateTimeFields := fieldsMap["date-time"].GetStructValue().GetFields()

	startTimeText := getFirstValue(dateTimeFields, startTimeKeys)
	endTimeText := getFirstValue(dateTimeFields, endTimeKeys)

	if startTimeText == "" {
		startTimeText = fieldsMap["date-time"].GetStringValue()
	}

	startTime_, err := time.Parse(time.RFC3339, startTimeText)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse startTimeText")
	}

	var endTime_ time.Time

	if endTimeText == "" {
		endTime_ = startTime_.Add(time.Hour)
	} else {
		endTime_, err = time.Parse(time.RFC3339, endTimeText)
	}

	location := fieldsMap["location"].GetStringValue()
	title := fieldsMap["title"].GetStringValue()

	params := &CreateEventParams{
		startTime: startTime_,
		endTime:   endTime_,
		title:     title,
		location:  location,
	}
	return params, nil
}

// getFirstValue returns the first value of the keys.
func getFirstValue(dateTimeFields map[string]*structpb.Value, keys []string) string {
	for _, key := range keys {
		if timeText := dateTimeFields[key].GetStringValue(); timeText != "" {
			return timeText
		}
	}

	return ""
}

func CreateEvent(params *CreateEventParams) (*calendar.Event, error) {
	if isTestMode {
		return &calendar.Event{Id: "test-event-ID"}, nil
	}

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
