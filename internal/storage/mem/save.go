package mem

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type StateSaver interface {
	Restore(ctx context.Context, FileStoragePath string, Restore bool) error
	Save(ctx context.Context, FileStoragePath string) error
	SaveTimer(ctx context.Context, FileStoragePath string, interval int64)
}

// Записывает все метрики в файл
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
	encoder.Encode(metrics)

	return nil
}

// Восстанавливает состояние хранилища из файла
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
		return err
	}
	defer file.Close()
	metrics := make([]*metrx.Metrics, 1)
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

// Сохраняет текущее состояние хранилища в файл с заданным интервалом. Также сохраняет всё при завершении контекста
func (ms *Storage) SaveTimer(ctx context.Context, FileStoragePath string, interval int64) {
	if interval <= 0 {
		interval = 1
	}

	exit, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	defer stop()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := ms.Save(ctx, FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
		case <-exit.Done():
			err := ms.Save(ctx, FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
			os.Exit(0)
		}
	}
}
