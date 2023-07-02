package server

import (
	"encoding/json"
	"os"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

func saveStorage(s *httpServer) error {
	fileName := s.config.FileStoragePath
	if fileName == "" {
		return nil
	}
	logger.Info("Write metrics to file")
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	metrics, err := s.storage.Values()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(metrics)

	return nil
}
