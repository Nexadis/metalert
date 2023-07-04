package storage

import (
	"encoding/json"
	"os"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func (s *MetricsStorage) SaveStorage(FileStoragePath string) error {
	fileName := FileStoragePath
	if fileName == "" {
		return nil
	}
	logger.Info("Write metrics to file")
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	metrics, err := s.GetAll()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(metrics)

	return nil
}

func (s *MetricsStorage) RestoreStorage(FileStoragePath string, Restore bool) error {
	fileName := FileStoragePath
	if fileName == "" {
		return nil
	}
	if !Restore {
		return nil
	}
	logger.Info("Read metrics from file")
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	metrics := make([]*metrx.MetricsString, 1)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metrics)
	if err != nil {
		return err
	}
	for _, m := range metrics {
		err = s.Set(m.MType, m.ID, m.Value)
		if err != nil {
			return err
		}
	}
	return nil
}
