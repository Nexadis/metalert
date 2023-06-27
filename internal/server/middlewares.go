package server

import (
	"net/http"
	"time"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

type responseData struct {
	size   int
	status int
}

type logWrite struct {
	w  http.ResponseWriter
	rd *responseData
}

func (lw logWrite) Write(b []byte) (int, error) {
	size, err := lw.w.Write(b)
	lw.rd.size = size
	return size, err
}

func (lw logWrite) WriteHeader(status int) {
	lw.rd.status = status
	lw.w.WriteHeader(status)
}

func (lw logWrite) Header() http.Header {
	return lw.w.Header()
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := logWrite{
			w:  w,
			rd: &responseData{},
		}
		h.ServeHTTP(lw, r)
		duration := time.Since(start)
		logger.Info("URI", r.RequestURI, "Method", r.Method, "Duration", duration, "Size", lw.rd.size, "Status", lw.rd.status)
	})
}
