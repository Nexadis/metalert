package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/verifier"
)

func TestMain(m *testing.M) {
	c := NewConfig()
	c.SetDefault()
	c.GRPC = ":1333"
	c.DB.Restore = false
	s, err := New(c)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Run(ctx)
	time.Sleep(time.Second)
	code := m.Run()
	os.Exit(code)
}

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
	values     []stringMetrics
	body       string
	headers    http.Header
}

type stringMetrics struct {
	name    string
	valType string
	value   string
}

type testReq struct {
	name    string
	request req
	want    want
}

func TestNewServer(t *testing.T) {
	c := Config{
		DB: storage.NewConfig(),
	}

	_, err := New(&c)
	assert.NoError(t, err)

	c.DB.DSN = "invalid dsn"

	_, err = New(&c)
	assert.NoError(t, err)
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
			url:    `/update/gauge/positiveg/2`,
		},
		want: want{
			statusCode: http.StatusOK,
			valType:    "gauge",
			name:       "positiveg",
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
			url:    `/value/gauge/positiveg`,
		},
		want: want{
			statusCode: http.StatusOK,
			valType:    "gauge",
			name:       "positiveg",
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
	config.SignKey = "test_key"
	server := &httpServer{
		nil,
		storage,
		config,
		nil,
		nil,
	}
	server.MountHandlers()
	return server
}

func TestUpdateURL(t *testing.T) {
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
			getted, err := server.storage.Get(ctx, test.want.valType, test.want.name)
			assert.NoError(t, err)
			val, err := getted.GetValue()
			assert.NoError(t, err)
			assert.Equal(t, val, test.want.value)
		})
	}
}

func TestValueURL(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range valueTests {
		t.Run(test.name, func(t *testing.T) {
			m, err := models.NewMetric(
				test.want.name,
				test.want.valType,
				test.want.value,
			)
			assert.NoError(t, err)
			err = server.storage.Set(ctx, m)
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

func TestValuesURL(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range valueTests {
		m, err := models.NewMetric(
			test.want.name,
			test.want.valType,
			test.want.value,
		)
		assert.NoError(t, err)
		err = server.storage.Set(ctx, m)
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

var JSONHeaders = http.Header{
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
			headers: JSONHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    models.GaugeType,
			value:      "1.23",
			body:       "",
			headers:    JSONHeaders,
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
			headers: JSONHeaders,
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
			headers: JSONHeaders,
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
			headers: JSONHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    models.CounterType,
			value:      "1423",
			body:       "",
			headers:    JSONHeaders,
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
			headers: JSONHeaders,
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
			headers: JSONHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "name",
			valType:    models.GaugeType,
			value:      "123.123",
			body:       `{"id":"name","type":"gauge","value":123.123}`,
			headers:    JSONHeaders,
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
			headers: JSONHeaders,
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
			headers: JSONHeaders,
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
			headers: JSONHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			name:       "ctr",
			valType:    models.CounterType,
			value:      "1423",
			body:       `{"id":"ctr","type":"counter","delta":1423}`,
			headers:    JSONHeaders,
		},
	},
}

func TestUpdateJSON(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range JSONUpdateTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
			signature, err := verifier.Sign([]byte(test.request.body), []byte(server.config.SignKey))
			assert.NoError(t, err)
			r.Header.Set(verifier.HashHeader, base64.StdEncoding.EncodeToString(signature))
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			if result.StatusCode == http.StatusOK {
				getted, err := server.storage.Get(ctx, test.want.valType, test.want.name)
				assert.NoError(t, err)
				val, err := getted.GetValue()
				assert.NoError(t, err)
				assert.Equal(t, val, test.want.value)
			}
		})
	}
}

func TestValueJSON(t *testing.T) {
	server := testServer()

	ctx := context.TODO()
	for _, test := range JSONValueTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
			r.Header = test.request.headers
			w := httptest.NewRecorder()
			if test.want.statusCode == http.StatusOK {
				m, err := models.NewMetric(
					test.want.name,
					test.want.valType,
					test.want.value,
				)
				assert.NoError(t, err)
				server.storage.Set(ctx, m)
			}
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

var JSONUpdatesTests = []testReq{
	{
		name: "Invalid type",
		request: req{
			method: http.MethodPost,
			url:    "/updates/",
			body: `[{
				"id": "name",
				"type": "counrer",
				"value": 1.23
			}]`,
		},
		want: want{
			statusCode: http.StatusBadRequest,
		},
	},
	{
		name: "Many valid values",
		request: req{
			method: http.MethodPost,
			url:    "/updates/",
			body: `
				[
{"id":"name","type":"gauge","value":123.123},
{"id":"someGauge","type":"gauge","value":435.435},
{"id":"ctr","type":"counter","delta":1423},
{"id":"SomeCtr","type":"counter","delta":1445309},
{"id":"ctr","type":"counter","delta":1423}
				]
			
			`,
			headers: JSONHeaders,
		},
		want: want{
			statusCode: http.StatusOK,
			values: []stringMetrics{
				{
					name:    "name",
					valType: "gauge",
					value:   "123.123",
				},
				{
					name:    "someGauge",
					valType: "gauge",
					value:   "435.435",
				},
				{
					name:    "ctr",
					valType: "counter",
					value:   "2846",
				},
				{
					name:    "SomeCtr",
					valType: "counter",
					value:   "1445309",
				},
			},
		},
	},
}

func TestUpdatesJSON(t *testing.T) {
	server := testServer()
	ctx := context.TODO()
	for _, test := range JSONUpdatesTests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
			r.Header = test.request.headers
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, r)
			result := w.Result()
			assert.Equal(t, test.want.statusCode, result.StatusCode)
			defer result.Body.Close()
			if result.StatusCode == http.StatusOK {
				for _, want := range test.want.values {
					m, err := server.storage.Get(ctx, want.valType, want.name)
					assert.NoError(t, err)
					value, err := m.GetValue()
					assert.NoError(t, err)
					assert.Equal(t, want.value, value)
				}
			}
		})
	}
}

func BenchmarkUpdateURL(b *testing.B) {
	server := testServer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		test := updateTests[0]
		r := httptest.NewRequest(test.request.method, test.request.url, nil)
		w := httptest.NewRecorder()
		b.StartTimer()
		server.router.ServeHTTP(w, r)
	}
}

func BenchmarkUpdateJSON(b *testing.B) {
	server := testServer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		test := JSONUpdateTests[0]
		r := httptest.NewRequest(test.request.method, test.request.url, strings.NewReader(test.request.body))
		r.Header = test.request.headers
		w := httptest.NewRecorder()
		b.StartTimer()
		server.router.ServeHTTP(w, r)
	}
}

func ExampleNewServer() {
	c := NewConfig()
	c.SetDefault()
	c.Address = ":9090"
	c.GRPC = ":1333"
	s, err := New(c)
	if err != nil {
		// ... Handle error
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	go s.Run(ctx)
}

func ExamplehttpServer_DBPing() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "ping")
	r, err := http.Get(addr)
	if err != nil {
		log.Fatal(err) // ... Handle error
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// ... Handle error
	}
	defer r.Body.Close()
	fmt.Println(string(body))

	// Output:
	// DB is not connected
}

func ExamplehttpServer_Update() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "update/gauge/name/123.123")
	r, err := http.Post(addr, "", nil)
	if err != nil {
		// ... Handle error
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// ... Handle error
	}
	defer r.Body.Close()
	fmt.Println(string(body))

	// Output:
	// Value name type gauge updated
}

func ExamplehttpServer_Value() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "value/gauge/name")
	r, err := http.Get(addr)
	if err != nil {
		// ... Handle error
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// ... Handle error
	}
	defer r.Body.Close()
	fmt.Println(string(body))

	// Output:
	// 123.123
}

func ExamplehttpServer_Values() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "value")
	r, err := http.Get(addr)
	if err != nil {
		// ... Handle error
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// ... Handle error
	}
	defer r.Body.Close()
	fmt.Println(string(body))

	// Output:
	// name=123.123
}

func ExamplehttpServer_UpdateJSON() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "update")
	m, err := models.NewMetric("name", "gauge", "123.123")
	if err != nil {
		// ... Handle error
	}
	data, err := json.Marshal(m)
	if err != nil {
		// ... Handle error
	}
	rd := bytes.NewReader(data)
	r, err := http.Post(addr, "application/json", rd)
	if err != nil {
		// ... Handle error
	}
	r.Body.Close()
	fmt.Println(r.Status)

	// Output:
	// 200 OK
}

func ExamplehttpServer_Updates() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "updates/")
	ms := make([]models.Metric, 0, 10)
	for i := 0; i < 10; i++ {
		val := fmt.Sprintf("%d", i)
		m, err := models.NewMetric(val, "gauge", val)
		if err != nil {
			// ... Handle error
		}
		ms = append(ms, m)
	}

	data, err := json.Marshal(ms)
	if err != nil {
		// ... Handle error
	}
	rd := bytes.NewReader(data)
	r, err := http.Post(addr, "application/json", rd)
	if err != nil {
		// ... Handle error
	}
	r.Body.Close()
	fmt.Println(r.Status)

	// Output:
	// 200 OK
}

func ExamplehttpServer_ValueJSON() {
	addr := fmt.Sprintf("http://localhost:8080/%s", "value")
	m, err := models.NewMetric("name", "gauge", "123.123")
	if err != nil {
		// ... Handle error
	}
	data, err := json.Marshal(m)
	if err != nil {
		// ... Handle error
	}
	rd := bytes.NewReader(data)
	r, err := http.Post(addr, "application/json", rd)
	if err != nil {
		// ... Handle error
	}
	r.Body.Close()
	fmt.Println(r.Status)

	// Output:
	// 200 OK
}
