package inbox

import (
	"time"
)

// EventStatus - статус обработки события в inbox
type EventStatus string

const (
	// StatusPending - событие ожидает обработки
	StatusPending EventStatus = "pending"

	// StatusProcessing - событие в процессе обработки
	StatusProcessing EventStatus = "processing"

	// StatusCompleted - событие успешно обработано
	StatusCompleted EventStatus = "completed"

	// StatusFailed - событие не удалось обработать после всех попыток (dead letter)
	StatusFailed EventStatus = "failed"
)

// Event - событие из Kafka
type Event struct {
	ID           int64
	Topic        string
	Partition    int32
	Offset       int64
	MessageKey   []byte
	MessageValue []byte
	Headers      map[string]string
	Status       EventStatus
	Attempts     int
	LastError    string // пустая строка если ошибки не было
	CreatedAt    time.Time
	ScheduledAt  time.Time
	ProcessedAt  time.Time // IsZero() если еще не обработано
}

// IsRetryable возвращает true, если событие можно повторить
func (e *Event) IsRetryable(maxRetries int) bool {
	return e.Attempts < maxRetries && e.Status != StatusCompleted
}

// ShouldProcess возвращает true, если событие готово к обработке
func (e *Event) ShouldProcess(now time.Time) bool {
	return (e.Status == StatusPending || e.Status == StatusFailed) &&
		!e.ScheduledAt.After(now)
}
