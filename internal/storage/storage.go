// Основной интерфейс для работы с хранилищами
package storage

import (
	"context"
	"errors"

	"github.com/Nexadis/metalert/internal/metrx"
)

type Getter interface {
	Get(ctx context.Context, mtype, id string) (metrx.Metrics, error)
	GetAll(ctx context.Context) ([]metrx.Metrics, error)
}

type Setter interface {
	Set(ctx context.Context, m metrx.Metrics) error
}

// Интерфейс для хранилищ. Позволяет использовать pg и mem хранилища.
type Storage interface {
	Getter
	Setter
}

// Ошибки при работе с хранилищем.
var (
	ErrNotFound    = errors.New(`value not found`)
	ErrInvalidType = errors.New(`invalid type`)
)
