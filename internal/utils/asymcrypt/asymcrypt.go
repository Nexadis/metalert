package asymcrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

func NewPem(filename string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	// кодируем ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	if err != nil {
		return err
	}
	privname := filename + "_priv.pem"
	pubname := filename + "_pub.pem"

	err = os.WriteFile(privname, privateKeyPEM.Bytes(), 0777)
	if err != nil {
		return err
	}
	logger.Info("Created Pivate key:", privname)
	logger.Info("Created PublicKey key:", pubname)
	return os.WriteFile(pubname, publicKeyPEM.Bytes(), 0777)
}

func ReadPem(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM file")
	}
	logger.Info(block.Headers)
	return block.Bytes, nil
}

func Decrypt(data []byte, privKey []byte) ([]byte, error) {
	key, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	decrypted, err := rsa.DecryptPKCS1v15(nil, key, data)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func Encrypt(body []byte, pubKey []byte) ([]byte, error) {
	key, err := x509.ParsePKCS1PublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, key, body)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}
