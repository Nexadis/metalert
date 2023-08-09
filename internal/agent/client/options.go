package client

func SetKey(key string) func(*httpClient) {
	return func(hc *httpClient) {
		hc.key = key
	}
}
