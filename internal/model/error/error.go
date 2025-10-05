// internal/model/error/error.go
package error

import "errors"

var (
	// ErrNoHandler - не найден обработчик для топика
	ErrNoHandler = errors.New("no handler found for topic")
)
