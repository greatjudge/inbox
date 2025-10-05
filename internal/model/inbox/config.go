package inbox

import "time"

// internal/model/inbox/config.go

// TopicConfig конфигурация обработки для конкретного топика
type TopicConfig struct {
	// OrderedByKey если true, обрабатывает события последовательно по message_key
	OrderedByKey bool

	// MaxRetries максимальное количество попыток для этого топика
	MaxRetries int

	// RetryBackoff базовая длительность для backoff
	RetryBackoff time.Duration
}

// Config содержит конфигурацию для inbox processor
type Config struct {
	// BatchSize количество событий для обработки за одну итерацию
	BatchSize int

	// TopicConfigs конфигурация для каждого топика
	TopicConfigs map[string]TopicConfig
}

// NewDefaultTopicConfig возвращает конфигурацию топика по умолчанию
func NewDefaultTopicConfig() TopicConfig {
	return TopicConfig{
		OrderedByKey: false,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Minute,
	}
}

// NewConfig создает новую конфигурацию processor
func NewConfig(batchSize int, topicConfigs map[string]TopicConfig) Config {
	if topicConfigs == nil {
		topicConfigs = make(map[string]TopicConfig)
	}

	return Config{
		BatchSize:    batchSize,
		TopicConfigs: topicConfigs,
	}
}

// NewDefaultConfig возвращает конфигурацию processor со значениями по умолчанию
func NewDefaultConfig() Config {
	return NewConfig(10, make(map[string]TopicConfig))
}
