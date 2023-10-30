package asymcrypt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Keys struct {
	private []byte
	public  []byte
}

func newKS(t *testing.T) Keys {
	name := os.TempDir() + "/key"
	err := NewPem(name)
	t.Log(err)
	public, err := ReadPem(name + "_pub.pem")
	t.Log(err)
	priv, err := ReadPem(name + "_priv.pem")
	t.Log(err)

	return Keys{
		priv,
		public,
	}
}

func TestEncryption(t *testing.T) {
	s := "Hello"
	keys := newKS(t)
	encrypted, err := Encrypt([]byte(s), keys.public)
	assert.NoError(t, err)
	decrypted, err := Decrypt(encrypted, keys.private)
	assert.NoError(t, err)
	assert.Equal(t, s, string(decrypted))
}
