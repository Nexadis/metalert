package mem

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type StateSaver interface {
	Restore(ctx context.Context, FileStoragePath string, Restore bool) error
	Save(ctx context.Context, FileStoragePath string) error
	SaveTimer(ctx context.Context, FileStoragePath string, interval int64)
}

// Save Записывает все метрики в файл
func (ms *Storage) Save(ctx context.Context, FileStoragePath string) error {
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
	metrics, err := ms.GetAll(ctx)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	logger.Error(encoder.Encode(metrics))

	return nil
}

// Restore Восстанавливает состояние хранилища из файла
func (ms *Storage) Restore(ctx context.Context, FileStoragePath string, Restore bool) error {
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
		logger.Error(err)
		return nil
	}
	defer file.Close()
	metrics := make([]*metrx.Metric, 1)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&metrics)
	if err != nil {
		return err
	}
	for _, m := range metrics {
		err = ms.Set(ctx, *m)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveTimer Сохраняет текущее состояние хранилища в файл с заданным интервалом. Также сохраняет всё при завершении контекста
func (ms *Storage) SaveTimer(ctx context.Context, FileStoragePath string, interval int64) {
	if interval <= 0 {
		interval = 1
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := ms.Save(ctx, FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
		case <-ctx.Done():
			err := ms.Save(ctx, FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
		}
	}
}
