package asymcrypt

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

func ReadPem(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM file")
	}
	return block.Bytes, nil
}

func Decrypter(body io.Reader, privKey []byte) (io.Reader, error) {
	key, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	decrypted, err := rsa.DecryptPKCS1v15(nil, key, data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(decrypted), nil
}

func Encrypter(body io.Reader, pubKey []byte) (io.Reader, error) {
	key, err := x509.ParsePKCS1PublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	encrypted, err := rsa.EncryptPKCS1v15(nil, key, data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(encrypted), nil
}
