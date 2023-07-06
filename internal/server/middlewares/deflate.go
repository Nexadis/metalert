package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

const StandardCompression = `gzip`

func isEncoded(r *http.Request, algorithm string) bool {
	contentEncoding := r.Header.Get("Content-Encoding")
	return strings.Contains(contentEncoding, algorithm)
}

func canEncode(r *http.Request, algorithm string) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, algorithm)
}

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *compressWriter) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

func WithDeflate(h http.Handler) http.Handler {
	deflate := func(w http.ResponseWriter, r *http.Request) {
		if isEncoded(r, StandardCompression) {
			logger.Info("Got compressed content")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		if canEncode(r, StandardCompression) {
			logger.Info("Send compressed content")
			w.Header().Set("Content-Encoding", StandardCompression)
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w = &compressWriter{
				ResponseWriter: w,
				Writer:         gz,
			}
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(deflate)
}
