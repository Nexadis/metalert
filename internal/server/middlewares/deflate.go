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

func Compress(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	var writer io.Writer
	if canEncode(r, StandardCompression) {
		logger.Info("Send compressed content")
		w.Header().Set("Content-Encoding", StandardCompression)
		gz := gzip.NewWriter(w)
		writer = gz
	} else {
		writer = w
	}
	return &compressWriter{
		ResponseWriter: w,
		Writer:         writer,
	}
}

func WithDeflate(h http.Handler) http.Handler {
	deflate := func(w http.ResponseWriter, r *http.Request) {
		reader, err := Decompress(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer reader.Close()
		r.Body = reader
		writer := Compress(w, r)
		h.ServeHTTP(writer, r)
	}
	return http.HandlerFunc(deflate)
}
