package event

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, e Event, headers map[string]string) error
	Close() error
}
