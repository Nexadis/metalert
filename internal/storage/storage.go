// Основной интерфейс для работы с хранилищами
package storage

import (
	"errors"
)

// Ошибки при работе с хранилищем.
var (
	ErrNotFound    = errors.New(`value not found`)
	ErrInvalidType = errors.New(`invalid type`)
)
