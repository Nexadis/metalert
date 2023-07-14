package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/stretchr/testify/assert"
)

type req struct {
	method  string
	url     string
	body    string
	headers http.Header
}
type want struct {
	statusCode int
	name       string
	valType    string
	value      string
	body       string
	headers    http.Header
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
	storage := mem.NewMetricsStorage()
	config := NewConfig()
	server := &httpServer{
		nil,
		storage,
		config,
	}
	server.MountHandlers()
	return server
}

func TestUpdateHandlerURL(t *testing.T) {
	server := testServer()
	ctx := context.TODO()

	for _, test := range updateTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, result.StatusCode, test.want.statusCode)
			defer result.Body.Close()
			getted, _ := server.storage.Get(ctx, test.want.valType, test.want.name)
			assert.Equal(t, getted.GetValue(), test.want.value)
		})
	}
}

func TestValueHandlerURL(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range valueTests {
		t.Run(test.name, func(t *testing.T) {
			err := server.storage.Set(ctx, test.want.valType, test.want.name, test.want.value)
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
	ctx := context.TODO()
	for _, test := range valueTests {
		err := server.storage.Set(ctx, test.want.valType, test.want.name, test.want.value)
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

var httpHeaders = http.Header{
	"Content-type": []string{"application/json"},
}

var JSONUpdateTests = []testReq{
	{
		name: "Gauge valid",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type": "gauge",
				"value": 1.23
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    metrx.GaugeType,
			value:      "1.23",
			body:       "",
			headers:    httpHeaders,
		},
	},
	{
		name: "Gauge invalid",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type: "gauge",
				"value": 1.23
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
	{
		name: "Invalid Gauge value",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type": "gauge",
				"delta": 1
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
	{
		name: "Counter valid",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type": "counter",
				"delta": 1423
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    metrx.CounterType,
			value:      "1423",
			body:       "",
			headers:    httpHeaders,
		},
	},
	{
		name: "Invalid Counter",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type": "counter",
				"value": 1.23
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
	{
		name: "Invalid Content-type",
		request: req{
			method: http.MethodPost,
			url:    "/update/",
			body: `{
				"id": "name",
				"type": "counter",
				"value": 1.23
			}`,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
}

var JSONValueTests = []testReq{
	{
		name: "Gauge valid",
		request: req{
			method: http.MethodPost,
			url:    "/value/",
			body: `{
				"id": "name",
				"type": "gauge"
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    metrx.GaugeType,
			value:      "123.123",
			body:       `{"id":"name","type":"gauge","value":123.123}`,
			headers:    httpHeaders,
		},
	},
	{
		name: "Gauge invalid",
		request: req{
			method: http.MethodPost,
			url:    "/value/",
			body: `{
				"id": "name",
				"type: "gauge"
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
	{
		name: "Not Found Gauge value",
		request: req{
			method: http.MethodPost,
			url:    "/value/",
			body: `{
				"id": "notfound",
				"type": "gauge"
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusNotFound,
		},
	},
	{
		name: "Counter valid",
		request: req{
			method: http.MethodPost,
			url:    "/value/",
			body: `{
				"id": "ctr",
				"type": "counter"
			}`,
			headers: httpHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "ctr",
			valType:    metrx.CounterType,
			value:      "1423",
			body:       `{"id":"ctr","type":"counter","delta":1423}`,
			headers:    httpHeaders,
		},
	},
}

func TestUpdateHandlerJSON(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range JSONUpdateTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
			r.Header = test.request.headers
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			if result.StatusCode == http.StatusOK {
				getted, _ := server.storage.Get(ctx, test.want.valType, test.want.name)
				assert.Equal(t, getted.GetValue(), test.want.value)
			}
		})
	}
}

func TestValueHandlerJSON(t *testing.T) {
	server := testServer()

	ctx := context.TODO()
	for _, test := range JSONValueTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
			r.Header = test.request.headers
			w := httptest.NewRecorder()
			server.storage.Set(ctx, test.want.valType, test.want.name, test.want.value)
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			if result.StatusCode == http.StatusOK {
				body, err := io.ReadAll(result.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.want.body, string(body))
				assert.EqualValues(t, test.want.headers, r.Header)
			}
		})
	}
}
