package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/hacker4257/go-ddd-template/internal/infra/mq/kafka"
	"github.com/hacker4257/go-ddd-template/internal/infra/persistence/mysql"
	"github.com/hacker4257/go-ddd-template/internal/pkg/metrics"
)

type OutboxDispatcher struct {
	log   *slog.Logger
	store *mysql.OutboxStore
	kpub  *kafka.Producer
}

func NewOutboxDispatcher(log *slog.Logger, store *mysql.OutboxStore, kpub *kafka.Producer) *OutboxDispatcher {
	return &OutboxDispatcher{log: log, store: store, kpub: kpub}
}

func (d *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.drainOnce(ctx)
		}
	}
}

func (d *OutboxDispatcher) drainOnce(ctx context.Context) {
	rows, err := d.store.ListUnsent(ctx, 50)
	if err != nil {
		d.log.Error("outbox_list_error", slog.Any("err", err))
		return
	}
	metrics.OutboxPolledTotal.Add(int64(len(rows)))
	for _, r := range rows {
		// headers json -> []kgo.RecordHeader
		var hm map[string]string
		_ = json.Unmarshal(r.Headers, &hm)
		var hs []kgo.RecordHeader
		for k, v := range hm {
			hs = append(hs, kgo.RecordHeader{Key: k, Value: []byte(v)})
		}

		err := d.kpub.PublishRaw(ctx, &kgo.Record{
			Topic:   r.Topic,
			Key:     []byte(r.MsgKey),
			Value:   r.Payload,
			Headers: hs,
		})
		if err != nil {
			d.log.Error("outbox_publish_error", slog.Uint64("id", r.ID), slog.Any("err", err))
			metrics.OutboxFailedTotal.Add(1)
			return // 停住，等下一轮重试
		}

		if err := d.store.MarkSent(ctx, r.ID); err != nil {
			d.log.Error("outbox_mark_sent_error", slog.Uint64("id", r.ID), slog.Any("err", err))
			metrics.OutboxFailedTotal.Add(1)
			return
		}
		metrics.OutboxSentTotal.Add(1)
	}
}
