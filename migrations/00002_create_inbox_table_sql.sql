-- +goose Up
-- +goose StatementBegin

CREATE TABLE tasks.inbox_events (
    -- Primary key
    id BIGSERIAL PRIMARY KEY,
    
    -- Kafka source metadata
    topic TEXT NOT NULL,
    partition INT NOT NULL,
    offset BIGINT NOT NULL,
    
    -- Message content
    message_key BYTEA,
    message_value BYTEA NOT NULL,
    headers_json JSONB DEFAULT '{}'::jsonb,
    
    -- Processing state
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    locked_at TIMESTAMP,  -- когда событие было заблокировано для обработки
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    scheduled_at TIMESTAMP NOT NULL DEFAULT NOW(),  -- когда событие готово к обработке (для retry with backoff)
    processed_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT unique_kafka_event UNIQUE(topic, partition, offset)
);

CREATE INDEX idx_inbox_processing ON tasks.inbox_events(status, scheduled_at, locked_at) 
    WHERE status IN ('pending', 'failed');

CREATE INDEX idx_inbox_ordered ON tasks.inbox_events(message_key, offset) 
    WHERE status IN ('pending', 'failed') AND message_key IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX tasks.idx_inbox_ordered;

DROP INDEX tasks.idx_inbox_processing;

DROP TABLE IF EXISTS tasks.inbox_events;

-- +goose StatementEnd
