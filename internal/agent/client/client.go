// client реализует клиента для отправки метрик по HTTP
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/verifier"
)

// TransportType создаёт тип для видов передачи метрик
type TransportType string

// Константы для определения способа передачи метрик
const (
	RESTType TransportType = "REST"
	JSONType TransportType = "JSON"
)

// MetricPoster интерфейс для отправки метрик как через URL, так и JSON-объектами.
type MetricPoster interface {
	Post(ctx context.Context, path string, m models.Metric) error
}

// httpClient отправляет метрики и подписывает их ключом key.
type httpClient struct {
	client    *resty.Client
	transport TransportType
	signkey   string
	pubkey    []byte
}

// NewHTTP - конструктор для httpClient, принимает в качестве аргументов функции, например:
//
// func SetKey(key string) func(*httpClient)
func NewHTTP(options ...FOption) *httpClient {
	client := &httpClient{
		client: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second),
		transport: RESTType,
	}
	for _, o := range options {
		o(client)
	}
	return client
}

func (c *httpClient) Post(ctx context.Context, path string, m models.Metric) error {
	switch c.transport {
	case RESTType:
		return c.postREST(ctx, path, m)
	case JSONType:
		return c.postJSON(ctx, path, m)
	}
	return fmt.Errorf("unknown transport type %s", c.transport)
}

// Post отправляет метрику через REST-запрос
//
// path - адрес сервера, например "http://localhost:8080/update"
func (c *httpClient) postREST(ctx context.Context, path string, m models.Metric) error {
	val, err := m.GetValue()
	if err != nil {
		return err
	}
	_, err = c.client.R().
		SetContext(ctx).
		SetHeader("Content-type", "text/plain").
		SetHeader("Accept-Encoding", "gzip").
		SetPathParams(map[string]string{
			"valType": m.MType,
			"name":    m.ID,
			"value":   val,
		}).
		Post(path)
	return err
}

// postJSON отправляет метрику в виде JSON-строки, дополнительно сжимая её с помощью gzip и подписывая с помощью httpClient.key.
func (c *httpClient) postJSON(ctx context.Context, path string, m models.Metric) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	encrypted := buf
	if c.pubkey != nil {
		encrypted, err = asymcrypt.Encrypt(buf, c.pubkey)
		if err != nil {
			return err
		}
	}
	body := &bytes.Buffer{}
	g := gzip.NewWriter(body)
	_, err = g.Write(encrypted)
	if err != nil {
		return err
	}
	err = g.Close()
	if err != nil {
		return err
	}
	Headers := map[string]string{
		"Content-type":     "application/json",
		"Accept-Encoding":  "gzip",
		"Content-Encoding": "gzip",
	}
	if c.signkey != "" {
		signature, err := verifier.Sign(buf, []byte(c.signkey))
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
