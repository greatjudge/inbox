package inbox

import (
	"context"
	"time"

	"github.com/greatjudge/inbox/internal/model/inbox"
)

// TopicSettings описывает обработчик событий и конфигурацию для конкретного топика
type TopicSettings struct {
	// Handler функция обработки события
	Handler topicHandler

	// Config конфигурация обработки для этого топика
	Config inbox.TopicConfig
}

type (
	// inboxRepository определяет методы для работы с inbox событиями в БД
	inboxRepository interface {
		// GetEventsForProcessing возвращает батч событий готовых к обработке (unordered режим)
		GetEventsForProcessing(ctx context.Context, topic string, limit int) ([]inbox.Event, error)

		// GetOrderedEventsForProcessing возвращает первые события для каждого message_key (ordered режим)
		GetOrderedEventsForProcessing(ctx context.Context, topic string, limit int) ([]inbox.Event, error)

		// UpdateStatus обновляет статус события
		UpdateStatus(ctx context.Context, id int64, status inbox.EventStatus, lastError string) error

		// ScheduleRetry планирует повторную обработку события с backoff
		ScheduleRetry(ctx context.Context, id int64, attempts int, scheduledAt time.Time, lastError string) error

		// MarkCompleted помечает событие как успешно обработанное
		MarkCompleted(ctx context.Context, id int64) error

		// MarkFailed помечает событие как окончательно failed (dead letter)
		MarkFailed(ctx context.Context, id int64, lastError string) error
	}

	// topicHandler обрабатывает события из конкретного топика
	topicHandler interface {
		// Handle обрабатывает одно событие
		Handle(ctx context.Context, event inbox.Event) error
	}
)

// Processor обрабатывает события из inbox
type Processor struct {
	repository    inboxRepository
	topicHandlers map[string]TopicSettings // топик -> хендлер + конфиг
	config        inbox.Config
}

// NewProcessor создает новый процессор inbox событий
func NewProcessor(
	repository inboxRepository,
	topicHandlers map[string]TopicSettings,
	config inbox.Config,
) *Processor {
	return &Processor{
		repository:    repository,
		topicHandlers: topicHandlers,
		config:        config,
	}
}
