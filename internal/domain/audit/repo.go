package audit

import "context"

type Repo interface {
	Insert(ctx context.Context, eventType, eventKey string, payload []byte) error
}
