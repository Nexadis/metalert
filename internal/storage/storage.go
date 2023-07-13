package storage

import "context"

type ObjectGetter interface {
	GetMType() string
	GetID() string
	GetValue() string
}

type Getter interface {
	Get(ctx context.Context, valType, name string) (ObjectGetter, error)
	GetAll(ctx context.Context) ([]ObjectGetter, error)
}

type Setter interface {
	Set(ctx context.Context, valType, name, value string) error
}

type Storage interface {
	Getter
	Setter
}
