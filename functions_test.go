package dialogflow_calendar_webhook

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMainHTTP(t *testing.T) {
	request := dialogflow.WebhookRequest{
		QueryResult: &dialogflow.QueryResult{
			Intent: &dialogflow.Intent{
				DisplayName: "event.create",
			},
			Parameters: must(structpb.NewStruct(map[string]interface{}{
				"date-time": map[string]interface{}{
					"startDate": time.Date(2021, 8, 15, 4, 49, 0, 0, time.UTC).Format(time.RFC3339),
					"endDate":   time.Date(2021, 8, 15, 5, 49, 0, 0, time.UTC).Format(time.RFC3339),
				},
				"title": "title",
			})),
		},
	}
	marshaller := jsonpb.Marshaler{}
	str, err := marshaller.MarshalToString(&request)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(str))

	respRecorder := httptest.NewRecorder()

	handler := http.HandlerFunc(MainHTTP)
	handler.ServeHTTP(respRecorder, req)

	assert.Equal(t, http.StatusOK, respRecorder.Code)

	var webhookResponse dialogflow.WebhookResponse
	err = jsonpb.Unmarshal(respRecorder.Body, &webhookResponse)
	assert.NoError(t, err)

	assert.Contains(t, webhookResponse.GetFulfillmentMessages()[0].GetText().GetText()[0], "Event (title) was created on Sun Aug 15 04:49:00 UTC 2021. You can find your event at")
}

func must(val *structpb.Struct, err error) *structpb.Struct {
	if err != nil {
		panic(err)
	}
	return val
}
