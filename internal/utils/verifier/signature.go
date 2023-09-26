// Реализация подписи
package verifier

import (
	"crypto/hmac"
	"crypto/sha256"
)

// HashHeader - Алгоритм для hmac
const HashHeader = `HashSHA256`

// Sign Создаёт подпись данных на ключе
func Sign(body []byte, key []byte) ([]byte, error) {
	sign := hmac.New(sha256.New, key)
	_, err := sign.Write(body)
	if err != nil {
		return nil, nil
	}
	return sign.Sum(nil), nil
}
