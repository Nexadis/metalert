package mem

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type StateSaver interface {
	Restore(FileStoragePath string, Restore bool) error
	Save(FileStoragePath string) error
	SaveTimer(FileStoragePath string, interval int64)
}

func (ms *Storage) Save(FileStoragePath string) error {
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
	metrics, err := ms.GetAll()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(metrics)

	return nil
}

func (ms *Storage) Restore(FileStoragePath string, Restore bool) error {
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
		err = ms.Set(m.MType, m.ID, m.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *Storage) SaveTimer(FileStoragePath string, interval int64) {
	if interval <= 0 {
		interval = 1
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := ms.Save(FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
		case <-exit:
			err := ms.Save(FileStoragePath)
			if err != nil {
				logger.Info("Can't save storage")
			}
			return
		}
	}
}
