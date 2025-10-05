// internal/processors/inbox/process.go
package inbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/greatjudge/inbox/internal/model/inbox"
	"github.com/jackc/pgx/v5"
)

// Process обрабатывает до taskCount событий из inbox
func (p *Processor) Process(ctx context.Context, taskCount int) {
	if err := p.process(ctx, taskCount); err != nil {
		log.Printf("Error during processing: %v", err)
	}
}

// process внутренняя функция обработки событий
func (p *Processor) process(ctx context.Context, taskCount int) error {
	// Обрабатываем события для каждого топика
	for topic, settings := range p.topicHandlers {
		if err := p.processTopic(ctx, topic, settings, taskCount); err != nil {
			log.Printf("Failed to process topic %s: %v", topic, err)
			// Продолжаем обработку остальных топиков
			continue
		}
	}
	return nil
}

// processTopic обрабатывает события для конкретного топика
func (p *Processor) processTopic(
	ctx context.Context,
	topic string,
	settings TopicSettings,
	limit int,
) error {
	// Получаем батч событий для топика
	var events []inbox.Event
	var err error

	if settings.Config.OrderedByKey {
		events, err = p.repository.GetOrderedEventsForProcessing(ctx, topic, limit)
	} else {
		events, err = p.repository.GetEventsForProcessing(ctx, topic, limit)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	log.Printf("Processing %d events for topic %s (ordered: %v)",
		len(events), topic, settings.Config.OrderedByKey)

	// Обрабатываем события
	for _, event := range events {
		if err := p.processEvent(ctx, event, settings.Handler, settings.Config); err != nil {
			log.Printf("Error processing event %d: %v", event.ID, err)
		}
	}

	return nil
}

// processEvent обрабатывает одно событие
func (p *Processor) processEvent(
	ctx context.Context,
	event inbox.Event,
	handler topicHandler,
	topicConfig inbox.TopicConfig,
) error {
	// Помечаем событие как "в обработке"
	if err := p.repository.UpdateStatus(ctx, event.ID, inbox.StatusProcessing, ""); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Вызываем обработчик
	if err := handler.Handle(ctx, event); err != nil {
		return p.handleProcessingError(ctx, event, err, topicConfig)
	}

	// Успешно обработали - помечаем как completed
	if err := p.repository.MarkCompleted(ctx, event.ID); err != nil {
		return fmt.Errorf("failed to mark event as completed: %w", err)
	}

	log.Printf("Successfully processed event %d from topic %s", event.ID, event.Topic)
	return nil
}

// handleProcessingError обрабатывает ошибку при обработке события
func (p *Processor) handleProcessingError(
	ctx context.Context,
	event inbox.Event,
	handlerErr error,
	topicConfig inbox.TopicConfig,
) error {
	newAttempts := event.Attempts + 1
	errorMsg := handlerErr.Error()

	// Проверяем, можно ли повторить
	if newAttempts >= topicConfig.MaxRetries {
		// Исчерпали все попытки - отправляем в dead letter
		if err := p.repository.MarkFailed(ctx, event.ID, errorMsg); err != nil {
			return fmt.Errorf("failed to mark event as failed: %w", err)
		}
		log.Printf("Event %d failed after %d attempts: %s", event.ID, newAttempts, errorMsg)
		return nil
	}

	// Планируем повторную обработку с экспоненциальным backoff
	backoff := p.calculateBackoff(newAttempts, topicConfig.RetryBackoff)
	scheduledAt := time.Now().Add(backoff)

	if err := p.repository.ScheduleRetry(ctx, event.ID, newAttempts, scheduledAt, errorMsg); err != nil {
		return fmt.Errorf("failed to schedule retry: %w", err)
	}

	log.Printf("Event %d scheduled for retry %d/%d in %v: %s",
		event.ID, newAttempts, topicConfig.MaxRetries, backoff, errorMsg)

	return nil
}

// calculateBackoff вычисляет время backoff для повторной попытки
func (p *Processor) calculateBackoff(attempts int, baseBackoff time.Duration) time.Duration {
	// Экспоненциальный backoff: baseBackoff * 2^(attempts-1)
	multiplier := 1 << (attempts - 1) // 2^(attempts-1)
	return baseBackoff * time.Duration(multiplier)
}

type TxCommiter interface {
	InWriteTx(context.Context, func(context.Context, pgx.Tx) error) error
}

type Inbox struct {
	db TxCommiter
}
