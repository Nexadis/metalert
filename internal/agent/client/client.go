// client реализует клиента для отправки метрик по HTTP
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nexadis/metalert/internal/utils/verifier"
)

// MetricPoster интерфейс для отправки метрик как через URL, так и JSON-объектами.
type MetricPoster interface {
	Post(ctx context.Context, path, valType, name, value string) error
	PostObj(ctx context.Context, path string, obj interface{}) error
}

// httpClient отправляет метрики и подписывает их ключом key.
type httpClient struct {
	client *resty.Client
	key    string
}

// NewHTTP - конструктор для httpClient, принимает в качестве аргументов функции, например:
//
//	func SetKey(key string) func(*httpClient)
func NewHTTP(options ...func(*httpClient)) *httpClient {
	client := &httpClient{
		client: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second),
	}
	for _, o := range options {
		o(client)
	}
	return client
}

// Post отправляет метрику через Post-запрос, генерируя url по параметрам valType, name, value.
//
// path - адрес сервера, например "http://localhost:8080/update"
func (c *httpClient) Post(ctx context.Context, path, valType, name, value string) error {
	_, err := c.client.R().
		SetContext(ctx).
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

// PostObj отправляет метрику в виде JSON-строки, дополнительно сжимая её с помощью gzip и подписывая с помощью httpClient.key.
func (c *httpClient) PostObj(ctx context.Context, path string, obj interface{}) error {
	buf, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	g := gzip.NewWriter(body)
	g.Write(buf)
	g.Close()
	Headers := map[string]string{
		"Content-type":     "application/json",
		"Accept-Encoding":  "gzip",
		"Content-Encoding": "gzip",
	}
	if c.key != "" {
		signature, err := verifier.Sign(buf, []byte(c.key))
		if err != nil {
			return err
		}
		Headers[verifier.HashHeader] = base64.StdEncoding.EncodeToString(signature)
	}

	_, err = c.client.R().
		SetContext(ctx).
		SetHeaders(Headers).
		SetBody(body).
		Post(path)
	return err
}
