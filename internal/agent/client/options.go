// Задает опции для конструктора httpClient.
package client

// SetKey определяет ключ для подписи отправляемых метрик.
func SetKey(key string) func(*httpClient) {
	return func(hc *httpClient) {
		hc.key = key
	}
}

// SetTransport определяет тип транспорта. Например REST
func SetTransport(transport TransportType) func(*httpClient) {
	return func(hc *httpClient) {
		hc.transport = transport
	}
}
