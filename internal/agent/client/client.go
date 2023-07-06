package client

import "github.com/go-resty/resty/v2"

type MetricPoster interface {
	Post(path, valType, name, value string) error
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
		SetPathParams(map[string]string{
			"valType": valType,
			"name":    name,
			"value":   value,
		}).
		Post(path)
	return err
}
