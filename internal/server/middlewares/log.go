package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

type responseData struct {
	body   string
	size   int
	status int
}

type logWrite struct {
	w  http.ResponseWriter
	rd *responseData
}

func (lw *logWrite) Write(b []byte) (int, error) {
	body := bytes.NewReader(b)
	read, err := io.ReadAll(body)
	lw.rd.body = string(read)
	if err != nil {
		panic(err)
	}

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

// Логирует информацию о запросе
// Method
// Status
// Duration
// Size
func WithLogging(h http.Handler) http.Handler {
	logFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{}
		lw := logWrite{
			w:  w,
			rd: responseData,
		}
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Error when read request", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recvBody, err := io.ReadAll(bytes.NewBuffer(buf))
		if err != nil {
			logger.Error("Error when read request", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBody := io.NopCloser(bytes.NewBuffer(buf))
		r.Body = newBody
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Info("URI", r.RequestURI,
			"Method", r.Method,
			"Status", responseData.status,
			"Duration", duration,
			"Size", responseData.size,
			"Recieved Body", string(recvBody),
			"Sended Body", responseData.body)
		logger.Info("Headers:")
		for k, v := range r.Header {
			logger.Info(k, strings.Join(v, ", "))
		}
	}
	return http.HandlerFunc(logFunc)
}
