package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func WithDecrypt(h http.Handler, privKey []byte) http.Handler {
	decrypt := func(w http.ResponseWriter, r *http.Request) {
		if privKey == nil {
			logger.Info("No key, no decrypt")
			h.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		if len(body) == 0 {
			return
		}
		logger.Info("Begin Decrypt")
		decrypted, err := asymcrypt.Decrypt(body, privKey)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(decrypted))
		logger.Info("Decrypted data:", string(decrypted))

		h.ServeHTTP(w, r)
		logger.Info("Decrypt middleware after serve")
	}
	return http.HandlerFunc(decrypt)
}
