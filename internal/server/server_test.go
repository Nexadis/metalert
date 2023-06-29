package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/stretchr/testify/assert"
)

type req struct {
	method string
	url    string
}
type want struct {
	statusCode int
	name       string
	valType    string
	value      string
	body       string
}

type testReq struct {
	name    string
	request req
	want    want
}

var updateTests = []testReq{
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

var valueTests = []testReq{
	{
		name: "Counter type, positive",
		request: req{
			method: http.MethodGet,
			url:    `/value/counter/positive`,
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
			method: http.MethodGet,
			url:    `/value/counter/big`,
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
			method: http.MethodGet,
			url:    `/value/gauge/positive`,
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
			method: http.MethodGet,
			url:    `/value/gauge/negative`,
		},
		want: want{
			statusCode: http.StatusOK,
			valType:    "gauge",
			name:       "negative",
			value:      "-2",
		},
	},
}

func getAll(tests []testReq) string {
	var resultValues string
	for _, test := range tests {
		resultValues += fmt.Sprintf("%s=%s\n", test.want.name, test.want.value)
	}
	return resultValues
}

var valuesTests = []testReq{
	{
		name: "All values",
		request: req{
			method: http.MethodGet,
			url:    `/value`,
		},
		want: want{
			statusCode: http.StatusOK,
			body:       getAll(valueTests),
		},
	},
}

func testServer() *httpServer {
	storage := metrx.NewMetricsStorage()
	server := &httpServer{
		"http://localhost:8080",
		nil,
		storage,
	}
	server.MountHandlers()
	return server
}

func TestUpdateHandlerURL(t *testing.T) {
	server := testServer()

	for _, test := range updateTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, result.StatusCode, test.want.statusCode)
			defer result.Body.Close()
			getted, _ := server.storage.Get(test.want.valType, test.want.name)
			assert.Equal(t, getted.Value, test.want.value)
		})
	}
}

func TestValueHandlerURL(t *testing.T) {
	server := testServer()
	for _, test := range valueTests {
		t.Run(test.name, func(t *testing.T) {
			err := server.storage.Set(test.want.valType, test.want.name, test.want.value)
			assert.NoError(t, err)
			r := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			getted, _ := io.ReadAll(result.Body)
			assert.Equal(t, test.want.value, string(getted))
		})
	}
}

func TestValuesHandlerURL(t *testing.T) {
	server := testServer()
	for _, test := range valueTests {
		err := server.storage.Set(test.want.valType, test.want.name, test.want.value)
		assert.NoError(t, err)
	}
	for _, test := range valuesTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			body, _ := io.ReadAll(result.Body)
			getted := strings.Split(string(body), "\n")
			wanted := strings.Split(test.want.body, "\n")

			assert.ElementsMatch(t, getted, wanted)
		})
	}
}
