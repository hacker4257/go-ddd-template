package metrics

import (
	"expvar"
	"sync/atomic"
	"time"
)

var (
	HTTPInFlight        = expvar.NewInt("http_in_flight")
	HTTPRequestsTotal   = expvar.NewInt("http_requests_total")
	HTTP2xxTotal        = expvar.NewInt("http_2xx_total")
	HTTP4xxTotal        = expvar.NewInt("http_4xx_total")
	HTTP5xxTotal        = expvar.NewInt("http_5xx_total")
	HTTPLastLatencyMs   = expvar.NewInt("http_last_latency_ms")

	OutboxSentTotal     = expvar.NewInt("outbox_sent_total")
	OutboxFailedTotal   = expvar.NewInt("outbox_failed_total")
	OutboxPolledTotal   = expvar.NewInt("outbox_polled_total")

	ConsumerProcessedTotal = expvar.NewInt("consumer_processed_total")
	ConsumerFailedTotal    = expvar.NewInt("consumer_failed_total")
	ConsumerDLQTotal       = expvar.NewInt("consumer_dlq_total")
)

func ObserveHTTPLatency(d time.Duration) {
	HTTPLastLatencyMs.Set(d.Milliseconds())
}

func IncStatus(code int) {
	switch {
	case code >= 200 && code < 300:
		HTTP2xxTotal.Add(1)
	case code >= 400 && code < 500:
		HTTP4xxTotal.Add(1)
	case code >= 500:
		HTTP5xxTotal.Add(1)
	}
}

// 可选：给 worker 打点时用（示例，避免 atomic.Int64 版本差异）
var _ = atomic.Int64{}
