package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandlerURL(t *testing.T) {

	type req struct {
		method string
		url    string
	}
	type want struct {
		statusCode int
		name       string
		valType    string
		value      string
	}
	testCases := []struct {
		name    string
		request req
		want    want
	}{
		{
			name: "Counter type, positive",
			request: req{
				method: http.MethodPost,
				url:    `/update/counter/positive/2`,
			},
			want: want{
				statusCode: http.StatusOK,
				valType:    "counter",
				name:       "positive",
				value:      "2",
			},
		},
		{
			name: "Counter type, big num",
			request: req{
				method: http.MethodPost,
				url:    `/update/counter/big/2985198054390`,
			},
			want: want{
				statusCode: http.StatusOK,
				valType:    "counter",
				name:       "big",
				value:      "2985198054390",
			},
		},
		{
			name: "Gauge type, positive",
			request: req{
				method: http.MethodPost,
				url:    `/update/gauge/positive/2`,
			},
			want: want{
				statusCode: http.StatusOK,
				valType:    "gauge",
				name:       "positive",
				value:      "2",
			},
		},
		{
			name: "Gauge type, negative",
			request: req{
				method: http.MethodPost,
				url:    `/update/gauge/negative/-2`,
			},
			want: want{
				statusCode: http.StatusOK,
				valType:    "gauge",
				name:       "negative",
				value:      "-2",
			},
		},
	}
	storage := metrx.NewMetricsStorage()
	server := httpServer{
		"",
		nil,
		storage,
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			server.UpdateHandler(w, r)
			result := w.Result()
			assert.Equal(t, result.StatusCode, test.want.statusCode)
			defer result.Body.Close()
			getted, _ := storage.Get(test.want.valType, test.want.name)
			assert.Equal(t, getted, test.want.value)
		})
	}
}
