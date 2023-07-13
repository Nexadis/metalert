package storage

type ObjectGetter interface {
	GetMType() string
	GetID() string
	GetValue() string
}

type Getter interface {
	Get(valType, name string) (ObjectGetter, error)
	GetAll() ([]ObjectGetter, error)
}

type Setter interface {
	Set(valType, name, value string) error
}

type Storage interface {
	Getter
	Setter
}
