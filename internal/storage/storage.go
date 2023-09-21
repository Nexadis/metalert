package storage

import (
	"context"
	"errors"

	"github.com/Nexadis/metalert/internal/metrx"
)

type ObjectGetter interface {
	GetMType() string
	GetID() string
	GetValue() string
}

type Getter interface {
	Get(ctx context.Context, mtype, id string) (metrx.Metrics, error)
	GetAll(ctx context.Context) ([]metrx.Metrics, error)
}

type Setter interface {
	Set(ctx context.Context, m metrx.Metrics) error
}

type Storage interface {
	Getter
	Setter
}

var (
	ErrNotFound    = errors.New(`value not found`)
	ErrInvalidType = errors.New(`invalid type`)
)
