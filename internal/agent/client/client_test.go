package client

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/stretchr/testify/assert"
)

func TestHTTP(t *testing.T) {
	c := NewHTTP(SetKey("test-key"))
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
		},
	},
}

func TestPostREST(t *testing.T) {
	r := reqLogger{}
	s := httptest.NewServer(http.HandlerFunc(r.showHandler))
	defer s.Close()
	c := NewHTTP(SetKey("test-key"), SetTransport(RESTType))
	ctx := context.Background()
	path := fmt.Sprintf("%s%s", s.URL, "/update/{valType}/{name}/{value}")
	for _, test := range postTests {
		t.Run(test.name, func(t *testing.T) {
			m, err := metrx.NewMetric(test.m.name, test.m.mtype, test.m.val)
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
		},
	},
}

func TestPostJSON(t *testing.T) {
	r := reqLogger{}
	s := httptest.NewServer(http.HandlerFunc(r.showHandler))
	defer s.Close()
	c := NewHTTP(SetKey("test-key"), SetTransport(JSONType))
	ctx := context.Background()
	path := fmt.Sprintf("%s%s", s.URL, "/update")
	for _, test := range postObjTests {
		t.Run(test.name, func(t *testing.T) {
			m, err := metrx.NewMetric(test.m.name, test.m.mtype, test.m.val)
			assert.NoError(t, err)
			err = c.Post(ctx, path, m)
			assert.NoError(t, err)
			assert.Equal(t, test.want.url, r.url)
			body := strings.NewReader(r.body)
			g, err := gzip.NewReader(body)
			assert.NoError(t, err)
			buf, err := io.ReadAll(g)
			assert.NoError(t, err)
			assert.NoError(t, g.Close())
			assert.JSONEq(t, test.want.body, string(buf))
			assert.Equal(t, test.want.method, r.method)
		},
		)
	}
}
