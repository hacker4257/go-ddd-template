package event

import "context"

type OutboxMessage struct {
	Topic   string            `json:"topic"`
	Key     string            `json:"key"`
	Type    string            `json:"type"`
	Payload map[string]any    `json:"payload"`
	Headers map[string]string `json:"headers"`
}

type Outbox interface {
	Add(ctx context.Context, m OutboxMessage) error
}
