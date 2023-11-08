// client реализует клиента для отправки метрик по HTTP
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/verifier"
)

type httpType int

const (
	JSONType httpType = iota
	RESTType
)

// Endpoint'ы для отправки метрик.
const (
	UpdateURL     = "/update/{valType}/{name}/{value}"
	JSONUpdateURL = "/update/"
	// Для отправки сразу пачки метрик
	JSONUpdatesURL = "/updates/"
)

// httpClient отправляет метрики и подписывает их ключом key.
type httpClient struct {
	client    *resty.Client
	signkey   string
	pubkey    []byte
	transport httpType
	server    string
}

func newClient(server string, options ...FOption) *httpClient {
	client := &httpClient{
		client: resty.New().
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second),
		server: server,
	}
	for _, o := range options {
		o(client)
	}
	return client
}

// NewREST - создаёт httpClient для отправки метрик через REST, принимает в качестве аргументов функции, например:
//
// func SetKey(key string) func(*httpClient)
func NewREST(server string, options ...FOption) *httpClient {
	c := newClient(server, options...)
	c.transport = RESTType
	return c
}

// NewJSON - создаёт httpClient для отправки метрик через REST, принимает в качестве аргументов функции, например:
//
// func SetKey(key string) func(*httpClient)
func NewJSON(server string, options ...FOption) *httpClient {
	c := newClient(server, options...)
	c.transport = JSONType
	return c
}

func (c *httpClient) Post(ctx context.Context, m models.Metric) error {
	switch c.transport {
	case RESTType:
		return c.postREST(ctx, c.server, m)
	case JSONType:
		return c.postJSON(ctx, c.server, m)
	}
	return fmt.Errorf("unknown transport type")
}

// Post отправляет метрику через REST-запрос
//
// path - адрес сервера, например "localhost:8080"
func (c *httpClient) postREST(ctx context.Context, server string, m models.Metric) error {
	val, err := m.GetValue()
	if err != nil {
		return err
	}
	realIP, err := getRealIP()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("http://%s%s", server, UpdateURL)
	_, err = c.client.R().
		SetContext(ctx).
		SetHeader("Content-type", "text/plain").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", realIP.String()).
		SetPathParams(map[string]string{
			"valType": m.MType,
			"name":    m.ID,
			"value":   val,
		}).Post(query)

	return err
}

// postJSON отправляет метрику в виде JSON-строки, дополнительно сжимая её с помощью gzip и подписывая с помощью httpClient.key.
func (c *httpClient) postJSON(ctx context.Context, server string, m models.Metric) error {
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
	realIP, err := getRealIP()
	if err != nil {
		return err
	}

	Headers := map[string]string{
		"Content-type":     "application/json",
		"Accept-Encoding":  "gzip",
		"Content-Encoding": "gzip",
		"X-Real-IP":        realIP.String(),
	}
	if c.signkey != "" {
		signature, err := verifier.Sign(buf, []byte(c.signkey))
		if err != nil {
			return err
		}
		Headers[verifier.HashHeader] = base64.StdEncoding.EncodeToString(signature)
	}
	query := fmt.Sprintf("http://%s%s", server, JSONUpdateURL)

	_, err = c.client.R().
		SetContext(ctx).
		SetHeaders(Headers).
		SetBody(body).
		Post(query)
	return err
}

func getRealIP() (net.Addr, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var realIP net.Addr
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				realIP = a
			}
		}
	}
	return realIP, nil
}
