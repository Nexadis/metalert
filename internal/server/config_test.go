package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Nexadis/metalert/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestLoadJSON(t *testing.T) {
	f, err := os.CreateTemp("", "test_config")
	name := f.Name()
	assert.NoError(t, err)
	testC := &Config{
		Address:   "some_address",
		Verbose:   true,
		SignKey:   "sign_key",
		CryptoKey: "crypto_key",
		Config:    name,
		DB: &storage.Config{
			DSN:             "dsn",
			Retry:           3,
			Restore:         true,
			Timeout:         5,
			StoreInterval:   2,
			FileStoragePath: "some_filepath",
		},
	}
	testC.SetDefault()
	data, err := json.Marshal(testC)
	assert.NoError(t, err)
	f.Write(data)
	f.Close()
	loadedConf := NewConfig()
	loadedConf.Config = name
	loadJSON(loadedConf)
	assert.Equal(t, testC, loadedConf)
}
