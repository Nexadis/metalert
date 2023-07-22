package verifier

import (
	"crypto/hmac"
	"crypto/sha256"
)

const HashHeader = `Hash`

func Sign(body []byte, key []byte) ([]byte, error) {
	sign := hmac.New(sha256.New, key)
	_, err := sign.Write(body)
	if err != nil {
		return nil, nil
	}
	return sign.Sum(nil), nil
}
