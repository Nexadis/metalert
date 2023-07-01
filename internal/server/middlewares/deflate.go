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
	encode := strings.Contains(acceptEncoding, algorithm)
	return encode
}

func Decompress(r *http.Request) (io.ReadCloser, error) {
	var reader io.ReadCloser
	if isEncoded(r, StandardCompression) {
		logger.Info("Got compressed content")
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		reader = io.NopCloser(gz)
	} else {
		reader = r.Body
	}
	return reader, nil
}

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *compressWriter) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

func WithDeflate(h http.Handler) http.Handler {
	var writer io.Writer
	deflate := func(w http.ResponseWriter, r *http.Request) {
		reader, err := Decompress(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = reader
		if canEncode(r, StandardCompression) {
			logger.Info("Send compressed content")
			w.Header().Set("Content-Encoding", StandardCompression)
			gz := gzip.NewWriter(w)
			writer = gz
			defer gz.Close()
		} else {
			writer = w
		}
		w = &compressWriter{
			ResponseWriter: w,
			Writer:         writer,
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(deflate)
}
