package client

import (
	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/go-resty/resty/v2"
)

type MetricPoster interface {
	Post(path, valType, name, value string) error
	PostJSON(path string, m *metrx.Metrics) error
	PostJSONs(path string, m []metrx.Metrics) error
}

type httpClient struct {
	client *resty.Client
}

func NewHTTP() MetricPoster {
	return &httpClient{
		client: resty.New(),
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

func (c *httpClient) PostJSON(path string, m *metrx.Metrics) error {
	_, err := c.client.R().
		SetHeader("Content-type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(*m).
		Post(path)
	return err
}

func (c *httpClient) PostJSONs(path string, m []metrx.Metrics) error {
	_, err := c.client.R().
		SetHeader("Content-type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(m).
		Post(path)
	return err
}
