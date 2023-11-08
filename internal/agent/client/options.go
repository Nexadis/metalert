// Задает опции для конструктора httpClient.
package client

// SetSignKey определяет ключ для подписи отправляемых метрик.
func SetSignKey(key string) FOption {
	return func(hc *httpClient) {
		hc.signkey = key
	}
}

// SetPubKey устанавливает публичный ключ, с помощью которого будет производиться шифрование трафика
func SetPubKey(key []byte) FOption {
	return func(hc *httpClient) {
		hc.pubkey = key
	}
}

type FOption func(*httpClient)
