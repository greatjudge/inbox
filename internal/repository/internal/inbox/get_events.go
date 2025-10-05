// internal/repository/internal/inbox/get_events.go
package inbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetEventsForProcessing возвращает row структуры событий готовых к обработке (unordered режим)
// и обновляет их статус и locked_at
func GetEventsForProcessing(
	ctx context.Context,
	tx pgx.Tx,
	topic string,
	limit int,
	lockTimeout string,
) ([]InboxEventRow, error) {
	// Сначала выбираем и блокируем события
	selectQuery := `
		SELECT 
			id, topic, partition, offset,
			message_key, message_value, headers_json,
			status, attempts, last_error, locked_at,
			created_at, scheduled_at, processed_at
		FROM tasks.inbox_events
		WHERE topic = @topic
		  AND status IN ('pending', 'failed')
		  AND scheduled_at <= NOW()
		  AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL 5 MIN')
		ORDER BY created_at ASC
		LIMIT @limit
		FOR UPDATE SKIP LOCKED
	`

	args := pgx.NamedArgs{
		"topic": topic,
		"limit": limit,
	}

	rows, err := tx.Query(ctx, selectQuery, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	rowStructs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[InboxEventRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows: %w", err)
	}

	if len(rowStructs) == 0 {
		return rowStructs, nil
	}

	// Обновляем статус и locked_at для выбранных событий
	ids := make([]int64, len(rowStructs))
	for i, row := range rowStructs {
		ids[i] = row.ID
	}

	updateQuery := `
		UPDATE tasks.inbox_events
		SET status = 'processing',
		    locked_at = NOW()
		WHERE id = ANY(@ids)
	`

	updateArgs := pgx.NamedArgs{
		"ids": ids,
	}

	_, err = tx.Exec(ctx, updateQuery, updateArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to update locked_at: %w", err)
	}

	return rowStructs, nil
}
