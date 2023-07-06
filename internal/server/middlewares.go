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

func (lw *logWrite) Write(b []byte) (int, error) {
	lw.WriteHeader(http.StatusOK)
	size, err := lw.w.Write(b)
	lw.rd.size += size
	return size, err
}

func (lw *logWrite) WriteHeader(statusCode int) {
	lw.rd.status = statusCode
	lw.w.WriteHeader(statusCode)
}

func (lw *logWrite) Header() http.Header {
	return lw.w.Header()
}

func WithLogging(h http.Handler) http.Handler {
	logFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{}
		lw := logWrite{
			w:  w,
			rd: responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Info("URI", r.RequestURI, "Method", r.Method, "Status", responseData.status, "Duration", duration, "Size", responseData.size)
	}
	return http.HandlerFunc(logFunc)
}
