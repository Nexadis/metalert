package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/stretchr/testify/assert"
)

func TestHTTP(t *testing.T) {
	c := NewHTTP(SetSignKey("test-key"))
	assert.NotNil(t, c)
}

type reqLogger struct {
	method string
	body   string
	url    string
}

func (l *reqLogger) showHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()
	l.body = string(body)
	l.url = r.URL.Path
	l.method = r.Method
}

type metric struct {
	name  string
	mtype string
	val   string
}

type want struct {
	method string
	body   string
	url    string
	err    error
}

type testReq struct {
	name string
	m    metric
	want want
}

var postTests = []testReq{
	{
		"Post URL gauge",
		metric{
			"name",
			"gauge",
			"123.123",
		},
		want{
			http.MethodPost,
			"",
			"/update/gauge/name/123.123",
			nil,
		},
	},
	{
		"Post URL counter",
		metric{
			"newname",
			"counter",
			"1230",
		},
		want{
			http.MethodPost,
			"",
			"/update/counter/newname/1230",
			nil,
		},
	},
	{
		"Post URL invalid",
		metric{
			"name",
			"invalid",
			"some-value",
		},
		want{
			http.MethodPost,
			"",
			"/update/invalid/name/some-value",
			models.ErrorType,
		},
	},
}

func TestPostREST(t *testing.T) {
	r := reqLogger{}
	s := httptest.NewServer(http.HandlerFunc(r.showHandler))
	defer s.Close()
	c := NewHTTP(SetSignKey("test-key"), SetTransport(RESTType))
	ctx := context.Background()
	path := s.URL[len("http://"):]
	for _, test := range postTests {
		t.Run(test.name, func(t *testing.T) {
			m, err := models.NewMetric(test.m.name, test.m.mtype, test.m.val)
			if test.want.err != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			err = c.Post(ctx, path, m)
			assert.NoError(t, err)
			assert.Equal(t, test.want.url, r.url)
			assert.Equal(t, test.want.body, r.body)
			assert.Equal(t, test.want.method, r.method)
		},
		)
	}
}

var postObjTests = []testReq{
	{
		"Just object",
		metric{
			"name",
			"",
			"",
		},
		want{
			http.MethodPost,
			"\"name\"",
			"/update",
			models.ErrorMetrics,
		},
	},
	{
		"Valid counter",
		metric{
			"somec",
			models.CounterType,
			"123",
		},
		want{
			http.MethodPost,
			"",
			"/update/",
			nil,
		},
	},
	{
		"Valid gauge",
		metric{
			"someg",
			models.GaugeType,
			"123",
		},
		want{
			http.MethodPost,
			"",
			"/update/",
			nil,
		},
	},
}

type testJSON struct {
	name   string
	metric models.Metric
	want   want
}

func prepareJSONtests(tests []testReq) []testJSON {
	jsons := make([]testJSON, 0, len(postObjTests))
	for _, test := range tests {

		m, err := models.NewMetric(test.m.name, test.m.mtype, test.m.val)
		if err != nil {
			continue
		}
		jsoned, err := json.Marshal(m)
		test.want.body = string(jsoned)
		if err != nil {
			continue
		}
		jsons = append(jsons, testJSON{
			test.name,
			m,
			test.want,
		})
	}
	return jsons
}

func TestPostJSON(t *testing.T) {
	r := reqLogger{}
	s := httptest.NewServer(http.HandlerFunc(r.showHandler))
	defer s.Close()
	keyname := os.TempDir() + "/test-key"
	err := asymcrypt.NewPem(keyname)
	assert.NoError(t, err)
	pub, err := asymcrypt.ReadPem(keyname + "_pub.pem")
	assert.NoError(t, err)
	priv, err := asymcrypt.ReadPem(keyname + "_priv.pem")
	assert.NoError(t, err)
	c := NewHTTP(SetSignKey("test-key"), SetTransport(JSONType), SetPubKey(pub))
	ctx := context.Background()
	path := s.URL[len("http://"):]
	tests := prepareJSONtests(postObjTests)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.NoError(t, err)
			err = c.Post(ctx, path, test.metric)
			assert.NoError(t, err)
			assert.Equal(t, test.want.url, r.url)
			body := strings.NewReader(r.body)
			g, err := gzip.NewReader(body)
			assert.NoError(t, err)
			buf, err := io.ReadAll(g)
			assert.NoError(t, err)
			decrypted, err := asymcrypt.Decrypt(buf, priv)
			assert.NoError(t, err)
			assert.NoError(t, g.Close())
			assert.JSONEq(t, test.want.body, string(decrypted))
			assert.Equal(t, test.want.method, r.method)
		},
		)
	}
}
