package asymcrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Keys struct {
	private []byte
	public  []byte
}

func newKS() Keys {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	PrivKey := x509.MarshalPKCS1PrivateKey(key)
	PubKey := x509.MarshalPKCS1PublicKey(&key.PublicKey)
	return Keys{
		PrivKey,
		PubKey,
	}
}

func TestEncryption(t *testing.T) {
	s := "Hello"
	keys := newKS()
	body := strings.NewReader(s)
	encrypted, err := Encrypt(body, keys.public)
	assert.NoError(t, err)
	decrypted, err := Decrypt(encrypted, keys.private)
	assert.NoError(t, err)
	result, err := io.ReadAll(decrypted)
	assert.NoError(t, err)
	assert.Equal(t, s, string(result))
}
