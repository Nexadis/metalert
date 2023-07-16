package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
)

type MetricPoster interface {
	Post(path, valType, name, value string) error
	PostObj(path string, obj interface{}) error
}

type httpClient struct {
	client *resty.Client
}

func NewHTTP() MetricPoster {
	return &httpClient{
		client: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second),
	}
}

func (c *httpClient) Post(path, valType, name, value string) error {
	_, err := c.client.R().
		SetHeader("Content-type", "text/plain").
		SetHeader("Accept-Encoding", "gzip").
		SetPathParams(map[string]string{
			"valType": valType,
			"name":    name,
			"value":   value,
		}).
		Post(path)
	return err
}

func (c *httpClient) PostObj(path string, obj interface{}) error {
	buf, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	g := gzip.NewWriter(body)
	g.Write(buf)
	g.Close()

	_, err = c.client.R().
		SetHeader("Content-type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		Post(path)
	return err
}
