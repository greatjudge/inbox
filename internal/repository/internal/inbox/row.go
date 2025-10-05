// internal/repository/internal/inbox/row.go
package inbox

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/greatjudge/inbox/internal/model/inbox"
)

// InboxEventRow представляет строку из таблицы tasks.inbox_events
type InboxEventRow struct {
	ID           int64          `db:"id"`
	Topic        string         `db:"topic"`
	Partition    int32          `db:"partition"`
	Offset       int64          `db:"offset"`
	MessageKey   sql.RawBytes   `db:"message_key"`
	MessageValue []byte         `db:"message_value"`
	HeadersJSON  []byte         `db:"headers_json"`
	Status       string         `db:"status"`
	Attempts     int            `db:"attempts"`
	LastError    sql.NullString `db:"last_error"`
	CreatedAt    time.Time      `db:"created_at"`
	ScheduledAt  time.Time      `db:"scheduled_at"`
	ProcessedAt  sql.NullTime   `db:"processed_at"`
}

// ToModel конвертирует row в доменную модель
func (r *InboxEventRow) ToModel() (inbox.Event, error) {
	var headers map[string]string
	if len(r.HeadersJSON) > 0 {
		if err := json.Unmarshal(r.HeadersJSON, &headers); err != nil {
			return inbox.Event{}, err
		}
	} else {
		headers = make(map[string]string)
	}

	event := inbox.Event{
		ID:           r.ID,
		Topic:        r.Topic,
		Partition:    r.Partition,
		Offset:       r.Offset,
		MessageKey:   []byte(r.MessageKey),
		MessageValue: r.MessageValue,
		Headers:      headers,
		Status:       inbox.EventStatus(r.Status),
		Attempts:     r.Attempts,
		CreatedAt:    r.CreatedAt,
		ScheduledAt:  r.ScheduledAt,
	}

	if r.LastError.Valid {
		event.LastError = r.LastError.String
	}

	if r.ProcessedAt.Valid {
		event.ProcessedAt = r.ProcessedAt.Time
	}

	return event, nil
}

// RowsToModels конвертирует слайс row структур в слайс доменных моделей
func RowsToModels(rows []InboxEventRow) ([]inbox.Event, error) {
	events := make([]inbox.Event, 0, len(rows))
	for _, row := range rows {
		event, err := row.ToModel()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
