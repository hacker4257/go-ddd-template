package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	auditapp "github.com/hacker4257/go-ddd-template/internal/app/audit"
	"github.com/hacker4257/go-ddd-template/internal/infra/idempotency"
	"github.com/hacker4257/go-ddd-template/internal/infra/mq/kafka"
	"github.com/hacker4257/go-ddd-template/internal/pkg/metrics"
)

type UserEvent struct {
	Type       string                 `json:"type"`
	Key        string                 `json:"key"`
	OccurredAt string                 `json:"occurred_at"`
	Payload    map[string]any         `json:"payload"`
}

type UserConsumer struct {
	log        *slog.Logger
	cl         *kgo.Client
	audit      *auditapp.Service
	idem       *idempotency.Store
	dlqTopic   string
	maxRetries int
	producer   *kafka.Producer
}

func NewUserConsumer(log *slog.Logger, brokers []string, group string, topic string, dlqTopic string, maxRetries int,
	audit *auditapp.Service, idem *idempotency.Store, producer *kafka.Producer,
) (*UserConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(), // 我们手动 commit：成功/跳过才 commit
	)
	if err != nil {
		return nil, err
	}

	return &UserConsumer{
		log: log, cl: cl,
		audit: audit, idem: idem,
		dlqTopic: dlqTopic, maxRetries: maxRetries,
		producer: producer,
	}, nil
}

func (c *UserConsumer) Close() {
	c.cl.Close()
}

func (c *UserConsumer) Run(ctx context.Context) {
	for {
		fetches := c.cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.log.Error("kafka_fetch_error", slog.Any("err", e.Err))
			}
			continue
		}

		fetches.EachRecord(func(r *kgo.Record) {
			c.handleRecord(ctx, r)
		})
	}
}

func (c *UserConsumer) handleRecord(ctx context.Context, r *kgo.Record) {
	// 幂等 key：topic-partition-offset（足够做演示；生产可用 message_id）
	idemKey := fmt.Sprintf("%s:%d:%d", r.Topic, r.Partition, r.Offset)

	first, err := c.idem.TryMarkProcessed(ctx, idemKey, 24*time.Hour)
	if err != nil {
		c.log.Error("idem_error", slog.Any("err", err))
		metrics.ConsumerFailedTotal.Add(1)
		return // 不 commit，后续重试
	}
	if !first {
		// 重复消息：直接 commit 跳过
		c.cl.CommitRecords(ctx, r)
		metrics.ConsumerFailedTotal.Add(1)
		return
	}

	// 重试次数（从 header 读取）
	retry := headerInt(r.Headers, "retry")

	// 解析事件
	var evt UserEvent
	if err := json.Unmarshal(r.Value, &evt); err != nil {
		c.log.Error("event_unmarshal_error", slog.Any("err", err))
		c.sendDLQ(ctx, r, "unmarshal_error", retry)
		metrics.ConsumerProcessedTotal.Add(1)
		c.cl.CommitRecords(ctx, r)
		return
	}

	// 示例：只处理 UserCreated
	if evt.Type == "UserCreated" {
		payloadBytes := r.Value // 原样落库，最简单
		if err := c.audit.Record(ctx, evt.Type, evt.Key, payloadBytes); err != nil {
			c.log.Error("audit_record_error", slog.Any("err", err))

			if retry+1 >= c.maxRetries {
				c.sendDLQ(ctx, r, "max_retries_exceeded", retry)
				c.cl.CommitRecords(ctx, r)
				return
			}

			// 重新投递到原 topic（带 retry+1 header），然后 commit 当前 offset，避免堵塞
			if err := c.requeue(ctx, r, retry+1); err != nil {
				// requeue 失败：不 commit，让它重试
				c.log.Error("requeue_error", slog.Any("err", err))
				return
			}
			c.cl.CommitRecords(ctx, r)
			return
		}
	}

	// 成功：commit
	c.cl.CommitRecords(ctx, r)
}

func (c *UserConsumer) requeue(ctx context.Context, r *kgo.Record, retry int) error {
	rec := &kgo.Record{
		Topic: r.Topic,
		Key:   r.Key,
		Value: r.Value,
		Headers: append(copyHeaders(r.Headers), kgo.RecordHeader{
			Key: "retry", Value: []byte(fmt.Sprintf("%d", retry)),
		}),
	}
	return c.producer.PublishRaw(ctx, rec)
}

func (c *UserConsumer) sendDLQ(ctx context.Context, r *kgo.Record, reason string, retry int) {
	rec := &kgo.Record{
		Topic: c.dlqTopic,
		Key:   r.Key,
		Value: r.Value,
		Headers: append(copyHeaders(r.Headers),
			kgo.RecordHeader{Key: "dlq_reason", Value: []byte(reason)},
			kgo.RecordHeader{Key: "retry", Value: []byte(fmt.Sprintf("%d", retry))},
		),
	}
	_ = c.producer.PublishRaw(ctx, rec) // best effort
}

func copyHeaders(hs []kgo.RecordHeader) []kgo.RecordHeader {
	out := make([]kgo.RecordHeader, 0, len(hs))
	for _, h := range hs {
		out = append(out, kgo.RecordHeader{Key: h.Key, Value: append([]byte(nil), h.Value...)})
	}
	return out
}

func headerInt(hs []kgo.RecordHeader, key string) int {
	for _, h := range hs {
		if h.Key == key {
			var n int
			_, _ = fmt.Sscanf(string(h.Value), "%d", &n)
			return n
		}
	}
	return 0
}
