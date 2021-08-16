package dialogflow_calendar_webhook

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const calendarID = "l9c4qhf35d3102u4mmsmlggeqo@group.calendar.google.com"
const outputContextNameForCalendarEvent = "calendar-event"

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
	webhookRequest := &dialogflowpb.WebhookRequest{}
	bs, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.WithError(err).Error("failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := protojson.Unmarshal(bs, webhookRequest); err != nil {
		log.WithError(err).Warn("failed to unmarshal dialogflowpb.WebhookRequest")
		return
	}

	log.WithField("webhookRequest", webhookRequest.String()).Info("received the request")

	switch webhookRequest.GetQueryResult().GetIntent().GetDisplayName() {
	case "event.new":
		handleEventCreate(webhookRequest, w)
		return
	case "ping":
		handlePing(webhookRequest, w)
		return
	case "event.delete":
		handleEventDelete(webhookRequest, w)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getContextNameForCalendarEvent(sessionID string) string {
	return fmt.Sprintf("%s/contexts/%s", sessionID, outputContextNameForCalendarEvent)
}

func findEventIDFromContext(request *dialogflowpb.WebhookRequest) string {
	for _, context := range request.GetQueryResult().GetOutputContexts() {
		if context.GetName() == getContextNameForCalendarEvent(request.GetSession()) {
			return context.GetParameters().GetFields()["eventID"].GetStringValue()
		}
	}
	return ""
}

func handleEventDelete(req *dialogflowpb.WebhookRequest, w http.ResponseWriter) {
	eventID := findEventIDFromContext(req)
	if eventID == "" {
		sendMessageToDialogflow(w, req, "I wasn't able to find the event", nil)
		return
	}

	if err := calendarService.Events.Delete(calendarID, eventID).Do(); err != nil {
		log.WithError(err).Warn("failed to delete the event")
		sendMessageToDialogflow(w, req, "I wasn't able to delete the event", nil)
		return
	}

	sendMessageToDialogflow(w, req, "Event was deleted", nil)
}

func handlePing(req *dialogflowpb.WebhookRequest, w http.ResponseWriter) {
	sendMessageToDialogflow(w, req, "pong", nil)
}

func sendMessageToDialogflow(w http.ResponseWriter, req *dialogflowpb.WebhookRequest, message string, payload *structpb.Struct) {
	resp := dialogflowpb.WebhookResponse{
		FulfillmentMessages: []*dialogflowpb.Intent_Message{
			{
				Message: &dialogflowpb.Intent_Message_Text_{
					Text: &dialogflowpb.Intent_Message_Text{
						Text: []string{message}}},
			},
		},
		OutputContexts: []*dialogflowpb.Context{
			{
				Name:          getContextNameForCalendarEvent(req.GetSession()),
				LifespanCount: 5,
				Parameters:    payload,
			},
		},
	}

	bs, err := protojson.Marshal(&resp)
	if err != nil {
		log.WithError(err).Error("failed to marshal response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(bs)
}

func handleEventCreate(webhookRequest *dialogflowpb.WebhookRequest, w http.ResponseWriter) {
	params, err := extractCreateEventParams(webhookRequest)
	if err != nil {
		log.WithError(err).Warn("failed to parse event parameters")
		return
	}

	event, err := CreateEvent(params)
	if err != nil {
		log.WithError(err).Warn("failed to create a calendar event")
		return
	}

	payload, err := structpb.NewStruct(map[string]interface{}{
		"eventID": event.Id,
	})
	if err != nil {
		log.WithError(err).Warn("failed to create a payload")
		return
	}

	sendMessageToDialogflow(w,
		webhookRequest,
		fmt.Sprintf("Event (%s) was created on %s. You can find your event at %s",
			params.title,
			params.startTime.Format(time.UnixDate),
			getCalendarLink(event.Id, calendarID)),
		payload)
}

func getCalendarLink(eventID, calendarID string) string {
	return "https://google.com/calendar/event?eid=" + strings.Trim(base64.StdEncoding.EncodeToString([]byte(eventID+" "+calendarID)), "=")
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
		endTime_ = startTime_.Add(time.Minute * 30)
	} else {
		endTime_, err = time.Parse(time.RFC3339, endTimeText)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse endTimeText")
		}
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
