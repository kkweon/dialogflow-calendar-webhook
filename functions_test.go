package dialogflow_calendar_webhook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMainHTTP(t *testing.T) {

	testCases := []struct {
		request *dialogflow.WebhookRequest
		want    string
		wantErr bool
	}{
		{
			request: &dialogflow.WebhookRequest{
				QueryResult: &dialogflow.QueryResult{
					Intent: &dialogflow.Intent{
						DisplayName: "event.new",
					},
					Parameters: must(structpb.NewStruct(map[string]interface{}{
						"date-time": map[string]interface{}{
							"startDate": time.Date(2021, 8, 15, 4, 49, 0, 0, time.UTC).Format(time.RFC3339),
							"endDate":   time.Date(2021, 8, 15, 5, 49, 0, 0, time.UTC).Format(time.RFC3339),
						},
						"title": "title",
					})),
				},
			},
			want: "Event (title) was created on Sun Aug 15 04:49:00 UTC 2021. You can find your event at",
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.request.QueryResult.Intent.DisplayName, func(t *testing.T) {
			str, err := protojson.Marshal(testCase.request)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(str))

			respRecorder := httptest.NewRecorder()

			handler := http.HandlerFunc(MainHTTP)
			handler.ServeHTTP(respRecorder, req)

			assert.Equal(t, http.StatusOK, respRecorder.Code)

			var webhookResponse dialogflow.WebhookResponse
			bs, err := ioutil.ReadAll(respRecorder.Body)
			assert.NoError(t, err)

			err = protojson.Unmarshal(bs, &webhookResponse)
			assert.NoError(t, err)

			assert.Contains(t, webhookResponse.GetFulfillmentMessages()[0].GetText().GetText()[0], testCase.want)
		})
	}

}

func must(val *structpb.Struct, err error) *structpb.Struct {
	if err != nil {
		panic(err)
	}
	return val
}
