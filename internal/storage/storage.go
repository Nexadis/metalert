// Основной интерфейс для работы с хранилищами
package storage

import (
	"context"
	"time"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Getter interface {
	Get(ctx context.Context, mtype, id string) (models.Metric, error)
	GetAll(ctx context.Context) ([]models.Metric, error)
}

type Setter interface {
	Set(ctx context.Context, m models.Metric) error
}

// Storage Интерфейс для хранилищ. Позволяет использовать pg и mem хранилища.
type Storage interface {
	Getter
	Setter
}

func ChooseStorage(ctx context.Context, config *Config) (Storage, error) {
	if config.DSN != "" {
		d := db.New()
		db.Configure(d,
			db.SetRetries(config.Retry),
			db.SetTimeout(time.Duration(config.Timeout)),
		)
		dbctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second))
		defer cancel()
		err := d.Open(dbctx, config.DSN)
		if err == nil {
			return d, nil
		}
		logger.Error(err)
	}
	return getMemStorage(ctx, config)
}

func getMemStorage(ctx context.Context, config *Config) (*mem.Storage, error) {
	logger.Info("Use in mem storage")
	metricsStorage := mem.NewMetricsStorage()
	if config.Restore {
		err := metricsStorage.Restore(ctx, config.FileStoragePath, config.Restore)
		if err != nil {
			logger.Info(err)
			return nil, err
		}
	}
	go metricsStorage.SaveTimer(ctx, config.FileStoragePath, config.StoreInterval)
	return metricsStorage, nil
}
