package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
	r.Body.Close()
}

func BenchmarkWithVerify(b *testing.B) {
	signKey := "TestKey"
	verifier := WithVerify(http.HandlerFunc(EmptyHandler), signKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		packet := strings.NewReader(`{"id":"name","type":"gauge","value":123.123}`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/update", packet)
		b.StartTimer()
		verifier(w, r)
	}
}
