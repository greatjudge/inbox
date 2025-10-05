// internal/repository/inbox.go
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/greatjudge/inbox/internal/model/inbox"
	inboxrepo "github.com/greatjudge/inbox/internal/repository/internal/inbox"
)

// TxCommiter интерфейс для работы с транзакциями
type TxCommiter interface {
	InWriteTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error
}

// Inbox репозиторий для работы с inbox событиями
type Inbox struct {
	db TxCommiter
}

// NewInbox создает новый репозиторий inbox
func NewInbox(db TxCommiter) *Inbox {
	return &Inbox{
		db: db,
	}
}

// GetEventsForProcessing возвращает батч событий готовых к обработке (unordered режим)
func (i *Inbox) GetEventsForProcessing(
	ctx context.Context,
	topic string,
	limit int,
) ([]inbox.Event, error) {
	var rowStructs []inboxrepo.InboxEventRow

	err := i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		rowStructs, err = inboxrepo.GetEventsForProcessing(ctx, tx, topic, limit)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Конвертация вне транзакции
	return inboxrepo.RowsToModels(rowStructs)
}

// GetOrderedEventsForProcessing возвращает первые события для каждого message_key (ordered режим)
func (i *Inbox) GetOrderedEventsForProcessing(
	ctx context.Context,
	topic string,
	limit int,
) ([]inbox.Event, error) {
	var rowStructs []inboxrepo.InboxEventRow

	err := i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		rowStructs, err = inboxrepo.GetOrderedEventsForProcessing(ctx, tx, topic, limit)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Конвертация вне транзакции
	return inboxrepo.RowsToModels(rowStructs)
}

// UpdateStatus обновляет статус события
func (i *Inbox) UpdateStatus(
	ctx context.Context,
	id int64,
	status inbox.EventStatus,
	lastError string,
) error {
	return i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return inboxrepo.UpdateStatus(ctx, tx, id, status, lastError)
	})
}

// ScheduleRetry планирует повторную обработку события с backoff
func (i *Inbox) ScheduleRetry(
	ctx context.Context,
	id int64,
	attempts int,
	scheduledAt time.Time,
	lastError string,
) error {
	return i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return inboxrepo.ScheduleRetry(ctx, tx, id, attempts, scheduledAt, lastError)
	})
}

// MarkCompleted помечает событие как успешно обработанное
func (i *Inbox) MarkCompleted(ctx context.Context, id int64) error {
	return i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return inboxrepo.MarkCompleted(ctx, tx, id)
	})
}

// MarkFailed помечает событие как окончательно failed (dead letter)
func (i *Inbox) MarkFailed(ctx context.Context, id int64, lastError string) error {
	return i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return inboxrepo.MarkFailed(ctx, tx, id, lastError)
	})
}

// Insert добавляет новое событие в inbox (для Kafka consumer)
func (i *Inbox) Insert(
	ctx context.Context,
	topic string,
	partition int32,
	offset int64,
	messageKey []byte,
	messageValue []byte,
	headers map[string]string,
) error {
	return i.db.InWriteTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return inboxrepo.Insert(ctx, tx, topic, partition, offset, messageKey, messageValue, headers)
	})
}
