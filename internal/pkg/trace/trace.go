package trace

import (
	"context"

	appmw "github.com/hacker4257/go-ddd-template/internal/api/http/middleware"
)

func RequestID(ctx context.Context) string {
	return appmw.GetRequestID(ctx)
}
