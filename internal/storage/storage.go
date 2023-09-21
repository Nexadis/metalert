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
	Get(ctx context.Context, mtype, id string) (*metrx.MetricsString, error)
	GetAll(ctx context.Context) ([]*metrx.MetricsString, error)
}

type Setter interface {
	Set(ctx context.Context, mtype, id, value string) error
}

type Storage interface {
	Getter
	Setter
}

var (
	ErrNotFound    = errors.New(`value not found`)
	ErrInvalidType = errors.New(`invalid type`)
)
