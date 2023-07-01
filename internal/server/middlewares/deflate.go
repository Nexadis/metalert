package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
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
		defer r.Body.Close()
		reader, err := Decompress(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = reader
		writer := Compress(w, r)
		h.ServeHTTP(writer, r)
	}
	return http.HandlerFunc(deflate)
}
