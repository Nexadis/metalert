package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/Nexadis/metalert/internal/utils/verifier"
)

var (
	ErrorCheckHash   = ("verify hash: %w")
	ErrorInvalidHash = errors.New("invalid hash")
)

type verifiedWriter struct {
	http.ResponseWriter
	Writer io.Writer
	key    string
}

func (vw *verifiedWriter) Write(data []byte) (int, error) {
	signature, err := verifier.Sign(data, []byte(vw.key))
	if err != nil {
		return 0, err
	}
	vw.Header().Set(verifier.HashHeader, base64.StdEncoding.EncodeToString(signature))
	return vw.Writer.Write(data)
}

func (s *httpServer) WithVerify(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.Key == "" {
			h.ServeHTTP(w, r)
			return
		}
		gotSignature := r.Header.Get(verifier.HashHeader)
		logger.Info("GOT HASH:", gotSignature, r.Header)
		if gotSignature == "" {
			h.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Errorf(ErrorCheckHash, err).Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		newBody := io.NopCloser(bytes.NewBuffer(body))
		r.Body = newBody
		signature, err := verifier.Sign(body, []byte(s.config.Key))
		if err != nil {
			http.Error(w, fmt.Errorf(ErrorCheckHash, err).Error(), http.StatusInternalServerError)
			return
		}
		strSignature := base64.StdEncoding.EncodeToString(signature)

		if gotSignature != strSignature {
			logger.Error(ErrorInvalidHash.Error(), gotSignature+"!="+strSignature)
			http.Error(w, ErrorInvalidHash.Error(), http.StatusBadRequest)
			return
		}
		logger.Info("Signature is good")
		w = &verifiedWriter{
			ResponseWriter: w,
			Writer:         w,
			key:            s.config.Key,
		}

		h.ServeHTTP(w, r)
	}
}
